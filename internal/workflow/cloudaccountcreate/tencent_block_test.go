package cloudaccountcreate

import (
	"reflect"
	"strings"
	"testing"
)

func tencentBlockSpec() Spec {
	return Spec{CloudType: "tencent_bs", StorageType: "block", AccessKeyID: "test-ak", AccessKeySecret: "test-sk", AppID: "1000000001", AuthRegionID: "ap-beijing"}
}

func TestTencentBlockRequestMatchesCapturedForm(t *testing.T) {
	spec := tencentBlockSpec()
	spec.MetadataOverrides = map[string]interface{}{"region_type_list_name": "华北地区(北京)"}
	path, body, err := BuildRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	want := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type": "tencent_bs", "cloud_auth_type": "aksk", "storage_type": nil,
			"metadata": map[string]interface{}{
				"access_key_id": "test-ak", "access_key_secret": "test-sk", "app_id": "1000000001",
				"skip_driver_fix": "false", "skip_driver_fix_name": "是",
				"region_type": "1", "region_type_name": "自动获取",
				"region_type_list": "ap-beijing", "region_type_list_name": "华北地区(北京)",
				"region_type_input": "", "region_text": "", "account_name": nil, "auth_region_id": "ap-beijing",
			},
		},
		"auto_upload_images": nil, "only_verify": nil,
	}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %#v, want %#v", body, want)
	}
	for _, storageType := range []string{"", "block", "HyperGate"} {
		if !HasRegisteredAdapter("tencent_bs", storageType) {
			t.Fatalf("adapter missing for %q", storageType)
		}
	}
	if HasRegisteredAdapter("tencent_bs", "objectstorage") {
		t.Fatal("block adapter registered for object storage")
	}
}

func TestTencentBlockValidatesRequiredFields(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Spec)
		want   string
	}{
		{"missing AK", func(s *Spec) { s.AccessKeyID = "" }, "access-key-id is required"},
		{"blank AK", func(s *Spec) { s.AccessKeyID = "  " }, "access-key-id is required"},
		{"missing SK", func(s *Spec) { s.AccessKeySecret = "" }, "access-key-secret is required"},
		{"blank SK", func(s *Spec) { s.AccessKeySecret = "  " }, "access-key-secret is required"},
		{"missing App ID", func(s *Spec) { s.AppID = "" }, "app-id is required"},
		{"missing region", func(s *Spec) { s.AuthRegionID = "" }, "auth-region-id is required"},
		{"override cannot bypass validation", func(s *Spec) { s.MetadataOverrides = map[string]interface{}{"access_key_secret": ""} }, "access-key-secret is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := tencentBlockSpec()
			tc.change(&spec)
			_, _, err := BuildRequest(spec)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
	for _, value := range []string{"0", "-1", "+1", "not-an-id", "1.5", "1e9", "18446744073709551616"} {
		t.Run(value, func(t *testing.T) {
			spec := tencentBlockSpec()
			spec.AppID = value
			_, _, err := BuildRequest(spec)
			if err == nil || !strings.Contains(err.Error(), "positive decimal integer") {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestTencentBlockMetadataFileAndNumericAppID(t *testing.T) {
	for _, appID := range []interface{}{"1000000001", 1000000001, float64(1000000001)} {
		spec := Spec{CloudType: "tencent_bs", MetadataOverrides: map[string]interface{}{
			"access_key_id": "test-ak", "access_key_secret": "test-sk", "app_id": appID,
			"auth_region_id": "ap-beijing", "account_name": "test-account",
		}}
		_, body, err := BuildRequest(NormalizeSpec(spec))
		if err != nil {
			t.Fatal(err)
		}
		account := body["cloud_account"].(map[string]interface{})
		metadata := account["metadata"].(map[string]interface{})
		if metadata["app_id"] != "1000000001" || metadata["account_name"] != "test-account" || metadata["region_type_list"] != "ap-beijing" {
			t.Fatalf("metadata = %#v", metadata)
		}
	}
}

func TestTencentBlockTrimsCredentialsAndRegion(t *testing.T) {
	spec := tencentBlockSpec()
	spec.AccessKeyID = " test-ak "
	spec.AccessKeySecret = " test-sk "
	spec.AppID = " 1000000001 "
	spec.AuthRegionID = " ap-beijing "
	_, body, err := BuildRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	for key, want := range map[string]string{"access_key_id": "test-ak", "access_key_secret": "test-sk", "app_id": "1000000001", "auth_region_id": "ap-beijing", "region_type_list": "ap-beijing", "region_type_list_name": "ap-beijing"} {
		if metadata[key] != want {
			t.Fatalf("%s = %v, want %s", key, metadata[key], want)
		}
	}
}
