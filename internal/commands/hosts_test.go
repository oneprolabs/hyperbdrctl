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

func TestHostsListPassesUnknownFlagsAsQuery(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"hosts": []map[string]interface{}{
					{"id": "host-1", "name": "test-host"},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"host", "list",
		"--page", "1",
		"--page-size", "10",
		"--step", "3",
		"--custom-filter=value",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "step=3") {
		t.Fatalf("query missing step: %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "custom_filter=value") {
		t.Fatalf("query missing custom_filter: %q", gotQuery)
	}
	if !strings.Contains(out.String(), "host-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHostsSyncWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"host", "sync",
		"--id", "host-1",
		"--mode", "full",
		"--transfer-speed", "100",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/batchSync" {
		t.Fatalf("path = %q", gotPath)
	}
	items := gotBody["batch_sync"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["migration_id"] != "host-1" || item["sync_mode"] != "full" || item["transfer_speed"].(float64) != 100 || item["do_snapshot"] != true {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsSyncRejectsDoSnapshotFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "sync", "--id", "host-1", "--do-snapshot"), &out, &errOut)
	if err == nil {
		t.Fatal("expected unknown flag")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsSyncModeAndTransferSpeedAreOptional(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "sync", "--id", "host-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	items := gotBody["batch_sync"].([]interface{})
	item := items[0].(map[string]interface{})
	if _, ok := item["sync_mode"]; ok {
		t.Fatalf("sync_mode should be omitted: %+v", item)
	}
	if _, ok := item["transfer_speed"]; ok {
		t.Fatalf("transfer_speed should be omitted: %+v", item)
	}
	if item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	if item["do_snapshot"] != true {
		t.Fatalf("do_snapshot should default to true: %+v", item)
	}
}

func TestHostsRegisterWithVMID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "200",
			"detail":  "",
			"message": "host added",
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "register", "--vm-id", "vm-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotPath != "/hypermotion/v1/hosts/register" {
		t.Fatalf("method=%q path=%q", gotMethod, gotPath)
	}
	ids := gotBody["ids"].([]interface{})
	if len(ids) != 1 || ids[0] != "vm-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "message") || !strings.Contains(out.String(), "host added") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHostsRegisterWithVMIDs(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "200"})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "register", "--vm-ids", "vm-1,vm-2"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	ids := gotBody["ids"].([]interface{})
	if len(ids) != 2 || ids[0] != "vm-1" || ids[1] != "vm-2" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsRegisterMergesVMIDAndVMIDs(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "200"})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "register", "--vm-id", "vm-1", "--vm-ids", "vm-2,vm-3"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	ids := gotBody["ids"].([]interface{})
	if len(ids) != 3 || ids[0] != "vm-1" || ids[1] != "vm-2" || ids[2] != "vm-3" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsRegisterRequiresVMIDOrVMIDs(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "register"), &out, &errOut)
	if err == nil || err.Error() != "vm-id or vm-ids is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsRegisterJSONOutputPreservesRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "200",
			"detail":  "",
			"message": "添加主机成功",
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "host", "register", "--vm-id", "vm-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"code": "200"`) || !strings.Contains(out.String(), `"message": "添加主机成功"`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHostsRegisterRejectsHostIDFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "register", "--id", "host-1"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}

	err = Execute(withHost(t, "https://example.invalid", "host", "register", "--ids", "host-1,host-2"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"host", "boot",
		"--id", "host-1",
		"--snapshot-id", "snap-1",
		"--boot-instance-purpose", "drill",
		"--cloud-type", "vmware_obs",
		"--custom-number", "3",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/batchBoot" {
		t.Fatalf("path = %q", gotPath)
	}
	items := gotBody["batch_boot"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["migration_id"] != "host-1" ||
		item["snapshot_id"] != "snap-1" ||
		item["boot_instance_purpose"] != "drill" ||
		item["cloud_type"] != "vmware_obs" ||
		item["custom_number"].(float64) != 3 {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsBootWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot.json")
	if err := os.WriteFile(bodyPath, []byte(`{"batch_boot":[{"migration_id":"file-host","snapshot_id":"snap-1","boot_instance_purpose":"drill","cloud_type":"vmware_obs"}]}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	items := gotBody["batch_boot"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["migration_id"] != "file-host" || item["snapshot_id"] != "snap-1" || item["boot_instance_purpose"] != "drill" || item["cloud_type"] != "vmware_obs" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsCleanupValidationHostWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "cleanup-validation-host", "--id", "host-1", "--ids", "host-2,host-3"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/batchDeleteInstances" {
		t.Fatalf("path = %q", gotPath)
	}
	items := gotBody["batch_delete_instances"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("body = %+v", gotBody)
	}
	for i, want := range []string{"host-1", "host-2", "host-3"} {
		item := items[i].(map[string]interface{})
		if item["migration_id"] != want {
			t.Fatalf("body = %+v", gotBody)
		}
	}
}

func TestHostsCleanupValidationHostWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "cleanup-validation-host.json")
	if err := os.WriteFile(bodyPath, []byte(`{"batch_delete_instances":[{"migration_id":"file-host"}]}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "cleanup-validation-host", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	items := gotBody["batch_delete_instances"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["migration_id"] != "file-host" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsDeregisterWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "deregister", "--id", "host-1", "--ids", "host-2,host-3", "--force"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/hypermotion/v1/hosts" {
		t.Fatalf("method=%q path=%q", gotMethod, gotPath)
	}
	ids := gotBody["ids"].([]interface{})
	if len(ids) != 3 || ids[0] != "host-1" || ids[1] != "host-2" || ids[2] != "host-3" {
		t.Fatalf("body = %+v", gotBody)
	}
	clean := gotBody["clean"].(map[string]interface{})
	if clean["force"] != true {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsDeregisterWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "deregister.json")
	if err := os.WriteFile(bodyPath, []byte(`{"ids":["file-host"],"clean":{"force":false}}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "deregister", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	ids := gotBody["ids"].([]interface{})
	if len(ids) != 1 || ids[0] != "file-host" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsWaitSyncUsesHostDetailStatus(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	statuses := []string{"host_register_done", "sync_doing", "sync_snapshot_done"}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/getHostDetail" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("host_id") != "host-1" {
			t.Fatalf("query = %q", r.URL.RawQuery)
		}
		status := statuses[calls]
		calls++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"id":             "host-1",
				"status":         status,
				"display_status": status,
				"task_id":        "task-sync",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "sync",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d", calls)
	}
	if !strings.Contains(out.String(), `"result": "success"`) || !strings.Contains(out.String(), `"status": "sync_snapshot_done"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsWaitBootUsesBootStatus(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	statuses := []string{"not_boot", "boot_doing", "boot_done"}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := statuses[calls]
		calls++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"id":                  "host-1",
				"boot_status":         status,
				"display_boot_status": status,
				"boot_task_id":        "task-boot",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "boot",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d", calls)
	}
	if !strings.Contains(out.String(), `"task_id": "task-boot"`) || !strings.Contains(out.String(), `"status": "boot_done"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsWaitFailureCanIncludeTaskStepError(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"id":                          "host-1",
					"boot_status":                 "boot_failed",
					"display_boot_status":         "failed",
					"boot_task_id":                "task-boot",
					"boot_task_error_description": "host detail error",
				},
			})
		case "/hypermotion/v1/job/steps":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": []map[string]interface{}{
					{"id": "step-1", "message": "step error"},
				},
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "boot",
		"--include-steps",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(out.String(), `"result": "failed"`) || !strings.Contains(out.String(), `"error": "step error"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsWaitDeregisterTreatsNotFoundAsSuccess(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"id":             "host-1",
					"status":         "clean_doing",
					"display_status": "cleaning",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":  "00003005",
			"title": "not found",
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "deregister",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
	if !strings.Contains(out.String(), `"result": "success"`) || !strings.Contains(out.String(), `"status": "not_found"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsWaitBatchContinuesAfterFailure(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		status := "sync_snapshot_done"
		if hostID == "host-2" {
			status = "sync_failed"
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"id":                     hostID,
				"status":                 status,
				"display_status":         status,
				"task_id":                "task-" + hostID,
				"task_error_description": "sync failed",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--ids", "host-1,host-2",
		"--operation", "sync",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(out.String(), `"id": "host-1"`) || !strings.Contains(out.String(), `"id": "host-2"`) {
		t.Fatalf("output = %s", out.String())
	}
	if !strings.Contains(out.String(), `"result": "success"`) || !strings.Contains(out.String(), `"result": "failed"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsBootConfigCreateWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-create.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","storage_name":"storage-name","pool_id":"pool-1","pool_name":"pool-name"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot-config", "create", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/batchBootConfigs" {
		t.Fatalf("path = %q", gotPath)
	}
	items := gotBody["batch_create"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	meta := item["metadata"].(map[string]interface{})
	if meta["storage_id"] != "storage-1" || meta["pool_id"] != "pool-1" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsBootConfigUpdateWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-update.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","storage_name":"storage-name","pool_id":"pool-1","pool_name":"pool-name"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_config_id": "cfg-1",
					"boot_config":    map[string]interface{}{},
				},
			})
		case "/api/v2/batchUpdateBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot-config", "update", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHostDetail" || gotPaths[1] != "/api/v2/batchUpdateBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	items := gotBody["batch_update"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["id"] != "cfg-1" || item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	meta := item["metadata"].(map[string]interface{})
	if meta["storage_id"] != "storage-1" || meta["pool_id"] != "pool-1" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsBootConfigApplyCreatesWhenBootConfigMissing(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-apply-create.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","pool_id":"pool-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"id":   "host-1",
					"name": "host-name",
				},
			})
		case "/api/v2/batchBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"status": "ok"}})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "boot-config-cli", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHostDetail" || gotPaths[1] != "/api/v2/batchBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	items := gotBody["batch_create"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "create") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHostsBootConfigApplyUpdatesWhenBootConfigExists(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-apply-update.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","pool_id":"pool-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_config_id": "cfg-1",
				},
			})
		case "/api/v2/batchUpdateBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"status": "ok"}})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "boot-config-cli", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHostDetail" || gotPaths[1] != "/api/v2/batchUpdateBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	items := gotBody["batch_update"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["id"] != "cfg-1" || item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "update") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHostsBootConfigApplyFallsBackToNestedBootConfigID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-apply-fallback.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_config": map[string]interface{}{"id": "cfg-1"},
				},
			})
		case "/api/v2/batchUpdateBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"status": "ok"}})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "boot-config-cli", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	items := gotBody["batch_update"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["id"] != "cfg-1" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsBootConfigGetWithID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPaths []string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_config_id": "cfg-1",
					"boot_config":    map[string]interface{}{},
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
						{
							"id":           "cfg-1",
							"migration_id": "host-1",
							"metadata": map[string]interface{}{
								"storage_name":       "storage-name",
								"pool_name":          "pool-name",
								"network_id":         "net-1",
								"repair_host_mapper": map[string]interface{}{"enable_repair_fs": "1"},
								"nics":               []map[string]interface{}{{"network_id": "net-1"}},
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot-config", "get", "--id", "host-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHostDetail" || gotPaths[1] != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	items := gotBody["batch_get"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["id"] != "cfg-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "host_id") || !strings.Contains(out.String(), "host-1") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "boot_config_id") || !strings.Contains(out.String(), "cfg-1") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "migration_id") || !strings.Contains(out.String(), "metadata.storage_name") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "storage-name") || !strings.Contains(out.String(), "pool-name") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), `{"enable_repair_fs":"1"}`) {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), `[{"network_id":"net-1"}]`) {
		t.Fatalf("output = %q", out.String())
	}
	hostIdx := strings.Index(out.String(), "host_id")
	bootIdx := strings.Index(out.String(), "boot_config_id")
	metaIdx := strings.Index(out.String(), "metadata.network_id")
	if hostIdx < 0 || bootIdx < 0 || metaIdx < 0 || !(hostIdx < bootIdx && bootIdx < metaIdx) {
		t.Fatalf("output order = %q", out.String())
	}
}

