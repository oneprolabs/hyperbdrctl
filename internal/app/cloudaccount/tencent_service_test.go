package cloudaccount

import "testing"

func TestTencentServiceNormalizesAppIDAndPreparesWithoutPosting(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)
	spec := CreateSpec{CloudType: "tencent_bs", MetadataOverrides: map[string]interface{}{
		"access_key_id": "test-ak", "access_key_secret": "test-sk", "app_id": 1000000001, "auth_region_id": "ap-beijing",
	}}
	prepared, err := service.PrepareCreate(spec)
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "" {
		t.Fatalf("prepare posted to %q", api.postPath)
	}
	metadata := prepared.Body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["app_id"] != "1000000001" {
		t.Fatalf("metadata = %#v", metadata)
	}
	if _, err := service.Create(spec); err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("post path = %q", api.postPath)
	}
}

func TestTencentServiceFetchResourcesIncludesAppID(t *testing.T) {
	api := &fakeAPI{}
	_, err := NewService(api).FetchResources(FetchResourcesSpec{
		Spec:     CreateSpec{CloudType: "tencent_bs", StorageType: "HyperGate", AccessKeyID: "test-ak", AccessKeySecret: "test-sk", AppID: "1000000001"},
		FetchRes: "regions",
	})
	if err != nil {
		t.Fatal(err)
	}
	body := api.postBody.(map[string]interface{})
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if api.postPath != "/api/v3/postCloudInfoForAuth" || metadata["app_id"] != "1000000001" {
		t.Fatalf("path=%q metadata=%#v", api.postPath, metadata)
	}
}
