package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTargetWaitCommandsRequireID(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{args: []string{"cloud-account", "wait"}},
		{args: []string{"cloud-sync-gateway", "wait"}},
		{args: []string{"oss", "wait"}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, "-"), func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)

			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid", tt.args...), &out, &errOut)
			if err == nil || err.Error() != "id is required" {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestCloudAccountWaitJSONOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypermotion/v1/cloud_accounts/account-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"status":         "active",
				"display_status": "Available",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-account", "wait",
		"--id", "account-1",
		"--interval-seconds", "0",
		"--timeout-seconds", "1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0]["result"] != "success" || rows[0]["operation"] != "create-account" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestCloudAccountWaitJSONOutputFromTopLevelCloudAccount(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypermotion/v1/cloud_accounts/account-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"cloud_account": map[string]interface{}{
				"status":         "available",
				"display_status": "Available",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-account", "wait",
		"--id", "account-1",
		"--interval-seconds", "0",
		"--timeout-seconds", "1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0]["result"] != "success" || rows[0]["status"] != "available" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestTargetCloudSyncGatewayWaitTypeMismatchReturnsError(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/getStorageDetailInfo" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storage": map[string]interface{}{
					"type":   "objectstorage",
					"status": "creating",
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "wait",
		"--id", "storage-1",
		"--interval-seconds", "0",
		"--timeout-seconds", "1",
	), &out, &errOut)
	if err == nil || err.Error() != "resource type mismatch: expected cloud-sync-gateway, got objectstorage" {
		t.Fatalf("err = %v", err)
	}

	var rows []map[string]interface{}
	if jsonErr := json.Unmarshal(out.Bytes(), &rows); jsonErr != nil {
		t.Fatal(jsonErr)
	}
	if rows[0]["result"] != "failed" {
		t.Fatalf("rows = %+v", rows)
	}
}