func TestHostsBootConfigGetFailsWithoutBootConfigID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot-config", "get", "--id", "host-1"), &out, &errOut)
	if err == nil || err.Error() != "host host-1 has no existing boot config" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigGetJSONOutputPreservesRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_config_id": "cfg-1",
					"boot_config":    map[string]interface{}{},
				},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{
							"id":           "cfg-1",
							"migration_id": "host-1",
							"storage_id":   "storage-1",
						},
					},
				},
				"trace_id": "trace-1",
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "host", "boot-config", "get", "--id", "host-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"boot_configs"`) || !strings.Contains(out.String(), `"migration_id"`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHostsBootConfigCreateRequiresFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "create", "--id", "host-1"), &out, &errOut)
	if err == nil || err.Error() != "file is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigUpdateRequiresFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "update", "--id", "host-1"), &out, &errOut)
	if err == nil || err.Error() != "file is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigGetRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "get"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigGetRejectsIDsFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "get", "--id", "host-1", "--ids", "host-2"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigGetRejectsFileFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "get", "--id", "host-1", "--file", "x.json"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestHostBootConfigCommandAvailable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"boot_config_id": "cfg-1",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot-config", "get", "--id", "host-1", "--output", "json"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestBootConfigWizardSubcommandRemoved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "wizard", "storages"), &out, &errOut)
	if err == nil || err.Error() != `unknown boot-config command "wizard"` {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigCreateRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-create.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "create", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigUpdateRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-update.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "update", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigApplyRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-apply.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config-cli", "apply", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigApplyAvailableFromBootConfigGroup(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-apply.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var callCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.URL.Path {
		case "/api/v2/getStorageDetailInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"storage_type": "HyperGate"},
			})
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{},
			})
		case "/api/v2/batchBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"created": true},
			})
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "boot-config", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 3 {
		t.Fatalf("callCount = %d", callCount)
	}
}

