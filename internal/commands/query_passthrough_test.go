package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetCommandsPassUnknownFlagsAsQuery(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{name: "tasks list", args: []string{"tasks", "list", "--custom-step", "3"}, path: "/api/v2/getTasks"},
		{name: "tasks steps", args: []string{"tasks", "steps", "--task-id", "task-1", "--custom-step", "3"}, path: "/hypermotion/v1/job/steps"},
		{name: "source list", args: []string{"source", "list", "--type", "vmware", "--custom-step", "3"}, path: "/hypermotion/v1/sources"},
		{name: "source detail", args: []string{"source", "detail", "--id", "conn-1", "--custom-step", "3"}, path: "/api/v2/getConnectionDetail"},
		{name: "source vms", args: []string{"source", "vms", "--connection-type", "vmware", "--connection-uuid", "conn-1", "--custom-step", "3"}, path: "/hypermotion/v1/sources/vms"},
		{name: "source agent-install", args: []string{"source", "agent-install", "--custom-step", "3"}, path: "/hypermotion/v1/sources"},
		{name: "source agentless-install", args: []string{"source", "agentless-install", "--custom-step", "3"}, path: "/hypermotion/v1/sources"},
		{name: "source sync-nodes", args: []string{"source", "sync-nodes", "--custom-step", "3"}, path: "/hypermotion/v1/synch_nodes"},
		{name: "target account list", args: []string{"target", "account", "list", "--custom-step", "3"}, path: "/api/v2/getCloudAccounts"},
		{name: "target account detail", args: []string{"target", "account", "detail", "--id", "account-1", "--custom-step", "3"}, path: "/hypermotion/v1/cloud_accounts/account-1"},
		{name: "target cloud-sync-gateway list", args: []string{"target", "cloud-sync-gateway", "list", "--custom-step", "3"}, path: "/api/v2/getStorages"},
		{name: "target cloud-sync-gateway detail", args: []string{"target", "cloud-sync-gateway", "detail", "--id", "storage-1", "--custom-step", "3"}, path: "/api/v2/getStorageDetailInfo"},
		{name: "target oss list", args: []string{"target", "oss", "list", "--custom-step", "3"}, path: "/api/v2/getStorages"},
		{name: "target oss detail", args: []string{"target", "oss", "detail", "--id", "storage-1", "--custom-step", "3"}, path: "/api/v2/getStorageDetailInfo"},
		{name: "licenses list", args: []string{"licenses", "list", "--custom-step", "3"}, path: "/api/v2/getLicenses"},
		{name: "licenses reg-code", args: []string{"licenses", "reg-code", "--custom-step", "3"}, path: "/api/v2/getLicenseRegCode"},
		{name: "boot-config-wizard storages", args: []string{"boot-config-wizard", "storages", "--custom-step", "3"}, path: "/api/v2/getStorages"},
		{name: "boot-config-wizard storage-detail", args: []string{"boot-config-wizard", "storage-detail", "--storage-id", "storage-1", "--custom-step", "3"}, path: "/api/v2/getStorageDetailInfo"},
		{name: "boot-config-wizard target-platforms", args: []string{"boot-config-wizard", "target-platforms", "--custom-step", "3"}, path: "/api/v2/getCloudAccounts"},
		{name: "boot-config-wizard target-accounts", args: []string{"boot-config-wizard", "target-accounts", "--custom-step", "3"}, path: "/api/v2/getCloudAccounts"},
		{name: "boot-config-wizard host-profile", args: []string{"boot-config-wizard", "host-profile", "--id", "host-1", "--custom-step", "3"}, path: "/api/v2/getHostDetail"},
		{name: "boot-config-wizard strategies", args: []string{"boot-config-wizard", "strategies", "--custom-step", "3"}, path: "/api/v2/getHostPolicyList"},
		{name: "upgrade host", args: []string{"upgrade", "host", "--custom-step", "3"}, path: "/api/v2/getUpgradeHostList"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			var gotPath, gotQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.RawQuery
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
			}))
			defer srv.Close()

			args := append(withHost(t, srv.URL), tt.args...)
			var out, errOut bytes.Buffer
			if err := Execute(args, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			if gotPath != tt.path {
				t.Fatalf("path = %q, want %q", gotPath, tt.path)
			}
			if !strings.Contains(gotQuery, "custom_step=3") {
				t.Fatalf("query = %q", gotQuery)
			}
		})
	}
}

func TestBootConfigWizardTargetPlatformsUsesCloudAccountsAlias(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	var gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config-wizard", "target-platforms",
		"--target-type", "recovery",
		"--cloud-type", "vmware_obs",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/getCloudAccounts" {
		t.Fatalf("path = %q, want %q", gotPath, "/api/v2/getCloudAccounts")
	}
	if gotQuery.Get("storage_type") != "objectstorage" {
		t.Fatalf("storage_type = %q, want %q", gotQuery.Get("storage_type"), "objectstorage")
	}
	if gotQuery.Get("status") != "available" {
		t.Fatalf("status = %q, want %q", gotQuery.Get("status"), "available")
	}
	if gotQuery.Get("cloud_type") != "vmware_obs" {
		t.Fatalf("cloud_type = %q, want %q", gotQuery.Get("cloud_type"), "vmware_obs")
	}
	if gotQuery.Get("target_type") != "" {
		t.Fatalf("target_type = %q, want empty", gotQuery.Get("target_type"))
	}
}

func TestBootConfigWizardStoragesDefaultsStatusOnly(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	var gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config-wizard", "storages",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/getStorages" {
		t.Fatalf("path = %q, want %q", gotPath, "/api/v2/getStorages")
	}
	if gotQuery.Get("type") != "" {
		t.Fatalf("type = %q, want empty", gotQuery.Get("type"))
	}
	if gotQuery.Get("status") != "available" {
		t.Fatalf("status = %q, want %q", gotQuery.Get("status"), "available")
	}
}
