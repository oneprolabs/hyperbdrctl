package cloudaccountcreate

import "fmt"

func buildAliyunObject(spec Spec) (string, map[string]interface{}, error) {
	if spec.AccessKeyID == "" {
		return "", nil, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return "", nil, fmt.Errorf("access-key-secret is required")
	}
	if spec.RegionID == "" {
		return "", nil, fmt.Errorf("region-id is required")
	}
	if spec.BootLoaderImageID == "" {
		return "", nil, fmt.Errorf("boot-loader-image-id is required when no boot_loader_images candidate can be auto-selected")
	}

	uploadUEFIImage := 1
	if spec.UploadUEFIImage != nil {
		uploadUEFIImage = *spec.UploadUEFIImage
	}

	resolveImage := func(id string) (string, string, bool) {
		if id != "" && id != "auto_upload" {
			return id, id, true
		}
		return "auto_upload", autoUploadLabel, false
	}
	linuxBootImageID, linuxBootImageName, linuxBootExplicit := resolveImage(spec.LinuxBootImageID)
	windowsBootImageID, windowsBootImageName, windowsBootExplicit := resolveImage(spec.WindowsBootImageID)
	linuxUEFIBootImageID, linuxUEFIBootImageName, linuxUEFIExplicit := resolveImage(spec.LinuxUEFIBootImageID)
	windowsUEFIBootImageID, windowsUEFIBootImageName, windowsUEFIExplicit := resolveImage(spec.WindowsUEFIBootImageID)

	autoUploadImages := 1
	if linuxBootExplicit || windowsBootExplicit || linuxUEFIExplicit || windowsUEFIExplicit {
		autoUploadImages = 0
	}

	regionName := firstNonEmptyString(spec.RegionName, spec.RegionID)
	customName := firstNonEmptyString(spec.CustomName, defaultAliyunObjectCustomName(regionName))

	metadata := map[string]interface{}{
		"cloud_type":                   "aliyun_obs",
		"access_key_id":                spec.AccessKeyID,
		"access_key_secret":            spec.AccessKeySecret,
		"region_type":                  "1",
		"region_type_list":             spec.RegionID,
		"region_type_input":            "",
		"region_id":                    spec.RegionID,
		"region_name":                  regionName,
		"linux_boot_image_id":          linuxBootImageID,
		"linux_boot_image_name":        linuxBootImageName,
		"windows_boot_image_id":        windowsBootImageID,
		"windows_boot_image_name":      windowsBootImageName,
		"linux_uefi_boot_image_id":     linuxUEFIBootImageID,
		"linux_uefi_boot_image_name":   linuxUEFIBootImageName,
		"windows_uefi_boot_image_id":   windowsUEFIBootImageID,
		"windows_uefi_boot_image_name": windowsUEFIBootImageName,
		"use_internal_ip_for_control":  firstNonEmptyString(spec.UseInternalIP, "0"),
		"custom_name":                  customName,
		"boot_loader_image_id":         spec.BootLoaderImageID,
		"boot_loader_image_name":       firstNonEmptyString(spec.BootLoaderImageName, spec.BootLoaderImageID),
		"skip_driver_fix":              "1",
		"upload_uefi_image":            uploadUEFIImage,
	}
	if spec.BootLoaderFlavorID != "" {
		metadata["boot_loader_flavor_id"] = spec.BootLoaderFlavorID
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "aliyun_obs",
			"cloud_auth_type": "aksk",
			"metadata":        metadata,
			"storage_type":    "objectstorage",
		},
		"auto_upload_images": autoUploadImages,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	}
	return "/hypermotion/v1/cloud_accounts", finalizeCreateBody(spec, body), nil
}
