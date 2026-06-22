package license

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath  string
	getQuery url.Values
	postPath string
	postBody interface{}
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	f.getQuery = q
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func (f *fakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postPath = path
	f.postBody = body
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceActivateRejectsMissingFields(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Activate(ActivateSpec{KKTY: "k"})
	if err == nil || err.Error() != "ddty is required" {
		t.Fatalf("err = %v", err)
	}
}
