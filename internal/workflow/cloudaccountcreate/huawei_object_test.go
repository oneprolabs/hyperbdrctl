package cloudaccountcreate

import "testing"

func TestBuildHuaweiObjectCreatesSingleMakeImageHostConfig(t *testing.T) {
	path, body, err := BuildRequest(Spec{
		CloudType:           "huawei_obs",
		StorageType:         "objectstorage",
		AccessKeyID:         "ak",
		AccessKeySecret:     "sk",
		RegionID:            "cn-north-1",
		RegionName:          "North China - Beijing 1",
		CustomName:          "custom",
		BootLoaderImageID:   "boot-image-1",
		BootLoaderImageName: "Windows transition image",
		BootLoaderFlavorID:  "boot-flavor-1",
		LinuxBootImageHostConfig: LinuxBootImageHostConfigSpec{
			ZoneID: "zone-1", ZoneName: "Zone 1",
			FlavorID: "flavor-1", FlavorName: "2C4G", FlavorIDArr: []string{"u-2", "u-2-m-4", "flavor-1"},
			NetworkID: "network-1", NetworkName: "vpc-a",
			SubnetID: "subnet-1", SubnetName: "subnet-a",
			ImageID: "image-1", ImageName: "Ubuntu",
			SystemDiskTypeID: "ssd", SystemDiskTypeName: "SSD",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if cloudAccount["cloud_auth_type"] != "aksk" || metadata["access_id"] != "ak" || metadata["access_secret"] != "sk" {
		t.Fatalf("cloudAccount = %+v", cloudAccount)
	}
	if metadata["linux_boot_image_id"] != "make_image" {
		t.Fatalf("metadata = %+v", metadata)
	}
	host := metadata["linux_boot_image_host_config"].(map[string]interface{})
	if host["zone_id"] != "zone-1" || host["flavor_id"] != "flavor-1" || host["image_id"] != "image-1" {
		t.Fatalf("host = %+v", host)
	}
	for _, forbidden := range []string{
		"boot_image_source", "skip_driver_fix", "linux_boot_image_name", "windows_boot_image_id",
		"linux_uefi_boot_image_id", "windows_uefi_boot_image_id", "upload_uefi_image", "region_type", "region_type_list",
		"control_access_ip_radio", "control_access_ip",
	} {
		if _, ok := metadata[forbidden]; ok {
			t.Fatalf("metadata must not contain %s: %+v", forbidden, metadata)
		}
	}
	for _, forbidden := range []string{"bandwidth_id", "bandwidth_name"} {
		if _, ok := host[forbidden]; ok {
			t.Fatalf("host must not contain %s: %+v", forbidden, host)
		}
	}
}
