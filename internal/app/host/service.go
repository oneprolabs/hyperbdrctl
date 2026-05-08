package host

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hyperbdr-client/internal/client"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
	Post(path string, body interface{}) (client.APIResponse, error)
	Delete(path string, body interface{}) (client.APIResponse, error)
}

type Service struct {
	api API
}

func NewService(api API) Service {
	return Service{api: api}
}

type ListSpec struct {
	Status     string
	BootStatus string
	KW         string
	CloudType  string
	IDs        string
	MACs       string
	Page       int
	PageSize   int
	Query      url.Values
}

type DetailSpec struct {
	ID    string
	Query url.Values
}

type SnapshotsSpec struct {
	ID         string
	Status     string
	SyncDetail bool
	Query      url.Values
}

type SyncSpec struct {
	RawBody       map[string]interface{}
	ID            string
	IDs           string
	Mode          *string
	TransferSpeed *int
}

type RegisterSpec struct {
	VMID  string
	VMIDs string
}

type BootSpec struct {
	RawBody map[string]interface{}
	ID      string
	Known   map[string]string
	Unknown map[string]interface{}
}

type CleanupValidationHostSpec struct {
	RawBody map[string]interface{}
	ID      string
	IDs     string
}

type DeregisterSpec struct {
	RawBody map[string]interface{}
	ID      string
	IDs     string
	Force   bool
}

type WaitSpec struct {
	ID           string
	IDs          string
	Operation    string
	Interval     time.Duration
	Timeout      time.Duration
	IncludeSteps bool
}

type WaitExecutionResult struct {
	Rows   []map[string]interface{}
	Failed bool
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("status", spec.Status)
	q.Set("boot_status", spec.BootStatus)
	q.Set("kw", spec.KW)
	q.Set("cloud_type", spec.CloudType)
	q.Set("ids", spec.IDs)
	q.Set("macs", spec.MACs)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getHosts", q)
}

func (s Service) Detail(spec DetailSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("host_id", spec.ID)
	return s.api.Get("/api/v2/getHostDetail", q)
}

func (s Service) Snapshots(spec SnapshotsSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("sheet", "snapshot")
	q.Set("host_id", spec.ID)
	q.Set("status", spec.Status)
	addBool(q, "sync_detail", spec.SyncDetail)
	return s.api.Get("/api/v2/getHostDetail", q)
}

func (s Service) Sync(spec SyncSpec) (client.APIResponse, error) {
	body, err := syncBody(spec.RawBody, spec.ID, spec.IDs, spec.Mode, spec.TransferSpeed)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post("/api/v2/batchSync", body)
}

func (s Service) Register(spec RegisterSpec) (client.APIResponse, error) {
	body, err := registerBody(spec.VMID, spec.VMIDs)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post("/hypermotion/v1/hosts/register", body)
}

func (s Service) Boot(spec BootSpec) (client.APIResponse, error) {
	body, err := bootBody(spec.RawBody, spec.ID, spec.Known, spec.Unknown)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post("/api/v2/batchBoot", body)
}

func (s Service) CleanupValidationHost(spec CleanupValidationHostSpec) (client.APIResponse, error) {
	body, err := cleanupValidationHostBody(spec.RawBody, spec.ID, spec.IDs)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post("/api/v2/batchDeleteInstances", body)
}

func (s Service) Deregister(spec DeregisterSpec) (client.APIResponse, error) {
	body, err := deregisterBody(spec.RawBody, spec.ID, spec.IDs, spec.Force)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Delete("/hypermotion/v1/hosts", body)
}

func (s Service) Wait(spec WaitSpec) (WaitExecutionResult, error) {
	hostIDs := splitCSV(spec.IDs)
	if spec.ID != "" {
		hostIDs = append([]string{spec.ID}, hostIDs...)
	}
	if len(hostIDs) == 0 {
		return WaitExecutionResult{}, fmt.Errorf("id or ids is required")
	}
	if !validWaitOperation(spec.Operation) {
		return WaitExecutionResult{}, fmt.Errorf("operation must be one of sync, boot, cleanup-validation-host, deregister")
	}
	if spec.Interval < 0 {
		return WaitExecutionResult{}, fmt.Errorf("interval-seconds must be >= 0")
	}
	if spec.Timeout < 0 {
		return WaitExecutionResult{}, fmt.Errorf("timeout-seconds must be >= 0")
	}

	result := WaitExecutionResult{
		Rows: make([]map[string]interface{}, 0, len(hostIDs)),
	}
	for _, hostID := range hostIDs {
		row := s.waitForHost(hostID, spec.Operation, spec.Interval, spec.Timeout, spec.IncludeSteps)
		if row["result"] != "success" {
			result.Failed = true
		}
		result.Rows = append(result.Rows, row)
	}
	return result, nil
}

