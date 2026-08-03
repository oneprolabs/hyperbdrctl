package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hyperbdr-client/catalog"
)

func TestCloudResourceCatalogTableOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-resource", "catalog"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"== Block Storage Clouds ==",
		"== Object Storage Clouds ==",
		"-------------------------",
		"Provider",
		"Name",
		"aliyun",
		"Alibaba Cloud",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloud-resource catalog output missing %q: %q", want, text)
		}
	}
}

func TestCloudResourceCatalogTableUsesLocalizedName(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "cloud-resource", "catalog"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"== 块存储支持列表 ==", "== 对象存储支持列表 ==", "-------------------------", "阿里云"} {
		if !strings.Contains(text, want) {
			t.Fatalf("zh cloud-resource catalog output missing %q: %q", want, text)
		}
	}
}

func TestCloudResourceCatalogJSONOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--output", "json", "cloud-resource", "catalog"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	var raw map[string][]map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["block_clouds"][0]["cloud_type"]; !ok {
		t.Fatalf("block_clouds item should expose cloud_type: %+v", raw["block_clouds"][0])
	}
	if _, ok := raw["object_clouds"][0]["name_zh_cn"]; !ok {
		t.Fatalf("object_clouds item should expose name_zh_cn: %+v", raw["object_clouds"][0])
	}

	var got map[string][]catalog.CloudEntry
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got["block_clouds"]) != len(catalog.BlockClouds) {
		t.Fatalf("block_clouds len = %d, want %d", len(got["block_clouds"]), len(catalog.BlockClouds))
	}
	if len(got["object_clouds"]) != len(catalog.ObjectClouds) {
		t.Fatalf("object_clouds len = %d, want %d", len(got["object_clouds"]), len(catalog.ObjectClouds))
	}
	if strings.Contains(out.String(), "== ") {
		t.Fatalf("json output should not contain section titles: %q", out.String())
	}
}

func TestCloudResourceCatalogHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-resource", "catalog", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"Usage: hyperbdrctl cloud-resource catalog [flags]", "\nFlags:\n", "Usage Notes:", "does not call the remote API", "hyperbdrctl cloud-resource fetch --help"} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloud-resource catalog help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "\nCommands:\n") {
		t.Fatalf("cloud-resource catalog help should not include commands section: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestCloudAccountHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-account", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"list",
		"detail",
		"wait",
		"create",
		"delete",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloud-account help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"create-block", "create-oss"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("cloud-account help should not include %q: %q", unwanted, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("cloud-account help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestCloudAccountLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"cloud-account", "list", "--help"},
			want: []string{"Usage Notes:", "--storage-type", "hyperbdrctl cloud-account list", "hyperbdrctl cloud-account detail --id <account_id>"},
		},
		{
			args: []string{"cloud-account", "detail", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl cloud-account detail --id <account_id>"},
		},
		{
			args: []string{"cloud-account", "wait", "--help"},
			want: []string{"Usage Notes:", "--id", "--interval-seconds", "--timeout-seconds"},
		},
		{
			args: []string{"cloud-account", "create", "--help"},
			want: []string{"Usage Notes:", "--preview-request", "cloud-account create --storage-type block --help"},
		},
		{
			args: []string{"cloud-account", "delete", "--help"},
			want: []string{"Usage Notes:", "--force", "hyperbdrctl cloud-account delete --id <account_id>"},
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
		if strings.Contains(text, "--vertical") {
			t.Fatalf("args=%v help should hide --vertical: %q", tt.args, text)
		}
		if strings.Contains(text, "\nCommands:\n") {
			t.Fatalf("leaf help should not include commands section args=%v: %q", tt.args, text)
		}
		for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("leaf help should not include %q args=%v: %q", unwanted, tt.args, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestCloudAccountDeleteHelpOmitsStorageType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-account", "delete", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "--storage-type") {
		t.Fatalf("cloud-account delete help should not include --storage-type: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestCloudAccountDeleteDoesNotAcceptStorageType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"cloud-account", "delete", "--id", "account-1", "--storage-type", "objectstorage"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountDeleteOmitsStorageTypeFromRequestBody(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"deleted": true},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "cloud-account", "delete", "--id", "account-1", "--force"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/cloud_accounts/account-1?force=true" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["id"] != "account-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	if _, ok := gotBody["storage_type"]; ok {
		t.Fatalf("body should omit storage_type: %+v", gotBody)
	}
}

func TestTargetAccountListSupportsVerticalOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/getCloudAccounts" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_accounts": []map[string]interface{}{
					{
						"id":                  "account-1",
						"uuid":                "uuid-1",
						"name":                "demo-object",
						"display_username":    "demo-user",
						"cloud_type":          "openstack",
						"storage_type":        "objectstorage",
						"status":              "available",
						"display_status":      "Available",
						"display_task_status": "Idle",
						"created_at":          "2026-06-18 10:00:00",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "cloud-account", "list", "--storage-type", "object", "-G"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "storage_type=objectstorage") {
		t.Fatalf("query=%q", gotQuery)
	}

	text := out.String()
	for _, want := range []string{
		"*************************** 1. row ***************************",
		"ID",
		"account-1",
		"Display Username",
		"demo-user",
		"Cloud Type",
		"openstack",
		"Status",
		"Available",
		"Task Status",
		"Idle",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("vertical output missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "UUID") {
		t.Fatalf("vertical output should not include UUID: %q", text)
	}
	if strings.Contains(text, "\navailable\n") {
		t.Fatalf("vertical output should use display_status instead of raw status: %q", text)
	}
}

func TestTargetAccountListMapsBlockStorageFilterToHyperGate(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/getCloudAccounts" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_accounts": []map[string]interface{}{},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "cloud-account", "list", "--storage-type", "block"), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "storage_type=HyperGate") {
		t.Fatalf("query=%q", gotQuery)
	}
}

func TestTargetAccountListVerticalDoesNotOverrideJSON(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_accounts": []map[string]interface{}{
					{"id": "account-1"},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "cloud-account", "list", "-G"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(text, "\"cloud_accounts\"") {
		t.Fatalf("expected json output: %q", text)
	}
	if strings.Contains(text, "1. row") {
		t.Fatalf("vertical output should not override json: %q", text)
	}
}

func TestUnifiedFlagDescriptionsInChineseHelp(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "cloud-resource", "catalog", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"输出请求调试日志",
		"显示语言，可选值 en / zh_cn，默认值 en",
		"输出格式，可选值 table / json，默认值 table",
		"显示帮助信息",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloud-resource catalog help missing %q: %q", want, text)
		}
	}

	out.Reset()
	errOut.Reset()
	if err := Execute([]string{"--lang", "zh_cn", "cloud-account", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "输出请求体，但不发送请求") {
		t.Fatalf("cloud-account create help missing unified preview-request text: %q", out.String())
	}
}

func TestCloudAccountCreateStorageHelpShowsProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"cloud-account", "create", "--storage-type", "block", "--help"},
			want: []string{"Usage Notes:", "Providers:", "aliyun", "openstack"},
		},
		{
			args: []string{"cloud-account", "create", "--storage-type", "object", "--help"},
			want: []string{"Usage Notes:", "Providers:", "aliyun", "huawei", "openstack"},
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
		assertNoHelpFooter(t, text)
	}
}

func TestTargetCloudSyncGatewayHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"list",
		"detail",
		"wait",
		"create",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloud-sync-gateway help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("cloud-sync-gateway help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTargetCloudSyncGatewayLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"cloud-sync-gateway", "list", "--help"},
			want: []string{"Usage Notes:", "--type string", "default HyperGate", "--cloud-account-id", "hyperbdrctl cloud-sync-gateway list --cloud-account-id <account_id>", "hyperbdrctl cloud-sync-gateway detail --id <storage_id>"},
		},
		{
			args: []string{"cloud-sync-gateway", "detail", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl cloud-sync-gateway detail --id <storage_id>"},
		},
		{
			args: []string{"cloud-sync-gateway", "wait", "--help"},
			want: []string{"Usage Notes:", "--id", "--interval-seconds", "--timeout-seconds"},
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
		for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("leaf help should not include %q args=%v: %q", unwanted, tt.args, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestTargetCloudSyncGatewayCreateHelpUsesAccountGuideLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"--cloud-account-id",
		"--preview-request",
		"Usage Notes:",
		"hyperbdrctl cloud-account list --storage-type block",
		"hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --help",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloud-sync-gateway create help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"--cloud-type", "Providers:", "aliyun", "openstack", "huawei", "\nCommands:\n", "\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("cloud-sync-gateway create help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTargetCloudSyncGatewayCreateOpenStackHelpUsesResourceCommandFlow(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "create", "--cloud-type", "openstack", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage Notes:",
		"Parameter Sources:",
		"cloud-resource fetch",
		"--cloud-account-id <account_id>",
		"--output json",
		"--boot-loader-image-id <windows_image_id>",
		"cloud-sync-gateway wait --id <storage_id>",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("openstack create help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "get_cloud_info") {
		t.Fatalf("openstack create help should describe CLI resource flow instead of raw API details: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestOSSHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"oss", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"list",
		"detail",
		"wait",
		"buckets",
		"catalog",
		"create",
		"delete",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("oss help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("oss help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestOSSLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
		omit []string
	}{
		{
			args: []string{"oss", "list", "--help"},
			want: []string{"Usage Notes:", "hyperbdrctl oss list", "`object` type"},
			omit: []string{"--type"},
		},
		{
			args: []string{"oss", "detail", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl oss detail --id <storage_id>"},
		},
		{
			args: []string{"oss", "wait", "--help"},
			want: []string{"Usage Notes:", "--id", "--interval-seconds", "--timeout-seconds"},
		},
		{
			args: []string{"oss", "catalog", "--help"},
			want: []string{"Usage Notes:", "--provider", "hyperbdrctl oss catalog --provider aliyun"},
		},
		{
			args: []string{"oss", "delete", "--help"},
			want: []string{"Usage Notes:", "--force", "hyperbdrctl oss delete --id <storage_id>"},
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
		for _, unwanted := range tt.omit {
			if strings.Contains(text, unwanted) {
				t.Fatalf("args=%v help should not expose %q: %q", tt.args, unwanted, text)
			}
		}
		if strings.Contains(text, "\nCommands:\n") {
			t.Fatalf("leaf help should not include commands section args=%v: %q", tt.args, text)
		}
		for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("leaf help should not include %q args=%v: %q", unwanted, tt.args, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestOSSCreateHelpProfiles(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name string
		args []string
		want []string
		omit []string
	}{
		{
			name: "default custom",
			args: []string{"oss", "create", "--help"},
			want: []string{"Usage Notes:", "creates custom object storage", "--auth-url", "--bucket-mode", "--bucket-name"},
			omit: []string{"--cloud-type-select"},
		},
		{
			name: "provider catalog",
			args: []string{"oss", "create", "--provider", "aliyun", "--help"},
			want: []string{"Usage Notes:", "hyperbdrctl oss catalog --provider aliyun", "--region-id", "--public-endpoint", "--internal-endpoint"},
			omit: []string{"--cloud-type-select"},
		},
		{
			name: "custom provider alias",
			args: []string{"oss", "create", "--provider", "custom", "--help"},
			want: []string{"Usage Notes:", "creates custom object storage", "--auth-url", "--region-id"},
			omit: []string{"--cloud-type-select", "localized provider name and region name"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
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
			for _, unwanted := range tt.omit {
				if strings.Contains(text, unwanted) {
					t.Fatalf("args=%v help should not expose %q: %q", tt.args, unwanted, text)
				}
			}
			assertNoHelpFooter(t, text)
		})
	}
}

func TestOSSBucketsHelpProfiles(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "default custom",
			args: []string{"oss", "buckets", "--help"},
			want: []string{"Usage Notes:", "--auth-url", "--bucket-lookup", "target bucket"},
		},
		{
			name: "provider catalog",
			args: []string{"oss", "buckets", "--provider", "aliyun", "--help"},
			want: []string{"Usage Notes:", "hyperbdrctl oss catalog --provider aliyun", "--region-id", "--auth-url"},
		},
		{
			name: "custom provider alias",
			args: []string{"oss", "buckets", "--provider", "custom", "--help"},
			want: []string{"Usage Notes:", "--auth-url", "target bucket"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
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
			assertNoHelpFooter(t, text)
		})
	}
}

func TestOSSCreateHelpUsesCustomModeInChinese(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "oss", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"使用说明", "默认按自定义对象存储方式创建", "--provider <provider_id>", "--region-id string"} {
		if !strings.Contains(text, want) {
			t.Fatalf("oss create zh help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"<cloud-type>-<region-id>", "基于 `--auth-url` 的兼容回退行为", "区域 ID（必须）", "其它平台"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("oss create zh help should not expose %q: %q", unwanted, text)
		}
	}
}

func TestOSSArchivedHelpCopyIsLocalized(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name string
		args []string
		want []string
		omit []string
	}{
		{
			name: "zh root",
			args: []string{"--lang", "zh_cn", "oss", "--help"},
			want: []string{"目标对象存储管理", "buckets                  验证存储桶可见性", "catalog                  查看云厂商和区域", "建议按下面顺序操作"},
		},
		{
			name: "zh buckets",
			args: []string{"--lang", "zh_cn", "oss", "buckets", "--help"},
			want: []string{"验证存储桶可见性", "云厂商 ID", "存储桶寻址方式，可选值 dns /", "当前凭据是否能访问目标存储桶"},
			omit: []string{"默认值 true"},
		},
		{
			name: "zh catalog",
			args: []string{"--lang", "zh_cn", "oss", "catalog", "--help"},
			want: []string{"查看云厂商和区域", "查看当前平台支持的云厂商和区域信息", "--provider string   云厂商 ID"},
		},
		{
			name: "zh create",
			args: []string{"--lang", "zh_cn", "oss", "create", "--help"},
			want: []string{"创建目标对象存储", "存储桶模式，可选值 existing /", "默认按自定义对象存储方式创建", "--preview-request"},
			omit: []string{"默认值 true"},
		},
		{
			name: "zh delete",
			args: []string{"--lang", "zh_cn", "oss", "delete", "--help"},
			want: []string{"删除目标对象存储", "CLI 会先检查该对象存储是否仍有关联主机", "目标对象存储 ID 通常来自"},
			omit: []string{"默认值 false"},
		},
		{
			name: "zh detail",
			args: []string{"--lang", "zh_cn", "oss", "detail", "--help"},
			want: []string{"查看目标对象存储详情", "根据目标对象存储 ID 读取单个目标对象存储详情", "hyperbdrctl oss create"},
		},
		{
			name: "zh list",
			args: []string{"--lang", "zh_cn", "oss", "list", "--help"},
			want: []string{"列出目标对象存储", "固定查询 `object` 类型", "--page-size int"},
			omit: []string{"固定查询 `objectstorage` 类型"},
		},
		{
			name: "zh wait",
			args: []string{"--lang", "zh_cn", "oss", "wait", "--help"},
			want: []string{"等待目标对象存储创建完成", "对象存储创建请求提交后", "超时只会停止本地轮询"},
		},
		{
			name: "en root",
			args: []string{"--lang", "en", "oss", "--help"},
			want: []string{"Target object storage management", "buckets                  Verify bucket visibility", "catalog                  View cloud providers and regions", "Follow this recommended sequence"},
		},
		{
			name: "en buckets",
			args: []string{"--lang", "en", "oss", "buckets", "--help"},
			want: []string{"Verify bucket visibility", "Cloud provider ID", "Bucket lookup style, allowed values dns / path", "current credentials can access the target bucket"},
			omit: []string{"default true"},
		},
		{
			name: "en catalog",
			args: []string{"--lang", "en", "oss", "catalog", "--help"},
			want: []string{"View cloud providers and regions", "supported by the current platform", "--provider string   Cloud provider ID"},
		},
		{
			name: "en create",
			args: []string{"--lang", "en", "oss", "create", "--help"},
			want: []string{"Create target object storage", "Bucket mode, allowed values existing / new", "creates custom object storage", "--preview-request"},
			omit: []string{"default true"},
		},
		{
			name: "en delete",
			args: []string{"--lang", "en", "oss", "delete", "--help"},
			want: []string{"Delete target object storage", "still has associated hosts", "target object storage ID usually comes from"},
			omit: []string{"default false"},
		},
		{
			name: "en detail",
			args: []string{"--lang", "en", "oss", "detail", "--help"},
			want: []string{"Show target object storage detail", "Read the details of one target object storage", "hyperbdrctl oss create"},
		},
		{
			name: "en list",
			args: []string{"--lang", "en", "oss", "list", "--help"},
			want: []string{"List target object storages", "always queries the `object` type", "--page-size int"},
			omit: []string{"always queries the `objectstorage` type"},
		},
		{
			name: "en wait",
			args: []string{"--lang", "en", "oss", "wait", "--help"},
			want: []string{"Wait for target object storage creation", "Continue polling for the final status", "does not cancel the backend create task"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
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
			for _, unwanted := range tt.omit {
				if strings.Contains(text, unwanted) {
					t.Fatalf("args=%v help should not contain %q: %q", tt.args, unwanted, text)
				}
			}
			assertNoHelpFooter(t, text)
		})
	}
}

func TestLegacyCommandGroupsAreRemoved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"object-storages", "list"}, want: `unknown command "object-storages"`},
		{args: []string{"cloud-accounts", "list"}, want: `unknown command "cloud-accounts"`},
		{args: []string{"block-storages", "list"}, want: `unknown command "block-storages"`},
		{args: []string{"target", "info"}, want: `unknown command "target"`},
		{args: []string{"target", "account", "fetch-regions"}, want: `unknown command "target"`},
		{args: []string{"target", "cloud-sync-gateway", "transition-images"}, want: `unknown command "target"`},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		err := Execute(tt.args, &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("args=%v err=%v want substring %q", tt.args, err, tt.want)
		}
	}
}

func TestRemovedTargetCommandsReturnUnknownCommand(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"target"},
		{"target", "supports"},
		{"target", "account"},
		{"target", "account", "list"},
		{"target", "account", "detail"},
		{"target", "account", "wait"},
		{"target", "account", "create"},
		{"target", "account", "delete"},
		{"target", "account", "create-block"},
		{"target", "account", "create-block", "aliyun"},
		{"target", "account", "create-oss"},
		{"target", "account", "create-oss", "aliyun"},
		{"target", "account", "fetch-resources"},
		{"target", "account", "fetch-block-resources"},
		{"target", "account", "fetch-oss-resources"},
		{"target", "oss"},
		{"target", "oss", "list"},
		{"target", "oss", "detail"},
		{"target", "oss", "wait"},
		{"target", "oss", "buckets"},
		{"target", "oss", "catalog"},
		{"target", "oss", "create"},
		{"target", "oss", "delete"},
		{"target", "cloud-sync-gateway"},
		{"target", "cloud-sync-gateway", "list"},
		{"target", "cloud-sync-gateway", "detail"},
		{"target", "cloud-sync-gateway", "delete"},
		{"target", "cloud-sync-gateway", "wait"},
		{"target", "cloud-sync-gateway", "create"},
		{"target", "cloud-sync-gateway", "create", "aliyun"},
		{"target", "cloud-sync-gateway", "resources"},
		{"target", "cloud-sync-gateway", "subnet-config"},
	}

	for _, args := range cases {
		var out, errOut bytes.Buffer
		err := Execute(args, &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), "unknown") {
			t.Fatalf("args=%v err=%v", args, err)
		}
	}
}

func TestRemovedTargetHelpReturnsUnknownCommand(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	for _, args := range [][]string{
		{"help", "target"},
		{"help", "target", "supports"},
		{"help", "target", "oss"},
	} {
		var out, errOut bytes.Buffer
		err := Execute(args, &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), "unknown") || !strings.Contains(err.Error(), "target") {
			t.Fatalf("args=%v err=%v", args, err)
		}
	}
}
