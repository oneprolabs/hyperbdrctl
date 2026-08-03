package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"hyperbdr-client/internal/config"
	"hyperbdr-client/internal/i18n"
)

func assertNoHelpFooter(t *testing.T, text string) {
	t.Helper()
	for _, unwanted := range []string{
		`Use "`,
		`for more information.`,
		`使用 "`,
		`查看更多说明`,
	} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include footer %q: %q", unwanted, text)
		}
	}
}

func TestExtractGlobalFlagsAnywhere(t *testing.T) {
	args, flags, err := extractGlobalFlags([]string{"config", "set", "--host", "https://x", "--password=secret", "--insecure", "--lang", "zh_cn", "-G"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(args, " ") != "config set --host https://x --password=secret --insecure" {
		t.Fatalf("args = %v", args)
	}
	if flags.Host != "" || flags.Password != "" || flags.Lang != "zh_cn" || flags.Insecure || flags.InsecureSet || !flags.Vertical {
		t.Fatalf("flags = %+v", flags)
	}
}

func TestUsageIsLocalized(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(out.String(), "用法") {
		t.Fatalf("usage = %q", out.String())
	}
	if !strings.Contains(out.String(), "license") {
		t.Fatalf("usage missing license = %q", out.String())
	}
	if !strings.Contains(out.String(), "--version") {
		t.Fatalf("usage missing version flag = %q", out.String())
	}
	if strings.Contains(out.String(), "--insecure") {
		t.Fatalf("usage should not include non-global insecure flag = %q", out.String())
	}
	for _, visible := range []string{"production-site", "agent", "sync-proxy"} {
		if !strings.Contains(out.String(), visible) {
			t.Fatalf("usage missing %s = %q", visible, out.String())
		}
	}
	for _, hidden := range []string{"\n  api", "\n  batch-boot-config", "\n  boot-config-wizard", "\n  tasks", "\n  upgrade", "\n  source "} {
		if strings.Contains(text, hidden) {
			t.Fatalf("usage should hide %q = %q", hidden, text)
		}
	}
	if strings.Contains(text, "boot-config              Deprecated top-level boot configuration command group") {
		t.Fatalf("usage should hide deprecated boot-config alias = %q", text)
	}
	if strings.Contains(text, "boot-config-cli          Deprecated compatibility apply command") {
		t.Fatalf("usage should hide deprecated boot-config-cli alias = %q", text)
	}
	if strings.Contains(text, "sources") {
		t.Fatalf("usage should hide deprecated sources alias = %q", text)
	}
	if strings.Contains(text, "licenses") {
		t.Fatalf("usage should hide deprecated licenses alias = %q", text)
	}
}

func TestVersionDoesNotRequireConfiguration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--version"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "hyperbdrctl version dev") {
		t.Fatalf("version output = %q", got)
	}
}

func TestVersionJSONDoesNotRequireConfiguration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--output", "json", "--version"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("version json = %q: %v", out.String(), err)
	}
	for _, key := range []string{"version", "commit", "build_date", "go_version", "platform"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("version json missing %q: %v", key, got)
		}
	}
	if got["version"] != "dev" {
		t.Fatalf("version = %v", got["version"])
	}
}

