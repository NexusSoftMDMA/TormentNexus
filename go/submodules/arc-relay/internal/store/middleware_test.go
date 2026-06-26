package store_test

import (
	"encoding/json"
	"testing"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func setupMiddlewareTest(t *testing.T) (*store.MiddlewareStore, string) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	mwStore := store.NewMiddlewareStore(db)

	// Insert a server for FK reference
	_, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test-server', 'Test Server', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("inserting test server: %v", err)
	}
	return mwStore, "srv-1"
}

func TestUpsert_CreatesAndUpdates(t *testing.T) {
	mwStore, serverID := setupMiddlewareTest(t)

	mc := &store.MiddlewareConfig{
		ServerID:   &serverID,
		Middleware: "sanitizer",
		Enabled:    true,
		Config:     json.RawMessage(`{"patterns":["SSN"]}`),
		Priority:   10,
	}

	if err := mwStore.Upsert(mc); err != nil {
		t.Fatalf("initial upsert: %v", err)
	}

	// Verify it was created
	configs, err := mwStore.GetForServer(serverID)
	if err != nil {
		t.Fatalf("GetForServer: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if string(configs[0].Config) != `{"patterns":["SSN"]}` {
		t.Errorf("config = %s, want {\"patterns\":[\"SSN\"]}", configs[0].Config)
	}

	// Update via upsert
	mc.Config = json.RawMessage(`{"patterns":["SSN","email"]}`)
	if err := mwStore.Upsert(mc); err != nil {
		t.Fatalf("update upsert: %v", err)
	}

	configs, _ = mwStore.GetForServer(serverID)
	if string(configs[0].Config) != `{"patterns":["SSN","email"]}` {
		t.Errorf("updated config = %s", configs[0].Config)
	}
}

func TestUpsertEnabled_PreservesConfig(t *testing.T) {
	mwStore, serverID := setupMiddlewareTest(t)

	// First, create a config with real data
	mc := &store.MiddlewareConfig{
		ServerID:   &serverID,
		Middleware: "archive",
		Enabled:    true,
		Config:     json.RawMessage(`{"url":"https://example.com","auth_type":"bearer","auth_value":"secret"}`),
		Priority:   40,
	}
	if err := mwStore.Upsert(mc); err != nil {
		t.Fatalf("initial upsert: %v", err)
	}

	// Now toggle enabled off - should NOT overwrite config
	if err := mwStore.UpsertEnabled(serverID, "archive", false, 40); err != nil {
		t.Fatalf("UpsertEnabled: %v", err)
	}

	configs, _ := mwStore.GetForServer(serverID)
	var found *store.MiddlewareConfig
	for _, c := range configs {
		if c.Middleware == "archive" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("archive config not found after toggle")
	}
	if found.Enabled {
		t.Error("expected enabled=false after toggle")
	}
	if string(found.Config) != `{"url":"https://example.com","auth_type":"bearer","auth_value":"secret"}` {
		t.Errorf("config was overwritten by toggle: %s", found.Config)
	}
}

func TestUpsertEnabled_CreatesNewRow(t *testing.T) {
	mwStore, serverID := setupMiddlewareTest(t)

	// Toggle on a middleware that doesn't exist yet
	if err := mwStore.UpsertEnabled(serverID, "archive", true, 40); err != nil {
		t.Fatalf("UpsertEnabled (new): %v", err)
	}

	configs, _ := mwStore.GetForServer(serverID)
	var found *store.MiddlewareConfig
	for _, c := range configs {
		if c.Middleware == "archive" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("archive config not found after first-time toggle")
	}
	if !found.Enabled {
		t.Error("expected enabled=true")
	}
	if string(found.Config) != "{}" {
		t.Errorf("expected empty config for new row, got %s", found.Config)
	}
}

func TestUpsertGlobal_CreateAndUpdate(t *testing.T) {
	mwStore, _ := setupMiddlewareTest(t)

	cfg := json.RawMessage(`{"url":"https://compliance.example.com","auth_type":"bearer","auth_value":"token123"}`)

	// Create global config
	mc := &store.MiddlewareConfig{
		Middleware: "archive",
		Enabled:    true,
		Config:     cfg,
		Priority:   40,
	}
	if err := mwStore.UpsertGlobal(mc); err != nil {
		t.Fatalf("UpsertGlobal (create): %v", err)
	}

	global, err := mwStore.GetGlobal("archive")
	if err != nil {
		t.Fatalf("GetGlobal: %v", err)
	}
	if global == nil {
		t.Fatal("global archive config not found")
	}
	if string(global.Config) != string(cfg) {
		t.Errorf("global config = %s, want %s", global.Config, cfg)
	}

	// Update global config
	newCfg := json.RawMessage(`{"url":"https://new.example.com","auth_type":"api_key","auth_value":"key456"}`)
	mc.Config = newCfg
	if err := mwStore.UpsertGlobal(mc); err != nil {
		t.Fatalf("UpsertGlobal (update): %v", err)
	}

	global, _ = mwStore.GetGlobal("archive")
	if string(global.Config) != string(newCfg) {
		t.Errorf("updated global config = %s, want %s", global.Config, newCfg)
	}
}

func TestUpsertGlobal_NoDuplicates(t *testing.T) {
	mwStore, _ := setupMiddlewareTest(t)

	mc := &store.MiddlewareConfig{
		Middleware: "archive",
		Enabled:    true,
		Config:     json.RawMessage(`{"url":"https://example.com"}`),
		Priority:   40,
	}

	// Call UpsertGlobal twice
	if err := mwStore.UpsertGlobal(mc); err != nil {
		t.Fatalf("first UpsertGlobal: %v", err)
	}
	mc.Config = json.RawMessage(`{"url":"https://updated.example.com"}`)
	if err := mwStore.UpsertGlobal(mc); err != nil {
		t.Fatalf("second UpsertGlobal: %v", err)
	}

	// Verify only one global row exists
	global, _ := mwStore.GetGlobal("archive")
	if global == nil {
		t.Fatal("no global config found")
	}
	if string(global.Config) != `{"url":"https://updated.example.com"}` {
		t.Errorf("global config = %s", global.Config)
	}
}

func TestGetForServer_InheritsGlobalConfig(t *testing.T) {
	mwStore, serverID := setupMiddlewareTest(t)

	// Create global archive config (enabled=false, config-only container)
	globalMC := &store.MiddlewareConfig{
		Middleware: "archive",
		Enabled:    false,
		Config:     json.RawMessage(`{"url":"https://global.example.com","auth_type":"bearer","auth_value":"globaltoken"}`),
		Priority:   40,
	}
	if err := mwStore.UpsertGlobal(globalMC); err != nil {
		t.Fatalf("UpsertGlobal: %v", err)
	}

	// Disabled global should NOT appear when no server row exists
	configs, err := mwStore.GetForServer(serverID)
	if err != nil {
		t.Fatalf("GetForServer (no server row): %v", err)
	}
	for _, c := range configs {
		if c.Middleware == "archive" {
			t.Error("disabled global archive should not appear without server row")
		}
	}

	// Create per-server toggle with empty config
	if err := mwStore.UpsertEnabled(serverID, "archive", true, 40); err != nil {
		t.Fatalf("UpsertEnabled: %v", err)
	}

	// GetForServer should return server's enabled state with global's config
	configs, err = mwStore.GetForServer(serverID)
	if err != nil {
		t.Fatalf("GetForServer: %v", err)
	}

	var found *store.MiddlewareConfig
	for _, c := range configs {
		if c.Middleware == "archive" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("archive not found in merged configs")
	}
	if !found.Enabled {
		t.Error("expected enabled=true (from server row)")
	}
	if string(found.Config) != `{"url":"https://global.example.com","auth_type":"bearer","auth_value":"globaltoken"}` {
		t.Errorf("expected inherited global config, got %s", found.Config)
	}
	// Verify server_id is set (server row, not global)
	if found.ServerID == nil {
		t.Error("expected server-specific row, got global")
	}
}

func TestGetForServer_ServerConfigOverridesGlobal(t *testing.T) {
	mwStore, serverID := setupMiddlewareTest(t)

	// Create global archive config
	globalMC := &store.MiddlewareConfig{
		Middleware: "archive",
		Enabled:    true,
		Config:     json.RawMessage(`{"url":"https://global.example.com"}`),
		Priority:   40,
	}
	if err := mwStore.UpsertGlobal(globalMC); err != nil {
		t.Fatalf("UpsertGlobal: %v", err)
	}

	// Create per-server config with actual config (not empty)
	serverMC := &store.MiddlewareConfig{
		ServerID:   &serverID,
		Middleware: "archive",
		Enabled:    true,
		Config:     json.RawMessage(`{"url":"https://server-specific.example.com"}`),
		Priority:   40,
	}
	if err := mwStore.Upsert(serverMC); err != nil {
		t.Fatalf("Upsert server config: %v", err)
	}

	configs, _ := mwStore.GetForServer(serverID)
	var found *store.MiddlewareConfig
	for _, c := range configs {
		if c.Middleware == "archive" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("archive not found")
	}
	// Server-specific config should win over global
	if string(found.Config) != `{"url":"https://server-specific.example.com"}` {
		t.Errorf("expected server config, got %s", found.Config)
	}
}
