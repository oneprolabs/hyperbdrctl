package targetresource

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPaths []string
	getQs    []url.Values
	getResps []client.APIResponse
	postPath string
	postBody interface{}
	postResp client.APIResponse
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPaths = append(f.getPaths, path)
	f.getQs = append(f.getQs, q)
	if len(f.getResps) == 0 {
		return client.APIResponse{Data: map[string]interface{}{}}, nil
	}
	resp := f.getResps[0]
	f.getResps = f.getResps[1:]
	return resp, nil
}

func (f *fakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postPath = path
	f.postBody = body
	if f.postResp.Data != nil || f.postResp.Raw != nil {
		return f.postResp, nil
	}
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestDirectAuthUsesGenericAuthEndpoint(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	result, err := service.DirectAuth(DirectAuthSpec{
		CloudType:   "aliyun_bs",
		StorageType: "HyperGate",
		FetchRes:    "regions",
		DynamicFields: map[string]string{
			"access_key_id":     "ak",
			"access_key_secret": "sk",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Route != routeDirectAuthGeneric || api.postPath != routeDirectAuthGeneric {
		t.Fatalf("route=%q path=%q", result.Route, api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "aliyun_bs" || cloudAccount["storage_type"] != "HyperGate" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account=%#v", cloudAccount)
	}
}

func TestDirectAuthOpenStackPasswordUsesTargetAuthEndpoint(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	result, err := service.DirectAuth(DirectAuthSpec{
		CloudType:   "openstack",
		StorageType: "objectstorage",
		DynamicFields: map[string]string{
			"auth_url":       "http://identity:5000/v3",
			"username":       "demo",
			"password":       "secret",
			"user_domain_id": "default",
			"project_id":     "project-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Route != routeDirectAuthOpenStack || api.postPath != routeDirectAuthOpenStack {
		t.Fatalf("route=%q path=%q", result.Route, api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	if body["project_id"] != "project-1" {
		t.Fatalf("body=%#v", body)
	}
}

func TestDirectAuthOpenStackObjectFlavorUsesGenericAuthCompatibilityEndpoint(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	result, err := service.DirectAuth(DirectAuthSpec{
		CloudType:   "openstack",
		StorageType: "objectstorage",
		FetchRes:    "flavor",
		DynamicFields: map[string]string{
			"auth_url":       "http://identity:5000/v3",
			"username":       "demo",
			"password":       "secret",
			"user_domain_id": "default",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Route != routeDirectAuthGeneric || api.postPath != routeDirectAuthGeneric {
		t.Fatalf("route=%q path=%q", result.Route, api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_account_id"] != "anonymous" {
		t.Fatalf("cloud_account=%#v", cloudAccount)
	}
	if body["fetch_res"] != "flavors" {
		t.Fatalf("body=%#v", body)
	}
	if body["rt_flatten"] != 1 {
		t.Fatalf("body=%#v", body)
	}
}

func TestDirectAuthHuaweiObjectPurposeIsTopLevel(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.DirectAuth(DirectAuthSpec{
		CloudType:   "huawei",
		StorageType: "objectstorage",
		FetchRes:    "flavors",
		Purpose:     "make_image",
		DynamicFields: map[string]string{
			"access_key_id":     "ak",
			"access_key_secret": "sk",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := api.postBody.(map[string]interface{})
	if body["purpose"] != "make_image" {
		t.Fatalf("body = %+v", body)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if _, ok := metadata["purpose"]; ok {
		t.Fatalf("metadata should not contain purpose: %+v", metadata)
	}
}

func TestDirectAuthKnownQueryFieldsAreTopLevel(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.DirectAuth(DirectAuthSpec{
		CloudType:   "huawei_obs",
		StorageType: "objectstorage",
		FetchRes:    "images",
		RegionID:    "cn-north-1",
		ZoneID:      "cn-north-1a",
		FlavorID:    "flavor-1",
		FlavorVCPUs: "2",
		FlavorRAM:   "4",
		NetworkID:   "network-1",
		OSType:      "linux",
		ImageType:   "system",
		BootMode:    "bios",
		Purpose:     "make_image",
		DynamicFields: map[string]string{
			"access_key_id":     "ak",
			"access_key_secret": "sk",
			"provider_option":   "kept-in-metadata",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := api.postBody.(map[string]interface{})
	want := map[string]string{
		"fetch_res": "images", "region_id": "cn-north-1", "zone_id": "cn-north-1a",
		"flavor_id": "flavor-1", "flavor_vcpus": "2", "flavor_ram": "4",
		"network_id": "network-1", "os_type": "linux", "image_type": "system",
		"boot_mode": "bios", "purpose": "make_image",
	}
	for key, value := range want {
		if body[key] != value {
			t.Fatalf("body[%q]=%v, want %q; body=%+v", key, body[key], value, body)
		}
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	for key := range want {
		if _, ok := metadata[key]; ok {
			t.Fatalf("metadata should not contain query field %q: %+v", key, metadata)
		}
	}
	for _, key := range []string{"region_type", "region_type_list"} {
		if _, ok := metadata[key]; ok {
			t.Fatalf("Huawei object metadata should not contain %q: %+v", key, metadata)
		}
	}
	if metadata["provider_option"] != "kept-in-metadata" {
		t.Fatalf("unmodeled provider field should remain in metadata: %+v", metadata)
	}
}

func TestDirectAuthAliyunRegionCompatibilityMetadataIsPreserved(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.DirectAuth(DirectAuthSpec{
		CloudType:   "aliyun_obs",
		StorageType: "objectstorage",
		FetchRes:    "images",
		RegionID:    "cn-beijing",
		DynamicFields: map[string]string{
			"access_key_id":     "ak",
			"access_key_secret": "sk",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := api.postBody.(map[string]interface{})
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["region_type"] != "1" || metadata["region_type_list"] != "cn-beijing" {
		t.Fatalf("Alibaba Cloud compatibility metadata changed: %+v", metadata)
	}
}

func TestFetchDefaultsToGetCloudInfo(t *testing.T) {
	api := &fakeAPI{
		getResps: []client.APIResponse{
			{Data: map[string]interface{}{
				"cloud_type":   "aliyun_bs",
				"storage_type": "HyperGate",
				"region_id":    "cn-beijing",
			}},
			{Data: map[string]interface{}{"cloud_info": map[string]interface{}{"regions": []interface{}{}}}},
		},
	}
	service := NewService(api)

	result, err := service.Fetch(AccountFetchSpec{
		CloudAccountID: "account-1",
		FetchRes:       "regions,zones",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Route != string(accountRouteGetCloudInfo) {
		t.Fatalf("route=%q", result.Route)
	}
	if len(api.getPaths) != 2 || api.getPaths[1] != routeAccountGetCloudInfo {
		t.Fatalf("paths=%v", api.getPaths)
	}
	if api.getQs[1].Get("rt_flatten") != "1" || api.getQs[1].Get("cloud_account_id") != "account-1" || api.getQs[1].Get("region_id") != "cn-beijing" {
		t.Fatalf("query=%v", api.getQs[1])
	}
}

func TestFetchSharedBlockResourcesStayOnGetCloudInfo(t *testing.T) {
	api := &fakeAPI{
		getResps: []client.APIResponse{
			{Data: map[string]interface{}{
				"cloud_type":   "aliyun_bs",
				"storage_type": "HyperGate",
			}},
			{Data: map[string]interface{}{}},
		},
	}
	service := NewService(api)

	result, err := service.Fetch(AccountFetchSpec{
		CloudAccountID: "account-1",
		FetchRes:       "images,networks",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Route != string(accountRouteGetCloudInfo) {
		t.Fatalf("route=%q", result.Route)
	}
	if api.postPath != "" {
		t.Fatalf("postPath=%q", api.postPath)
	}
}

func TestFetchActionRuleUsesCloudAccountAction(t *testing.T) {
	api := &fakeAPI{
		getResps: []client.APIResponse{
			{Data: map[string]interface{}{
				"cloud_type":   "aliyun_bs",
				"storage_type": "HyperGate",
				"region_id":    "cn-beijing",
			}},
		},
	}
	service := NewService(api)

	result, err := service.Fetch(AccountFetchSpec{
		CloudAccountID: "account-1",
		FetchRes:       "abilities",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Route != string(accountRouteAction) {
		t.Fatalf("route=%q", result.Route)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts/account-1/action" {
		t.Fatalf("path=%q", api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	getCloudInfo := body["get_cloud_info"].(map[string]interface{})
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	if _, ok := resourceOptions["abilities"]; !ok {
		t.Fatalf("resourceOptions=%#v", resourceOptions)
	}
}

func TestFetchInfersStorageTypeFromCloudType(t *testing.T) {
	api := &fakeAPI{
		getResps: []client.APIResponse{
			{Data: map[string]interface{}{
				"cloud_type": "huawei_obs",
			}},
			{Data: map[string]interface{}{}},
		},
	}
	service := NewService(api)

	_, err := service.Fetch(AccountFetchSpec{
		CloudAccountID: "account-1",
		FetchRes:       "regions",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := api.getQs[1].Get("storage_type"); got != "objectstorage" {
		t.Fatalf("storage_type=%q", got)
	}
}
