package objectstorage

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

func TestServiceBucketsBuildsValidatedRequest(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Buckets(BucketsSpec{
		AuthURL:         "oss-cn-beijing.aliyuncs.com",
		RegionID:        "oss-cn-beijing",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		Protocol:        "s3",
		BucketLookup:    "virtual-hosted-style",
		UseTLS:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/api/v2/objectStorageBuckets" {
		t.Fatalf("path = %q", api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	if body["bucket_lookup"] != "dns" {
		t.Fatalf("body = %+v", body)
	}
}

func TestServiceListBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.List(ListSpec{
		Page:        1,
		PageSize:    10,
		StorageType: "objectstorage",
		Query:       url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getStorages" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("type") != "objectstorage" || api.getQuery.Get("page") != "1" || api.getQuery.Get("page_size") != "10" || api.getQuery.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}
