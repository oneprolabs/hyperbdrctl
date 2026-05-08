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
	args, flags, err := extractGlobalFlags([]string{"config", "set", "--host", "https://x", "--password=secret", "--insecure", "--lang", "zh_cn"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(args, " ") != "config set --host https://x --password=secret --insecure" {
		t.Fatalf("args = %v", args)
	}
	if flags.Host != "" || flags.Password != "" || flags.Lang != "zh_cn" || flags.Insecure || flags.InsecureSet {
		t.Fatalf("flags = %+v", flags)
	}
}

func TestUsageIsLocalized(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "用法") {
		t.Fatalf("usage = %q", out.String())
	}
	if !strings.Contains(out.String(), "licenses") {
		t.Fatalf("usage missing licenses = %q", out.String())
	}
	if !strings.Contains(out.String(), "api") {
		t.Fatalf("usage missing api = %q", out.String())
	}
	if !strings.Contains(out.String(), "batch-boot-config") {
		t.Fatalf("usage missing batch-boot-config = %q", out.String())
	}
	if !strings.Contains(out.String(), "boot-config-wizard") {
		t.Fatalf("usage missing boot-config-wizard = %q", out.String())
	}
	if !strings.Contains(out.String(), "source") {
		t.Fatalf("usage missing source = %q", out.String())
	}
	if strings.Contains(out.String(), "boot-config              Deprecated top-level boot configuration command group") {
		t.Fatalf("usage should hide deprecated boot-config alias = %q", out.String())
	}
	if strings.Contains(out.String(), "boot-config-cli          Deprecated compatibility apply command") {
		t.Fatalf("usage should hide deprecated boot-config-cli alias = %q", out.String())
	}
	if strings.Contains(out.String(), "sources") {
		t.Fatalf("usage should hide deprecated sources alias = %q", out.String())
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

func TestTopLevelBootConfigWizardIsRouted(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config-wizard"), &out, &errOut)
	if err == nil || err.Error() != "boot-config-wizard requires subcommand" {
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

func TestHostBootConfigHelpShowsCanonicalSubcommands(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "host", "boot-config"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"apply", "create", "update", "Canonical boot configuration command group"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
}

func TestDeprecatedBootConfigCLIHelpShowsDeprecation(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "boot-config-cli"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "Deprecated") || !strings.Contains(got, "host boot-config apply") {
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
		"command-line flags > environment variables > config file > defaults",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
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

func TestNonConfigCommandsRejectLegacyHostFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"--host", "https://example.invalid", "licenses", "list"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --host") {
		t.Fatalf("err = %v", err)
	}
}
