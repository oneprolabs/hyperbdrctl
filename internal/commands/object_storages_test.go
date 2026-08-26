package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testObjectStorageCatalogJSON = `[
  {
    "id": "aliyun",
    "name": "阿里云",
    "name_en": "Alibaba Cloud",
    "regions": [
      {
        "id": "oss-cn-beijing",
        "name": "华北2",
        "name_en": "Beijing",
        "auth_url": "oss-cn-beijing.aliyuncs.com",
        "external_endpoint": "oss-cn-beijing.aliyuncs.com",
        "internal_endpoint": "oss-cn-beijing-internal.aliyuncs.com",
        "protocol": "s3",
        "bucket_lookup": "path"
      }
    ]
  },
  {
    "id": "huaweicloud",
    "name": "华为云",
    "name_en": "Huawei Cloud",
    "regions": [
      {
        "id": "cn-north-4",
        "name": "北京四",
        "name_en": "Beijing 4",
        "auth_url": "obs.cn-north-4.myhuaweicloud.com",
        "external_endpoint": "obs.cn-north-4.myhuaweicloud.com",
        "internal_endpoint": "obs.internal.cn-north-4.myhuaweicloud.com",
        "protocol": "obs",
        "bucket_lookup": "dns"
      }
    ]
  },
  {
    "id": "ens",
    "name": "天翼云 ENS",
    "name_en": "ENS",
    "regions": [
      {
        "id": "ens-region-1",
        "name": "ENS 区域",
        "name_en": "ENS Region",
        "auth_url": "obs.ens.example.com",
        "external_endpoint": "obs.ens.example.com",
        "internal_endpoint": "obs.internal.ens.example.com",
        "protocol": "eos"
      }
    ]
  }
]`

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
		"oss", "buckets",
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

