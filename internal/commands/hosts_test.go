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
	"time"
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

func TestHostsBootAllowsFormerLegacyConfigFlagsAsBodyOverrides(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args    []string
		wantKey string
		want    interface{}
	}{
		{args: []string{"--host", "https://legacy.invalid"}, wantKey: "host", want: "https://legacy.invalid"},
		{args: []string{"--username", "legacy-user"}, wantKey: "username", want: "legacy-user"},
		{args: []string{"--password", "legacy-pass"}, wantKey: "password", want: "legacy-pass"},
		{args: []string{"--scene", "migration"}, wantKey: "scene", want: "migration"},
		{args: []string{"--insecure"}, wantKey: "insecure", want: true},
	}

	for _, tc := range cases {
		var gotBody map[string]interface{}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v2/getHostDetail":
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{
						"snapshots": []map[string]interface{}{
							{"id": "snap-1", "created_at": "2026-06-16T10:00:00Z"},
						},
					},
				})
			case "/api/v2/batchBoot":
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Fatal(err)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
			default:
				t.Fatalf("path = %q", r.URL.Path)
			}
		}))

		var out, errOut bytes.Buffer
		args := append(withHost(t, srv.URL, "host", "boot", "--id", "host-1"), tc.args...)
		err := Execute(args, &out, &errOut)
		srv.Close()
		if err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		item := gotBody["batch_boot"].([]interface{})[0].(map[string]interface{})
		if item[tc.wantKey] != tc.want {
			t.Fatalf("args=%v item=%+v want %s=%v", args, item, tc.wantKey, tc.want)
		}
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

func TestHostsBootUsesLatestSnapshotByDefault(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPaths []string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path+"?"+r.URL.RawQuery)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			if r.URL.Query().Get("host_id") != "host-1" || r.URL.Query().Get("sheet") != "snapshot" {
				t.Fatalf("query = %q", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"snapshots": []map[string]interface{}{
						{"id": "snap-1", "created_at": "2026-06-15T10:00:00Z"},
						{"id": "snap-2", "created_at": "2026-06-16T10:00:00Z"},
					},
				},
			})
		case "/api/v2/batchBoot":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"accepted": true}})
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot", "--id", "host-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 {
		t.Fatalf("paths = %+v", gotPaths)
	}
	item := gotBody["batch_boot"].([]interface{})[0].(map[string]interface{})
	if item["migration_id"] != "host-1" || item["snapshot_id"] != "snap-2" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestHostsBootFailsWhenLatestSnapshotCannotBeResolved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/getHostDetail" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"snapshots": []map[string]interface{}{},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "host", "boot", "--id", "host-1"), &out, &errOut)
	if err == nil || err.Error() != "host has no snapshots" {
		t.Fatalf("err = %v", err)
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

func TestHostsCleanWithFlags(t *testing.T) {
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
	err := Execute(withHost(t, srv.URL, "host", "clean", "--id", "host-1", "--ids", "host-2,host-3"), &out, &errOut)
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

func TestHostsCleanWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "clean-hosts.json")
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
	err := Execute(withHost(t, srv.URL, "host", "clean", "--file", bodyPath), &out, &errOut)
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

func TestHostsWaitCleanUsesBootStatus(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	statuses := []string{"clean_doing", "clean_done"}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := statuses[calls]
		calls++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"id":                          "host-1",
				"boot_status":                 status,
				"display_boot_status":         status,
				"boot_task_id":                "task-clean",
				"boot_task_error_description": "",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "clean",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
	if !strings.Contains(out.String(), `"task_id": "task-clean"`) || !strings.Contains(out.String(), `"status": "clean_done"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsWaitCleanTreatsNotBootAsSuccess(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	statuses := []string{"clean_doing", "not_boot"}
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
				"boot_task_id":        "task-clean",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "clean",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
	if !strings.Contains(out.String(), `"result": "success"`) || !strings.Contains(out.String(), `"status": "not_boot"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestHostsWaitCleanTreatsBootStatusesAsRunningUntilNotBoot(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	statuses := []string{"boot_doing", "boot_done", "not_boot"}
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
				"boot_task_id":        "task-clean",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"host", "wait",
		"--id", "host-1",
		"--operation", "clean",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d", calls)
	}
	if !strings.Contains(out.String(), `"result": "success"`) || !strings.Contains(out.String(), `"status": "not_boot"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestResolveHostWaitTimeoutUsesCleanDefaultWhenUnset(t *testing.T) {
	if got := resolveHostWaitTimeout("clean", 3600, false); got != 300*time.Second {
		t.Fatalf("timeout = %s", got)
	}
}

func TestResolveHostWaitTimeoutKeepsExplicitCleanTimeout(t *testing.T) {
	if got := resolveHostWaitTimeout("clean", 10, true); got != 10*time.Second {
		t.Fatalf("timeout = %s", got)
	}
}

func TestHostsWaitRejectsLegacyCleanupOperation(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"host", "wait",
		"--id", "host-1",
		"--operation", "cleanup-validation-host",
		"--interval-seconds", "0",
		"--timeout-seconds", "10",
	), &out, &errOut)
	if err == nil || err.Error() != "operation must be one of sync, boot, clean, deregister" {
		t.Fatalf("err = %v", err)
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
		{"host", "cleanup-validation-host"},
		{"host", "boot-config", "get"},
		{"host", "boot-config", "apply"},
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
