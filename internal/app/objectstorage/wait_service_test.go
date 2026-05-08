package objectstorage

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

func TestServiceWaitObjectStorageSucceedsAfterPolling(t *testing.T) {
	api := &fakeWaitAPI{
		responses: []client.APIResponse{
			{Data: map[string]interface{}{"storage": map[string]interface{}{"type": "objectstorage", "status": "creating"}}},
			{Data: map[string]interface{}{"storage": map[string]interface{}{"type": "objectstorage", "status": "available", "display_status": "Available"}}},
		},
	}

	result, err := NewService(api).Wait(WaitSpec{
		ID:       "storage-1",
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
	if row["result"] != "success" || row["operation"] != "create-oss" {
		t.Fatalf("row = %+v", row)
	}
}

func TestServiceWaitObjectStorageTypeMismatchFails(t *testing.T) {
	api := &fakeWaitAPI{
		responses: []client.APIResponse{
			{Data: map[string]interface{}{"storage": map[string]interface{}{"type": "HyperGate", "status": "creating"}}},
		},
	}

	result, err := NewService(api).Wait(WaitSpec{
		ID:       "storage-1",
		Interval: 0,
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Failed {
		t.Fatalf("result should fail: %+v", result)
	}
	row := result.Rows[0]
	if row["result"] != "failed" {
		t.Fatalf("row = %+v", row)
	}
	if row["error"] != "resource type mismatch: expected objectstorage, got hypergate" {
		t.Fatalf("row = %+v", row)
	}
}