func TestObjectStoragesBucketsWithProviderUsesCatalogDefaults(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/static/json/s3.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
		case "/api/v2/objectStorageBuckets":
			gotPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"buckets": []map[string]interface{}{
						{"name": "bucket-1", "location": "oss-cn-beijing"},
					},
				},
			})
		default:
			t.Fatalf("path=%q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"oss", "buckets",
		"--provider", "aliyun",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/objectStorageBuckets" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["auth_url"] != "oss-cn-beijing.aliyuncs.com" || gotBody["region_id"] != "oss-cn-beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
	if gotBody["protocol"] != "s3" || gotBody["bucket_lookup"] != "path" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestObjectStoragesBucketsWithProviderRequiresRegion(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"oss", "buckets",
		"--provider", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || err.Error() != "region-id is required" {
		t.Fatalf("err = %v", err)
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
		"oss", "create",
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
	if gotBody["display_name"] != "aliyun-beijing" || gotBody["cloud_type"] != "custom" || gotBody["type"] != "objectstorage" {
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
	if metadata["cloud_type_select"] != "custom" || metadata["app_id"] != "" {
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
		"oss", "create",
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
	if gotBody["cloud_type"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
	config := gotBody["config"].(map[string]interface{})
	if config["need_creation"] != true || config["bucket_name"] != "data-sync-storage-20260527163153-iioq38" {
		t.Fatalf("config = %+v", config)
	}
	metadata := gotBody["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
	if config["use_tls"] != false {
		t.Fatalf("custom mode should default use_tls=false, body = %+v", gotBody)
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
		"oss", "create",
		"--auth-url", "oss-cn-beijing.aliyuncs.com",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["display_name"] != "Custom" || gotBody["cloud_type"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
	metadata := gotBody["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestObjectStoragesCreateAutoGeneratesZhCNCustomDisplayNameWithoutFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"storage": map[string]interface{}{"uuid": "storage-1"}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--lang", "zh_cn",
		"oss", "create",
		"--auth-url", "192.168.8.171:9000",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["display_name"] != "其它平台" || gotBody["cloud_type"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestObjectStoragesCreateRejectsRemovedCloudTypeFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"oss", "create",
		"--cloud-type", "huawei",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined: -cloud-type") {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesCreateRejectsRemovedCloudTypeSelectFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"oss", "create",
		"--cloud-type-select", "aliyun,oss-cn-beijing",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined: -cloud-type-select") {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesCreateRejectsFileFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"oss", "create",
		"--file", "./tmp/object-storage-create.json",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined: -file") {
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
	err := Execute(withHost(t, srv.URL, "oss", "delete", "--id", "storage-1"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "oss", "delete", "--id", "storage-1", "--force"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "oss", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "type=objectstorage") {
		t.Fatalf("query = %q", gotQuery)
	}
}

func TestObjectStoragesListRejectsRemovedTypeFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"oss", "list",
		"--type", "objectstorage",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined: -type") {
		t.Fatalf("err = %v", err)
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
	err := Execute(withHost(t, srv.URL, "oss", "list"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "--output", "json", "oss", "list"), &out, &errOut)
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
	err := Execute(withHost(t, "https://example.invalid", "oss", "detail"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesCatalogListsProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "oss", "catalog"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"Name", "Alibaba Cloud", "Huawei Cloud", "Region Count", "aliyun", "huaweicloud"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"Name (EN)", "阿里云", "华为云"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("output should not contain %q: %q", unwanted, text)
		}
	}
}

func TestObjectStoragesCatalogListsRegionsForProvider(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "oss", "catalog", "--provider", "aliyun"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"oss-cn-beijing", "Beijing", "oss-cn-beijing.aliyuncs.com", "s3", "path"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"英文名称", "Name (EN)", "华北2"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("output should not contain %q: %q", unwanted, text)
		}
	}
}

func TestObjectStoragesCatalogUsesLocalizedNameForZhCN(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--lang", "zh_cn", "oss", "catalog", "--provider", "aliyun"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"名称", "华北2", "oss-cn-beijing.aliyuncs.com"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"英文名称", "Name (EN)", "Beijing"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("output should not contain %q: %q", unwanted, text)
		}
	}
}

func TestObjectStoragesCatalogJSONOutputPreservesRawProvider(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "oss", "catalog", "--provider", "ens"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "ens" {
		t.Fatalf("provider = %+v", got)
	}
	regions, ok := got["regions"].([]interface{})
	if !ok || len(regions) != 1 {
		t.Fatalf("regions = %+v", got["regions"])
	}
	region := regions[0].(map[string]interface{})
	if region["protocol"] != "eos" {
		t.Fatalf("region = %+v", region)
	}
	if _, exists := region["bucket_lookup"]; exists {
		t.Fatalf("raw region should preserve missing bucket_lookup: %+v", region)
	}
}

func TestObjectStoragesCatalogJSONOutputListsFullProviderArray(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "oss", "catalog"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0]["id"] != "aliyun" || got[1]["id"] != "huaweicloud" || got[2]["id"] != "ens" {
		t.Fatalf("providers = %+v", got)
	}
}

func TestObjectStoragesCatalogProviderNotFound(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "oss", "catalog", "--provider", "missing"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "provider") || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesCatalogRejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"bad":`))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "oss", "catalog"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "JSON") && !strings.Contains(err.Error(), "unexpected end") {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesCreateWithProviderUsesCatalogDefaults(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/static/json/s3.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
		case "/api/v2/createStorage":
			gotPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"storage": map[string]interface{}{"uuid": "storage-1"}},
			})
		default:
			t.Fatalf("path=%q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"oss", "create",
		"--provider", "aliyun",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/createStorage" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["cloud_type"] != "aliyun" || gotBody["display_name"] != "Alibaba Cloud-Beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
	config := gotBody["config"].(map[string]interface{})
	if config["auth_url"] != "oss-cn-beijing.aliyuncs.com" || config["public_endpoint"] != "oss-cn-beijing.aliyuncs.com" || config["internal_endpoint"] != "oss-cn-beijing-internal.aliyuncs.com" {
		t.Fatalf("config = %+v", config)
	}
	if config["protocol"] != "s3" || config["bucket_lookup"] != "path" {
		t.Fatalf("config = %+v", config)
	}
	metadata := gotBody["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "aliyun,oss-cn-beijing" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestObjectStoragesCreateWithProviderPreviewAllowsOverrides(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"oss", "create",
		"--provider", "huaweicloud",
		"--region-id", "cn-north-4",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
		"--internal-endpoint", "override.internal.example.com",
		"--bucket-lookup", "path",
		"--protocol", "obs",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &gotBody); err != nil {
		t.Fatal(err)
	}
	if gotBody["cloud_type"] != "huaweicloud" {
		t.Fatalf("body = %+v", gotBody)
	}
	if gotBody["display_name"] != "Huawei Cloud-Beijing 4" {
		t.Fatalf("body = %+v", gotBody)
	}
	config := gotBody["config"].(map[string]interface{})
	if config["auth_url"] != "obs.cn-north-4.myhuaweicloud.com" || config["public_endpoint"] != "obs.cn-north-4.myhuaweicloud.com" {
		t.Fatalf("config = %+v", config)
	}
	if config["internal_endpoint"] != "override.internal.example.com" || config["protocol"] != "obs" || config["bucket_lookup"] != "path" {
		t.Fatalf("config = %+v", config)
	}
	metadata := gotBody["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "huaweicloud,cn-north-4" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestObjectStoragesCreateWithoutProviderAllowsEmptyRegion(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"storage": map[string]interface{}{"uuid": "storage-1"}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"oss", "create",
		"--auth-url", "192.168.8.171:9000",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	config := gotBody["config"].(map[string]interface{})
	if config["region_id"] != "" || gotBody["cloud_type"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestObjectStoragesCreateWithProviderCustomMatchesDirectMode(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"storage": map[string]interface{}{"uuid": "storage-1"}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"oss", "create",
		"--provider", "custom",
		"--auth-url", "192.168.8.171:9000",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["cloud_type"] != "custom" || gotBody["display_name"] != "Custom" {
		t.Fatalf("body = %+v", gotBody)
	}
	metadata := gotBody["metadata"].(map[string]interface{})
	if metadata["cloud_type_select"] != "custom" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestObjectStoragesCreateWithProviderRequiresRegion(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"oss", "create",
		"--provider", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err == nil || err.Error() != "region-id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestObjectStoragesCreateWithProviderRejectsUnknownProviderOrRegion(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/json/s3.json" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testObjectStorageCatalogJSON))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"oss", "create",
		"--provider", "missing",
		"--region-id", "oss-cn-beijing",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "provider") || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("provider err = %v", err)
	}

	out.Reset()
	errOut.Reset()
	err = Execute(withHost(t, srv.URL,
		"oss", "create",
		"--provider", "aliyun",
		"--region-id", "missing-region",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--bucket-name", "bucket-1",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "missing-region") {
		t.Fatalf("region err = %v", err)
	}
}
