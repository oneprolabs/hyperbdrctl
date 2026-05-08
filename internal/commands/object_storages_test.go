package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObjectStoragesBucketsBuildsValidatedAliyunRequest(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"buckets": []map[string]interface{}{
					{"name": "data-sync-storage-20260515101326-uhxtv2", "location": "oss-cn-beijing", "created_at": "2026-05-15T10:13:26Z"},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "oss", "buckets",
		"--auth-url", "oss-cn-beijing.aliyuncs.com",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--protocol", "s3",
		"--bucket-lookup", "virtual-hosted-style",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/objectStorageBuckets" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["auth_type"] != "aksk" || gotBody["auth_url"] != "oss-cn-beijing.aliyuncs.com" || gotBody["region_id"] != "oss-cn-beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
	if gotBody["bucket_lookup"] != "dns" || gotBody["protocol"] != "s3" || gotBody["use_tls"] != true {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "data-sync-storage-20260515101326-uhxtv2") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestObjectStoragesCreateBuildsValidatedAliyunNewBucketPayload(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storage": map[string]interface{}{
					"uuid":   "7f5119ad-6beb-4370-99d8-956fad8a3cc2",
					"status": "creating",
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"target", "oss", "create",
		"--display-name", "aliyun-beijing",
		"--cloud-type", "aliyun",
		"--auth-url", "oss-cn-beijing.aliyuncs.com",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--protocol", "s3",
		"--bucket-lookup", "dns",
		"--bucket-mode", "new",
		"--bucket-name", "data-sync-storage-20260527163153-iioq38",
		"--public-endpoint", "oss-cn-beijing.aliyuncs.com",
		"--internal-endpoint", "oss-cn-beijing-internal.aliyuncs.com",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/createStorage" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["display_name"] != "aliyun-beijing" || gotBody["cloud_type"] != "aliyun" || gotBody["type"] != "objectstorage" {
		t.Fatalf("body = %+v", gotBody)
	}
	config := gotBody["config"].(map[string]interface{})
	if config["need_creation"] != true || config["bucket_name"] != "data-sync-storage-20260527163153-iioq38" {
		t.Fatalf("body = %+v", gotBody)
	}
	if config["auth_type"] != "aksk" || config["bucket_lookup"] != "dns" || config["public_endpoint"] != "oss-cn-beijing.aliyuncs.com" || config["internal_endpoint"] != "oss-cn-beijing-internal.aliyuncs.com" {
		t.Fatalf("body = %+v", gotBody)
	}
	metadata := gotBody["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "aliyun,oss-cn-beijing" || metadata["app_id"] != "" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), `"storage"`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestObjectStoragesCreatePreviewRequestPrintsRequestBody(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"--output", "json",
		"target", "oss", "create",
		"--display-name", "aliyun-beijing",
		"--cloud-type", "aliyun",
		"--auth-url", "oss-cn-beijing.aliyuncs.com",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--protocol", "s3",
		"--bucket-lookup", "dns",
		"--bucket-mode", "new",
		"--bucket-name", "data-sync-storage-20260527163153-iioq38",
		"--public-endpoint", "oss-cn-beijing.aliyuncs.com",
		"--internal-endpoint", "oss-cn-beijing-internal.aliyuncs.com",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &gotBody); err != nil {
		t.Fatal(err)
	}
	if gotBody["display_name"] != "aliyun-beijing" || gotBody["type"] != "objectstorage" {
		t.Fatalf("body = %+v", gotBody)
	}
	config := gotBody["config"].(map[string]interface{})
	if config["need_creation"] != true || config["bucket_name"] != "data-sync-storage-20260527163153-iioq38" {
		t.Fatalf("config = %+v", config)
	}
}

func TestObjectStoragesCreateRequiresDisplayNameWithoutFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "oss", "create",
		"--auth-url", "oss-cn-beijing.aliyuncs.com",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err == nil || err.Error() != "display-name is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesListDefaultsToObjectStorage(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"storages": []map[string]interface{}{}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "target", "oss", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "type=objectstorage") {
		t.Fatalf("query = %q", gotQuery)
	}
}

func TestObjectStoragesDetailRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "target", "oss", "detail"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}
