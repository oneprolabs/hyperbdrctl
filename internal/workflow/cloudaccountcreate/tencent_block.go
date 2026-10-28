package cloudaccountcreate

import (
	"fmt"
	"strconv"
	"strings"
)

// buildTencentBlock follows the Tencent Cloud (recommended) account form.
// The account is created independently of the subsequent HyperGate wizard.
func buildTencentBlock(spec Spec) (string, map[string]interface{}, error) {
	spec.AuthRegionID = strings.TrimSpace(spec.AuthRegionID)
	regionName := firstNonEmptyString(spec.RegionName, overrideStringValue(spec.MetadataOverrides, "region_type_list_name"), spec.AuthRegionID)
	body := finalizeCreateBody(spec, map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "tencent_bs",
			"cloud_auth_type": "aksk",
			"storage_type":    nil,
			"metadata": map[string]interface{}{
				"access_key_id":         spec.AccessKeyID,
				"access_key_secret":     spec.AccessKeySecret,
				"app_id":                spec.AppID,
				"skip_driver_fix":       "false",
				"skip_driver_fix_name":  yesLabel,
				"region_type":           "1",
				"region_type_name":      autoRegionLabel,
				"region_type_list":      spec.AuthRegionID,
				"region_type_list_name": regionName,
				"region_type_input":     "",
				"region_text":           "",
				"account_name":          valueOrNil(spec.AccountName),
				"auth_region_id":        spec.AuthRegionID,
			},
		},
		"auto_upload_images": nil,
		"only_verify":        boolOrNil(spec.OnlyVerify),
	})
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	for _, field := range []struct{ key, flag string }{
		{"access_key_id", "access-key-id"},
		{"access_key_secret", "access-key-secret"},
		{"app_id", "app-id"},
		{"auth_region_id", "auth-region-id"},
	} {
		value, ok := metadata[field.key].(string)
		// --set infers JSON numbers; App ID is still a string on the wire.
		if field.key == "app_id" {
			value = overrideStringValue(metadata, field.key)
			ok = value != ""
		}
		if !ok || strings.TrimSpace(value) == "" {
			return "", nil, fmt.Errorf("%s is required", field.flag)
		}
		metadata[field.key] = strings.TrimSpace(value)
	}
	appID := metadata["app_id"].(string)
	if n, err := strconv.ParseUint(appID, 10, 64); err != nil || n == 0 || strings.HasPrefix(appID, "+") {
		return "", nil, fmt.Errorf("app-id must be a positive decimal integer")
	}
	return "/hypermotion/v1/cloud_accounts", body, nil
}
