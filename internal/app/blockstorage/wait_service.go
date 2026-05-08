package blockstorage

import (
	"fmt"
	"net/url"
	"time"

	"hyperbdr-client/internal/app/waitutil"
	"hyperbdr-client/internal/client"
)

type WaitSpec = waitutil.WaitSpec

type WaitExecutionResult = waitutil.WaitExecutionResult

func (s Service) Wait(spec WaitSpec) (WaitExecutionResult, error) {
	if err := waitutil.ValidateSingleWaitSpec(spec); err != nil {
		return WaitExecutionResult{}, err
	}

	row := s.waitForCloudSyncGateway(spec.ID, spec.Interval, spec.Timeout)
	result := WaitExecutionResult{
		Rows: []map[string]interface{}{row},
	}
	if row["result"] != "success" {
		result.Failed = true
	}
	return result, nil
}

func (s Service) waitForCloudSyncGateway(storageID string, interval, timeout time.Duration) map[string]interface{} {
	start := time.Now()
	deadline := start.Add(timeout)
	if timeout == 0 {
		deadline = start
	}

	for {
		storage, notFound, err := s.fetchStorageDetail(storageID)
		elapsed := int(time.Since(start).Seconds())
		if err != nil {
			return waitutil.WaitResult(storageID, "create-cloud-sync-gateway", "failed", "", "", elapsed, err.Error())
		}
		if notFound {
			return waitutil.WaitResult(storageID, "create-cloud-sync-gateway", "failed", "not_found", "", elapsed, "cloud-sync-gateway not found")
		}

		if actualType := storageKind(storage); actualType != "" && actualType != "hypergate" {
			errText := fmt.Sprintf("resource type mismatch: expected cloud-sync-gateway, got %s", actualType)
			return waitutil.WaitResult(storageID, "create-cloud-sync-gateway", "failed", waitutil.FirstNonEmptyString(storage["status"]), waitutil.FirstNonEmptyString(storage["display_status"], storage["displayStatus"]), elapsed, errText)
		}

		status := waitutil.FirstNonEmptyString(storage["status"])
		displayStatus := waitutil.FirstNonEmptyString(storage["display_status"], storage["displayStatus"])
		switch {
		case waitutil.IsFailureStatus(status, displayStatus):
			return waitutil.WaitResult(storageID, "create-cloud-sync-gateway", "failed", status, displayStatus, elapsed, "")
		case waitutil.IsSuccessStatus(status, displayStatus):
			return waitutil.WaitResult(storageID, "create-cloud-sync-gateway", "success", status, displayStatus, elapsed, "")
		}

		if !time.Now().Before(deadline) {
			return waitutil.WaitResult(storageID, "create-cloud-sync-gateway", "timeout", status, displayStatus, elapsed, "wait timed out")
		}
		if interval > 0 {
			time.Sleep(interval)
		}
	}
}

func (s Service) fetchStorageDetail(storageID string) (map[string]interface{}, bool, error) {
	q := url.Values{}
	q.Set("storage_id", storageID)
	resp, err := s.api.Get("/api/v2/getStorageDetailInfo", q)
	if err != nil {
		if httpErr, ok := err.(client.HTTPError); ok && httpErr.StatusCode == 404 {
			return nil, true, nil
		}
		return nil, false, err
	}
	data := waitutil.MapFromData(resp.Data)
	if nested := waitutil.MapFromData(data["storage"]); nested != nil {
		return nested, false, nil
	}
	if nested := waitutil.MapFromData(data["storage_detail"]); nested != nil {
		return nested, false, nil
	}
	if data == nil {
		return nil, false, fmt.Errorf("cloud-sync-gateway detail response is empty")
	}
	return data, false, nil
}

func storageKind(storage map[string]interface{}) string {
	return waitutil.NormalizeStorageKind(waitutil.FirstNonEmptyString(
		storage["type"],
		storage["storage_type"],
		storage["display_type"],
		storage["display_storage_type"],
	))
}
