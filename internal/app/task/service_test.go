package task

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath  string
	getQuery url.Values
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	f.getQuery = q
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceStepsRequiresTaskID(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Steps(StepsSpec{})
	if err == nil || err.Error() != "task-id is required" {
		t.Fatalf("err = %v", err)
	}
}
