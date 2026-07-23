package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordEncodingRule(t *testing.T) {
	const want = "enc:v1:12010e1b0b15"
	got, err := encodePassword("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("encodePassword() = %q, want %q", got, want)
	}
	decoded, err := decodePassword("admin", got)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "secret" {
		t.Fatalf("decodePassword() = %q", decoded)
	}
}

func TestPasswordEncodingRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "ASCII special characters", username: "operator", password: `p@ss-W0rd!#$%`},
		{name: "Chinese", username: "管理员", password: "密码-安全-123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encodePassword(tt.username, tt.password)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := decodePassword(tt.username, encoded)
			if err != nil {
				t.Fatal(err)
			}
			if decoded != tt.password {
				t.Fatalf("decodePassword() = %q, want %q", decoded, tt.password)
			}
		})
	}
}

func TestSaveEncodesPasswordAndLoadDecodesIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := Save(path, Config{Host: "https://example.com", Username: "admin", Password: "secret"}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"password": "secret"`) {
		t.Fatalf("config contains plaintext password: %s", b)
	}
	var stored Config
	if err := json.Unmarshal(b, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Password != "enc:v1:12010e1b0b15" {
		t.Fatalf("stored password = %q", stored.Password)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Password != "secret" {
		t.Fatalf("loaded password = %q", loaded.Password)
	}
}

func TestLegacyPlaintextPasswordMigratesOnNextSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	legacy := []byte(`{"username":"admin","password":"legacy-secret"}`)
	if err := os.WriteFile(path, legacy, 0600); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Password != "legacy-secret" {
		t.Fatalf("loaded password = %q", loaded.Password)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string(legacy) {
		t.Fatalf("Load modified legacy config: %q", b)
	}

	if err := Save(path, loaded); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "legacy-secret") || !strings.Contains(string(b), `"password": "enc:v1:`) {
		t.Fatalf("legacy password was not migrated: %s", b)
	}
}

func TestPasswordEncodingErrors(t *testing.T) {
	if _, err := encodePassword("", "secret"); err == nil || !strings.Contains(err.Error(), "username is required") {
		t.Fatalf("encode error = %v", err)
	}
	if _, err := decodePassword("", "enc:v1:00"); err == nil || !strings.Contains(err.Error(), "username is required") {
		t.Fatalf("decode username error = %v", err)
	}
	if _, err := decodePassword("admin", "enc:v1:not-hex"); err == nil || !strings.Contains(err.Error(), "decode stored password") {
		t.Fatalf("decode hex error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := Save(path, Config{Password: "secret"}); err == nil || !strings.Contains(err.Error(), "username is required to encode password") {
		t.Fatalf("Save error = %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"username":"admin","password":"enc:v1:not-hex"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "decode stored password") {
		t.Fatalf("Load error = %v", err)
	}
}

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
