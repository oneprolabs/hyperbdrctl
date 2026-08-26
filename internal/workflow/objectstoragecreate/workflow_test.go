package objectstoragecreate

import (
	"reflect"
	"testing"
)

func TestBuildRequestCustomProfile(t *testing.T) {
	prepared, err := BuildRequest(CustomProfile("Custom"), Spec{
		AuthURL:         "192.168.8.171:9000",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		BucketName:      "bucket-1",
		UseTLS:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Path != "/api/v2/createStorage" {
		t.Fatalf("path = %q", prepared.Path)
	}
	if prepared.Body["cloud_type"] != "custom" || prepared.Body["display_name"] != "Custom" {
		t.Fatalf("body = %+v", prepared.Body)
	}
	metadata := prepared.Body["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "custom" {
		t.Fatalf("metadata = %+v", metadata)
	}
	config := prepared.Body["config"].(map[string]interface{})
	if config["need_creation"] != false || config["protocol"] != "s3" || config["bucket_lookup"] != "dns" {
		t.Fatalf("config = %+v", config)
	}
}

func TestBuildRequestCatalogProfileAppliesDefaults(t *testing.T) {
	profile := Profile{
		Mode:               ModeCatalog,
		ProviderID:         "aliyun",
		RegionID:           "oss-cn-beijing",
		AuthURL:            "oss-cn-beijing.aliyuncs.com",
		PublicEndpoint:     "oss-cn-beijing.aliyuncs.com",
		InternalEndpoint:   "oss-cn-beijing-internal.aliyuncs.com",
		Protocol:           "s3",
		BucketLookup:       "dns",
		DefaultDisplayName: "Alibaba Cloud-China (Beijing)",
	}
	prepared, err := BuildRequest(profile, Spec{
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		BucketMode:      "new",
		BucketName:      "bucket-1",
		UseTLS:          true,
		AppID:           "app-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Body["cloud_type"] != "aliyun" || prepared.Body["display_name"] != profile.DefaultDisplayName {
		t.Fatalf("body = %+v", prepared.Body)
	}
	metadata := prepared.Body["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "aliyun,oss-cn-beijing" || metadata["app_id"] != "app-1" {
		t.Fatalf("metadata = %+v", metadata)
	}
	config := prepared.Body["config"].(map[string]interface{})
	if config["need_creation"] != true || config["auth_url"] != profile.AuthURL || config["internal_endpoint"] != profile.InternalEndpoint {
		t.Fatalf("config = %+v", config)
	}
}

func TestBuildRequestExplicitValuesOverrideCatalog(t *testing.T) {
	profile := Profile{
		Mode:             ModeCatalog,
		ProviderID:       "huaweicloud",
		RegionID:         "cn-north-4",
		AuthURL:          "catalog.example.com",
		PublicEndpoint:   "catalog-public.example.com",
		InternalEndpoint: "catalog-internal.example.com",
		Protocol:         "s3",
		BucketLookup:     "dns",
	}
	prepared, err := BuildRequest(profile, Spec{
		AuthURL:          "override.example.com",
		AccessKeyID:      "ak",
		AccessKeySecret:  "sk",
		BucketName:       "bucket-1",
		InternalEndpoint: "override-internal.example.com",
		Protocol:         "obs",
		BucketLookup:     "path",
		ExplicitFields: map[string]bool{
			"auth-url":          true,
			"internal-endpoint": true,
			"protocol":          true,
			"bucket-lookup":     true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	config := prepared.Body["config"].(map[string]interface{})
	want := map[string]interface{}{
		"auth_url":          "override.example.com",
		"internal_endpoint": "override-internal.example.com",
		"protocol":          "obs",
		"bucket_lookup":     "path",
	}
	for key, value := range want {
		if !reflect.DeepEqual(config[key], value) {
			t.Fatalf("config[%q] = %#v, want %#v; config=%+v", key, config[key], value, config)
		}
	}
}

func TestBuildRequestRejectsUnsupportedProfile(t *testing.T) {
	_, err := BuildRequest(Profile{Mode: "unknown"}, Spec{})
	if err == nil || err.Error() != `unsupported object storage create profile mode "unknown"` {
		t.Fatalf("err = %v", err)
	}
}
