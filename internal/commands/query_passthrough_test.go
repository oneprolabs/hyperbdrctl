package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCommandsPassUnknownFlagsAsQuery(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
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
		{name: "license list", args: []string{"license", "list", "--custom-step", "3"}, path: "/api/v2/getLicenses"},
		{name: "license reg-code", args: []string{"license", "reg-code", "--custom-step", "3"}, path: "/api/v2/getLicenseRegCode"},
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
