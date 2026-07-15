package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/config"
	"hyperbdr-client/internal/i18n"
)

func TestSourcesAgentInstallDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": sampleAgentInstallData(),
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "agent", "install", "--custom-step", "3"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/sources" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "type=agent") || !strings.Contains(gotQuery, "custom_step=3") {
		t.Fatalf("query = %q", gotQuery)
	}

	text := out.String()
	for _, want := range []string{
		"== Agent Install ==",
		"Install agent on Linux or Windows source hosts.",
		"+ Linux",
		"  Description: Install agent on Linux hosts.",
		"  Best Choice: Default",
		"  Variants: Default, DKMS",
		"  Supported Systems:",
		"    - CentOS Linux 7",
		"    - Ubuntu 20.04 LTS",
		"    - Ubuntu 22.04 LTS",
		"    - CentOS 8",
		"  Recommendations:",
		"    - Ubuntu 20.04 / 22.04 / 24.04: Recommend DKMS",
		"    - CentOS 8 / 9: Recommend DKMS",
		"  [Default]",
		"  [DKMS]",
		"  curl -k 'https://example.invalid/linux' | bash",
		"  curl -k 'https://example.invalid/linux' | bash -s -- --enable-dkms",
		"+ Windows",
		"  Description: Run this command on the Windows source host.",
		"  Best Choice: Install Command",
		"  Variants: Install Command",
		"    - Windows Server 2016",
		"    - Windows Server 2019",
		"    - Windows Server X86",
		"    - Windows Server X64",
		"  Recommendations:",
		"    - Windows Server 2019 and later: Recommend Install Command",
		"  [Install Command]",
		"  curl -k -o winagent_dl.bat \"https://example.invalid/windows-install-cmd\" && winagent_dl.bat",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "Windows_install_cmd") {
		t.Fatalf("output should hide Windows_install_cmd placeholder: %q", text)
	}
	assertContainsInOrder(t, text, "+ Linux", "Supported Systems:", "Recommendations:", "[Default]", "[DKMS]", "+ Windows", "Supported Systems:", "Recommendations:", "[Install Command]")
}

func TestSourceListSupportsVerticalOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypermotion/v1/sources" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"sources": []map[string]interface{}{
					{
						"id":                    "source-1",
						"uuid":                  "uuid-1",
						"name":                  "vmware-prod",
						"type":                  "vmware",
						"display_source_status": "binding",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "production-site", "list", "--type", "vmware", "-G"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"*************************** 1. row ***************************",
		"ID",
		"source-1",
		"UUID",
		"uuid-1",
		"Name",
		"vmware-prod",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("vertical output missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "ID        UUID") {
		t.Fatalf("expected vertical output, got table-like output: %q", text)
	}
}

