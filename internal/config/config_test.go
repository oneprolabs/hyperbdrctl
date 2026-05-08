package config

import (
	"path/filepath"
	"testing"
)

func TestResolvePrecedence(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	tokenPath := filepath.Join(dir, "token.json")
	if err := Save(configPath, Config{
		Host:     "https://config.example",
		Username: "config-user",
		Password: "config-pass",
		Scene:    "dr",
		Lang:     "en",
		Output:   "table",
	}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HYPERBDR_HOST", "https://env.example")
	t.Setenv("HYPERBDR_USERNAME", "env-user")
	t.Setenv("HYPERBDR_PASSWORD", "env-pass")
	t.Setenv("HYPERBDR_LANG", "zh_cn")
	t.Setenv("HYPERBDR_OUTPUT", "json")
	t.Setenv("HYPERBDR_INSECURE", "true")
	t.Setenv("HYPERBDR_DEBUG", "true")

	resolved, err := ResolveWithPaths(Flags{
		Host:        "https://flag.example",
		Password:    "flag-pass",
		Insecure:    false,
		InsecureSet: true,
		Debug:       false,
		DebugSet:    true,
	}, configPath, tokenPath)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Host != "https://flag.example" {
		t.Fatalf("host = %q", resolved.Host)
	}
	if resolved.Username != "env-user" {
		t.Fatalf("username = %q", resolved.Username)
	}
	if resolved.Password != "flag-pass" {
		t.Fatalf("password = %q", resolved.Password)
	}
	if resolved.Lang != "zh_cn" {
		t.Fatalf("lang = %q", resolved.Lang)
	}
	if resolved.Output != "json" {
		t.Fatalf("output = %q", resolved.Output)
	}
	if resolved.Insecure {
		t.Fatal("flag insecure=false should override env insecure=true")
	}
	if resolved.Debug {
		t.Fatal("flag debug=false should override env debug=true")
	}
}

func TestResolveDefaults(t *testing.T) {
	dir := t.TempDir()
	resolved, err := ResolveWithPaths(Flags{}, filepath.Join(dir, "missing.json"), filepath.Join(dir, "token.json"))
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Scene != DefaultScene || resolved.Lang != DefaultLang || resolved.Output != DefaultOutput {
		t.Fatalf("defaults = scene:%q lang:%q output:%q", resolved.Scene, resolved.Lang, resolved.Output)
	}
}

func TestResolveLangWithPathUsesPrecedence(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := Save(configPath, Config{Lang: "zh_cn"}); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HYPERBDR_LANG", "en")

	lang, err := ResolveLangWithPath(Flags{Lang: "zh_cn"}, configPath)
	if err != nil {
		t.Fatal(err)
	}
	if lang != "zh_cn" {
		t.Fatalf("lang = %q", lang)
	}
}

func TestValidateRejectsAPILangAlias(t *testing.T) {
	if err := Validate(Config{Lang: "zh"}); err == nil {
		t.Fatal("expected zh to remain unsupported as a CLI config lang")
	}
}
