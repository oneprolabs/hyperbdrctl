package cloudaccountcreate

import "fmt"

func buildOpenStackObject(spec Spec) (string, map[string]interface{}, error) {
	if spec.AuthURL == "" {
		return "", nil, fmt.Errorf("auth-url is required")
	}
	if spec.CloudAccountUsername == "" {
		return "", nil, fmt.Errorf("cloud-account-username is required")
	}
	if spec.CloudAccountPassword == "" {
		return "", nil, fmt.Errorf("cloud-account-password is required")
	}
	if spec.UserDomainID == "" {
		return "", nil, fmt.Errorf("user-domain-id is required")
	}
	if spec.ProjectDomainID == "" {
		return "", nil, fmt.Errorf("project-domain-id is required")
	}
	if spec.ProjectID == "" {
		return "", nil, fmt.Errorf("project-id is required")
	}
	if spec.ProjectName == "" {
		return "", nil, fmt.Errorf("project-name is required")
	}
	if spec.RegionID == "" {
		return "", nil, fmt.Errorf("region-id is required")
	}
	if spec.RegionName == "" {
		return "", nil, fmt.Errorf("region-name is required")
	}
	if spec.BootLoaderImageID == "" {
		return "", nil, fmt.Errorf("boot-loader-image-id is required")
	}
	if spec.BootLoaderFlavorID == "" {
		return "", nil, fmt.Errorf("boot-loader-flavor-id is required")
	}
	if spec.DiskBusTypeID == "" {
		return "", nil, fmt.Errorf("disk-bus-type-id is required")
	}
	if spec.DiskBusTypeName == "" {
		return "", nil, fmt.Errorf("disk-bus-type-name is required")
	}

	autoUploadImages := 1
	if spec.AutoUploadImages != nil {
		autoUploadImages = *spec.AutoUploadImages
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "openstack",
			"cloud_auth_type": "password",
			"metadata": map[string]interface{}{
				"cloud_type":                  "openstack",
				"auth_url":                    spec.AuthURL,
				"user_domain_id":              spec.UserDomainID,
				"username":                    spec.CloudAccountUsername,
				"password":                    spec.CloudAccountPassword,
				"project_domain_id":           spec.ProjectDomainID,
				"project_id":                  spec.ProjectID,
				"project_name":                spec.ProjectName,
				"region_id":                   spec.RegionID,
				"region_name":                 spec.RegionName,
				"linux_boot_image_id":         firstNonEmptyString(spec.LinuxBootImageID, "auto_upload"),
				"linux_boot_image_name":       autoUploadLabel,
				"windows_boot_image_id":       firstNonEmptyString(spec.WindowsBootImageID, "auto_upload"),
				"windows_boot_image_name":     autoUploadLabel,
				"use_internal_ip_for_control": firstNonEmptyString(spec.UseInternalIP, "0"),
				"custom_name":                 spec.CustomName,
				"skip_driver_fix":             "0",
				"boot_loader_image_id":        spec.BootLoaderImageID,
				"boot_loader_image_name":      firstNonEmptyString(spec.BootLoaderImageName, spec.BootLoaderImageID),
				"disk_bus_type_id":            spec.DiskBusTypeID,
				"disk_bus_type_name":          spec.DiskBusTypeName,
				"boot_loader_flavor_id":       spec.BootLoaderFlavorID,
			},
			"storage_type": "objectstorage",
		},
		"auto_upload_images": autoUploadImages,
	}
	return "/hypermotion/v1/cloud_accounts", body, nil
}
