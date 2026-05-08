package blockstorage

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath  string
	getQ     url.Values
	getResp  client.APIResponse
	getErr   error
	postPath string
	postBody interface{}
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	f.getQ = q
	if f.getErr != nil {
		return client.APIResponse{}, f.getErr
	}
	if f.getResp.Data != nil || f.getResp.Raw != nil {
		return f.getResp, nil
	}
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func (f *fakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postPath = path
	f.postBody = body
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceResourcesBuildsGatewayActionRequest(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	result, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
		FetchRes:       "regions,zones",
		RegionID:       "cn-beijing",
		Purpose:        "make_hg",
		ImageType:      "system",
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts/account-1/action" {
		t.Fatalf("path = %q", api.postPath)
	}
	if len(result.Resources) != 2 || result.Resources[0] != "regions" || result.Resources[1] != "zones" {
		t.Fatalf("resources = %+v", result.Resources)
	}
}

func TestServiceResourcesBuildsHelperImageRequests(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
		FetchRes:       "win_hd_images,linux_hd_images",
		RegionID:       "cn-beijing",
		Purpose:        "make_hg",
	})
	if err != nil {
		t.Fatal(err)
	}

	body, ok := api.postBody.(map[string]interface{})
	if !ok {
		t.Fatalf("body = %#v", api.postBody)
	}
	getCloudInfo := body["get_cloud_info"].(map[string]interface{})
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	winOptions := resourceOptions["win_hd_images"].(map[string]interface{})
	linuxOptions := resourceOptions["linux_hd_images"].(map[string]interface{})
	if winOptions["image_type"] != "system" {
		t.Fatalf("win_hd_images = %#v", winOptions)
	}
	if linuxOptions["image_type"] != "user_create" {
		t.Fatalf("linux_hd_images = %#v", linuxOptions)
	}
}

func TestServiceResourcesUsesAccountRegionWhenRegionFlagIsOmitted(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)
	api.getResp = client.APIResponse{
		Data: map[string]interface{}{
			"region_id": "cn-beijing",
		},
	}

	result, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
		FetchRes:       "zones",
		Purpose:        "make_hg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Resources) != 1 || result.Resources[0] != "zones" {
		t.Fatalf("resources = %+v", result.Resources)
	}

	body, ok := api.postBody.(map[string]interface{})
	if !ok {
		t.Fatalf("body = %#v", api.postBody)
	}
	getCloudInfo := body["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if domain["region_id"] != "cn-beijing" {
		t.Fatalf("domain = %#v", domain)
	}
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	zoneOptions := resourceOptions["zones"].(map[string]interface{})
	if zoneOptions["region_id"] != "cn-beijing" {
		t.Fatalf("zones = %#v", zoneOptions)
	}
}

func TestServiceResourcesAllowsEmptyFetchResAndUsesAccountRegionWhenAvailable(t *testing.T) {
	api := &fakeAPI{
		getResp: client.APIResponse{
			Data: map[string]interface{}{
				"region_id": "cn-beijing",
			},
		},
	}
	service := NewService(api)

	result, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Resources) != 0 {
		t.Fatalf("resources = %+v", result.Resources)
	}

	body := api.postBody.(map[string]interface{})
	getCloudInfo := body["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if domain["region_id"] != "cn-beijing" {
		t.Fatalf("domain = %#v", domain)
	}
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	if len(resourceOptions) != 0 {
		t.Fatalf("resources_options = %#v", resourceOptions)
	}
}

func TestServiceResourcesAllowsEmptyFetchResWithoutRegion(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	result, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Resources) != 0 {
		t.Fatalf("resources = %+v", result.Resources)
	}

	body := api.postBody.(map[string]interface{})
	getCloudInfo := body["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if len(domain) != 0 {
		t.Fatalf("domain = %#v", domain)
	}
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	if len(resourceOptions) != 0 {
		t.Fatalf("resources_options = %#v", resourceOptions)
	}
}

