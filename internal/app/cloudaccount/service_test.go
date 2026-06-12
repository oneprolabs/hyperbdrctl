package cloudaccount

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath    string
	getQuery   url.Values
	postPath   string
	postBody   interface{}
	deletePath string
	deleteBody interface{}
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

func (f *fakeAPI) Delete(path string, body interface{}) (client.APIResponse, error) {
	f.deletePath = path
	f.deleteBody = body
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceCreateBuildsAndPostsRequest(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	resp, err := service.Create(CreateSpec{
		CloudType:       "aliyun_bs",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		RegionID:        "cn-beijing",
		RegionName:      "Beijing",
		OnlyVerify:      boolPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", api.postPath)
	}
	if resp.Data == nil {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestServicePrepareCreateRetainsBlockRequestOverrides(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	prepared, err := service.PrepareCreate(CreateSpec{
		CloudType:       "aliyun_bs",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		RegionID:        "cn-beijing",
		RequestOverrides: map[string]interface{}{
			"only_verify": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if prepared.Path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", prepared.Path)
	}
	if prepared.Body["only_verify"] != true {
		t.Fatalf("body = %+v", prepared.Body)
	}
}

func TestServiceCreateRawRoutesBlockPayload(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.CreateRaw(CreateRawSpec{
		Body: map[string]interface{}{
			"cloud_account": map[string]interface{}{
				"storage_type": "HyperGate",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", api.postPath)
	}
}

func TestServiceCreateRawRoutesObjectPayload(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.CreateRaw(CreateRawSpec{
		Body: map[string]interface{}{
			"cloud_account": map[string]interface{}{
				"storage_type": "objectstorage",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", api.postPath)
	}
}

func TestServiceCreateRawRejectsMissingStorageType(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.CreateRaw(CreateRawSpec{
		Body: map[string]interface{}{
			"cloud_account": map[string]interface{}{},
		},
	})
	if err == nil || err.Error() != "cloud account create body storage_type is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestServiceFetchResourcesBuildsRegionDiscoveryRequestWithoutRegionOrBootMode(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.FetchResources(FetchResourcesSpec{
		CloudType:       "aliyun_bs",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		StorageType:     "HyperGate",
		FetchRes:        "regions",
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/api/v3/postCloudInfoForAuth" {
		t.Fatalf("path = %q", api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	if body["fetch_res"] != "regions" {
		t.Fatalf("body = %+v", body)
	}
	if _, ok := body["region_id"]; ok {
		t.Fatalf("body should not contain region_id: %+v", body)
	}
	if _, ok := body["boot_mode"]; ok {
		t.Fatalf("body should not contain boot_mode: %+v", body)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if _, ok := metadata["region_type"]; ok {
		t.Fatalf("metadata should not contain region_type: %+v", metadata)
	}
	if _, ok := metadata["region_type_list"]; ok {
		t.Fatalf("metadata should not contain region_type_list: %+v", metadata)
	}
}

func TestServiceFetchResourcesIncludesBootModeWhenProvided(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.FetchResources(FetchResourcesSpec{
		CloudType:       "aliyun_obs",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		RegionID:        "cn-beijing",
		FetchRes:        "boot_loader_images",
		BootMode:        "bios",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := api.postBody.(map[string]interface{})
	if body["boot_mode"] != "bios" || body["region_id"] != "cn-beijing" {
		t.Fatalf("body = %+v", body)
	}
}

func TestServiceListBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.List(ListSpec{
		Page:        2,
		PageSize:    50,
		StorageType: "objectstorage",
		Query:       url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getCloudAccounts" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("storage_type") != "objectstorage" || api.getQuery.Get("page") != "2" || api.getQuery.Get("page_size") != "50" || api.getQuery.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceListOmitsStorageTypeWhenNotProvided(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.List(ListSpec{
		Page:     1,
		PageSize: 10,
		Query:    url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getQuery.Get("storage_type") != "" {
		t.Fatalf("query should omit storage_type when not provided: %+v", api.getQuery)
	}
	if api.getQuery.Get("page") != "1" || api.getQuery.Get("page_size") != "10" || api.getQuery.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceDetailRequiresID(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Detail(DetailSpec{})
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestServiceDeleteBuildsPathAndBody(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Delete(DeleteSpec{
		ID:          "account-1",
		StorageType: "HyperGate",
		Force:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.deletePath != "/hypermotion/v1/cloud_accounts/account-1?force=true" {
		t.Fatalf("path = %q", api.deletePath)
	}
	body := api.deleteBody.(map[string]interface{})
	if body["id"] != "account-1" || body["storage_type"] != "HyperGate" {
		t.Fatalf("body = %+v", body)
	}
}

func boolPtr(v bool) *bool { return &v }
