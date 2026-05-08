package cloudaccount

import (
	"net/url"
	"testing"
	"time"

	"hyperbdr-client/internal/client"
)

type fakeWaitAPI struct {
	responses []client.APIResponse
	errs      []error
	calls     int
	path      string
	query     url.Values
}

func (f *fakeWaitAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.path = path
	f.query = q
	idx := f.calls
	f.calls++
	if idx < len(f.errs) && f.errs[idx] != nil {
		return client.APIResponse{}, f.errs[idx]
	}
	if idx < len(f.responses) {
		return f.responses[idx], nil
	}
	if len(f.responses) > 0 {
		return f.responses[len(f.responses)-1], nil
	}
	return client.APIResponse{}, nil
}

func (f *fakeWaitAPI) Post(string, interface{}) (client.APIResponse, error) {
	return client.APIResponse{}, nil
}

func (f *fakeWaitAPI) Delete(string, interface{}) (client.APIResponse, error) {
	return client.APIResponse{}, nil
}

func TestServiceWaitCloudAccountSucceedsAfterPolling(t *testing.T) {
	api := &fakeWaitAPI{
		responses: []client.APIResponse{
			{Data: map[string]interface{}{"status": "creating"}},
			{Data: map[string]interface{}{"status": "active", "display_status": "Available"}},
		},
	}

	result, err := NewService(api).Wait(WaitSpec{
		ID:       "account-1",
		Interval: 0,
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed {
		t.Fatalf("result should succeed: %+v", result)
	}
	if api.calls != 2 {
		t.Fatalf("calls = %d, want 2", api.calls)
	}
	row := result.Rows[0]
	if row["result"] != "success" || row["operation"] != "create-account" {
		t.Fatalf("row = %+v", row)
	}
}

func TestServiceWaitCloudAccountTimeout(t *testing.T) {
	api := &fakeWaitAPI{
		responses: []client.APIResponse{
			{Data: map[string]interface{}{"status": "creating"}},
		},
	}

	result, err := NewService(api).Wait(WaitSpec{
		ID:       "account-1",
		Interval: 0,
		Timeout:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Failed {
		t.Fatalf("result should fail: %+v", result)
	}
	row := result.Rows[0]
	if row["result"] != "timeout" || row["error"] != "wait timed out" {
		t.Fatalf("row = %+v", row)
	}
}

func TestServiceWaitCloudAccountAcceptsTopLevelCloudAccountPayload(t *testing.T) {
	api := &fakeWaitAPI{
		responses: []client.APIResponse{
			{
				Raw: map[string]interface{}{
					"cloud_account": map[string]interface{}{
						"status":         "available",
						"display_status": "Available",
					},
				},
			},
		},
	}

	result, err := NewService(api).Wait(WaitSpec{
		ID:       "account-1",
		Interval: 0,
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed {
		t.Fatalf("result should succeed: %+v", result)
	}
	row := result.Rows[0]
	if row["result"] != "success" || row["status"] != "available" {
		t.Fatalf("row = %+v", row)
	}
}
