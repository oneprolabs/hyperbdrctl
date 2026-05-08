package cloudaccountcreate

import "testing"

func TestBuildRequestFallsBackToGenericBlock(t *testing.T) {
	path, body, err := BuildRequest(Spec{
		CloudType:       "huawei_bs",
		CloudAuthType:   "aksk",
		StorageType:     "block",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		RegionID:        "cn-north-1",
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "huawei_bs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %#v", cloudAccount)
	}
}

func TestBuildRequestFallsBackToGenericObject(t *testing.T) {
	path, body, err := BuildRequest(Spec{
		CloudType:            "vmware_obs",
		CloudAuthType:        "password",
		StorageType:          "objectstorage",
		AuthURL:              "https://vc.example.invalid",
		CloudAccountUsername: "admin",
		CloudAccountPassword: "secret",
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "vmware_obs" || cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %#v", cloudAccount)
	}
}

func TestBuildRequestAliyunObjectDefaultsImageSlotsToAutoUpload(t *testing.T) {
	_, body, err := BuildRequest(Spec{
		CloudType:         "aliyun_obs",
		StorageType:       "objectstorage",
		AccessKeyID:       "ak",
		AccessKeySecret:   "sk",
		RegionID:          "cn-beijing",
		BootLoaderImageID: "boot-loader-image",
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	if body["auto_upload_images"] != 1 {
		t.Fatalf("body = %+v", body)
	}

	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	for _, key := range []string{
		"linux_boot_image_id",
		"windows_boot_image_id",
		"linux_uefi_boot_image_id",
		"windows_uefi_boot_image_id",
	} {
		if metadata[key] != "auto_upload" {
			t.Fatalf("metadata[%s] = %v", key, metadata[key])
		}
	}
	if metadata["region_name"] != "cn-beijing" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if metadata["custom_name"] != "阿里云(推荐使用，SDK v2.0)-cn-beijing" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if _, ok := metadata["boot_loader_flavor_id"]; ok {
		t.Fatalf("metadata = %+v", metadata)
	}
}