func TestSourcesAgentInstallJSONOutputPreservesRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": sampleAgentInstallData(),
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "agent", "install"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{`"agent"`, `"install_link_dkms"`, `"Windows_install_cmd"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesAgentInstallDefaultOutputSupportsBarePayload(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(sampleAgentInstallData())
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "agent", "install"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"== Agent Install ==", "Linux", "Default", "DKMS", "Windows", "Install Command"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, `"agent"`) {
		t.Fatalf("output should not fall back to json: %q", text)
	}
}

func TestSourcesListTypeAgentReturnsGuidanceError(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "production-site", "list", "--type", "agent"), &out, &errOut)
	if err == nil {
		t.Fatal("expected guidance error")
	}
	if err.Error() != "agent install info is available via agent install" {
		t.Fatalf("err = %v", err)
	}
}

func TestSourcesListRequiresType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "production-site", "list"), &out, &errOut)
	if err == nil {
		t.Fatal("expected missing type error")
	}
	if err.Error() != "type is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestSourcesListDefaultOutputSupportsWrappedSourcesPayload(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"sources": map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"uuid":                  "eb5ab5af-980e-4303-9d19-3f19a7750008",
						"displayname":           "https://192.168.10.2:443 (User: zhangtianjie@vsphere.local)",
						"display_source_status": "正常",
						"type":                  "vsphere",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "production-site", "list", "--type", "vmware"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"ID",
		"UUID",
		"Name",
		"Type",
		"Status",
		"eb5ab5af-980e-4303-9d19-3f19a7750008",
		"https://192.168.10.2:443 (User: zhangtianjie@vsphere.local)",
		"vmware",
		"正常",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesAgentlessInstallDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":        "00000000",
			"title":       "无代理方式",
			"description": "通过无代理方法迁移",
			"data":        sampleAgentlessInstallWorkflowData(),
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "sync-proxy", "install", "--custom-step", "3"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/sources" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "type=agentless") || !strings.Contains(gotQuery, "custom_step=3") {
		t.Fatalf("query = %q", gotQuery)
	}
	text := out.String()
	for _, want := range []string{
		"== 无代理方式 ==",
		"Description",
		"通过无代理方法迁移",
		"Operating System",
		"Ubuntu 24.04",
		"Minimum Requirements",
		"At least 2 CPU cores, 4 GB RAM, and 100 GB disk space",
		"File System",
		"XFS or EXT4 only; LVM is not supported",
		"Network Access",
		"The host must be able to reach the source API endpoint",
		"Execute As",
		"root (all operations must run as root)",
		"Install Command",
		"curl -k 'https://example.invalid/proxy' | bash",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesAgentlessInstallJSONOutputPreservesRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":        "00000000",
			"title":       "无代理方式",
			"description": "通过无代理方法迁移",
			"data":        sampleAgentlessInstallWorkflowData(),
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "sync-proxy", "install"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{`"description"`, `"install_link"`, `下载同步节点OVA文件`} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesSyncNodesDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"synch_nodes": []interface{}{
					map[string]interface{}{
						"uuid":           "node-1",
						"name":           "proxy-1",
						"host_ip":        "192.168.8.22",
						"version":        "7.4.0",
						"type":           "proxy",
						"status":         "online",
						"display_status": "正常",
						"connections": []interface{}{
							map[string]interface{}{
								"type":     "vsphere",
								"status":   "active",
								"auth_url": "https://192.168.10.2:443",
								"uuid":     "conn-1",
							},
							map[string]interface{}{
								"type":     "aws",
								"status":   "idle",
								"auth_url": "https://aws.example.com",
								"uuid":     "conn-2",
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "sync-proxy", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/synch_nodes" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "type=proxy") || !strings.Contains(gotQuery, "status=online") {
		t.Fatalf("query = %q", gotQuery)
	}
	text := out.String()
	for _, want := range []string{
		"Sync Node 1",
		"ID",
		"Name",
		"Node IP",
		"Version",
		"Status",
		"Connections:",
		"node-1",
		"proxy-1",
		"192.168.8.22",
		"7.4.0",
		"正常",
		"1. vmware",
		"Status  active",
		"Address  https://192.168.10.2:443",
		"UUID    conn-1",
		"2. aws",
		"Status  idle",
		"Address  https://aws.example.com",
		"UUID    conn-2",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesSyncNodesDefaultOutputShowsNoneWhenNoConnections(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"synch_nodes": []interface{}{
					map[string]interface{}{
						"uuid":           "node-2",
						"name":           "proxy-2",
						"host_ip":        "192.168.8.23",
						"version":        "7.4.0",
						"status":         "offline",
						"display_status": "offline",
						"connections":    []interface{}{},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "sync-proxy", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"Sync Node 1",
		"node-2",
		"proxy-2",
		"Connections:",
		"  - None",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesSyncNodesJSONOutputPreservesRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"synch_nodes": []interface{}{
					map[string]interface{}{
						"uuid":    "node-3",
						"name":    "proxy-3",
						"host_ip": "192.168.8.24",
						"connections": []interface{}{
							map[string]interface{}{
								"type":     "vsphere",
								"status":   "active",
								"auth_url": "https://192.168.10.3:443",
								"uuid":     "conn-3",
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "sync-proxy", "list"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{`"synch_nodes"`, `"connections"`, `"auth_url"`, `"type": "vsphere"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesListSupportsExplicitVerificationFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"sources": []interface{}{}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "production-site", "list", "--type", "vmware", "--kw", "test-vm", "--binding-status", "binding", "--page", "1", "--page-size", "10"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type=vmware", "kw=test-vm", "binding_status=binding", "page=1", "page_size=10"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, missing %q", gotQuery, want)
		}
	}
}

