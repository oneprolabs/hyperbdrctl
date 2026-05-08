package cloudaccountcreate

import "fmt"

func buildOpenStackBlock(spec Spec) (string, map[string]interface{}, error) {
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
	if spec.ProjectName == "" {
		return "", nil, fmt.Errorf("project-name is required")
	}
	if spec.RegionName == "" {
		return "", nil, fmt.Errorf("region-name is required")
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "openstack",
			"cloud_auth_type": "password",
			"metadata": map[string]interface{}{
				"auth_url":             spec.AuthURL,
				"user_domain_id":       spec.UserDomainID,
				"username":             spec.CloudAccountUsername,
				"password":             spec.CloudAccountPassword,
				"project_domain_id":    spec.ProjectDomainID,
				"project_name":         spec.ProjectName,
				"region_name":          spec.RegionName,
				"skip_driver_fix":      "false",
				"skip_driver_fix_name": yesLabel,
				"ssh_port":             emptyStringOrNil(spec.SSHPort),
				"ssh_pass":             emptyStringOrNil(spec.SSHPass),
				"linux_hd_username":    spec.LinuxHDUsername,
				"linux_hd_password":    spec.LinuxHDPassword,
				"linux_hd_port":        emptyStringOrNil(spec.LinuxHDPort),
			},
			"storage_type": nil,
		},
		"auto_upload_images": nil,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	}
	return "/hypermotion/v1/cloud_accounts", body, nil
}
