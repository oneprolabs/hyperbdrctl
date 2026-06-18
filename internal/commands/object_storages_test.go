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

func TestObjectStoragesCreateAutoGeneratesDisplayNameWithoutFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storage": map[string]interface{}{
					"uuid": "storage-1",
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "oss", "create",
		"--auth-url", "oss-cn-beijing.aliyuncs.com",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["display_name"] != "aliyun-oss-cn-beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestObjectStoragesCreateRejectsRemovedCloudTypeFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "oss", "create",
		"--cloud-type", "huawei",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined: -cloud-type") {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesDeleteRequiresForceWhenAssociatedHostsExist(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var associatedCalls, deleteCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getStorageAssociatedResources":
			associatedCalls++
			if got := r.URL.Query().Get("storage_id"); got != "storage-1" {
				t.Fatalf("storage_id = %q", got)
			}
			if got := r.URL.Query().Get("with_statistics"); got != "false" {
				t.Fatalf("with_statistics = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"resources": []map[string]interface{}{
						{
							"host_id":   "host-1",
							"host_name": "DESKTOP-QD7LPO1-856c",
						},
					},
				},
			})
		case "/api/v2/deleteStorage":
			deleteCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"deleted": true}})
		default:
			t.Fatalf("path=%q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "target", "oss", "delete", "--id", "storage-1"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "--force") || !strings.Contains(err.Error(), "DESKTOP-QD7LPO1-856c") {
		t.Fatalf("err = %v", err)
	}
	if associatedCalls != 1 {
		t.Fatalf("associatedCalls = %d", associatedCalls)
	}
	if deleteCalls != 0 {
		t.Fatalf("deleteCalls = %d", deleteCalls)
	}
}

func TestObjectStoragesDeleteForcePostsDeleteRequest(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var calls []string
	var gotDeleteBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/deleteStorage":
			if err := json.NewDecoder(r.Body).Decode(&gotDeleteBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"deleted": true}})
		default:
			t.Fatalf("path=%q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "target", "oss", "delete", "--id", "storage-1", "--force"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "/api/v2/deleteStorage" {
		t.Fatalf("calls = %+v", calls)
	}
	if gotDeleteBody["storage_id"] != "storage-1" || gotDeleteBody["force"] != true {
		t.Fatalf("body = %+v", gotDeleteBody)
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

func TestObjectStoragesListRendersFlattenedTableColumns(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storages": []map[string]interface{}{
					{
						"id":           "storage-1",
						"uuid":         "uuid-1",
						"display_name": "华为对象存储-北京一",
						"status":       "available",
						"config": map[string]interface{}{
							"bucket_name": "data-sync-storage-cli-test",
							"region_id":   "cn-north-1",
							"auth_url":    "obs.cn-north-1.myhuaweicloud.com",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "target", "oss", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"ID",
		"Name",
		"Bucket Name",
		"Region",
		"Auth URL",
		"Status",
		"storage-1",
		"华为对象存储-北京一",
		"data-sync-storage-cli-test",
		"cn-north-1",
		"obs.cn-north-1.myhuaweicloud.com",
		"available",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	for _, unwanted := range []string{"UUID", "Storage Type"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("output = %q, should not contain %q", text, unwanted)
		}
	}
}

func TestObjectStoragesListJSONKeepsRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storages": []map[string]interface{}{
					{
						"id":           "storage-1",
						"display_name": "aliyun-beijing",
						"config": map[string]interface{}{
							"bucket_name": "bucket-1",
							"auth_url":    "oss-cn-beijing.aliyuncs.com",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "target", "oss", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{`"storages"`, `"config"`, `"bucket_name"`, `"auth_url"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
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