func TestSourcesCreatePreviewRequestBuildsVMwarePayload(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"production-site", "create",
		"--type", "vmware",
		"--synch-node-id", "node-1",
		"--synch-node-ids", "node-2,node-1",
		"--auth-url", "https://192.168.10.2:443",
		"--auth-key", "zhangtianjie@vsphere.local",
		"--auth-cert", "2b24",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		`"type": "vsphere"`,
		`"synch_node_ids": [`,
		`"node-1"`,
		`"node-2"`,
		`"vsphere": {`,
		`"auth_url": "https://192.168.10.2:443"`,
		`"auth_key": "zhangtianjie@vsphere.local"`,
		`"auth_cert": "2b24"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestSourcesCreateSendsAWSPayload(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotMethod string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"id": "conn-1", "status": "binding"},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"production-site", "create",
		"--type", "aws",
		"--synch-node-ids", "node-1,node-2",
		"--auth-url", "test",
		"--auth-key", "testak",
		"--auth-cert", "1e2334261801",
		"--region-id", "test-region",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v2/createConnection" {
		t.Fatalf("method=%q path=%q", gotMethod, gotPath)
	}
	connection := gotBody["connection"].(map[string]interface{})
	if connection["type"] != "aws" {
		t.Fatalf("connection.type = %v", connection["type"])
	}
	aws := connection["aws"].(map[string]interface{})
	if aws["auth_key"] != "testak" || aws["auth_cert"] != "1e2334261801" || aws["region_id"] != "test-region" {
		t.Fatalf("aws payload = %#v", aws)
	}
}

func TestSourceSplitHelpUsesGroupLayouts(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args     []string
		want     []string
		unwanted []string
	}{
		{
			args:     []string{"help", "production-site"},
			want:     []string{"Usage:", "\nFlags:\n", "\nCommands:\n", "list", "detail", "delete", "vm-list", "create", "hyperbdrctl production-site create --help"},
			unwanted: []string{"agent-install", "agentless-install", "sync-nodes", "sync-node-delete"},
		},
		{
			args:     []string{"help", "agent"},
			want:     []string{"Usage:", "\nFlags:\n", "\nCommands:\n", "install", "hyperbdrctl agent install"},
			unwanted: []string{"agent-install"},
		},
		{
			args:     []string{"help", "sync-proxy"},
			want:     []string{"Usage:", "\nFlags:\n", "\nCommands:\n", "install", "list", "delete", "hyperbdrctl sync-proxy list"},
			unwanted: []string{"agentless-install", "sync-nodes", "sync-node-delete"},
		},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}
		text := out.String()
		for _, want := range tt.want {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v help = %q, missing %q", tt.args, text, want)
			}
		}
		for _, unwanted := range append(tt.unwanted, "\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n") {
			if strings.Contains(text, unwanted) {
				t.Fatalf("args=%v help should not include %q: %q", tt.args, unwanted, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestSourceLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"production-site", "list", "--help"},
			want: []string{"Usage Notes:", "--type", "--binding-status", "hyperbdrctl agent install"},
		},
		{
			args: []string{"production-site", "create", "--help"},
			want: []string{"Usage Notes:", "--type", "--synch-node-id", "--synch-node-ids", "--preview-request", "hyperbdrctl sync-proxy list"},
		},
		{
			args: []string{"production-site", "delete", "--help"},
			want: []string{"Usage Notes:", "--id", "--force", "hyperbdrctl production-site detail --id <source_id>"},
		},
		{
			args: []string{"sync-proxy", "delete", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl sync-proxy list"},
		},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}

		text := out.String()
		for _, want := range tt.want {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v help missing %q: %q", tt.args, want, text)
			}
		}
		if strings.Contains(text, "\nCommands:\n") {
			t.Fatalf("leaf help should not include commands section args=%v: %q", tt.args, text)
		}
		for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n", "\nMinimum Flags:\n", "\nCommon Optional Flags:\n"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("leaf help should not include %q args=%v: %q", unwanted, tt.args, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestSourceDeleteUsesDeleteEndpoint(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.RequestURI()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"deleted": true, "id": "source-1"},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "production-site", "delete", "--id", "source-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/hypermotion/v1/sources/source-1" {
		t.Fatalf("path = %q", gotPath)
	}
	if strings.Contains(gotPath, "force=") {
		t.Fatalf("path should not include force by default: %q", gotPath)
	}
	if !strings.Contains(out.String(), `"deleted": true`) && !strings.Contains(out.String(), `"deleted":true`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestSourceDeleteUsesForceQueryWhenRequested(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000"})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "production-site", "delete", "--id", "source-1", "--force"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/sources/source-1?force=true" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestSourceDeleteRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "production-site", "delete"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestSourceSyncNodeDeleteUsesDeleteEndpoint(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.RequestURI()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"deleted": true, "id": "node-1"},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "sync-proxy", "delete", "--id", "node-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/hypermotion/v1/synch_nodes/node-1" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(out.String(), `"deleted": true`) && !strings.Contains(out.String(), `"deleted":true`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestSourceSyncNodeDeleteRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "sync-proxy", "delete"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestSourceSplitHelpUsesDirectPaths(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args     []string
		want     []string
		unwanted []string
	}{
		{
			args:     []string{"--lang", "en", "production-site", "create", "--help"},
			want:     []string{"Usage: hyperbdrctl production-site create", "--type", "--synch-node-id", "hyperbdrctl sync-proxy list", "hyperbdrctl production-site list"},
			unwanted: []string{"hyperbdrctl source", "source create"},
		},
		{
			args:     []string{"--lang", "en", "agent", "install", "--help"},
			want:     []string{"Usage: hyperbdrctl agent install", "Usage Notes:", "hyperbdrctl agent install"},
			unwanted: []string{"source agent-install", "hyperbdrctl source"},
		},
		{
			args:     []string{"--lang", "en", "sync-proxy", "list", "--help"},
			want:     []string{"Usage: hyperbdrctl sync-proxy list", "--type", "--status", "hyperbdrctl production-site create --help"},
			unwanted: []string{"source sync-nodes", "hyperbdrctl source"},
		},
		{
			args:     []string{"--lang", "zh_cn", "production-site", "vm-list", "--help"},
			want:     []string{"用法: hyperbdrctl production-site vm-list", "--connection-type", "hyperbdrctl host register --vm-id <vm_id>"},
			unwanted: []string{"source vms", "hyperbdrctl source"},
		},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}
		got := normalizeHelpText(out.String())
		for _, want := range tt.want {
			if !strings.Contains(got, want) {
				t.Fatalf("args=%v help missing %q: %q", tt.args, want, got)
			}
		}
		for _, unwanted := range tt.unwanted {
			if strings.Contains(got, unwanted) {
				t.Fatalf("args=%v help should not include %q: %q", tt.args, unwanted, got)
			}
		}
		assertNoHelpFooter(t, got)
	}
}

func TestSourceSplitHelpMatchesArchivedCopyInBothLanguages(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name string
		path []string
		zh   []string
		en   []string
	}{
		{
			name: "production-site group",
			path: []string{"production-site"},
			zh:   []string{"生产站点管理", "管理生产站点，包括查询、创建、删除和生产虚拟机发现。", "如需创建 Agentless 生产站点"},
			en:   []string{"Production site management", "Manage production sites, including queries, creation, deletion, and production VM discovery.", "To create an Agentless production site"},
		},
		{
			name: "production-site list",
			path: []string{"production-site", "list"},
			zh:   []string{"列出生产站点", "生产站点类型（必须）", "最小查询命令如下：", "校验新建生产站点"},
			en:   []string{"List production sites", "Production site type (required)", "The minimum query is:", "validating a newly created production site"},
		},
		{
			name: "production-site detail",
			path: []string{"production-site", "detail"},
			zh:   []string{"查看生产站点详情", "查看单个生产站点详情。", "--binding-status binding"},
			en:   []string{"Show production site details", "View the details of one production site.", "--binding-status binding"},
		},
		{
			name: "production-site delete",
			path: []string{"production-site", "delete"},
			zh:   []string{"删除生产站点", "删除前确认生产站点：", "最小删除命令如下："},
			en:   []string{"Delete production site", "Confirm the production site before deletion:", "The minimum delete command is:"},
		},
		{
			name: "production-site vm-list",
			path: []string{"production-site", "vm-list"},
			zh:   []string{"列出生产站点虚拟机", "注册状态过滤，可选值 0 / 1", "参数来源：", "生产站点 UUID 通常来自："},
			en:   []string{"List production site VMs", "Registration filter, allowed values 0 / 1", "Parameter Sources:", "The production site UUID usually comes from:"},
		},
		{
			name: "production-site create",
			path: []string{"production-site", "create"},
			zh:   []string{"创建生产站点", "生产站点类型（必须），可选值 vmware / aws", "多个同步节点 ID 使用英文逗号分隔。", "--type vmware --binding-status binding"},
			en:   []string{"Create production site", "Production site type (required), allowed values", "vmware / aws", "Separate multiple sync node IDs with commas.", "--type vmware --binding-status binding"},
		},
		{
			name: "sync-proxy group",
			path: []string{"sync-proxy"},
			zh:   []string{"同步代理管理", "管理 Agentless 同步代理准备流程和已注册节点。", "输出同步代理安装说明："},
			en:   []string{"Sync proxy management", "Manage Agentless sync proxy preparation and registered nodes.", "Output sync proxy installation guidance:"},
		},
		{
			name: "sync-proxy install",
			path: []string{"sync-proxy", "install"},
			zh:   []string{"输出 Agentless 同步代理安装说明", "查看默认安装说明：", "确认同步节点在线后，继续创建生产站点："},
			en:   []string{"Output Agentless sync proxy installation guidance", "View the default installation guidance:", "After confirming the sync node is online, continue by creating a production site:"},
		},
		{
			name: "sync-proxy list",
			path: []string{"sync-proxy", "list"},
			zh:   []string{"列出同步代理节点", "在创建 Agentless 生产站点前", "--type string     类型过滤，默认值 proxy"},
			en:   []string{"List sync proxy nodes", "before creating an Agentless production site", "--type string     Type filter, default proxy"},
		},
		{
			name: "sync-proxy delete",
			path: []string{"sync-proxy", "delete"},
			zh:   []string{"删除同步代理节点", "删除前确认同步代理节点：", "删除成功后，继续校验剩余节点："},
			en:   []string{"Delete sync proxy node", "Confirm the sync proxy node before deletion:", "After successful deletion, validate the remaining nodes:"},
		},
		{
			name: "agent group",
			path: []string{"agent"},
			zh:   []string{"Agent 源端代理管理", "管理 Agent 模式下的源端主机准备流程。", "输出源端代理安装说明："},
			en:   []string{"Agent source proxy management", "Manage the source host preparation workflow for Agent mode.", "Output source proxy installation guidance:"},
		},
		{
			name: "agent install",
			path: []string{"agent", "install"},
			zh:   []string{"输出 Agent 源端代理安装说明", "默认输出按 Linux 和 Windows 分段展示安装命令。", "安装完成后，可返回 CLI 查看源端主机："},
			en:   []string{"Output Agent source proxy installation guidance", "separate Linux and Windows sections", "After installation, return to the CLI to view source hosts:"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, lang := range []struct {
				name string
				want []string
			}{
				{name: "zh_cn", want: tc.zh},
				{name: "en", want: tc.en},
			} {
				t.Run(lang.name, func(t *testing.T) {
					args := append([]string{"--lang", lang.name}, tc.path...)
					args = append(args, "--help")
					var out, errOut bytes.Buffer
					if err := Execute(args, &out, &errOut); err != nil {
						t.Fatal(err)
					}
					text := normalizeHelpText(out.String())
					for _, want := range lang.want {
						if !strings.Contains(text, want) {
							t.Fatalf("help missing %q: %q", want, text)
						}
					}
				})
			}
		})
	}
}

func normalizeHelpText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "--vertical") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func TestSourceCommandsAreRemoved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"source"},
		{"source", "list"},
		{"source", "detail"},
		{"source", "delete"},
		{"source", "sync-node-delete"},
		{"source", "vms"},
		{"source", "agent-install"},
		{"source", "agentless-install"},
		{"source", "sync-nodes"},
		{"source", "create"},
		{"help", "source"},
		{"help", "sources"},
	}

	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute(args, &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), "unknown") {
				t.Fatalf("args=%v err=%v", args, err)
			}
		})
	}
}

func TestWriteAgentInstallResponseSkipsMissingRows(t *testing.T) {
	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("en"),
	}
	resp := client.APIResponse{
		Data: map[string]interface{}{
			"agent": map[string]interface{}{
				"linux": map[string]interface{}{
					"install_link": "curl -k 'https://example.invalid/linux' | bash",
				},
			},
		},
	}

	if err := writeAgentInstallResponse(ctx, resp); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if strings.Contains(text, "DKMS") || strings.Contains(text, "Windows") {
		t.Fatalf("output should skip missing rows: %q", text)
	}
	for _, want := range []string{
		"Linux",
		"Best Choice: Default",
		"Variants: Default",
		"Supported Systems:",
		"    - None",
		"[Default]",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "Recommendations:") {
		t.Fatalf("output should omit empty recommendations: %q", text)
	}
}

func TestWriteAgentInstallResponseFallsBackToHumanOutput(t *testing.T) {
	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("en"),
	}
	resp := client.APIResponse{
		Data: map[string]interface{}{
			"agent": map[string]interface{}{
				"title": "Agent",
			},
		},
	}

	if err := writeAgentInstallResponse(ctx, resp); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if strings.Contains(text, "{") || strings.Contains(text, `"title": "Agent"`) {
		t.Fatalf("output should stay human-readable: %q", text)
	}
	if !strings.Contains(text, "== Agent ==") {
		t.Fatalf("output = %q", text)
	}
}

func TestWriteAgentInstallResponseSupportsUnwrappedAgentObject(t *testing.T) {
	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("en"),
	}
	resp := client.APIResponse{
		Data: map[string]interface{}{
			"linux": map[string]interface{}{
				"install_link": "curl -k 'https://example.invalid/linux' | bash",
			},
			"windows": map[string]interface{}{
				"system_version": map[string]interface{}{
					"url_list": []interface{}{
						map[string]interface{}{
							"system": "Windows Server 2019",
							"link":   "https://example.invalid/windows",
						},
						map[string]interface{}{
							"system": "Windows_install_cmd",
							"link":   "curl -k -o winagent_dl.bat \"https://example.invalid/windows-install-cmd\" && winagent_dl.bat",
						},
					},
				},
			},
		},
	}

	if err := writeAgentInstallResponse(ctx, resp); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"Linux",
		"Windows",
		"Best Choice: Default",
		"Best Choice: Install Command",
		"[Default]",
		"[Install Command]",
		"    - Windows Server 2019",
		"  Recommendations:",
		"    - Windows Server 2019 and later: Recommend Install Command",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "Windows_install_cmd") {
		t.Fatalf("output should hide Windows_install_cmd placeholder: %q", text)
	}
}

func TestWriteAgentInstallResponseSupportsLiveAPIShape(t *testing.T) {
	payload := map[string]interface{}{
		"description": "live api shape",
		"title":       "agent mode",
		"linux": map[string]interface{}{
			"description":       "linux description",
			"install_link":      "curl -k 'https://example.invalid/linux' | (cd ~ && bash)",
			"install_link_dkms": "curl -k 'https://example.invalid/linux' | (cd ~ && bash -s -- --enable-dkms)",
			"system_version": map[string]interface{}{
				"url_list": []interface{}{
					map[string]interface{}{"system": "CentOS Linux 6/7/8"},
					map[string]interface{}{"system": "Ubuntu 12.04/14.04/16.04 LTS"},
				},
			},
			"title": "Linux",
		},
		"windows": map[string]interface{}{
			"description": "windows description",
			"system_version": map[string]interface{}{
				"url_list": []interface{}{
					map[string]interface{}{"link": "", "system": "缁嬪啿鐣鹃悧鍫熸拱"},
					map[string]interface{}{"link": "curl -k -o winagent_dl.bat \"https://example.invalid/windows-install-cmd\" && winagent_dl.bat", "system": "Windows_install_cmd"},
				},
			},
			"title": "Windows",
		},
	}

	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("zh_cn"),
	}
	resp := client.APIResponse{Data: payload}
	defaultVariant := ctx.loc.T("value.agent_install.variant.default")
	installVariant := ctx.loc.T("value.agent_install.variant.install_cmd")
	supportedSystems := ctx.loc.T("label.agent_install.supported_systems")
	recommendations := ctx.loc.T("label.agent_install.recommendations")
	ubuntuRecommendation := ctx.loc.T("value.agent_install.recommendation.linux_ubuntu_dkms")
	centosRecommendation := ctx.loc.T("value.agent_install.recommendation.linux_centos_dkms")
	windowsRecommendation := ctx.loc.T("value.agent_install.recommendation.windows_install_cmd")

	if err := writeAgentInstallResponse(ctx, resp); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"Linux",
		defaultVariant,
		"DKMS",
		"Windows",
		installVariant,
		supportedSystems,
		recommendations,
		ubuntuRecommendation,
		centosRecommendation,
		windowsRecommendation,
		"    - CentOS Linux 6/7/8",
		"    - Ubuntu 12.04/14.04/16.04 LTS",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, `"linux"`) || strings.Contains(text, `"windows"`) {
		t.Fatalf("output should not fall back to json: %q", text)
	}
}

func assertContainsInOrder(t *testing.T, text string, wants ...string) {
	t.Helper()

	pos := 0
	for _, want := range wants {
		idx := strings.Index(text[pos:], want)
		if idx < 0 {
			t.Fatalf("output = %q, missing ordered fragment %q", text, want)
		}
		pos += idx + len(want)
	}
}

func sampleAgentInstallData() map[string]interface{} {
	return map[string]interface{}{
		"agent": map[string]interface{}{
			"title":       "Agent Install",
			"description": "Install agent on Linux or Windows source hosts.",
			"linux": map[string]interface{}{
				"title":             "Linux",
				"description":       "Install agent on Linux hosts.",
				"install_link":      "curl -k 'https://example.invalid/linux' | bash",
				"install_link_dkms": "curl -k 'https://example.invalid/linux' | bash -s -- --enable-dkms",
				"system_version": map[string]interface{}{
					"url_list": []interface{}{
						map[string]interface{}{"system": "CentOS Linux 7"},
						map[string]interface{}{"system": "Ubuntu 20.04 LTS"},
						map[string]interface{}{"system": "Ubuntu 22.04 LTS"},
						map[string]interface{}{"system": "CentOS 8"},
					},
				},
			},
			"windows": map[string]interface{}{
				"title":       "Windows",
				"description": "Run this command on the Windows source host.",
				"system_version": map[string]interface{}{
					"url_list": []interface{}{
						map[string]interface{}{
							"system": "Windows Server 2016",
							"link":   "https://example.invalid/windows",
						},
						map[string]interface{}{
							"system": "Windows Server 2019",
							"link":   "https://example.invalid/windows-2019",
						},
						map[string]interface{}{
							"system": "Windows_server_32bit",
							"link":   "https://example.invalid/windows-x86",
						},
						map[string]interface{}{
							"system": "Windows_server_64bit",
							"link":   "https://example.invalid/windows-x64",
						},
						map[string]interface{}{
							"system": "Windows_install_cmd",
							"link":   "curl -k -o winagent_dl.bat \"https://example.invalid/windows-install-cmd\" && winagent_dl.bat",
						},
					},
				},
			},
		},
	}
}

func sampleAgentlessInstallWorkflowData() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"title": "支持VMware版本：vCenter/ESXi 5.1/5.5/6.0/6.5/6.7",
		},
		map[string]interface{}{
			"install_link": "https://example.invalid/proxy-agent_BaseOS.ova",
			"title":        "(1) 下载同步节点OVA文件",
		},
		map[string]interface{}{
			"description":   "支持以下Linux操作系统",
			"install_link":  "curl -k 'https://example.invalid/proxy' | bash",
			"install_title": "登录通过OVA导入的主机后执行安装命令：",
			"system_version": map[string]interface{}{
				"url_list": []interface{}{
					map[string]interface{}{"system": "CentOS Linux 7"},
					map[string]interface{}{"system": "Red Hat Enterprise Linux 7"},
				},
			},
			"title": "(2) 安装同步节点",
		},
	}
}
