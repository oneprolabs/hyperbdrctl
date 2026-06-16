package bootconfigquery

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

func TestServiceResourcesBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
		CloudType:      "aliyun_obs",
		StorageType:    "objectstorage",
		FetchRes:       "regions,zones",
		ZoneID:         "cn-beijing-h",
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
	if api.q.Get("cloud_type") != "aliyun_obs" || api.q.Get("storage_type") != "objectstorage" {
		t.Fatalf("query = %+v", api.q)
	}
}

func TestServiceResourcesRequiresCloudAccountID(t *testing.T) {
	service := NewService(&fakeAPI{})
	_, err := service.Resources(ResourcesSpec{CloudType: "aliyun_obs", StorageType: "objectstorage"})
	if err == nil || err.Error() != "cloud-account-id is required" {
		t.Fatalf("err = %v", err)
	}
}
