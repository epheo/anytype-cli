package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileGivesDefaults(t *testing.T) {
	t.Setenv(EnvAppKey, "")
	t.Setenv(EnvBaseURL, "")
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != DefaultBaseURL || cfg.AppKey != "" || cfg.Authenticated() {
		t.Fatalf("unexpected config %+v", cfg)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("Load must not create the config file")
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	t.Setenv(EnvAppKey, "")
	t.Setenv(EnvBaseURL, "")
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")

	cfg := &Config{AppKey: "k", BaseURL: "http://x:1", path: path}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("perm = %o, want 600", perm)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.AppKey != "k" || got.BaseURL != "http://x:1" {
		t.Fatalf("round trip mismatch %+v", got)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("app_key: file\nbase_url: http://file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvAppKey, "env")
	t.Setenv(EnvBaseURL, "http://env")
	t.Setenv(EnvSpace, "work")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppKey != "env" || cfg.BaseURL != "http://env" || cfg.DefaultSpace != "work" {
		t.Fatalf("env must win, got %+v", cfg)
	}
}
