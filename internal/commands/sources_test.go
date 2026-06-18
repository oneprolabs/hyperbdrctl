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
	err := Execute(withHost(t, srv.URL, "source", "agent-install", "--custom-step", "3"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "list", "--type", "vmware", "-G"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "--output", "json", "source", "agent-install"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "agent-install"), &out, &errOut)
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
	err := Execute(withHost(t, "https://example.invalid", "source", "list", "--type", "agent"), &out, &errOut)
	if err == nil {
		t.Fatal("expected guidance error")
	}
	if err.Error() != "agent install info is available via source agent-install" {
		t.Fatalf("err = %v", err)
	}
}

func TestSourcesListRequiresType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "source", "list"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "list", "--type", "vmware"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "agentless-install", "--custom-step", "3"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "--output", "json", "source", "agentless-install"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "sync-nodes"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "sync-nodes"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "--output", "json", "source", "sync-nodes"), &out, &errOut)
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
	err := Execute(withHost(t, srv.URL, "source", "list", "--type", "vmware", "--kw", "test-vm", "--binding-status", "binding", "--page", "1", "--page-size", "10"), &out, &errOut)
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
		"source", "create",
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
		"source", "create",
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

func TestSourceHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "source"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"\nCommands:\n",
		"list",
		"detail",
		"vms",
		"agent-install",
		"agentless-install",
		"sync-nodes",
		"create",
		"Usage Notes:",
		"hyperbdrctl source create --help",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help = %q, missing %q", text, want)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestSourceLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"source", "list", "--help"},
			want: []string{"Usage Notes:", "--type", "--binding-status", "hyperbdrctl source agent-install"},
		},
		{
			args: []string{"source", "create", "--help"},
			want: []string{"Usage Notes:", "--type", "--synch-node-id", "--synch-node-ids", "--preview-request", "hyperbdrctl source sync-nodes"},
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

func TestSourceEnHelpMatchesExpected(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want string
	}{
		{
			args: []string{"--lang", "en", "source", "--help"},
			want: joinHelpLines(
				"Production site configuration",
				"",
				"Usage: hyperbdrctl source <command> [flags]",
				"",
				"Flags:",
				"      --debug           Output request debug logs",
				"      --lang string     Display language, choices en / zh_cn, default en",
				"  -o, --output string   Output format, choices table / json, default table",
				"  -h, --help            Show help information",
				"",
				"Commands:",
				"  agent-install            Show Agent source proxy install commands",
				"  agentless-install        Show Agentless sync proxy install commands",
				"  create                   Create production platform",
				"  detail                   Show production platform detail",
				"  list                     List production platforms",
				"  sync-nodes               List sync proxy nodes",
				"  vms                      List Agentless VMs",
				"",
				"Usage Notes:",
				"  Use this command group for production platform queries, production host preparation, and production platform creation.",
				"",
				"  To inspect existing production platforms first, run:",
				"    hyperbdrctl source list --type vmware",
				"",
				"  To locate candidate production VM IDs before `host register`, run:",
				"    hyperbdrctl source vms --connection-type vmware",
				"",
				"  To prepare Agent or Agentless installation commands, run:",
				"    hyperbdrctl source agent-install",
				"    hyperbdrctl source agentless-install",
				"",
				"  To check available sync nodes before Agentless create, run:",
				"    hyperbdrctl source sync-nodes",
				"",
				"  To inspect the write flow and required provider inputs, continue with:",
				"    hyperbdrctl source create --help",
			),
		},
		{
			args: []string{"--lang", "en", "source", "list", "--help"},
			want: joinHelpLines(
				"List production platforms",
				"",
				"Usage: hyperbdrctl source list [flags]",
				"",
				"Flags:",
				"      --type string             Type filter (required)",
				"      --kw string               Keyword filter",
				"      --binding-status string   Binding status filter",
				"      --page int                Page number, default 1",
				"      --page-size int           Page size, default 10",
				"      --debug                   Output request debug logs",
				"      --lang string             Display language, choices en / zh_cn, default en",
				"  -o, --output string           Output format, choices table / json, default table",
				"  -h, --help                    Show help information",
				"",
				"Usage Notes:",
				"  Use this command to query configured production platforms by type, keyword, and binding state.",
				"",
				"  The minimum query is:",
				"    hyperbdrctl source list --type vmware",
				"",
				"  To narrow the result while checking a newly created production platform, run:",
				"    hyperbdrctl source list --type vmware --binding-status binding",
				"",
				"  To inspect raw API fields for scripting, add:",
				"    --output json",
				"",
				"  Do not use `--type agent` here. To inspect installation metadata, run:",
				"    hyperbdrctl source agent-install",
			),
		},
		{
			args: []string{"--lang", "en", "source", "detail", "--help"},
			want: joinHelpLines(
				"Show production platform detail",
				"",
				"Usage: hyperbdrctl source detail [flags]",
				"",
				"Flags:",
				"      --id string               Resource ID (required)",
				"      --type string             Type filter",
				"      --binding-status string   Binding status filter",
				"      --page int                Page number, default 1",
				"      --page-size int           Page size, default 10",
				"      --debug                   Output request debug logs",
				"      --lang string             Display language, choices en / zh_cn, default en",
				"  -o, --output string           Output format, choices table / json, default table",
				"  -h, --help                    Show help information",
				"",
				"Usage Notes:",
				"  Use this command to inspect one production platform in detail.",
				"",
				"  The minimum query is:",
				"    hyperbdrctl source detail --id <source_id>",
				"",
				"  When you need the related binding view together with the platform detail, add:",
				"    --binding-status binding",
				"",
				"  To script against the raw response fields, add:",
				"    --output json",
			),
		},
		{
			args: []string{"--lang", "en", "source", "vms", "--help"},
			want: joinHelpLines(
				"List production platform VMs",
				"",
				"Usage: hyperbdrctl source vms [flags]",
				"",
				"Flags:",
				"      --connection-type string   Production platform type (required)",
				"      --connection-uuid string   Production platform UUID",
				"      --registered string        Registration filter",
				"      --kw string                Keyword filter",
				"      --page int                 Page number, default 1",
				"      --page-size int            Page size, default 10",
				"      --debug                    Output request debug logs",
				"      --lang string              Display language, choices en / zh_cn, default en",
				"  -o, --output string            Output format, choices table / json, default table",
				"  -h, --help                     Show help information",
				"",
				"Usage Notes:",
				"  Use this command to query production VMs under one production platform.",
				"",
				"  Before running it, confirm the production platform type and, when available, the production platform UUID.",
				"",
				"  A minimal discovery query is:",
				"    hyperbdrctl source vms --connection-type vmware",
				"",
				"  To focus on unregistered VMs before `host register`, run:",
				"    hyperbdrctl source vms --connection-type vmware --registered 0",
				"",
				"  After you obtain a VM ID, continue with:",
				"    hyperbdrctl host register --vm-id <vm_id>",
			),
		},
		{
			args: []string{"--lang", "en", "source", "agent-install", "--help"},
			want: joinHelpLines(
				"Show Agent source proxy install commands",
				"",
				"Usage: hyperbdrctl source agent-install [flags]",
				"",
				"Flags:",
				"      --debug           Output request debug logs",
				"      --lang string     Display language, choices en / zh_cn, default en",
				"  -o, --output string   Output format, choices table / json, default table",
				"  -h, --help            Show help information",
				"",
				"Usage Notes:",
				"  Use this command to inspect Agent-mode source proxy installation guidance.",
				"",
				"  To read the simplified operator-facing install commands directly, run:",
				"    hyperbdrctl source agent-install",
				"",
				"  Default output groups Linux and Windows guidance separately.",
				"",
				"  To inspect the full raw per-OS metadata for scripting or diffing, add:",
				"    --output json",
			),
		},
		{
			args: []string{"--lang", "en", "source", "agentless-install", "--help"},
			want: joinHelpLines(
				"Show Agentless sync proxy install commands",
				"",
				"Usage: hyperbdrctl source agentless-install [flags]",
				"",
				"Flags:",
				"      --debug           Output request debug logs",
				"      --lang string     Display language, choices en / zh_cn, default en",
				"  -o, --output string   Output format, choices table / json, default table",
				"  -h, --help            Show help information",
				"",
				"Usage Notes:",
				"  Use this command to inspect Agentless sync-proxy installation prerequisites and install commands.",
				"",
				"  To read the simplified host preparation guidance directly, run:",
				"    hyperbdrctl source agentless-install",
				"",
				"  Then verify the required sync node is online with:",
				"    hyperbdrctl source sync-nodes",
				"",
				"  To inspect the raw metadata payload for scripting, add:",
				"    --output json",
			),
		},
		{
			args: []string{"--lang", "en", "source", "sync-nodes", "--help"},
			want: joinHelpLines(
				"List production platform sync proxy nodes",
				"",
				"Usage: hyperbdrctl source sync-nodes [flags]",
				"",
				"Flags:",
				"      --type string     Type filter, default proxy",
				"      --status string   Status filter, default online",
				"      --debug           Output request debug logs",
				"      --lang string     Display language, choices en / zh_cn, default en",
				"  -o, --output string   Output format, choices table / json, default table",
				"  -h, --help            Show help information",
				"",
				"Usage Notes:",
				"  Use this command to inspect registered source-side sync proxy nodes before Agentless create.",
				"",
				"  The default query already filters to online proxy nodes:",
				"    hyperbdrctl source sync-nodes",
				"",
				"  If you need to inspect other node types or statuses, override the defaults with:",
				"    --type <node_type>",
				"    --status <status>",
				"",
				"  After you confirm a usable node ID, continue with:",
				"    hyperbdrctl source create --help",
			),
		},
		{
			args: []string{"--lang", "en", "source", "create", "--help"},
			want: joinHelpLines(
				"Create production platform connection",
				"",
				"Usage: hyperbdrctl source create [flags]",
				"",
				"Flags:",
				"      --type string             Type filter (required), allowed values vmware / aws",
				"      --synch-node-id string    Single sync node ID",
				"      --synch-node-ids string   Comma-separated sync node IDs",
				"      --auth-url string         Cloud auth URL, source auth endpoint, or",
				"                                object-storage auth endpoint",
				"      --auth-key string         Source auth key or username",
				"      --auth-cert string        Source auth secret or certificate",
				"      --region-id string        Region ID",
				"      --preview-request         Output request body without sending request",
				"      --debug                   Output request debug logs",
				"      --lang string             Display language, choices en / zh_cn, default en",
				"  -o, --output string           Output format, choices table / json, default table",
				"  -h, --help                    Show help information",
				"",
				"Usage Notes:",
				"  Use this command to create an Agentless production platform connection for `vmware` or `aws`.",
				"",
				"  Before create, prepare the sync node and install metadata first:",
				"    hyperbdrctl source agentless-install",
				"    hyperbdrctl source sync-nodes",
				"",
				"  The minimum VMware create flow is:",
				"    hyperbdrctl source create \\",
				"      --type vmware \\",
				"      --synch-node-id <node_id> \\",
				"      --auth-url https://vcenter.example:443 \\",
				"      --auth-key <username> \\",
				"      --auth-cert <password>",
				"",
				"  When you need to inspect the final request body without sending `/api/v2/createConnection`, add:",
				"    --preview-request",
				"",
				"  After create returns, verify the connection state with:",
				"    hyperbdrctl source list --type <vmware|aws> --binding-status binding",
			),
		},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}

		got := normalizeHelpText(out.String())
		if got != tt.want {
			t.Fatalf("args=%v help mismatch\nwant:\n%s\n\ngot:\n%s", tt.args, tt.want, got)
		}
		assertNoHelpFooter(t, got)
	}
}

