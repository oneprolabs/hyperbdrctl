package cloudaccountcreate

import (
	"fmt"
	"strconv"
)

func NormalizeSpec(spec Spec) Spec {
	metadata := spec.MetadataOverrides

	spec.AccessKeyID = firstNonEmptyString(spec.AccessKeyID, overrideStringValue(metadata, "access_key_id"))
	spec.AccessKeySecret = firstNonEmptyString(spec.AccessKeySecret, overrideStringValue(metadata, "access_key_secret"))
	spec.RegionID = firstNonEmptyString(spec.RegionID, overrideStringValue(metadata, "region_id"))
	spec.RegionName = firstNonEmptyString(spec.RegionName, overrideStringValue(metadata, "region_name"))
	spec.AccountName = firstNonEmptyString(spec.AccountName, overrideStringValue(metadata, "account_name"))
	spec.AuthRegionID = firstNonEmptyString(spec.AuthRegionID, overrideStringValue(metadata, "auth_region_id"))
	spec.AuthURL = firstNonEmptyString(spec.AuthURL, overrideStringValue(metadata, "auth_url"))
	spec.CloudAccountUsername = firstNonEmptyString(spec.CloudAccountUsername, overrideStringValue(metadata, "username", "cloud_account_username"))
	spec.CloudAccountPassword = firstNonEmptyString(spec.CloudAccountPassword, overrideStringValue(metadata, "password", "cloud_account_password"))
	spec.UserDomainID = firstNonEmptyString(spec.UserDomainID, overrideStringValue(metadata, "user_domain_id"))
	spec.ProjectDomainID = firstNonEmptyString(spec.ProjectDomainID, overrideStringValue(metadata, "project_domain_id"))
	spec.ProjectID = firstNonEmptyString(spec.ProjectID, overrideStringValue(metadata, "project_id"))
	spec.ProjectName = firstNonEmptyString(spec.ProjectName, overrideStringValue(metadata, "project_name"))
	spec.UseInternalIP = firstNonEmptyString(spec.UseInternalIP, overrideStringValue(metadata, "use_internal_ip_for_control"))
	spec.BootLoaderImageID = firstNonEmptyString(spec.BootLoaderImageID, overrideStringValue(metadata, "boot_loader_image_id"))
	spec.BootLoaderImageName = firstNonEmptyString(spec.BootLoaderImageName, overrideStringValue(metadata, "boot_loader_image_name"))
	spec.BootLoaderFlavorID = firstNonEmptyString(spec.BootLoaderFlavorID, overrideStringValue(metadata, "boot_loader_flavor_id"))
	spec.LinuxBootImageID = firstNonEmptyString(spec.LinuxBootImageID, overrideStringValue(metadata, "linux_boot_image_id"))
	spec.WindowsBootImageID = firstNonEmptyString(spec.WindowsBootImageID, overrideStringValue(metadata, "windows_boot_image_id"))
	spec.LinuxUEFIBootImageID = firstNonEmptyString(spec.LinuxUEFIBootImageID, overrideStringValue(metadata, "linux_uefi_boot_image_id"))
	spec.WindowsUEFIBootImageID = firstNonEmptyString(spec.WindowsUEFIBootImageID, overrideStringValue(metadata, "windows_uefi_boot_image_id"))
	spec.CustomName = firstNonEmptyString(spec.CustomName, overrideStringValue(metadata, "custom_name"))
	spec.DiskBusTypeID = firstNonEmptyString(spec.DiskBusTypeID, overrideStringValue(metadata, "disk_bus_type_id"))
	spec.DiskBusTypeName = firstNonEmptyString(spec.DiskBusTypeName, overrideStringValue(metadata, "disk_bus_type_name"))
	spec.SSHPort = firstNonEmptyString(spec.SSHPort, overrideStringValue(metadata, "ssh_port"))
	spec.SSHPass = firstNonEmptyString(spec.SSHPass, overrideStringValue(metadata, "ssh_pass"))
	spec.LinuxHDUsername = firstNonEmptyString(spec.LinuxHDUsername, overrideStringValue(metadata, "linux_hd_username"))
	spec.LinuxHDPassword = firstNonEmptyString(spec.LinuxHDPassword, overrideStringValue(metadata, "linux_hd_password"))
	spec.LinuxHDPort = firstNonEmptyString(spec.LinuxHDPort, overrideStringValue(metadata, "linux_hd_port"))

	if spec.AutoUploadImages == nil {
		if value, ok := overrideIntValue(spec.RequestOverrides, "auto_upload_images"); ok {
			spec.AutoUploadImages = &value
		}
	}
	if spec.OnlyVerify == nil {
		if value, ok := overrideBoolValue(spec.RequestOverrides, "only_verify"); ok {
			spec.OnlyVerify = &value
		}
	}
	if spec.UploadUEFIImage == nil {
		if value, ok := overrideIntValue(metadata, "upload_uefi_image"); ok {
			spec.UploadUEFIImage = &value
		}
	}

	return spec
}

func overrideStringValue(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if values == nil {
			return ""
		}
		if value, ok := values[key]; ok {
			switch typed := value.(type) {
			case string:
				if typed != "" {
					return typed
				}
			case fmt.Stringer:
				text := typed.String()
				if text != "" {
					return text
				}
			case int:
				return strconv.Itoa(typed)
			case int64:
				return strconv.FormatInt(typed, 10)
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			case bool:
				return strconv.FormatBool(typed)
			}
		}
	}
	return ""
}

func overrideIntValue(values map[string]interface{}, key string) (int, bool) {
	if values == nil {
		return 0, false
	}
	value, ok := values[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case string:
		parsed, err := strconv.Atoi(typed)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func overrideBoolValue(values map[string]interface{}, key string) (bool, bool) {
	if values == nil {
		return false, false
	}
	value, ok := values[key]
	if !ok {
		return false, false
	}
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err == nil {
			return parsed, true
		}
	}
	return false, false
}
