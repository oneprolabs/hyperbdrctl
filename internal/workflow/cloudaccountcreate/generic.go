package cloudaccountcreate

import (
	"fmt"
	"strings"
)

func buildGenericBlock(spec Spec) (string, map[string]interface{}, error) {
	authType, metadata, err := buildGenericMetadata(spec, false)
	if err != nil {
		return "", nil, err
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": authType,
			"metadata":        metadata,
			"storage_type":    nil,
		},
		"auto_upload_images": nil,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	}
	return "/hypermotion/v1/cloud_accounts", finalizeCreateBody(spec, body), nil
}

func buildGenericObject(spec Spec) (string, map[string]interface{}, error) {
	authType, metadata, err := buildGenericMetadata(spec, true)
	if err != nil {
		return "", nil, err
	}

	autoUploadImages := 1
	if spec.AutoUploadImages != nil {
		autoUploadImages = *spec.AutoUploadImages
	} else if hasExplicitObjectImage(spec) {
		autoUploadImages = 0
	}

	uploadUEFIImage := 1
	if spec.UploadUEFIImage != nil {
		uploadUEFIImage = *spec.UploadUEFIImage
	}
	metadata["upload_uefi_image"] = uploadUEFIImage

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": authType,
			"metadata":        metadata,
			"storage_type":    "objectstorage",
		},
		"auto_upload_images": autoUploadImages,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	}
	return "/hypermotion/v1/cloud_accounts", finalizeCreateBody(spec, body), nil
}

func buildGenericMetadata(spec Spec, isObject bool) (string, map[string]interface{}, error) {
	authType, err := resolveGenericAuthType(spec)
	if err != nil {
		return "", nil, err
	}

	metadata := map[string]interface{}{}
	switch authType {
	case "aksk":
		if err := setGenericAKSKMetadata(metadata, spec); err != nil {
			return "", nil, err
		}
	case "password":
		if spec.AuthURL == "" {
			return "", nil, fmt.Errorf("auth-url is required")
		}
		if spec.CloudAccountUsername == "" {
			return "", nil, fmt.Errorf("username is required")
		}
		if spec.CloudAccountPassword == "" {
			return "", nil, fmt.Errorf("password is required")
		}
		metadata["auth_url"] = spec.AuthURL
		metadata["username"] = spec.CloudAccountUsername
		metadata["password"] = spec.CloudAccountPassword
	default:
		return "", nil, fmt.Errorf("cloud-auth-type must be aksk or password")
	}

	setMetadataString(metadata, "cloud_type", spec.CloudType)
	setMetadataString(metadata, "account_name", spec.AccountName)
	setMetadataString(metadata, "auth_region_id", spec.AuthRegionID)
	setMetadataString(metadata, "user_domain_id", spec.UserDomainID)
	setMetadataString(metadata, "project_domain_id", spec.ProjectDomainID)
	setMetadataString(metadata, "project_id", spec.ProjectID)
	setMetadataString(metadata, "project_name", spec.ProjectName)
	setMetadataString(metadata, "region_id", spec.RegionID)
	setMetadataString(metadata, "region_name", spec.RegionName)
	setMetadataString(metadata, "ssh_port", spec.SSHPort)
	setMetadataString(metadata, "ssh_pass", spec.SSHPass)
	setMetadataString(metadata, "linux_hd_username", spec.LinuxHDUsername)
	setMetadataString(metadata, "linux_hd_password", spec.LinuxHDPassword)
	setMetadataString(metadata, "linux_hd_port", spec.LinuxHDPort)

	if spec.RegionID != "" {
		metadata["region_type"] = "1"
		metadata["region_type_list"] = spec.RegionID
		metadata["region_type_input"] = ""
		metadata["region_text"] = ""
		if spec.RegionName != "" {
			metadata["region_type_list_name"] = spec.RegionName
		}
		if authType == "aksk" && spec.AuthRegionID == "" {
			metadata["auth_region_id"] = spec.RegionID
		}
	}

	if isObject {
		metadata["use_internal_ip_for_control"] = firstNonEmptyString(spec.UseInternalIP, "0")
		setMetadataString(metadata, "custom_name", spec.CustomName)
		setMetadataString(metadata, "boot_loader_image_id", spec.BootLoaderImageID)
		setMetadataString(metadata, "boot_loader_image_name", firstNonEmptyString(spec.BootLoaderImageName, spec.BootLoaderImageID))
		setMetadataString(metadata, "boot_loader_flavor_id", spec.BootLoaderFlavorID)
		setMetadataString(metadata, "disk_bus_type_id", spec.DiskBusTypeID)
		setMetadataString(metadata, "disk_bus_type_name", spec.DiskBusTypeName)

		resolveObjectImage(metadata, "linux_boot_image", spec.LinuxBootImageID)
		resolveObjectImage(metadata, "windows_boot_image", spec.WindowsBootImageID)
		resolveObjectImage(metadata, "linux_uefi_boot_image", spec.LinuxUEFIBootImageID)
		resolveObjectImage(metadata, "windows_uefi_boot_image", spec.WindowsUEFIBootImageID)
	}

	return authType, metadata, nil
}

