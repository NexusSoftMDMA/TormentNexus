package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		RelayURL: "http://127.0.0.1:8080",
		APIKey:   "test-token-123",
	}

	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.RelayURL != cfg.RelayURL {
		t.Errorf("RelayURL = %q, want %q", loaded.RelayURL, cfg.RelayURL)
	}
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, cfg.APIKey)
	}
}

func TestSaveConfigPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission checks not applicable on Windows")
	}

	dir := t.TempDir()
	cfg := &Config{RelayURL: "http://example.com", APIKey: "key"}

	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	info, err := os.Stat(ConfigPath(dir))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("config file permissions = %04o, want 0600", perm)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	// TempDir creates with its own perms, so we check the nested dir if created
	_ = dirInfo
}

func TestLoadConfigNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestLoadConfigMissingURL(t *testing.T) {
	dir := t.TempDir()
	data := `{"api_key": "key"}`
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing relay_url, got nil")
	}
}

func TestLoadConfigMissingKey(t *testing.T) {
	dir := t.TempDir()
	data := `{"relay_url": "http://example.com"}`
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing api_key, got nil")
	}
}

func TestLoadConfigMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestResolveCredentialsFromEnv(t *testing.T) {
	t.Setenv("ARC_SYNC_URL", "http://env-relay:8080")
	t.Setenv("ARC_SYNC_API_KEY", "env-token")

	creds, err := ResolveCredentials(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}

	if creds.Source != "environment" {
		t.Errorf("Source = %q, want %q", creds.Source, "environment")
	}
	if creds.RelayURL != "http://env-relay:8080" {
		t.Errorf("RelayURL = %q, want http://env-relay:8080", creds.RelayURL)
	}
	if creds.APIKey != "env-token" {
		t.Errorf("APIKey = %q, want env-token", creds.APIKey)
	}
}

func TestResolveCredentialsFromFile(t *testing.T) {
	// Ensure env vars are cleared
	t.Setenv("ARC_SYNC_URL", "")
	t.Setenv("ARC_SYNC_API_KEY", "")

	dir := t.TempDir()
	cfg := &Config{RelayURL: "http://file-relay:8080", APIKey: "file-token"}
	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	creds, err := ResolveCredentials(dir)
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}

	if creds.Source == "environment" {
		t.Error("expected source to be config file, got environment")
	}
	if creds.RelayURL != "http://file-relay:8080" {
		t.Errorf("RelayURL = %q, want http://file-relay:8080", creds.RelayURL)
	}
}

func TestResolveCredentialsEnvPartial(t *testing.T) {
	// Only URL set in env, no key — should fall through to config file
	t.Setenv("ARC_SYNC_URL", "http://partial:8080")
	t.Setenv("ARC_SYNC_API_KEY", "")

	dir := t.TempDir()
	cfg := &Config{RelayURL: "http://file:8080", APIKey: "file-key"}
	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}

	creds, err := ResolveCredentials(dir)
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}

	// Should fall through to file since both env vars must be set
	if creds.Source == "environment" {
		t.Error("expected file source when only partial env vars set")
	}
}

func TestCheckPermissionsSecure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission checks not applicable on Windows")
	}

	dir := t.TempDir()
	cfg := &Config{RelayURL: "http://example.com", APIKey: "key"}
	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}

	warning := CheckPermissions(dir)
	if warning != "" {
		t.Errorf("expected no warning for 0600 file, got: %s", warning)
	}
}

func TestCheckPermissionsInsecure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission checks not applicable on Windows")
	}

	dir := t.TempDir()
	cfg := &Config{RelayURL: "http://example.com", APIKey: "key"}
	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}

	// Make insecure
	if err := os.Chmod(ConfigPath(dir), 0644); err != nil {
		t.Fatal(err)
	}

	warning := CheckPermissions(dir)
	if warning == "" {
		t.Error("expected warning for 0644 file, got empty")
	}
}
