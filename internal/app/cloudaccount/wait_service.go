package cloudaccount

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

	row := s.waitForAccount(spec.ID, spec.Interval, spec.Timeout)
	result := WaitExecutionResult{
		Rows: []map[string]interface{}{row},
	}
	if row["result"] != "success" {
		result.Failed = true
	}
	return result, nil
}

func (s Service) waitForAccount(accountID string, interval, timeout time.Duration) map[string]interface{} {
	start := time.Now()
	deadline := start.Add(timeout)
	if timeout == 0 {
		deadline = start
	}

	for {
		account, notFound, err := s.fetchAccountDetail(accountID)
		elapsed := int(time.Since(start).Seconds())
		if err != nil {
			return waitutil.WaitResult(accountID, "create-account", "failed", "", "", elapsed, err.Error())
		}
		if notFound {
			return waitutil.WaitResult(accountID, "create-account", "failed", "not_found", "", elapsed, "cloud account not found")
		}

		status := waitutil.FirstNonEmptyString(account["status"])
		displayStatus := waitutil.FirstNonEmptyString(account["display_status"], account["displayStatus"])
		switch {
		case waitutil.IsFailureStatus(status, displayStatus):
			return waitutil.WaitResult(accountID, "create-account", "failed", status, displayStatus, elapsed, "")
		case waitutil.IsSuccessStatus(status, displayStatus):
			return waitutil.WaitResult(accountID, "create-account", "success", status, displayStatus, elapsed, "")
		}

		if !time.Now().Before(deadline) {
			return waitutil.WaitResult(accountID, "create-account", "timeout", status, displayStatus, elapsed, "wait timed out")
		}
		if interval > 0 {
			time.Sleep(interval)
		}
	}
}

func (s Service) fetchAccountDetail(accountID string) (map[string]interface{}, bool, error) {
	resp, err := s.api.Get("/hypermotion/v1/cloud_accounts/"+url.PathEscape(accountID), url.Values{})
	if err != nil {
		if httpErr, ok := err.(client.HTTPError); ok && httpErr.StatusCode == 404 {
			return nil, true, nil
		}
		return nil, false, err
	}
	data := waitutil.MapFromData(resp.Data)
	if nested := waitutil.MapFromData(data["cloud_account"]); nested != nil {
		return nested, false, nil
	}
	if nested := waitutil.MapFromData(resp.Raw["cloud_account"]); nested != nil {
		return nested, false, nil
	}
	if data == nil {
		return nil, false, fmt.Errorf("cloud account detail response is empty")
	}
	return data, false, nil
}
