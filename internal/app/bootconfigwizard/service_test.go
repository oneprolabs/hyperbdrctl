package bootconfigwizard

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	path string
	q    url.Values
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.path = path
	f.q = q
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceTargetAuthInfoBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.TargetAuthInfo(TargetAuthInfoSpec{
		CloudAccountID: "account-1",
		CloudType:      "aliyun_obs",
		StorageType:    "objectstorage",
		FetchRes:       "regions,zones",
		Query:          url.Values{"custom_number": []string{"3"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.path != "/api/v3/getCloudInfo" {
		t.Fatalf("path = %q", api.path)
	}
	if api.q.Get("cloud_account_id") != "account-1" || api.q.Get("custom_number") != "3" {
		t.Fatalf("query = %+v", api.q)
	}
}
