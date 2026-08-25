package cloudaccountcreate

import "fmt"

func buildHuaweiObject(spec Spec) (string, map[string]interface{}, error) {
	accessID := firstNonEmptyString(spec.AccessID, spec.AccessKeyID)
	accessSecret := firstNonEmptyString(spec.AccessSecret, spec.AccessKeySecret)
	if accessID == "" {
		return "", nil, fmt.Errorf("access-key-id is required")
	}
	if accessSecret == "" {
		return "", nil, fmt.Errorf("access-key-secret is required")
	}
	if spec.RegionID == "" {
		return "", nil, fmt.Errorf("region-id is required")
	}

	host, err := buildHuaweiLinuxBootImageHostConfig(spec.LinuxBootImageHostConfig)
	if err != nil {
		return "", nil, err
	}
	regionName := firstNonEmptyString(spec.RegionName, spec.RegionID)
	metadata := map[string]interface{}{
		"cloud_type":                   "huawei_obs",
		"access_id":                    accessID,
		"access_secret":                accessSecret,
		"auth_region_id":               firstNonEmptyString(spec.AuthRegionID, spec.RegionID),
		"auth_project_id":              spec.AuthProjectID,
		"region_id":                    spec.RegionID,
		"region_name":                  regionName,
		"custom_name":                  firstNonEmptyString(spec.CustomName, "Huawei Cloud(Recommended, SDK v3.1.86)-"+regionName),
		"use_internal_ip_for_control":  firstNonEmptyString(spec.UseInternalIP, "0"),
		"linux_boot_image_id":          "make_image",
		"linux_boot_image_host_config": host,
	}
	if spec.ControlAccessIP != "" {
		metadata["control_access_ip"] = spec.ControlAccessIP
	}
	if spec.BootLoaderImageID != "" {
		metadata["boot_loader_image_id"] = spec.BootLoaderImageID
		metadata["boot_loader_image_name"] = firstNonEmptyString(spec.BootLoaderImageName, spec.BootLoaderImageID)
	}
	if spec.BootLoaderFlavorID != "" {
		metadata["boot_loader_flavor_id"] = spec.BootLoaderFlavorID
	}

	autoUploadImages := 1
	if spec.AutoUploadImages != nil {
		autoUploadImages = *spec.AutoUploadImages
	}
	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "huawei_obs",
			"cloud_auth_type": "aksk",
			"metadata":        metadata,
			"storage_type":    "objectstorage",
		},
		"auto_upload_images": autoUploadImages,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	}
	return "/hypermotion/v1/cloud_accounts", finalizeCreateBody(spec, body), nil
}

func buildHuaweiLinuxBootImageHostConfig(spec LinuxBootImageHostConfigSpec) (map[string]interface{}, error) {
	required := []struct {
		name  string
		value string
	}{
		{"zone-id", spec.ZoneID},
		{"flavor-id", spec.FlavorID},
		{"network-id", spec.NetworkID},
		{"subnet-id", spec.SubnetID},
		{"image-id", spec.ImageID},
		{"system-disk-type-id", spec.SystemDiskTypeID},
	}
	for _, field := range required {
		if field.value == "" {
			return nil, fmt.Errorf("linux-boot-image-host-config-%s could not be resolved", field.name)
		}
	}

	host := map[string]interface{}{
		"zone_id":               spec.ZoneID,
		"zone_name":             firstNonEmptyString(spec.ZoneName, spec.ZoneID),
		"flavor_id":             spec.FlavorID,
		"flavor_name":           firstNonEmptyString(spec.FlavorName, spec.FlavorID),
		"flavor_id_arr":         spec.FlavorIDArr,
		"network_id":            spec.NetworkID,
		"network_name":          firstNonEmptyString(spec.NetworkName, spec.NetworkID),
		"subnet_id":             spec.SubnetID,
		"subnet_name":           firstNonEmptyString(spec.SubnetName, spec.SubnetID),
		"image_id":              spec.ImageID,
		"image_name":            firstNonEmptyString(spec.ImageName, spec.ImageID),
		"system_disk_type_id":   spec.SystemDiskTypeID,
		"system_disk_type_name": firstNonEmptyString(spec.SystemDiskTypeName, spec.SystemDiskTypeID),
	}
	return host, nil
}
