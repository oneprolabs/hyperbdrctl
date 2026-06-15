package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"hyperbdr-client/catalog"
)

func TestTargetSupportsTableOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "supports"}, &out, &errOut); err != nil {
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
			t.Fatalf("target supports output missing %q: %q", want, text)
		}
	}
}

func TestTargetSupportsTableUsesLocalizedName(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "target", "supports"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"== 块存储支持列表 ==", "== 对象存储支持列表 ==", "-------------------------", "阿里云"} {
		if !strings.Contains(text, want) {
			t.Fatalf("zh target supports output missing %q: %q", want, text)
		}
	}
}

func TestTargetSupportsJSONOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--output", "json", "target", "supports"}, &out, &errOut); err != nil {
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

func TestTargetHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"\nCommands:\n",
		"supports",
		"cloud-sync-gateway",
		"oss",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("target help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("target help should not include %q: %q", unwanted, text)
		}
	}
	flags := strings.Index(text, "\nFlags:\n")
	commands := strings.Index(text, "\nCommands:\n")
	notes := strings.Index(text, "Usage Notes:")
	if flags < 0 || commands < 0 || notes < 0 || !(flags < commands && commands < notes) {
		t.Fatalf("target help order mismatch: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestTargetSupportsHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "supports", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"Usage:", "\nFlags:\n", "Usage Notes:", "does not call the remote API"} {
		if !strings.Contains(text, want) {
			t.Fatalf("target supports help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "\nCommands:\n") {
		t.Fatalf("target supports help should not include commands section: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestTargetAccountHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "account", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"list",
		"detail",
		"wait",
		"fetch-block-resources",
		"fetch-oss-resources",
		"create",
		"create-block",
		"create-oss",
		"delete",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("target account help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("target account help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTargetAccountLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"target", "account", "list", "--help"},
			want: []string{"Usage Notes:", "--storage-type", "hyperbdrctl target account list", "hyperbdrctl target account detail --id <account_id>"},
		},
		{
			args: []string{"target", "account", "detail", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl target account detail --id <account_id>"},
		},
		{
			args: []string{"target", "account", "wait", "--help"},
			want: []string{"Usage Notes:", "--id", "--interval-seconds", "--timeout-seconds"},
		},
		{
			args: []string{"target", "account", "create", "--help"},
			want: []string{"Usage Notes:", "--preview-request", "target account create-block --help"},
		},
		{
			args: []string{"target", "account", "delete", "--help"},
			want: []string{"Usage Notes:", "--force", "hyperbdrctl target account delete --id <account_id>"},
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
		if len(tt.args) >= 3 && tt.args[0] == "target" && tt.args[1] == "oss" && tt.args[2] == "create" && strings.Contains(text, "--cloud-type") {
			t.Fatalf("target oss create help should not expose --cloud-type: %q", text)
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

func TestUnifiedFlagDescriptionsInChineseHelp(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "target", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"输出请求调试日志",
		"显示语言，可选 en / zh_cn，默认值 en",
		"输出格式，可选 table / json，默认值 table",
		"显示帮助信息",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("target help missing %q: %q", want, text)
		}
	}

	out.Reset()
	errOut.Reset()
	if err := Execute([]string{"--lang", "zh_cn", "target", "account", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "输出请求体，但不发送请求") {
		t.Fatalf("target account create help missing unified preview-request text: %q", out.String())
	}
}

func TestTargetAccountFetchResourcesHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"target", "account", "fetch-block-resources", "--help"},
			want: []string{"\nCommands:\n", "aliyun", "openstack", "Usage Notes:"},
		},
		{
			args: []string{"target", "account", "fetch-oss-resources", "--help"},
			want: []string{"\nCommands:\n", "aliyun", "openstack", "Usage Notes:"},
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

func TestTargetAccountFetchResourcesProviderHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"target", "account", "fetch-block-resources", "aliyun", "--help"},
			want: []string{"Usage Notes:", "--cloud-auth-type", "--fetch-res", "fetch-block-resources aliyun", "--access-id + --access-secret => aksk"},
		},
		{
			args: []string{"target", "account", "fetch-block-resources", "openstack", "--help"},
			want: []string{"Usage Notes:", "--auth-url", "--username", "--user-domain-id", "fetch-block-resources openstack", "regions,projects"},
		},
		{
			args: []string{"target", "account", "fetch-oss-resources", "aliyun", "--help"},
			want: []string{"Usage Notes:", "--cloud-auth-type", "--fetch-res", "fetch-oss-resources aliyun", "--access-id + --access-secret => aksk"},
		},
		{
			args: []string{"target", "account", "fetch-oss-resources", "openstack", "--help"},
			want: []string{"Usage Notes:", "--auth-url", "--username", "--user-domain-id", "fetch-oss-resources openstack", "--output json", "optional"},
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
			t.Fatalf("provider help should not include commands section args=%v: %q", tt.args, text)
		}
		if strings.Contains(text, "--cloud-type") || strings.Contains(text, "--storage-type") {
			t.Fatalf("provider help should not expose legacy flags args=%v: %q", tt.args, text)
		}
		assertNoHelpFooter(t, text)
	}
}

func TestTargetAccountCreateProviderHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"target", "account", "create-block", "--help"},
			want: []string{"\nCommands:\n", "aliyun", "openstack", "Usage Notes:"},
		},
		{
			args: []string{"target", "account", "create-oss", "--help"},
			want: []string{"\nCommands:\n", "aliyun", "openstack", "Usage Notes:"},
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
	if err := Execute([]string{"target", "cloud-sync-gateway", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"list",
		"detail",
		"wait",
		"resources",
		"subnet-config",
		"create",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("target cloud-sync-gateway help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("target cloud-sync-gateway help should not include %q: %q", unwanted, text)
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
			args: []string{"target", "cloud-sync-gateway", "list", "--help"},
			want: []string{"Usage Notes:", "default `HyperGate`", "--cloud-account-id", "hyperbdrctl target cloud-sync-gateway detail --id <storage_id>"},
		},
		{
			args: []string{"target", "cloud-sync-gateway", "detail", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl target cloud-sync-gateway detail --id <storage_id>"},
		},
		{
			args: []string{"target", "cloud-sync-gateway", "wait", "--help"},
			want: []string{"Usage Notes:", "--id", "--interval-seconds", "--timeout-seconds"},
		},
		{
			args: []string{"target", "cloud-sync-gateway", "resources", "--help"},
			want: []string{"Usage Notes:", "--fetch-res", "--cloud-account-id <account_id>", "--output json"},
		},
		{
			args: []string{"target", "cloud-sync-gateway", "subnet-config", "--help"},
			want: []string{"Usage Notes:", "--network-id", "--cloud-type"},
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

func TestTargetCloudSyncGatewayCreateHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "cloud-sync-gateway", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"aliyun",
		"openstack",
		"huawei",
		"Usage Notes:",
		"hyperbdrctl target account list",
		"hyperbdrctl target cloud-sync-gateway resources --cloud-account-id <account_id> --output json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("target cloud-sync-gateway create help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("target cloud-sync-gateway create help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTargetCloudSyncGatewayCreateOpenStackHelpUsesResourceCommandFlow(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "cloud-sync-gateway", "create", "openstack", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage Notes:",
		"Parameter Sources:",
		"target cloud-sync-gateway resources",
		"--cloud-account-id <account_id>",
		"--output json",
		"--boot-loader-image-id <windows_image_id>",
		"target cloud-sync-gateway wait --id <storage_id>",
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

func TestTargetOSSHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "oss", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"\nCommands:\n",
		"list",
		"detail",
		"wait",
		"buckets",
		"create",
		"Usage Notes:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("target oss help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("target oss help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTargetOSSLeafHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"target", "oss", "list", "--help"},
			want: []string{"Usage Notes:", "hyperbdrctl target oss list", "default objectstorage"},
		},
		{
			args: []string{"target", "oss", "detail", "--help"},
			want: []string{"Usage Notes:", "--id", "hyperbdrctl target oss detail --id <storage_id>"},
		},
		{
			args: []string{"target", "oss", "wait", "--help"},
			want: []string{"Usage Notes:", "--id", "--interval-seconds", "--timeout-seconds"},
		},
		{
			args: []string{"target", "oss", "buckets", "--help"},
			want: []string{"Usage Notes:", "--auth-url", "--bucket-lookup"},
		},
		{
			args: []string{"target", "oss", "create", "--help"},
			want: []string{"Usage Notes:", "--display-name", "<cloud-type>-<region-id>", "existing / new"},
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

func TestLegacyObjectStoragesAndTargetInfoAreRemoved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"object-storages", "list"}, want: `unknown command "object-storages"`},
		{args: []string{"cloud-accounts", "list"}, want: `unknown command "cloud-accounts"`},
		{args: []string{"block-storages", "list"}, want: `unknown command "block-storages"`},
		{args: []string{"target", "info"}, want: `unknown target command "info"`},
		{args: []string{"target", "account", "fetch-regions"}, want: `unknown target account command "fetch-regions"`},
		{args: []string{"target", "cloud-sync-gateway", "transition-images"}, want: `unknown target cloud-sync-gateway command "transition-images"`},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		err := Execute(tt.args, &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("args=%v err=%v want substring %q", tt.args, err, tt.want)
		}
	}
}

func TestTargetAccountLegacyFetchResourcesShowsMigration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"target", "account", "fetch-resources", "--cloud-type", "aliyun_bs"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "use target account fetch-block-resources <provider> or target account fetch-oss-resources <provider>") {
		t.Fatalf("err = %v", err)
	}
}