func resolveObjectImage(metadata map[string]interface{}, prefix, id string) {
	if id == "" {
		metadata[prefix+"_id"] = "auto_upload"
		metadata[prefix+"_name"] = autoUploadLabel
		return
	}
	metadata[prefix+"_id"] = id
	metadata[prefix+"_name"] = id
}

func setMetadataString(metadata map[string]interface{}, key, value string) {
	if value != "" {
		metadata[key] = value
	}
}

func hasExplicitObjectImage(spec Spec) bool {
	return isManualObjectImage(spec.LinuxBootImageID) ||
		isManualObjectImage(spec.WindowsBootImageID) ||
		isManualObjectImage(spec.LinuxUEFIBootImageID) ||
		isManualObjectImage(spec.WindowsUEFIBootImageID)
}

func isManualObjectImage(id string) bool {
	return id != "" && id != "auto_upload" && id != "make_image"
}

func normalizeAuthType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func resolveGenericAuthType(spec Spec) (string, error) {
	authType := normalizeAuthType(spec.CloudAuthType)
	if authType != "" {
		switch authType {
		case "aksk", "password":
			return authType, nil
		default:
			return "", fmt.Errorf("cloud-auth-type must be aksk or password")
		}
	}

	if spec.HasDirectAKSKStyle || spec.HasDirectPasswordStyle {
		switch {
		case spec.HasDirectAKSKStyle && spec.HasDirectPasswordStyle:
			return "", fmt.Errorf("multiple credential styles provided; pass --cloud-auth-type explicitly")
		case spec.HasDirectAKSKStyle:
			return "aksk", nil
		default:
			return "password", nil
		}
	}

	hasAKSKStyle := hasAnyValue(spec.AccessKeyID, spec.AccessKeySecret, spec.AccessID, spec.AccessSecret)
	hasPasswordStyle := hasAnyValue(spec.CloudAccountUsername, spec.CloudAccountPassword)

	switch {
	case hasAKSKStyle && hasPasswordStyle:
		return "", fmt.Errorf("multiple credential styles provided; pass --cloud-auth-type explicitly")
	case hasAKSKStyle:
		return "aksk", nil
	case hasPasswordStyle:
		return "password", nil
	default:
		return "", fmt.Errorf("cloud-auth-type is required")
	}
}

func setGenericAKSKMetadata(metadata map[string]interface{}, spec Spec) error {
	switch {
	case spec.AccessKeyID != "" || spec.AccessKeySecret != "":
		if spec.AccessKeyID == "" {
			return fmt.Errorf("access-key-id is required")
		}
		if spec.AccessKeySecret == "" {
			return fmt.Errorf("access-key-secret is required")
		}
		metadata["access_key_id"] = spec.AccessKeyID
		metadata["access_key_secret"] = spec.AccessKeySecret
		return nil
	default:
		if spec.AccessID == "" {
			return fmt.Errorf("access-id is required")
		}
		if spec.AccessSecret == "" {
			return fmt.Errorf("access-secret is required")
		}
		metadata["access_id"] = spec.AccessID
		metadata["access_secret"] = spec.AccessSecret
		return nil
	}
}

func hasAnyValue(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}
