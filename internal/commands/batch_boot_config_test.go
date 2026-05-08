package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBatchBootConfigGetWithIDs(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPaths []string
	var gotQuery string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHosts":
			gotQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"hosts": []map[string]interface{}{
						{"id": "host-2", "boot_config_id": "cfg-2"},
						{"id": "host-1", "boot_config_id": "cfg-1"},
					},
				},
			})
		case "/api/v2/batchGetBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{"id": "cfg-1", "migration_id": "host-1", "storage_id": "storage-1", "storage_name": "storage-name-1", "pool_id": "pool-1", "pool_name": "pool-name-1"},
						{"id": "cfg-2", "migration_id": "host-2", "storage_id": "storage-2", "storage_name": "storage-name-2", "pool_id": "pool-2", "pool_name": "pool-name-2"},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "batch-boot-config", "get", "--ids", "host-1,host-2"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHosts" || gotPaths[1] != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	if !strings.Contains(gotQuery, "ids=host-1%2Chost-2") {
		t.Fatalf("query = %q", gotQuery)
	}
	items := gotBody["batch_get"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("body = %+v", gotBody)
	}
	if items[0].(map[string]interface{})["id"] != "cfg-1" || items[1].(map[string]interface{})["id"] != "cfg-2" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "storage-name-1") || !strings.Contains(out.String(), "storage-name-2") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestBatchBootConfigGetFallsBackToBootConfigObject(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getHosts":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"hosts": []map[string]interface{}{
						{"id": "host-1", "boot_config": map[string]interface{}{"id": "cfg-1"}},
					},
				},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"boot_configs": []map[string]interface{}{{"id": "cfg-1", "migration_id": "host-1"}}},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "batch-boot-config", "get", "--ids", "host-1"), &out, &errOut); err != nil {
		t.Fatal(err)
	}
}

func TestBatchBootConfigGetRequiresIDs(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "batch-boot-config", "get"), &out, &errOut)
	if err == nil || err.Error() != "ids is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigGetRejectsIDFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "batch-boot-config", "get", "--id", "host-1"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigGetFailsWithoutExistingBootConfig(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"hosts": []map[string]interface{}{{"id": "host-1"}}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "batch-boot-config", "get", "--ids", "host-1"), &out, &errOut)
	if err == nil || err.Error() != "host host-1 has no existing boot config" {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigCreateWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "batch-boot-config-create.json")
	raw := `{"batch_create":[{"migration_id":"host-1","metadata":{"storage_id":"storage-1"}}]}`
	if err := os.WriteFile(bodyPath, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "batch-boot-config", "create", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["batch_create"] == nil {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestBatchBootConfigUpdateWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "batch-boot-config-update.json")
	raw := `{"batch_update":[{"id":"cfg-1","migration_id":"host-1","metadata":{"storage_id":"storage-1"}}]}`
	if err := os.WriteFile(bodyPath, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "batch-boot-config", "update", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["batch_update"] == nil {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestBatchBootConfigCreateRequiresFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "batch-boot-config", "create"), &out, &errOut)
	if err == nil || err.Error() != "file is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigUpdateRequiresFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "batch-boot-config", "update"), &out, &errOut)
	if err == nil || err.Error() != "file is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigCreateRejectsIDsFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "batch-boot-config", "create", "--ids", "host-1"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigUpdateRejectsIDFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "batch-boot-config", "update", "--id", "host-1"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchBootConfigGetJSONOutputPreservesRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getHosts":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"hosts": []map[string]interface{}{
						{"id": "host-1", "boot_config_id": "cfg-1"},
					},
				},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{"id": "cfg-1", "migration_id": "host-1"},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "batch-boot-config", "get", "--ids", "host-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"boot_configs"`) || !strings.Contains(out.String(), `"migration_id"`) {
		t.Fatalf("output = %q", out.String())
	}
}