func validWaitOperation(operation string) bool {
	switch operation {
	case "sync", "boot", "cleanup-validation-host", "deregister":
		return true
	default:
		return false
	}
}

func (s Service) waitForHost(hostID, operation string, interval, timeout time.Duration, includeSteps bool) map[string]interface{} {
	start := time.Now()
	deadline := start.Add(timeout)
	if timeout == 0 {
		deadline = start
	}
	for {
		host, notFound, err := s.fetchHostDetail(hostID)
		elapsed := int(time.Since(start).Seconds())
		if notFound && operation == "deregister" {
			return waitResult(hostID, operation, "success", "not_found", "", "", elapsed, "")
		}
		if err != nil {
			return waitResult(hostID, operation, "failed", "", "", "", elapsed, err.Error())
		}
		status, displayStatus, taskID, taskError := waitHostState(host, operation)
		result := classifyWaitState(operation, status, notFound)
		switch result {
		case "success":
			return waitResult(hostID, operation, result, status, displayStatus, taskID, elapsed, "")
		case "failed":
			if includeSteps && taskID != "" {
				if stepError := s.taskStepError(taskID); stepError != "" {
					taskError = stepError
				}
			}
			return waitResult(hostID, operation, result, status, displayStatus, taskID, elapsed, taskError)
		}
		if !time.Now().Before(deadline) {
			return waitResult(hostID, operation, "timeout", status, displayStatus, taskID, elapsed, "wait timed out")
		}
		if interval > 0 {
			time.Sleep(interval)
		}
	}
}

func (s Service) fetchHostDetail(hostID string) (map[string]interface{}, bool, error) {
	q := url.Values{}
	q.Set("host_id", hostID)
	resp, err := s.api.Get("/api/v2/getHostDetail", q)
	if err != nil {
		if httpErr, ok := err.(client.HTTPError); ok && httpErr.StatusCode == 404 {
			return nil, true, nil
		}
		return nil, false, err
	}
	return mapFromData(resp.Data), false, nil
}

func waitHostState(host map[string]interface{}, operation string) (status, displayStatus, taskID, taskError string) {
	if host == nil {
		return "", "", "", ""
	}
	switch operation {
	case "boot":
		return stringValue(host["boot_status"]), stringValue(host["display_boot_status"]), stringValue(host["boot_task_id"]), stringValue(host["boot_task_error_description"])
	default:
		return stringValue(host["status"]), stringValue(host["display_status"]), stringValue(host["task_id"]), stringValue(host["task_error_description"])
	}
}

func classifyWaitState(operation, status string, notFound bool) string {
	if notFound && operation == "deregister" {
		return "success"
	}
	switch operation {
	case "sync":
		if status == "sync_snapshot_done" {
			return "success"
		}
		if status == "sync_doing" || status == "" || status == "host_register_done" {
			return "running"
		}
	case "boot":
		if status == "boot_done" {
			return "success"
		}
		if status == "boot_doing" || status == "" || status == "not_boot" {
			return "running"
		}
	case "cleanup-validation-host":
		if status == "clean_done" {
			return "success"
		}
		if status == "clean_doing" || status == "" {
			return "running"
		}
	case "deregister":
		if status == "host_unregister_done" {
			return "success"
		}
		if status == "clean_doing" || status == "" {
			return "running"
		}
	}
	if isFailureStatus(status) {
		return "failed"
	}
	return "running"
}

func isFailureStatus(status string) bool {
	status = strings.ToLower(status)
	return strings.Contains(status, "fail") || strings.Contains(status, "error")
}

func (s Service) taskStepError(taskID string) string {
	q := url.Values{}
	q.Set("task_id", taskID)
	resp, err := s.api.Get("/hypermotion/v1/job/steps", q)
	if err != nil {
		return ""
	}
	steps := listFromData(resp.Data, "steps")
	for i := len(steps) - 1; i >= 0; i-- {
		if msg := stringValue(steps[i]["message"]); msg != "" {
			return msg
		}
		if logs, ok := steps[i]["logs"].([]interface{}); ok {
			for j := len(logs) - 1; j >= 0; j-- {
				log, ok := logs[j].(map[string]interface{})
				if !ok {
					continue
				}
				parts := []string{}
				for _, key := range []string{"detail", "error_reason", "traceback"} {
					if value := stringValue(log[key]); value != "" {
						parts = append(parts, value)
					}
				}
				if len(parts) > 0 {
					return strings.Join(parts, " - ")
				}
			}
		}
	}
	return ""
}

