package cloudaccountcreate

import "fmt"

func buildAliyunBlock(spec Spec) (string, map[string]interface{}, error) {
	if spec.AccessKeyID == "" {
		return "", nil, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return "", nil, fmt.Errorf("access-key-secret is required")
	}
	if spec.RegionID == "" {
		return "", nil, fmt.Errorf("region-id is required")
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "aliyun_bs",
			"cloud_auth_type": "aksk",
			"metadata": map[string]interface{}{
				"access_key_id":         spec.AccessKeyID,
				"access_key_secret":     spec.AccessKeySecret,
				"skip_driver_fix":       "false",
				"skip_driver_fix_name":  yesLabel,
				"region_type":           "1",
				"region_type_name":      autoRegionLabel,
				"region_type_list":      spec.RegionID,
				"region_type_list_name": valueOrNil(spec.RegionName),
				"region_type_input":     "",
				"region_text":           "",
				"account_name":          valueOrNil(spec.AccountName),
				"auth_region_id":        firstNonEmptyString(spec.AuthRegionID, spec.RegionID),
			},
			"storage_type": nil,
		},
		"auto_upload_images": nil,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	}
	return "/hypermotion/v1/cloud_accounts", finalizeCreateBody(spec, body), nil
}
