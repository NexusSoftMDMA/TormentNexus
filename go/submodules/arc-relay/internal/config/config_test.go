package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddr(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Host: "127.0.0.1", Port: 9090},
	}
	got := cfg.Addr()
	want := "127.0.0.1:9090"
	if got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestPublicBaseURL(t *testing.T) {
	t.Run("returns BaseURL when set", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{Port: 8080, BaseURL: "https://mcp.example.com"},
		}
		got := cfg.PublicBaseURL()
		want := "https://mcp.example.com"
		if got != want {
			t.Errorf("PublicBaseURL() = %q, want %q", got, want)
		}
	})

	t.Run("falls back to localhost", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{Port: 3000},
		}
		got := cfg.PublicBaseURL()
		want := "http://localhost:3000"
		if got != want {
			t.Errorf("PublicBaseURL() = %q, want %q", got, want)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("defaults with empty path", func(t *testing.T) {
		// Neutralise PaaS-style env vars so this test does not depend
		// on the caller's environment (CI runners often set PORT).
		t.Setenv("ARC_RELAY_PORT", "")
		t.Setenv("PORT", "")
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
		}
		if cfg.Database.Path != "arc-relay.db" {
			t.Errorf("Database.Path = %q, want %q", cfg.Database.Path, "arc-relay.db")
		}
		if cfg.Docker.Network != "arc-relay" {
			t.Errorf("Docker.Network = %q, want %q", cfg.Docker.Network, "arc-relay")
		}
	})

	t.Run("loads TOML file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		content := `
[server]
host = "10.0.0.1"
port = 9999

[database]
path = "/tmp/test.db"
`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("writing config file: %v", err)
		}

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.Host != "10.0.0.1" {
			t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "10.0.0.1")
		}
		if cfg.Server.Port != 9999 {
			t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9999)
		}
		if cfg.Database.Path != "/tmp/test.db" {
			t.Errorf("Database.Path = %q, want %q", cfg.Database.Path, "/tmp/test.db")
		}
	})

	t.Run("env var overrides", func(t *testing.T) {
		t.Setenv("ARC_RELAY_PORT", "4444")
		t.Setenv("ARC_RELAY_BASE_URL", "https://override.example.com")
		t.Setenv("ARC_RELAY_DB_PATH", "/override/db.sqlite")
		t.Setenv("ARC_RELAY_ENCRYPTION_KEY", "secret-key")
		t.Setenv("ARC_RELAY_SESSION_SECRET", "session-secret")
		t.Setenv("ARC_RELAY_ADMIN_PASSWORD", "admin-pass")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.Port != 4444 {
			t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 4444)
		}
		if cfg.Server.BaseURL != "https://override.example.com" {
			t.Errorf("Server.BaseURL = %q, want %q", cfg.Server.BaseURL, "https://override.example.com")
		}
		if cfg.Database.Path != "/override/db.sqlite" {
			t.Errorf("Database.Path = %q, want %q", cfg.Database.Path, "/override/db.sqlite")
		}
		if cfg.Encryption.Key != "secret-key" {
			t.Errorf("Encryption.Key = %q, want %q", cfg.Encryption.Key, "secret-key")
		}
		if cfg.Auth.SessionSecret != "session-secret" {
			t.Errorf("Auth.SessionSecret = %q, want %q", cfg.Auth.SessionSecret, "session-secret")
		}
		if cfg.Auth.AdminPassword != "admin-pass" {
			t.Errorf("Auth.AdminPassword = %q, want %q", cfg.Auth.AdminPassword, "admin-pass")
		}
	})

	t.Run("missing file error", func(t *testing.T) {
		_, err := Load("/nonexistent/config.toml")
		if err == nil {
			t.Error("Load() should return error for missing file")
		}
	})

	t.Run("PORT env var fallback for PaaS", func(t *testing.T) {
		t.Setenv("PORT", "5555")
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.Port != 5555 {
			t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 5555)
		}
	})

	t.Run("RENDER_EXTERNAL_URL sets BaseURL when ARC_RELAY_BASE_URL unset", func(t *testing.T) {
		t.Setenv("RENDER_EXTERNAL_URL", "https://arc-relay-abcd.onrender.com")
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.BaseURL != "https://arc-relay-abcd.onrender.com" {
			t.Errorf("Server.BaseURL = %q, want %q", cfg.Server.BaseURL, "https://arc-relay-abcd.onrender.com")
		}
	})

	t.Run("RAILWAY_PUBLIC_DOMAIN sets BaseURL with https prefix", func(t *testing.T) {
		t.Setenv("RAILWAY_PUBLIC_DOMAIN", "arc-relay.up.railway.app")
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.BaseURL != "https://arc-relay.up.railway.app" {
			t.Errorf("Server.BaseURL = %q, want %q", cfg.Server.BaseURL, "https://arc-relay.up.railway.app")
		}
	})

	t.Run("ARC_RELAY_BASE_URL takes precedence over platform env vars", func(t *testing.T) {
		t.Setenv("ARC_RELAY_BASE_URL", "https://manual.example.com")
		t.Setenv("RENDER_EXTERNAL_URL", "https://render.example.com")
		t.Setenv("RAILWAY_PUBLIC_DOMAIN", "railway.example.com")
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.BaseURL != "https://manual.example.com" {
			t.Errorf("Server.BaseURL = %q, want %q", cfg.Server.BaseURL, "https://manual.example.com")
		}
	})

	t.Run("ARC_RELAY_PORT takes precedence over PORT", func(t *testing.T) {
		t.Setenv("ARC_RELAY_PORT", "6666")
		t.Setenv("PORT", "5555")
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Server.Port != 6666 {
			t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 6666)
		}
	})
}