func waitResult(hostID, operation, result, status, displayStatus, taskID string, elapsedSeconds int, errText string) map[string]interface{} {
	return map[string]interface{}{
		"id":              hostID,
		"operation":       operation,
		"result":          result,
		"status":          status,
		"display_status":  displayStatus,
		"task_id":         taskID,
		"elapsed_seconds": elapsedSeconds,
		"error":           errText,
	}
}

func syncBody(rawBody map[string]interface{}, id, ids string, mode *string, transferSpeed *int) (map[string]interface{}, error) {
	if rawBody != nil {
		return rawBody, nil
	}
	hostIDs := splitCSV(ids)
	if id != "" {
		hostIDs = append([]string{id}, hostIDs...)
	}
	if len(hostIDs) == 0 {
		return nil, fmt.Errorf("id or ids is required")
	}
	items := make([]map[string]interface{}, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		item := map[string]interface{}{"migration_id": hostID, "do_snapshot": true}
		if mode != nil {
			item["sync_mode"] = *mode
		}
		if transferSpeed != nil {
			item["transfer_speed"] = *transferSpeed
		}
		items = append(items, item)
	}
	return map[string]interface{}{"batch_sync": items}, nil
}

func registerBody(vmID, vmIDs string) (map[string]interface{}, error) {
	ids := splitCSV(vmIDs)
	if vmID != "" {
		ids = append([]string{vmID}, ids...)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("vm-id or vm-ids is required")
	}
	return map[string]interface{}{"ids": ids}, nil
}

func bootBody(rawBody map[string]interface{}, id string, known map[string]string, unknown map[string]interface{}) (map[string]interface{}, error) {
	if rawBody != nil {
		return rawBody, nil
	}
	item := map[string]interface{}{}
	for key, value := range known {
		if value != "" {
			item[key] = value
		}
	}
	for key, value := range unknown {
		item[key] = value
	}
	if id != "" {
		item["migration_id"] = id
	}
	if item["migration_id"] == nil || item["migration_id"] == "" {
		return nil, fmt.Errorf("id is required")
	}
	return map[string]interface{}{"batch_boot": []map[string]interface{}{item}}, nil
}

func cleanupValidationHostBody(rawBody map[string]interface{}, id, ids string) (map[string]interface{}, error) {
	if rawBody != nil {
		return rawBody, nil
	}
	hostIDs := splitCSV(ids)
	if id != "" {
		hostIDs = append([]string{id}, hostIDs...)
	}
	if len(hostIDs) == 0 {
		return nil, fmt.Errorf("id or ids is required")
	}
	items := make([]map[string]interface{}, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		items = append(items, map[string]interface{}{"migration_id": hostID})
	}
	return map[string]interface{}{"batch_delete_instances": items}, nil
}

func deregisterBody(rawBody map[string]interface{}, id, ids string, force bool) (map[string]interface{}, error) {
	if rawBody != nil {
		return rawBody, nil
	}
	hostIDs := splitCSV(ids)
	if id != "" {
		hostIDs = append([]string{id}, hostIDs...)
	}
	if len(hostIDs) == 0 {
		return nil, fmt.Errorf("id or ids is required")
	}
	return map[string]interface{}{
		"ids": hostIDs,
		"clean": map[string]interface{}{
			"force": force,
		},
	}, nil
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func cloneValues(v url.Values) url.Values {
	cloned := url.Values{}
	for key, values := range v {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}
	return cloned
}

func addInt(q url.Values, key string, value int) {
	if value > 0 {
		q.Set(key, fmt.Sprintf("%d", value))
	}
}

func addBool(q url.Values, key string, value bool) {
	if value {
		q.Set(key, "true")
	}
}

func stringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func mapFromData(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func listFromData(data interface{}, key string) []map[string]interface{} {
	if items, ok := data.([]interface{}); ok {
		rows := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
		return rows
	}
	m := mapFromData(data)
	if m == nil {
		return nil
	}
	items, ok := m[key].([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func InferValue(value string) interface{} {
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return value
}
