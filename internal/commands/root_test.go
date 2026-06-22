package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	if !strings.Contains(out.String(), "source") {
		t.Fatalf("usage missing source = %q", out.String())
	}
	for _, hidden := range []string{"\n  api", "\n  batch-boot-config", "\n  boot-config-wizard", "\n  tasks", "\n  upgrade"} {
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

func TestTopLevelAPIIsRouted(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"api", "request"}, &out, &errOut)
	if err == nil || err.Error() != "path is required" {
		t.Fatalf("err = %v", err)
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

func TestTopLevelBootConfigCLIIsRouted(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config-cli"), &out, &errOut)
	if err == nil || err.Error() != "boot-config-cli requires subcommand" {
		t.Fatalf("err = %v", err)
	}
}

func TestHiddenTopLevelCommandHelpRemainsAvailable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{args: []string{"help", "api"}, want: []string{"Usage:", "request"}},
		{args: []string{"help", "tasks"}, want: []string{"Usage:", "list", "steps"}},
		{args: []string{"help", "upgrade"}, want: []string{"Usage:", "host"}},
	}

	for _, tc := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tc.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tc.args, err)
		}
		text := out.String()
		for _, want := range tc.want {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v help missing %q: %q", tc.args, want, text)
			}
		}
	}
}

func TestRemovedTopLevelCommandHelpIsUnavailable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"help", "boot-config-wizard"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown") || !strings.Contains(err.Error(), "boot-config-wizard") {
		t.Fatalf("err = %v", err)
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
	assertNoHelpFooter(t, got)
}

func TestHostWaitHelpShowsCleanOperationChoice(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "host", "wait"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"--operation", "allowed values sync / boot / clean /", "default 60", "default 3600"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	if strings.Contains(got, "cleanup-validation-host") {
		t.Fatalf("help should not mention legacy operation: %q", got)
	}
	assertNoHelpFooter(t, got)
}

func TestDeprecatedBootConfigCLIHelpShowsDeprecation(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "boot-config-cli"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "Deprecated") || !strings.Contains(got, "boot-config apply") {
		t.Fatalf("help = %q", got)
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
		"hyperbdrctl config get",
		"hyperbdrctl config set \\",
		"HYPERBDR_HOST",
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
	notes := strings.Index(text, "Usage Notes:")
	if usage < 0 || flags < 0 || notes < 0 || !(usage < flags && flags < notes) {
		t.Fatalf("help order mismatch: %q", text)
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
		"default false",
		"default en",
		"default table",
		"hyperbdrctl config set \\",
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
	assertNoHelpFooter(t, text)
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
		"Global Flags:",
		"Usage Notes:",
		"Generate shell completion scripts",
		"hyperbdrctl config set \\",
		"HYPERBDR_HOST",
		"--vertical",
		"command-line flags > environment variables > config file > defaults",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, hidden := range []string{"\n  api", "\n  batch-boot-config", "\n  boot-config-wizard", "\n  tasks", "\n  upgrade", "hyperbdrctl tasks list"} {
		if strings.Contains(text, hidden) {
			t.Fatalf("help should not include %q: %q", hidden, text)
		}
	}
	for _, unwanted := range []string{"Examples:", "\nFlags:\n", "help for hyperbdrctl"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestNonConfigHelpDoesNotShowFooter(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"target", "account", "create", "--help"},
		{"target", "account", "create-block", "aliyun", "--help"},
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

func TestParseQueryFlagsIntoRejectsUnknownFlagsByDefault(t *testing.T) {
	fs := newFlagSet("test")
	fs.String("known", "", "")
	q := queryFromPairs()

	err := parseQueryFlagsInto(fs, []string{"--custom-step", "3"}, q)
	if err == nil || err.Error() != "unknown flag: --custom-step" {
		t.Fatalf("err = %v", err)
	}
}
