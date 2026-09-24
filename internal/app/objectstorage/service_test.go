package objectstorage

import (
	"net/url"
	"reflect"
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

func TestServiceAssociatedResourcesBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.AssociatedResources(AssociatedResourcesSpec{ID: "storage-1"})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getStorageAssociatedResources" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("storage_id") != "storage-1" || api.getQuery.Get("with_statistics") != "false" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServicePrepareCreateDefaultsCloudTypeToCustom(t *testing.T) {
	service := NewService(&fakeAPI{})

	prepared, err := service.PrepareCreate(CreateProfile{Mode: "custom", ProviderID: "custom"}, CreateSpec{
		AuthURL:         "oss-cn-beijing.aliyuncs.com",
		RegionID:        "oss-cn-beijing",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		BucketName:      "bucket-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := prepared.Body
	if body["display_name"] != "custom-oss-cn-beijing" {
		t.Fatalf("body = %+v", body)
	}
	if body["cloud_type"] != "custom" {
		t.Fatalf("body = %+v", body)
	}
	metadata := body["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "custom" {
		t.Fatalf("body = %+v", body)
	}
}

func TestServicePrepareCreateAllowsEmptyRegionForCustom(t *testing.T) {
	service := NewService(&fakeAPI{})

	prepared, err := service.PrepareCreate(CreateProfile{Mode: "custom", ProviderID: "custom"}, CreateSpec{
		AuthURL:         "object-storage.example.invalid:9000",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		BucketName:      "bucket-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := prepared.Body
	if body["cloud_type"] != "custom" || body["display_name"] != "custom" {
		t.Fatalf("body = %+v", body)
	}
	config := body["config"].(map[string]interface{})
	if config["region_id"] != "" {
		t.Fatalf("body = %+v", body)
	}
	metadata := body["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "custom" {
		t.Fatalf("body = %+v", body)
	}
}

func TestServicePrepareCreatePreservesExplicitCloudType(t *testing.T) {
	service := NewService(&fakeAPI{})

	prepared, err := service.PrepareCreate(CreateProfile{
		Mode:       "catalog",
		ProviderID: "huaweicloud",
		RegionID:   "cn-north-1",
	}, CreateSpec{
		AuthURL:         "obs.cn-north-1.myhuaweicloud.com",
		RegionID:        "cn-north-1",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		BucketName:      "bucket-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := prepared.Body
	if body["cloud_type"] != "huaweicloud" || body["display_name"] != "huaweicloud-cn-north-1" {
		t.Fatalf("body = %+v", body)
	}
	metadata := body["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "huaweicloud,cn-north-1" {
		t.Fatalf("body = %+v", body)
	}
}

func TestServiceCreateUsesPreparedRequest(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)
	profile := CreateProfile{
		Mode:               "catalog",
		ProviderID:         "aliyun",
		RegionID:           "oss-cn-beijing",
		AuthURL:            "oss-cn-beijing.aliyuncs.com",
		PublicEndpoint:     "oss-cn-beijing.aliyuncs.com",
		InternalEndpoint:   "oss-cn-beijing-internal.aliyuncs.com",
		Protocol:           "s3",
		BucketLookup:       "dns",
		DefaultDisplayName: "Alibaba Cloud-Beijing",
	}
	spec := CreateSpec{
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		BucketName:      "bucket-1",
		UseTLS:          true,
	}
	prepared, err := service.PrepareCreate(profile, spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(profile, spec); err != nil {
		t.Fatal(err)
	}
	if api.postPath != prepared.Path || !reflect.DeepEqual(api.postBody, prepared.Body) {
		t.Fatalf("post path/body = %q/%+v, prepared = %q/%+v", api.postPath, api.postBody, prepared.Path, prepared.Body)
	}
}

func TestServiceDeleteBuildsRequest(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Delete(DeleteSpec{
		ID:    "storage-1",
		Force: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/api/v2/deleteStorage" {
		t.Fatalf("path = %q", api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	if body["storage_id"] != "storage-1" || body["force"] != true {
		t.Fatalf("body = %+v", body)
	}
}
