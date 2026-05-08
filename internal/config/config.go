package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultScene  = "dr"
	DefaultLang   = "en"
	DefaultOutput = "table"
)

type Config struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Scene    string `json:"scene,omitempty"`
	Lang     string `json:"lang,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
	Output   string `json:"output,omitempty"`
	Debug    bool   `json:"debug,omitempty"`
}

type Flags struct {
	Host        string
	Username    string
	Password    string
	Scene       string
	Lang        string
	Insecure    bool
	InsecureSet bool
	Output      string
	Debug       bool
	DebugSet    bool
}

type Resolved struct {
	Config
	ConfigPath string
	CachePath  string
}

func Resolve(flags Flags) (Resolved, error) {
	path, err := ConfigPath()
	if err != nil {
		return Resolved{}, err
	}
	cachePath, err := TokenPath()
	if err != nil {
		return Resolved{}, err
	}
	return ResolveWithPaths(flags, path, cachePath)
}

func ResolveLang(flags Flags) (string, error) {
	path, err := ConfigPath()
	if err != nil {
		return "", err
	}
	return ResolveLangWithPath(flags, path)
}

func ResolveLangWithPath(flags Flags, path string) (string, error) {
	cfg, err := Load(path)
	if err != nil {
		return "", err
	}
	cfg = mergeEnv(cfg)
	cfg = mergeFlags(cfg, flags)
	applyDefaults(&cfg)
	return cfg.Lang, nil
}

func ResolveWithPaths(flags Flags, path, cachePath string) (Resolved, error) {
	cfg, err := Load(path)
	if err != nil {
		return Resolved{}, err
	}

	cfg = mergeEnv(cfg)
	cfg = mergeFlags(cfg, flags)
	applyDefaults(&cfg)
	if err := Validate(cfg); err != nil {
		return Resolved{}, err
	}
	return Resolved{Config: cfg, ConfigPath: path, CachePath: cachePath}, nil
}

func Load(path string) (Config, error) {
	var cfg Config
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	normalized, err := Normalize(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func Normalize(cfg Config) (Config, error) {
	applyDefaults(&cfg)
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hyperbdrctl", "config.json"), nil
}

func TokenPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hyperbdrctl", "token.json"), nil
}

func Validate(cfg Config) error {
	if cfg.Lang != "" && cfg.Lang != "en" && cfg.Lang != "zh_cn" {
		return fmt.Errorf("unsupported lang %q, expected en or zh_cn", cfg.Lang)
	}
	if cfg.Output != "" && cfg.Output != "table" && cfg.Output != "json" {
		return fmt.Errorf("unsupported output %q, expected table or json", cfg.Output)
	}
	return nil
}

func mergeEnv(cfg Config) Config {
	if v := os.Getenv("HYPERBDR_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("HYPERBDR_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("HYPERBDR_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("HYPERBDR_SCENE"); v != "" {
		cfg.Scene = v
	}
	if v := os.Getenv("HYPERBDR_LANG"); v != "" {
		cfg.Lang = v
	}
	if v := os.Getenv("HYPERBDR_OUTPUT"); v != "" {
		cfg.Output = v
	}
	if v := os.Getenv("HYPERBDR_INSECURE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Insecure = b
		}
	}
	if v := os.Getenv("HYPERBDR_DEBUG"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Debug = b
		}
	}
	return cfg
}

func mergeFlags(cfg Config, flags Flags) Config {
	if flags.Host != "" {
		cfg.Host = flags.Host
	}
	if flags.Username != "" {
		cfg.Username = flags.Username
	}
	if flags.Password != "" {
		cfg.Password = flags.Password
	}
	if flags.Scene != "" {
		cfg.Scene = flags.Scene
	}
	if flags.Lang != "" {
		cfg.Lang = flags.Lang
	}
	if flags.Output != "" {
		cfg.Output = flags.Output
	}
	if flags.InsecureSet {
		cfg.Insecure = flags.Insecure
	}
	if flags.DebugSet {
		cfg.Debug = flags.Debug
	}
	return cfg
}

func applyDefaults(cfg *Config) {
	if cfg.Scene == "" {
		cfg.Scene = DefaultScene
	}
	if cfg.Lang == "" {
		cfg.Lang = DefaultLang
	}
	if cfg.Output == "" {
		cfg.Output = DefaultOutput
	}
}
