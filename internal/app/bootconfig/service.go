package bootconfig

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"

	"hyperbdr-client/internal/client"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
	Post(path string, body interface{}) (client.APIResponse, error)
}

type Service struct {
	api API
}

type MutationResult struct {
	Response  client.APIResponse
	Operation string
}

type DetailField struct {
	Key   string
	Value interface{}
}

func NewService(api API) Service {
	return Service{api: api}
}

func (s Service) Create(hostID string, metadata map[string]interface{}) (client.APIResponse, error) {
	return s.api.Post("/api/v2/batchBootConfigs", map[string]interface{}{
		"batch_create": []map[string]interface{}{
			{
				"migration_id": hostID,
				"metadata":     metadata,
			},
		},
	})
}

func (s Service) Get(hostID string) (client.APIResponse, error) {
	bootConfigID, err := s.bootConfigIDByHostID(hostID)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.batchGetByIDs([]string{bootConfigID})
}

func (s Service) Update(hostID string, metadata map[string]interface{}) (client.APIResponse, error) {
	bootConfigID, err := s.bootConfigIDByHostID(hostID)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.updateWithID(hostID, bootConfigID, metadata)
}

func (s Service) Apply(hostID string, metadata map[string]interface{}) (MutationResult, error) {
	bootConfigID, exists, err := s.bootConfigIDByHostIDOptional(hostID)
	if err != nil {
		return MutationResult{}, err
	}
	if exists {
		resp, err := s.updateWithID(hostID, bootConfigID, metadata)
		return MutationResult{Response: resp, Operation: "update"}, err
	}
	resp, err := s.Create(hostID, metadata)
	return MutationResult{Response: resp, Operation: "create"}, err
}

func (s Service) BatchCreate(body map[string]interface{}) (client.APIResponse, error) {
	return s.api.Post("/api/v2/batchBootConfigs", body)
}

func (s Service) BatchGet(hostIDs []string) (client.APIResponse, error) {
	bootConfigIDs, err := s.batchBootConfigIDsByHostIDs(hostIDs)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.batchGetByIDs(bootConfigIDs)
}

func (s Service) BatchUpdate(body map[string]interface{}) (client.APIResponse, error) {
	return s.api.Post("/api/v2/batchUpdateBootConfigs", body)
}

func MetadataFile(file string) (map[string]interface{}, error) {
	if file == "" {
		return nil, fmt.Errorf("file is required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	switch t := raw.(type) {
	case []interface{}:
		return nil, fmt.Errorf("file must contain single metadata object")
	case map[string]interface{}:
		if _, ok := t["batch_create"]; ok {
			return nil, fmt.Errorf("file must contain metadata object, not batch_create wrapper")
		}
		if _, ok := t["batch_update"]; ok {
			return nil, fmt.Errorf("file must contain metadata object, not batch_update wrapper")
		}
		return t, nil
	default:
		return nil, fmt.Errorf("file must contain metadata object")
	}
}

func DetailFields(hostID string, row map[string]interface{}) []DetailField {
	fields := []DetailField{
		{Key: "host_id", Value: hostID},
	}
	if bootConfigID, _ := row["id"].(string); bootConfigID != "" {
		fields = append(fields, DetailField{Key: "boot_config_id", Value: bootConfigID})
	}
	if migrationID, ok := row["migration_id"]; ok {
		fields = append(fields, DetailField{Key: "migration_id", Value: migrationID})
	}
	if metadata, ok := row["metadata"].(map[string]interface{}); ok {
		keys := make([]string, 0, len(metadata))
		for key := range metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fields = append(fields, DetailField{Key: "metadata." + key, Value: metadata[key]})
		}
	}
	return fields
}

func MutationView(result MutationResult) map[string]interface{} {
	if data := mapFromData(result.Response.Data); data != nil {
		view := cloneMap(data)
		view["operation"] = result.Operation
		return view
	}
	if result.Response.Data == nil && result.Response.Raw != nil {
		view := cloneMap(result.Response.Raw)
		view["operation"] = result.Operation
		return view
	}
	return map[string]interface{}{"operation": result.Operation}
}

func (s Service) updateWithID(hostID, bootConfigID string, metadata map[string]interface{}) (client.APIResponse, error) {
	return s.api.Post("/api/v2/batchUpdateBootConfigs", map[string]interface{}{
		"batch_update": []map[string]interface{}{
			{
				"id":           bootConfigID,
				"migration_id": hostID,
				"metadata":     metadata,
			},
		},
	})
}

func (s Service) batchGetByIDs(bootConfigIDs []string) (client.APIResponse, error) {
	items := make([]map[string]interface{}, 0, len(bootConfigIDs))
	for _, bootConfigID := range bootConfigIDs {
		items = append(items, map[string]interface{}{"id": bootConfigID})
	}
	return s.api.Post("/api/v2/batchGetBootConfigs", map[string]interface{}{"batch_get": items})
}

func (s Service) bootConfigIDByHostID(hostID string) (string, error) {
	bootConfigID, exists, err := s.bootConfigIDByHostIDOptional(hostID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("host %s has no existing boot config", hostID)
	}
	return bootConfigID, nil
}

func (s Service) bootConfigIDByHostIDOptional(hostID string) (string, bool, error) {
	resp, err := s.api.Get("/api/v2/getHostDetail", queryFromPairs("host_id", hostID))
	if err != nil {
		return "", false, err
	}
	data := mapFromData(resp.Data)
	if data == nil {
		return "", false, fmt.Errorf("host %s did not return detail object", hostID)
	}
	bootConfigID := findBootConfigID(data)
	if bootConfigID == "" {
		return "", false, nil
	}
	return bootConfigID, true, nil
}

func (s Service) batchBootConfigIDsByHostIDs(hostIDs []string) ([]string, error) {
	resp, err := s.api.Get("/api/v2/getHosts", queryFromPairs("ids", joinCSV(hostIDs)))
	if err != nil {
		return nil, err
	}
	rows := listFromData(resp.Data, "hosts")
	if len(rows) == 0 {
		return nil, fmt.Errorf("no hosts matched provided ids")
	}
	byHostID := make(map[string]string, len(rows))
	for _, row := range rows {
		hostID, _ := row["id"].(string)
		if hostID == "" {
			continue
		}
		if bootConfigID := findBootConfigID(row); bootConfigID != "" {
			byHostID[hostID] = bootConfigID
		}
	}
	out := make([]string, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		bootConfigID := byHostID[hostID]
		if bootConfigID == "" {
			return nil, fmt.Errorf("host %s has no existing boot config", hostID)
		}
		out = append(out, bootConfigID)
	}
	return out, nil
}

func findBootConfigID(data map[string]interface{}) string {
	if bootConfigID, _ := data["boot_config_id"].(string); bootConfigID != "" {
		return bootConfigID
	}
	bootConfig, ok := data["boot_config"].(map[string]interface{})
	if !ok || bootConfig == nil {
		return ""
	}
	bootConfigID, _ := bootConfig["id"].(string)
	return bootConfigID
}

func queryFromPairs(pairs ...string) url.Values {
	q := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			q.Set(pairs[i], pairs[i+1])
		}
	}
	return q
}

func joinCSV(values []string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for i := 1; i < len(values); i++ {
		out += "," + values[i]
	}
	return out
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

func cloneMap(src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(src)+1)
	for key, value := range src {
		out[key] = value
	}
	return out
}