func TestServiceOpenStackResourcesDoNotRequireRegionForHelperImages(t *testing.T) {
	api := &fakeAPI{
		getResp: client.APIResponse{
			Data: map[string]interface{}{
				"cloud_type": "openstack",
			},
		},
	}
	service := NewService(api)

	_, err := service.Resources(ResourcesSpec{
		CloudAccountID: "account-1",
		FetchRes:       "win_hd_images",
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts/account-1/action" {
		t.Fatalf("path = %q", api.postPath)
	}
	body := api.postBody.(map[string]interface{})
	getCloudInfo := body["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if _, ok := domain["region_id"]; ok {
		t.Fatalf("domain should omit region_id: %#v", domain)
	}
}

func TestServiceListBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.List(ListSpec{
		Page:        3,
		PageSize:    20,
		StorageType: "HyperGate",
		Query:       url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getStorages" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQ.Get("type") != "HyperGate" || api.getQ.Get("page") != "3" || api.getQ.Get("page_size") != "20" || api.getQ.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQ)
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

func TestServiceTransitionImagesBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.TransitionImages(TransitionImagesSpec{
		CloudAccountID: "account-1",
		CloudType:      "aliyun_bs",
		RegionID:       "cn-beijing",
		ZoneID:         "cn-beijing-h",
		Purpose:        "rebuild_bcd",
		ImageType:      "system",
		OSType:         "windows",
		BootMode:       "bios",
		Query:          url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v3/getCloudInfo" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQ.Get("cloud_type") != "aliyun_bs" || api.getQ.Get("fetch_res") != "images" || api.getQ.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQ)
	}
}

func TestServiceSubnetConfigUsesAccountRegionWhenRegionFlagIsOmitted(t *testing.T) {
	api := &fakeAPI{
		getResp: client.APIResponse{
			Data: map[string]interface{}{
				"region_id":  "cn-beijing",
				"cloud_type": "aliyun_bs",
			},
		},
	}
	service := NewService(api)

	_, err := service.SubnetConfig(SubnetConfigSpec{
		CloudAccountID: "account-1",
		ZoneID:         "cn-beijing-h",
		NetworkID:      "vpc-1",
		SubnetID:       "vsw-1",
		Query:          url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v3/getSubnetConfig" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQ.Get("region_id") != "cn-beijing" || api.getQ.Get("cloud_type") != "aliyun_bs" || api.getQ.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQ)
	}
}

func TestServiceSubnetConfigOmitsRegionWhenNeitherFlagNorAccountProvidesIt(t *testing.T) {
	api := &fakeAPI{
		getResp: client.APIResponse{
			Data: map[string]interface{}{
				"cloud_type": "aliyun_bs",
			},
		},
	}
	service := NewService(api)

	_, err := service.SubnetConfig(SubnetConfigSpec{
		CloudAccountID: "account-1",
		ZoneID:         "cn-beijing-h",
		NetworkID:      "vpc-1",
		SubnetID:       "vsw-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v3/getSubnetConfig" {
		t.Fatalf("path = %q", api.getPath)
	}
	if _, ok := api.getQ["region_id"]; ok {
		t.Fatalf("query should omit region_id: %+v", api.getQ)
	}
}

func TestServiceSubnetConfigRejectsOpenStack(t *testing.T) {
	api := &fakeAPI{
		getResp: client.APIResponse{
			Data: map[string]interface{}{
				"cloud_type": "openstack",
			},
		},
	}
	service := NewService(api)

	_, err := service.SubnetConfig(SubnetConfigSpec{
		CloudAccountID: "account-1",
		ZoneID:         "nova",
		NetworkID:      "net-1",
	})
	if err == nil || err.Error() != "subnet-config is not supported for openstack; use target cloud-sync-gateway create openstack --preview-request to inspect the resolved defaults" {
		t.Fatalf("err = %v", err)
	}
}