func TestHostsBootConfigMetadataFileRejectsArray(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-metadata-array.json")
	if err := os.WriteFile(bodyPath, []byte(`[{"storage_id":"storage-1"}]`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "create", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "file must contain single metadata object" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigMetadataFileRejectsNonObject(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-metadata-string.json")
	if err := os.WriteFile(bodyPath, []byte(`"storage-1"`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config-cli", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "file must contain metadata object" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigMetadataFileRejectsBatchCreateWrapper(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-batch-create.json")
	if err := os.WriteFile(bodyPath, []byte(`{"batch_create":[{"migration_id":"host-1"}]}`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "create", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "file must contain metadata object, not batch_create wrapper" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigMetadataFileRejectsBatchUpdateWrapper(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-batch-update.json")
	if err := os.WriteFile(bodyPath, []byte(`{"batch_update":[{"id":"cfg-1","migration_id":"host-1"}]}`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "boot-config", "update", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "file must contain metadata object, not batch_update wrapper" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigWizardTargetAuthInfoWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"regions": []map[string]interface{}{{"id": "dc-1", "name": "Development"}}}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config-wizard", "target-auth-info",
		"--cloud-account-id", "account-1",
		"--cloud-type", "vmware_obs",
		"--storage-type", "object_storage",
		"--fetch-res", "regions,zones",
		"--custom-number", "3",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v3/getCloudInfo" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"cloud_account_id=account-1",
		"cloud_type=vmware_obs",
		"storage_type=object_storage",
		"fetch_res=regions%2Czones",
		"custom_number=3",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
}

func TestHostsBootConfigWizardTargetAuthInfoOmitsEmptyOptionalQueryFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"regions": []map[string]interface{}{{"id": "cn-beijing", "name": "Beijing"}}}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config-wizard", "target-auth-info",
		"--cloud-account-id", "account-1",
		"--cloud-type", "aliyun_obs",
		"--storage-type", "object_storage",
		"--fetch-res", "regions,zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"cloud_account_id=account-1",
		"cloud_type=aliyun_obs",
		"storage_type=object_storage",
		"fetch_res=regions%2Czones",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
	for _, unwanted := range []string{
		"arch=",
		"boot_loader_flavor_id=",
		"cloud_account_use_public=",
		"cloud_account_username=",
		"default_pool_id=",
		"default_volume_type_id=",
		"dest_boot_mode=",
		"flavor_id=",
		"flavor_ram=",
		"flavor_vcpus=",
		"flavors=",
		"host_id=",
		"max_nic_num=",
		"network_addr_for_read_data=",
		"network_addr_for_write_data=",
		"network_id=",
		"os_type=",
		"os_type_id=",
		"region_id=",
		"storage_id=",
		"system_volume_type_id=",
		"volume_type_id=",
		"zone_id=",
	} {
		if strings.Contains(gotQuery, unwanted) {
			t.Fatalf("query = %q, must not contain %q", gotQuery, unwanted)
		}
	}
}

func TestHostsBootConfigWizardTargetAuthInfoPassesAlibabaDependentResourceFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"cloud_info": map[string]interface{}{}}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config-wizard", "target-auth-info",
		"--cloud-account-id", "account-1",
		"--cloud-type", "aliyun_obs",
		"--storage-type", "objectstorage",
		"--fetch-res", "networks,subnets,security_groups",
		"--host-id", "host-1",
		"--storage-id", "storage-1",
		"--network-addr-for-write-data", "public_endpoint",
		"--network-addr-for-read-data", "internal_endpoint",
		"--region-id", "cn-beijing",
		"--zone-id", "cn-beijing-h",
		"--cloud-account-username", "ak",
		"--cloud-account-use-public", "0",
		"--flavor-id", "ecs.u1-c1m2.large",
		"--boot-loader-flavor-id", "ecs.c6t.large",
		"--arch", "x86_64",
		"--os-type-id", "id-Linux",
		"--os-type", "Linux",
		"--flavors", "id-ecs.u1-c1m2.large",
		"--flavor-vcpus", "2",
		"--flavor-ram", "4",
		"--max-nic-num", "2",
		"--system-volume-type-id", "cloud_essd_entry",
		"--volume-type-id", "cloud_essd_entry",
		"--default-volume-type-id", "cloud_essd_entry",
		"--default-pool-id", "pool-1",
		"--dest-boot-mode", "uefi",
		"--network-id", "vpc-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"host_id=host-1",
		"storage_id=storage-1",
		"network_addr_for_write_data=public_endpoint",
		"network_addr_for_read_data=internal_endpoint",
		"region_id=cn-beijing",
		"zone_id=cn-beijing-h",
		"cloud_account_username=ak",
		"cloud_account_use_public=0",
		"flavor_id=ecs.u1-c1m2.large",
		"boot_loader_flavor_id=ecs.c6t.large",
		"arch=x86_64",
		"os_type_id=id-Linux",
		"os_type=Linux",
		"flavors=id-ecs.u1-c1m2.large",
		"flavor_vcpus=2",
		"flavor_ram=4",
		"max_nic_num=2",
		"system_volume_type_id=cloud_essd_entry",
		"volume_type_id=cloud_essd_entry",
		"default_volume_type_id=cloud_essd_entry",
		"default_pool_id=pool-1",
		"dest_boot_mode=uefi",
		"network_id=vpc-1",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
}

func TestHostsBootConfigWizardTargetAuthInfoRequiresCloudAccount(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config-wizard", "target-auth-info"), &out, &errOut)
	if err == nil || err.Error() != "cloud-account is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestHostsBootConfigWizardSubnetConfigWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"cidr": "192.168.0.0/24"}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config-wizard", "subnet-config",
		"--cloud-account-id", "account-1",
		"--cloud-type", "aliyun_obs",
		"--region-id", "cn-beijing",
		"--zone-id", "cn-beijing-h",
		"--network-id", "vpc-1",
		"--subnet-id", "vsw-1",
		"--custom-number", "3",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v3/getSubnetConfig" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"cloud_account_id=account-1",
		"cloud_type=aliyun_obs",
		"region_id=cn-beijing",
		"zone_id=cn-beijing-h",
		"network_id=vpc-1",
		"subnet_id=vsw-1",
		"custom_number=3",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
}

func TestHostsBootConfigWizardSubnetConfigRequiresFields(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "cloud account", args: []string{"--cloud-type", "aliyun_obs", "--region-id", "cn-beijing", "--zone-id", "cn-beijing-h", "--network-id", "vpc-1"}, want: "cloud-account is required"},
		{name: "cloud type", args: []string{"--cloud-account-id", "account-1", "--region-id", "cn-beijing", "--zone-id", "cn-beijing-h", "--network-id", "vpc-1"}, want: "cloud-type is required"},
		{name: "region", args: []string{"--cloud-account-id", "account-1", "--cloud-type", "aliyun_obs", "--zone-id", "cn-beijing-h", "--network-id", "vpc-1"}, want: "region-id is required"},
		{name: "zone", args: []string{"--cloud-account-id", "account-1", "--cloud-type", "aliyun_obs", "--region-id", "cn-beijing", "--network-id", "vpc-1"}, want: "zone-id is required"},
		{name: "network", args: []string{"--cloud-account-id", "account-1", "--cloud-type", "aliyun_obs", "--region-id", "cn-beijing", "--zone-id", "cn-beijing-h"}, want: "network-id is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			var out, errOut bytes.Buffer
			args := append(withHost(t, "https://example.invalid", "boot-config-wizard", "subnet-config"), tc.args...)
			err := Execute(args, &out, &errOut)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestHostsDRConfigIsRejected(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "host", "dr-config", "get", "--id", "cfg-1"), &out, &errOut)
	if err == nil || err.Error() != `unknown host command "dr-config"` {
		t.Fatalf("err = %v", err)
	}
}

func TestRemovedCommandsAreNotAccepted(t *testing.T) {
	tests := [][]string{
		{"hosts", "list"},
		{"hosts", "sync"},
		{"host", "clean"},
		{"host", "delete-boot"},
		{"login"},
		{"upgrade", "hosts"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			var out, errOut bytes.Buffer
			if err := Execute(withHost(t, "https://example.invalid", args...), &out, &errOut); err == nil {
				t.Fatal("expected removed command to fail")
			}
		})
	}
}