func TestTopLevelBootConfigIsRouted(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config", "apply"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestTopLevelBootConfigWizardIsRemoved(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config-wizard"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), `unknown command "boot-config-wizard"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestRemovedHiddenTopLevelCommandsAreUnavailable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	for _, args := range [][]string{
		{"api"},
		{"boot-config-cli"},
		{"batch-boot-config"},
		{"tasks"},
		{"upgrade"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid", args...), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), "unknown") || !strings.Contains(err.Error(), args[0]) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestRemovedTopLevelCommandHelpIsUnavailable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	for _, name := range []string{"api", "boot-config-cli", "batch-boot-config", "tasks", "upgrade", "boot-config-wizard"} {
		t.Run(name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute([]string{"help", name}, &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), "unknown") || !strings.Contains(err.Error(), name) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestHostHelpShowsModernGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "host"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"Usage:", "\nFlags:\n", "\nCommands:\n", "Usage Notes:", "clean", "wait", "hyperbdrctl boot-config get --id <host_id>"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	for _, unwanted := range []string{"\n  boot-config", "\nExamples:\n", "\nNotes:\n", "\nGlobal Flags:\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, got)
		}
	}
	assertNoHelpFooter(t, got)
}

func TestHostCleanHelpUsesModernLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "host", "clean"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"Usage:", "\nFlags:\n", "--id", "Usage Notes:", "host wait --id <host_id> --operation clean"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	for _, unwanted := range []string{"\nCommands:\n", "\nExamples:\n", "\nNotes:\n", "\nGlobal Flags:\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, got)
		}
	}
	if strings.Contains(got, "--file string") {
		t.Fatalf("help should hide file flag from parameter block: %q", got)
	}
	assertNoHelpFooter(t, got)
}

func TestHostFileDrivenCommandHelpHidesFileFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	for _, args := range [][]string{
		{"help", "host", "sync"},
		{"help", "host", "boot"},
		{"help", "host", "deregister"},
	} {
		t.Run(strings.Join(args[1:], " "), func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(args, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			if strings.Contains(got, "--file string") {
				t.Fatalf("help should hide file flag from parameter block: %q", got)
			}
			assertNoHelpFooter(t, got)
		})
	}
}

func TestHostWaitHelpShowsCleanOperationChoice(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "host", "wait"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"--operation", "allowed values sync / boot /", "clean / deregister", "default 60", "default 3600"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	if strings.Contains(got, "cleanup-validation-host") {
		t.Fatalf("help should not mention legacy operation: %q", got)
	}
	assertNoHelpFooter(t, got)
}

func TestHostHelpMatchesArchivedChineseGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	tests := []struct {
		args []string
		want []string
	}{
		{[]string{"host", "--help"}, []string{"管理主机生命周期，包括查询、注册、同步、启动、清理和注销。", "hyperbdrctl boot-config get --id <host_id>"}},
		{[]string{"host", "list", "--help"}, []string{"页码，默认值 1", "每页数量，默认值 10", "host list --status sync_snapshot_done --boot-status not_boot", "如需保留原始 API 字段名，便于脚本处理，可附加下面的参数："}},
		{[]string{"host", "detail", "--help"}, []string{"资源 ID（必须）", "host detail --id <host_id>", "如需查看原始 API 字段，便于后续脚本处理，可附加下面的参数："}},
		{[]string{"host", "snapshots", "--help"}, []string{"资源 ID（必须）", "只有在需要同步细节字段时，才附加下面的参数", "--sync-detail"}},
		{[]string{"host", "sync", "--help"}, []string{"参数来源：", "--id 和 --ids 至少传入一个。", "host sync --ids <host_id_1,host_id_2>"}},
		{[]string{"host", "register", "--help"}, []string{"hyperbdrctl production-site vm-list", "host register --vm-ids <vm_id_1,vm_id_2>"}},
		{[]string{"host", "boot", "--help"}, []string{"--snapshot-id string", "hyperbdrctl host snapshots --id <host_id>", "host wait --id <host_id> --operation boot"}},
		{[]string{"host", "clean", "--help"}, []string{"删除一个或多个主机的验证启动实例。", "--id 和 --ids 至少传入一个。"}},
		{[]string{"host", "deregister", "--help"}, []string{"--id 和 --ids 至少传入一个。", "只有在确认目标主机无误后，才附加下面的参数"}},
		{[]string{"host", "wait", "--help"}, []string{"等待的操作，可选值 sync / boot / clean /", "deregister", "--id 和 --ids 至少传入一个。", "--include-steps"}},
	}

	for _, tt := range tests {
		var out, errOut bytes.Buffer
		args := append([]string{"--lang", "zh_cn"}, tt.args...)
		if err := Execute(args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		text := out.String()
		for _, want := range tt.want {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v missing %q: %q", args, want, text)
			}
		}
		for _, unwanted := range []string{"默认值 false", "--file string"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("args=%v should not contain %q: %q", args, unwanted, text)
			}
		}
	}
}

func TestHostListAndDetailHelpUseArchivedEnglishGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"host", "list", "--help"}, "When another tool needs raw API field names, add the following flag:"},
		{[]string{"host", "detail", "--help"}, "If you need raw API field names for follow-up scripting, add the following flag:"},
	}
	for _, tt := range tests {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}
		if !strings.Contains(out.String(), tt.want) {
			t.Fatalf("args=%v missing %q: %q", tt.args, tt.want, out.String())
		}
	}
}

func TestFindListItemsFallback(t *testing.T) {
	items, ok := findListItems(map[string]interface{}{
		"records": []interface{}{
			map[string]interface{}{"id": "1"},
		},
	}, "host")
	if !ok || len(items) != 1 {
		t.Fatalf("items=%v ok=%v", items, ok)
	}
}

func TestConfigGetMasksPassword(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	path, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(path, config.Config{Host: "https://x", Username: "u", Password: "secret"}); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := Execute([]string{"config", "get"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "secret") || !strings.Contains(out.String(), "********") {
		t.Fatalf("masked output = %q", out.String())
	}

	out.Reset()
	if err := Execute([]string{"config", "get", "--show-secret"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "secret") {
		t.Fatalf("show secret output = %q", out.String())
	}
}

func TestConfigHelpUsesConfiguredLang(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	path, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(path, config.Config{Lang: "zh_cn"}); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := Execute([]string{"config", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), i18n.New("zh_cn").T("help.section_usage")) {
		t.Fatalf("help = %q", out.String())
	}
	assertNoHelpFooter(t, out.String())
}

func TestConfigHelpShowsGuidedSections(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"config", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage Notes:",
		"Commands:",
		"get                      Show the saved local configuration",
		"set                      Validate and save local CLI configuration",
		"hyperbdrctl config get",
		"hyperbdrctl config set \\",
		"set environment variables using the syntax for your current shell",
		"command-line flags > environment variables > config file > defaults",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{
		"\nExamples:\n",
		"\nNotes:\n",
		"\nWorkflow:\n",
		"\nRelated Commands:\n",
		"\nNext Steps:\n",
		"\nGlobal Flags:\n",
	} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	usage := strings.Index(text, "Usage:")
	flags := strings.Index(text, "\nFlags:\n")
	commands := strings.Index(text, "\nCommands:\n")
	notes := strings.Index(text, "Usage Notes:")
	if usage < 0 || flags < 0 || commands < 0 || notes < 0 || !(usage < flags && flags < commands && commands < notes) {
		t.Fatalf("help order mismatch: %q", text)
	}
	if strings.Contains(text, "HYPERBDR_HOST") {
		t.Fatalf("config group help should defer environment variable details to config set: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestConfigGetHelpShowsSafeReadGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"config", "get", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage Notes:",
		"--show-secret",
		"does not call the remote API",
		"hyperbdrctl --output json config get",
		"hyperbdrctl config set \\",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nRelated Commands:\n", "\nNext Steps:\n", "\nWorkflow:\n", "\nMinimum Flags:\n", "\nGlobal Flags:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should stay lightweight: %q", text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestConfigSetHelpShowsGuidedSectionsInStyleOrder(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"config", "set", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage Notes:",
		"--host",
		"--username",
		"--password",
		"--scene",
		"--insecure",
		"--lang",
		"--output",
		"required for first-time initialization",
		"allowed values dr /",
		"default dr",
		"default en",
		"default table",
		"hyperbdrctl config set \\",
		"export HYPERBDR_HOST=https://example:10443",
		"validates the login before saving",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{
		"\nMinimum Flags:\n",
		"\nAutomatic behavior:\n",
		"\nExamples:\n",
		"\nNotes:\n",
		"\nWorkflow:\n",
		"\nCommon Optional Flags:\n",
		"\nRelated Commands:\n",
		"\nNext Steps:\n",
		"\nGlobal Flags:\n",
	} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	flagsStart := strings.Index(text, "\nFlags:\n")
	notes := strings.Index(text, "Usage Notes:")
	if flagsStart < 0 || notes < 0 {
		t.Fatalf("help = %q", text)
	}
	flagsSection := text[flagsStart:notes]
	if !strings.Contains(flagsSection, "-h, --help") {
		t.Fatalf("local flags should keep help flag: %q", flagsSection)
	}
	for _, want := range []string{"--host", "--username", "--password", "--scene", "--insecure"} {
		if !strings.Contains(flagsSection, want) {
			t.Fatalf("local flags should expose %q: %q", want, flagsSection)
		}
	}
	for _, want := range []string{"--lang", "--output", "--debug"} {
		if !strings.Contains(flagsSection, want) {
			t.Fatalf("flags should expose %q: %q", want, flagsSection)
		}
	}
	if !(flagsStart < notes) {
		t.Fatalf("help order mismatch: %q", text)
	}
	if strings.Contains(flagsSection, "default false") {
		t.Fatalf("boolean flags should not show a false default: %q", flagsSection)
	}
	assertNoHelpFooter(t, text)
}

func TestConfigHelpMatchesArchivedCopyInBothLanguages(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	tests := []struct {
		name      string
		args      []string
		want      []string
		unwanted  []string
		countText string
	}{
		{
			name: "zh_cn group",
			args: []string{"--lang", "zh_cn", "config", "--help"},
			want: []string{
				"查看或更新 CLI 默认配置",
				"\n命令:\n",
				"get                      显示已保存的本地配置",
				"set                      校验并保存本地 CLI 配置",
				"如果不希望把地址或凭证写入本地配置，也可以按当前 Shell 的方式设置环境变量。",
				"完整参数、环境变量和校验规则请继续查看：",
				"--username <username>",
				"--password <password>",
			},
			unwanted: []string{"HYPERBDR_HOST"},
		},
		{
			name: "en group",
			args: []string{"--lang", "en", "config", "--help"},
			want: []string{
				"View or update CLI default configuration",
				"\nCommands:\n",
				"get                      Show the saved local configuration",
				"set                      Validate and save local CLI configuration",
				"set environment variables using the syntax for your current shell",
				"For complete flags, environment variables, and validation rules",
				"--username <username>",
				"--password <password>",
			},
			unwanted: []string{"HYPERBDR_HOST"},
		},
		{
			name: "zh_cn get",
			args: []string{"--lang", "zh_cn", "config", "get", "--help"},
			want: []string{
				"读取当前机器上已保存的本地配置，并显示配置文件路径和 token 缓存路径。",
				"查看已保存配置：",
				"如需明文查看已保存密码：",
				"把结果交给其他工具处理：",
				"--username <username>",
				"--password <password>",
			},
		},
		{
			name: "en get",
			args: []string{"--lang", "en", "config", "get", "--help"},
			want: []string{
				"Read the local configuration saved on the current machine and show the config file and token cache paths.",
				"View the saved configuration:",
				"To view the saved password in plain text:",
				"Pass the result to another tool:",
				"--username <username>",
				"--password <password>",
			},
		},
		{
			name: "zh_cn set",
			args: []string{"--lang", "zh_cn", "config", "set", "--help"},
			want: []string{
				"平台场景，可选值 dr /",
				"--insecure          测试环境跳过 TLS 证书校验",
				"--debug             输出请求调试日志",
				"首次初始化本地配置：",
				"如果命令行省略某个字段，CLI 会按下面的优先级补齐：",
				"以 Bash 为例，可以先设置环境变量，再执行保存：",
				"export HYPERBDR_USERNAME=<username>",
				"保存前会先执行登录校验；如果登录失败，不会保存新的配置。",
			},
			unwanted:  []string{"默认值 false", "配置读取优先级如下："},
			countText: "命令行参数 > 环境变量 > 配置文件 > 默认值",
		},
		{
			name: "en set",
			args: []string{"--lang", "en", "config", "set", "--help"},
			want: []string{
				"Platform scene, allowed values dr /",
				"--insecure          Skip TLS certificate verification for test environments",
				"--debug             Output request debug logs",
				"Initialize the local configuration for the first time:",
				"If a field is omitted from the command line, the CLI fills it according to the following priority order:",
				"Using Bash as an example, set the environment variables before saving:",
				"export HYPERBDR_USERNAME=<username>",
				"The CLI validates the login before saving; if login fails, the new configuration is not saved.",
			},
			unwanted:  []string{"default false", "Configuration is read in the following priority order:"},
			countText: "command-line flags > environment variables > config file > defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(tt.args, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			text := out.String()
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			for _, unwanted := range tt.unwanted {
				if strings.Contains(text, unwanted) {
					t.Fatalf("help should not include %q: %q", unwanted, text)
				}
			}
			if tt.countText != "" && strings.Count(text, tt.countText) != 1 {
				t.Fatalf("help should include %q exactly once: %q", tt.countText, text)
			}
		})
	}
}

func TestRootHelpShowsModernGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Flags:",
		"Usage Notes:",
		"Generate shell completion scripts",
		"hyperbdrctl config set \\",
		"HYPERBDR_HOST",
		"command-line flags > environment variables > config file > defaults",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, hiddenFlag := range []string{"--vertical", "--insecure"} {
		if strings.Contains(text, hiddenFlag) {
			t.Fatalf("help should hide %s from flags: %q", hiddenFlag, text)
		}
	}
	for _, hidden := range []string{"\n  api", "\n  batch-boot-config", "\n  boot-config-wizard", "\n  target", "\n  tasks", "\n  upgrade", "hyperbdrctl tasks list"} {
		if strings.Contains(text, hidden) {
			t.Fatalf("help should not include %q: %q", hidden, text)
		}
	}
	for _, unwanted := range []string{"Examples:", "\nGlobal Flags:\n", "help for hyperbdrctl"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestRootHelpUsesLocalizedCopyInBothLanguages(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "zh_cn",
			args: []string{"--lang", "zh_cn", "--help"},
			want: []string{
				"HyperBDR 与 HyperMotion 命令行客户端\n\n用法: hyperbdrctl [全局参数] <命令> [参数]",
				"\n参数:\n",
				"--lang string     显示语言，可选值 en / zh_cn，默认值 en",
				"--output string   输出格式，可选值 table / json，默认值 table",
				"--version         显示 CLI 版本信息",
				"cloud-sync-gateway       云同步网关管理",
				"oss                      目标对象存储管理",
				"production-site          生产站点管理",
				"--host https://<server>:10443",
				"以 Bash 为例，先准备环境变量，再直接执行查询命令：",
				"配置完成后，可以先执行下面的命令确认配置和连接是否正常：\n    hyperbdrctl config get",
				"hyperbdrctl boot-config apply --cloud-account-id <account_id> --help",
				"hyperbdrctl cloud-account create --help",
			},
		},
		{
			name: "en",
			args: []string{"--lang", "en", "--help"},
			want: []string{
				"HyperBDR and HyperMotion command-line client\n\nUsage: hyperbdrctl [global flags] <command> [flags]",
				"\nFlags:\n",
				"--lang string     Display language, allowed values en / zh_cn, default en",
				"--output string   Output format, allowed values table / json, default table",
				"--version         Show CLI version information",
				"cloud-sync-gateway       Cloud sync gateway management",
				"oss                      Target object storage management",
				"production-site          Production site management",
				"--host https://<server>:10443",
				"Using Bash as an example, prepare the environment variables before running a query command:",
				"After configuration, run the following commands to verify the configuration and connection:\n    hyperbdrctl config get",
				"hyperbdrctl boot-config apply --cloud-account-id <account_id> --help",
				"hyperbdrctl cloud-account create --help",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(tt.args, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			text := out.String()
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			usage := strings.Index(text, i18n.New(tt.name).T("help.section_usage")+":")
			flags := strings.Index(text, "\n"+i18n.New(tt.name).T("help.section_flags")+":\n")
			commands := strings.Index(text, "\n"+i18n.New(tt.name).T("help.section_commands")+":\n")
			notes := strings.Index(text, "\n"+i18n.New(tt.name).T("help.section_usage_notes")+":\n")
			if usage < 0 || flags < 0 || commands < 0 || notes < 0 || !(usage < flags && flags < commands && commands < notes) {
				t.Fatalf("help order mismatch: %q", text)
			}
			for _, unwanted := range []string{"--insecure", "--vertical", "\nGlobal Flags:\n", "\n全局参数:\n"} {
				if strings.Contains(text, unwanted) {
					t.Fatalf("help should not include %q: %q", unwanted, text)
				}
			}
		})
	}
}

func TestNonConfigHelpDoesNotShowFooter(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"cloud-account", "create", "--help"},
		{"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
	}

	for _, args := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		assertNoHelpFooter(t, out.String())
	}
}

func TestConfigSetLogsInBeforeSaving(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotUsername, gotPassword string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/loginNoImageCaptcha" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		gotUsername = body["username"]
		gotPassword = body["password"]
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"token": "saved-token"},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute([]string{
		"config", "set",
		"--host", srv.URL,
		"--username", "admin",
		"--password", "secret",
		"--scene", "dr",
		"--lang", "zh_cn",
	}, &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotUsername != "admin" || gotPassword != "secret" {
		t.Fatalf("login body username=%q password=%q", gotUsername, gotPassword)
	}
	path, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	rawConfig, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rawConfig), `"password": "secret"`) || !strings.Contains(string(rawConfig), `"password": "enc:v1:12010e1b0b15"`) {
		t.Fatalf("password was not encoded in config file: %s", rawConfig)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != srv.URL || cfg.Username != "admin" || cfg.Password != "secret" || cfg.Lang != "zh_cn" {
		t.Fatalf("config = %+v", cfg)
	}
	tokenPath, err := config.TokenPath()
	if err != nil {
		t.Fatal(err)
	}
	tok, err := config.LoadToken(tokenPath)
	if err != nil {
		t.Fatal(err)
	}
	if tok.Token != "saved-token" {
		t.Fatalf("token = %+v", tok)
	}
}

func TestConfigSetReencodesPasswordWhenUsernameChanges(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotUsername, gotPassword string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		gotUsername = body["username"]
		gotPassword = body["password"]
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"token": "updated-token"},
		})
	}))
	defer srv.Close()

	path, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(path, config.Config{Host: srv.URL, Username: "old", Password: "secret"}); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := Execute([]string{"config", "set", "--username", "new"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if gotUsername != "new" || gotPassword != "secret" {
		t.Fatalf("login body username=%q password=%q", gotUsername, gotPassword)
	}

	rawConfig, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rawConfig), `"password": "enc:v1:1d00141c0003"`) {
		t.Fatalf("password was not re-encoded with the new username: %s", rawConfig)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Username != "new" || loaded.Password != "secret" {
		t.Fatalf("config = %+v", loaded)
	}
}

func TestConfigSetDoesNotSaveOnLoginFailure(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	path, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(path, config.Config{Host: "https://old.example", Username: "old", Password: "old-pass"}); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err = Execute([]string{
		"config", "set",
		"--host", srv.URL,
		"--username", "new",
		"--password", "bad",
	}, &out, &errOut)
	if err == nil {
		t.Fatal("expected login failure")
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://old.example" || cfg.Username != "old" || cfg.Password != "old-pass" {
		t.Fatalf("config should not have changed: %+v", cfg)
	}
}

func TestNonPassthroughCommandsStillRejectUnknownFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "license", "activate", "--ddty", "d", "--custom-step", "3"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}

func TestQueryPassthroughCommandsAllowFormerLegacyConfigFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args      []string
		wantQuery string
	}{
		{args: []string{"--host", "https://legacy.invalid"}, wantQuery: "host=https%3A%2F%2Flegacy.invalid"},
		{args: []string{"--username", "legacy-user"}, wantQuery: "username=legacy-user"},
		{args: []string{"--password", "legacy-pass"}, wantQuery: "password=legacy-pass"},
		{args: []string{"--scene", "migration"}, wantQuery: "scene=migration"},
		{args: []string{"--insecure"}, wantQuery: "insecure=true"},
	}

	for _, tc := range cases {
		var gotQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"pages": []interface{}{}},
			})
		}))

		var out, errOut bytes.Buffer
		args := append(withHost(t, srv.URL, "license", "list"), tc.args...)
		err := Execute(args, &out, &errOut)
		srv.Close()
		if err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		if !strings.Contains(gotQuery, tc.wantQuery) {
			t.Fatalf("args=%v query=%q missing %q", args, gotQuery, tc.wantQuery)
		}
	}
}

func TestLicenseHelpShowsModernGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"license", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage Notes:",
		"hyperbdrctl license list",
		"hyperbdrctl license reg-code",
		"hyperbdrctl license activate --ddty <activation_code>",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nGlobal Flags:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestLicenseActivateHelpShowsRequiredDDTYOnly(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"license", "activate", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"--ddty",
		"(required)",
		"hyperbdrctl license activate --ddty <activation_code>",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"--kkty", "--file", "required unless --file is used"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nGlobal Flags:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestLicenseHelpOmitsArchivedRemovedNotesInBothLanguages(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name     string
		args     []string
		unwanted []string
	}{
		{
			name: "zh group",
			args: []string{"--lang", "zh_cn", "license", "--help"},
			unwanted: []string{
				"`license activate` 会自动查询当前环境的注册码",
			},
		},
		{
			name: "zh reg-code",
			args: []string{"--lang", "zh_cn", "license", "reg-code", "--help"},
			unwanted: []string{
				"不要把注册码写入共享报告、工单或聊天记录。",
			},
		},
		{
			name: "zh activate",
			args: []string{"--lang", "zh_cn", "license", "activate", "--help"},
			unwanted: []string{
				"CLI 会在发送激活请求前自动查询当前环境的 `kkty`",
				"避免把激活码输出到共享日志、报告或工单。",
			},
		},
		{
			name: "en group",
			args: []string{"license", "--help"},
			unwanted: []string{
				"`license activate` automatically fetches the current environment registration code",
			},
		},
		{
			name: "en reg-code",
			args: []string{"license", "reg-code", "--help"},
			unwanted: []string{
				"Do not copy registration codes into shared reports, tickets, or chat logs.",
			},
		},
		{
			name: "en activate",
			args: []string{"license", "activate", "--help"},
			unwanted: []string{
				"The CLI automatically fetches the current environment `kkty` before it sends the activation request",
				"Avoid printing activation codes into shared logs, reports, or tickets.",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(tc.args, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			text := out.String()
			for _, unwanted := range tc.unwanted {
				if strings.Contains(text, unwanted) {
					t.Fatalf("help should not include %q: %q", unwanted, text)
				}
			}
		})
	}
}

func TestLicenseHelpMatchesArchivedCopyInBothLanguages(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name     string
		args     []string
		want     []string
		unwanted []string
	}{
		{
			name: "zh group",
			args: []string{"--lang", "zh_cn", "license", "--help"},
			want: []string{
				"查看当前许可证状态：",
				"获取当前环境的注册码：",
				"使用注册码生成激活码后，提交激活：",
			},
			unwanted: []string{"该命令组用于"},
		},
		{
			name: "zh list",
			args: []string{"--lang", "zh_cn", "license", "list", "--help"},
			want: []string{
				"按分页查看当前平台上的许可证摘要。",
				"查看默认分页结果：",
				"指定分页参数：",
				"输出 JSON 供其他工具处理：",
				"hyperbdrctl license list --output json",
			},
			unwanted: []string{"hyperbdrctl --output json license list"},
		},
		{
			name: "zh reg-code",
			args: []string{"--lang", "zh_cn", "license", "reg-code", "--help"},
			want: []string{
				"读取当前环境许可证激活所需的注册码。",
				"获取注册码：",
				"该命令只读取注册码，不会修改服务端状态。",
				"使用注册码生成激活码后，继续执行：",
			},
		},
		{
			name: "zh activate",
			args: []string{"--lang", "zh_cn", "license", "activate", "--help"},
			want: []string{
				"提交许可证激活码，完成当前环境的许可证激活。",
				"参数来源：",
				"--ddty string\n      使用基于注册码生成的许可证激活码。",
				"执行前，先获取当前环境的注册码：",
				"激活完成后，查看许可证状态：",
			},
		},
		{
			name: "en group",
			args: []string{"--lang", "en", "license", "--help"},
			want: []string{
				"View the current license status:",
				"Get the registration code for the current environment:",
				"After using the registration code to generate an activation code, submit the activation:",
			},
			unwanted: []string{"Use this command group"},
		},
		{
			name: "en list",
			args: []string{"--lang", "en", "license", "list", "--help"},
			want: []string{
				"View paginated license summaries on the current platform.",
				"View the default page:",
				"Specify pagination parameters:",
				"Output JSON for other tools:",
				"hyperbdrctl license list --output json",
			},
			unwanted: []string{"hyperbdrctl --output json license list"},
		},
		{
			name: "en reg-code",
			args: []string{"--lang", "en", "license", "reg-code", "--help"},
			want: []string{
				"Read the registration code required to activate the license in the current environment.",
				"Get the registration code:",
				"This command only reads the registration code and does not modify server-side state.",
				"After using the registration code to generate an activation code, continue with:",
			},
		},
		{
			name: "en activate",
			args: []string{"--lang", "en", "license", "activate", "--help"},
			want: []string{
				"Submit a license activation code to activate the license in the current environment.",
				"Parameter Sources:",
				"--ddty string\n      Use the license activation code generated from the registration code.",
				"Before activation, get the registration code for the current environment:",
				"After activation, view the license status:",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(tc.args, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			text := out.String()
			for _, want := range tc.want {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			for _, unwanted := range tc.unwanted {
				if strings.Contains(text, unwanted) {
					t.Fatalf("help should not include %q: %q", unwanted, text)
				}
			}
		})
	}
}

func TestParseQueryFlagsIntoRejectsUnknownFlagsByDefault(t *testing.T) {
	fs := newFlagSet("test")
	fs.String("known", "", "")
	q := queryFromPairs()

	err := parseQueryFlagsInto(fs, []string{"--custom-step", "3"}, q)
	if err == nil || err.Error() != "unknown flag: --custom-step" {
		t.Fatalf("err = %v", err)
	}
}