func TestSourceZhHelpMatchesArchive(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want string
	}{
		{
			args: []string{"--lang", "zh_cn", "source", "--help"},
			want: joinHelpLines(
				"生产站点配置",
				"",
				"用法: hyperbdrctl source <命令> [参数]",
				"",
				"参数:",
				"      --debug           输出请求调试日志",
				"      --lang string     显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string   输出格式，可选 table / json，默认值 table",
				"  -h, --help            显示帮助信息",
				"",
				"命令:",
				"  agent-install            查看 Agent 源端代理安装命令",
				"  agentless-install        查看 Agentless 同步代理安装命令",
				"  create                   创建生产平台",
				"  detail                   查看生产平台详情",
				"  list                     列出生产平台",
				"  sync-nodes               列出同步代理节点",
				"  vms                      列出 Agentless 虚拟机",
				"",
				"使用说明:",
				"  该命令组用于生产平台查询、生产主机准备和生产平台创建。",
				"",
				"  如需先查看当前已有的生产平台，可以执行：",
				"    hyperbdrctl source list --type vmware",
				"",
				"  如需在 `host register` 前先定位可注册的生产虚拟机 ID，可以执行：",
				"    hyperbdrctl source vms --connection-type vmware",
				"",
				"  如需准备 Agent 或 Agentless 安装命令，可以执行：",
				"    hyperbdrctl source agent-install",
				"    hyperbdrctl source agentless-install",
				"",
				"  如需在 Agentless 创建前确认可用同步节点，可以执行：",
				"    hyperbdrctl source sync-nodes",
				"",
				"  如需继续查看写入流程和所需参数，可以执行：",
				"    hyperbdrctl source create --help",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "list", "--help"},
			want: joinHelpLines(
				"列出生产平台",
				"",
				"用法: hyperbdrctl source list [参数]",
				"",
				"参数:",
				"      --type string             类型过滤（必须）",
				"      --kw string               关键字过滤",
				"      --binding-status string   绑定状态过滤",
				"      --page int                页码，默认值 1",
				"      --page-size int           每页数量，默认值 10",
				"      --debug                   输出请求调试日志",
				"      --lang string             显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string           输出格式，可选 table / json，默认值 table",
				"  -h, --help                    显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于按类型、关键字和绑定状态查询已配置的生产平台。",
				"",
				"  最小查询方式如下：",
				"    hyperbdrctl source list --type vmware",
				"",
				"  如需在校验新建生产平台时只看已绑定结果，可以执行：",
				"    hyperbdrctl source list --type vmware --binding-status binding",
				"",
				"  如需脚本化读取原始字段，可以附加：",
				"    --output json",
				"",
				"  不要在这里使用 `--type agent`。如需查看安装元数据，请执行：",
				"    hyperbdrctl source agent-install",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "detail", "--help"},
			want: joinHelpLines(
				"查看生产平台详情",
				"",
				"用法: hyperbdrctl source detail [参数]",
				"",
				"参数:",
				"      --id string               资源 ID（必须）",
				"      --type string             类型过滤",
				"      --binding-status string   绑定状态过滤",
				"      --page int                页码，默认值 1",
				"      --page-size int           每页数量，默认值 10",
				"      --debug                   输出请求调试日志",
				"      --lang string             显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string           输出格式，可选 table / json，默认值 table",
				"  -h, --help                    显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于查看单个生产平台详情。",
				"",
				"  最小查询方式如下：",
				"    hyperbdrctl source detail --id <source_id>",
				"",
				"  如需同时查看相关绑定视图，可附加：",
				"    --binding-status binding",
				"",
				"  如需脚本化读取原始返回字段，可附加：",
				"    --output json",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "vms", "--help"},
			want: joinHelpLines(
				"列出生产平台虚拟机",
				"",
				"用法: hyperbdrctl source vms [参数]",
				"",
				"参数:",
				"      --connection-type string   生产平台类型（必须）",
				"      --connection-uuid string   生产平台 UUID",
				"      --registered string        注册状态过滤",
				"      --kw string                关键字过滤",
				"      --page int                 页码，默认值 1",
				"      --page-size int            每页数量，默认值 10",
				"      --debug                    输出请求调试日志",
				"      --lang string              显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string            输出格式，可选 table / json，默认值 table",
				"  -h, --help                     显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于查询某个生产平台下的生产虚拟机。",
				"",
				"  执行前建议先确认生产平台类型，能拿到生产平台 UUID 时也可一并带上。",
				"",
				"  最小查询方式如下：",
				"    hyperbdrctl source vms --connection-type vmware",
				"",
				"  如需在 `host register` 前只看未注册虚拟机，可以执行：",
				"    hyperbdrctl source vms --connection-type vmware --registered 0",
				"",
				"  拿到 VM ID 后，可以继续执行：",
				"    hyperbdrctl host register --vm-id <vm_id>",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "agent-install", "--help"},
			want: joinHelpLines(
				"查看 Agent 源端代理安装命令",
				"",
				"用法: hyperbdrctl source agent-install [参数]",
				"",
				"参数:",
				"      --debug           输出请求调试日志",
				"      --lang string     显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string   输出格式，可选 table / json，默认值 table",
				"  -h, --help            显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于查看 Agent 模式下源端代理安装所需的操作说明。",
				"",
				"  如需直接查看精简后的安装命令，可以执行：",
				"    hyperbdrctl source agent-install",
				"",
				"  默认输出会分 Linux 和 Windows 展示。",
				"",
				"  如需脚本化比对完整原始元数据，可附加：",
				"    --output json",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "agentless-install", "--help"},
			want: joinHelpLines(
				"查看 Agentless 同步代理安装命令",
				"",
				"用法: hyperbdrctl source agentless-install [参数]",
				"",
				"参数:",
				"      --debug           输出请求调试日志",
				"      --lang string     显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string   输出格式，可选 table / json，默认值 table",
				"  -h, --help            显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于查看 Agentless 同步代理安装前置要求和安装命令。",
				"",
				"  如需直接查看主机准备要求和精简安装命令，可以执行：",
				"    hyperbdrctl source agentless-install",
				"",
				"  随后建议继续确认同步节点在线状态：",
				"    hyperbdrctl source sync-nodes",
				"",
				"  如需脚本化读取原始元数据，可附加：",
				"    --output json",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "sync-nodes", "--help"},
			want: joinHelpLines(
				"列出生产平台同步代理节点",
				"",
				"用法: hyperbdrctl source sync-nodes [参数]",
				"",
				"参数:",
				"      --type string     类型过滤，默认值 proxy",
				"      --status string   状态过滤，默认值 online",
				"      --debug           输出请求调试日志",
				"      --lang string     显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string   输出格式，可选 table / json，默认值 table",
				"  -h, --help            显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于在 Agentless 创建前查看已注册的源端同步代理节点。",
				"",
				"  默认查询已经限定为在线 proxy 节点：",
				"    hyperbdrctl source sync-nodes",
				"",
				"  如需查看其它类型或状态，可按需覆盖默认值：",
				"    --type <node_type>",
				"    --status <status>",
				"",
				"  确认可用节点 ID 后，可以继续执行：",
				"    hyperbdrctl source create --help",
			),
		},
		{
			args: []string{"--lang", "zh_cn", "source", "create", "--help"},
			want: joinHelpLines(
				"创建生产平台连接",
				"",
				"用法: hyperbdrctl source create [参数]",
				"",
				"参数:",
				"      --type string             类型过滤（必须），可选值 vmware / aws",
				"      --synch-node-id string    单个同步节点 ID",
				"      --synch-node-ids string   逗号分隔的同步节点 ID 列表",
				"      --auth-url string         云平台、源端或对象存储鉴权地址",
				"      --auth-key string         源端鉴权用户名或密钥",
				"      --auth-cert string        源端鉴权密码、密钥或证书",
				"      --region-id string        区域 ID",
				"      --preview-request         输出请求体，但不发送请求",
				"      --debug                   输出请求调试日志",
				"      --lang string             显示语言，可选 en / zh_cn，默认值 en",
				"  -o, --output string           输出格式，可选 table / json，默认值 table",
				"  -h, --help                    显示帮助信息",
				"",
				"使用说明:",
				"  该命令用于为 `vmware` 或 `aws` 创建 Agentless 生产平台连接。",
				"",
				"  创建前建议先准备同步节点和安装信息：",
				"    hyperbdrctl source agentless-install",
				"    hyperbdrctl source sync-nodes",
				"",
				"  VMware 最小创建方式如下：",
				"    hyperbdrctl source create \\",
				"      --type vmware \\",
				"      --synch-node-id <node_id> \\",
				"      --auth-url https://vcenter.example:443 \\",
				"      --auth-key <username> \\",
				"      --auth-cert <password>",
				"",
				"  如需先查看最终请求体但不发送 `/api/v2/createConnection`，可附加：",
				"    --preview-request",
				"",
				"  创建完成后，建议执行下面的命令校验连接状态：",
				"    hyperbdrctl source list --type <vmware|aws> --binding-status binding",
			),
		},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}

		got := normalizeHelpText(out.String())
		if got != tt.want {
			t.Fatalf("args=%v help mismatch\nwant:\n%s\n\ngot:\n%s", tt.args, tt.want, got)
		}
		assertNoHelpFooter(t, got)
	}
}

func joinHelpLines(lines ...string) string {
	return strings.Join(lines, "\n")
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

func TestSourcesCommandIsRemoved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"help", "sources"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), `unknown command "sources"`) {
		t.Fatalf("err = %v", err)
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
