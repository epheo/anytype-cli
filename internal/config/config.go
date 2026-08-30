// Package config reads and writes the CLI credentials file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL = "http://localhost:31009"
	EnvAppKey      = "ANYTYPE_APP_KEY"
	EnvBaseURL     = "ANYTYPE_BASE_URL"
	EnvSpace       = "ANYTYPE_SPACE"
)

type Config struct {
	AppKey  string `yaml:"app_key"`
	BaseURL string `yaml:"base_url"`
	// DefaultSpace is a space ID or name used when --space is absent.
	DefaultSpace string `yaml:"default_space,omitempty"`

	path string
}

// DefaultPath keeps the pre-refactor location so existing credentials stay valid.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".anytype-cli", "config.yaml"), nil
}

// Load never writes to disk; a missing file yields defaults.
// Environment variables override file values.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		if path, err = DefaultPath(); err != nil {
			return nil, err
		}
	}

	cfg := &Config{path: path}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return nil, err
	}

	if v := os.Getenv(EnvAppKey); v != "" {
		cfg.AppKey = v
	}
	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(EnvSpace); v != "" {
		cfg.DefaultSpace = v
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	return cfg, nil
}

func (c *Config) Path() string { return c.path }

func (c *Config) Authenticated() bool { return c != nil && c.AppKey != "" }

// Save restricts permissions because the file holds the API key.
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600)
}
