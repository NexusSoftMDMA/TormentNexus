package mcp

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestLoadConfigServersFromJSONC(t *testing.T) {
	configDir := t.TempDir()
	content := `// comment
{
  "mcpServers": {
    "core": {
      "command": "node",
      "args": ["server.js"],
      "env": {"MODE": "test"},
      "_meta": {
        "tools": [
          {"name": "search_tools", "description": "Search tools", "inputSchema": {"type": "object"}}
        ]
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.jsonc"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write mcp.jsonc: %v", err)
	}

	servers, err := loadConfigServers(configDir)
	if err != nil {
		t.Fatalf("loadConfigServers returned error: %v", err)
	}
	core, ok := servers["core"]
	if !ok {
		t.Fatalf("expected core server, got %#v", servers)
	}
	if core.Command != "node" || len(core.Args) != 1 || core.Args[0] != "server.js" {
		t.Fatalf("unexpected core config: %#v", core)
	}
	if len(core.Meta.Tools) != 1 || core.Meta.Tools[0].Name != "search_tools" {
		t.Fatalf("expected search_tools metadata, got %#v", core.Meta.Tools)
	}
}

func TestLoadInventoryFromConfigAndDatabase(t *testing.T) {
	workspace := t.TempDir()
	configDir := t.TempDir()

	configContent := `{
  "mcpServers": {
    "config-core": {
      "command": "node",
      "args": ["config-server.js"],
      "_meta": {
        "tools": [
          {"name": "config_search", "description": "Config search", "inputSchema": {"type": "object"}}
        ]
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(workspace, "tormentnexus.db"))
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE mcp_servers (
			uuid TEXT PRIMARY KEY,
			name TEXT,
			type TEXT,
			command TEXT,
			args BLOB,
			env BLOB,
			url TEXT,
			description TEXT,
			enabled BOOLEAN,
			always_on BOOLEAN
		);
		CREATE TABLE tools (
			name TEXT,
			description TEXT,
			mcp_server_uuid TEXT,
			always_on BOOLEAN,
			tool_schema BLOB
		);
		INSERT INTO mcp_servers (uuid, name, type, command, args, env, url, description, enabled, always_on)
		VALUES ('srv-db-1', 'db-core', 'STDIO', 'node', '["db-server.js"]', '{"MODE":"db"}', '', 'DB core', 1, 1);
		INSERT INTO tools (name, description, mcp_server_uuid, always_on, tool_schema)
		VALUES ('db_search', 'DB search', 'srv-db-1', 1, '{"type":"object"}');
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	inventory, err := LoadInventory(workspace, configDir)
	if err != nil {
		t.Fatalf("LoadInventory returned error: %v", err)
	}
	if inventory.Source != "database" {
		t.Fatalf("expected inventory source database, got %q", inventory.Source)
	}
	if len(inventory.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %#v", inventory.Servers)
	}
	if len(inventory.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %#v", inventory.Tools)
	}

	foundConfigTool := false
	foundDBTool := false
	for _, tool := range inventory.Tools {
		if tool.OriginalName == "config_search" && tool.Server == "config-core" {
			foundConfigTool = true
		}
		if tool.OriginalName == "db_search" && tool.Server == "db-core" && tool.AlwaysOn {
			foundDBTool = true
		}
	}
	if !foundConfigTool || !foundDBTool {
		t.Fatalf("expected both config and db tools, got %#v", inventory.Tools)
	}
}

func TestLoadInventoryWithCachePersistsAndReloadsWithoutLiveSources(t *testing.T) {
	workspace := t.TempDir()
	configDir := t.TempDir()
	cachePath := filepath.Join(t.TempDir(), "mcp_inventory_cache.json")

	configContent := `{
  "mcpServers": {
    "cache-core": {
      "command": "node",
      "args": ["cache-server.js"],
      "_meta": {
        "tools": [
          {"name": "search_tools", "description": "Cache search", "inputSchema": {"type": "object"}}
        ]
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	inventory, err := LoadInventoryWithCache(workspace, configDir, cachePath)
	if err != nil {
		t.Fatalf("LoadInventoryWithCache returned error: %v", err)
	}
	if inventory.Source != "config" {
		t.Fatalf("expected live inventory source config, got %q", inventory.Source)
	}
	if len(inventory.Tools) != 1 || inventory.Tools[0].OriginalName != "search_tools" {
		t.Fatalf("expected live inventory tool, got %#v", inventory.Tools)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected inventory cache file to be written: %v", err)
	}

	if err := os.Remove(filepath.Join(configDir, "mcp.json")); err != nil {
		t.Fatalf("failed to remove live config: %v", err)
	}

	reloaded, err := LoadInventoryWithCache(workspace, configDir, cachePath)
	if err != nil {
		t.Fatalf("LoadInventoryWithCache cache reload returned error: %v", err)
	}
	if reloaded.Source != "cache" {
		t.Fatalf("expected cached inventory source, got %q", reloaded.Source)
	}
	if reloaded.CachedAt == "" {
		t.Fatalf("expected cachedAt to be populated on cached inventory reload")
	}
	if len(reloaded.Tools) != 1 || reloaded.Tools[0].OriginalName != "search_tools" {
		t.Fatalf("expected cached inventory tool after reload, got %#v", reloaded.Tools)
	}
}

func TestSyncInventoryCacheFromLiveSourcesRemovesStaleCacheWhenSourcesAreEmpty(t *testing.T) {
	workspace := t.TempDir()
	configDir := t.TempDir()
	cachePath := filepath.Join(t.TempDir(), "mcp_inventory_cache.json")
	staleCache := `{
  "version": 1,
  "cachedAt": "2026-04-04T00:00:00Z",
  "inventory": {
    "servers": [{"uuid":"config:core","name":"core","displayName":"core","type":"STDIO","enabled":true}],
    "tools": [{"name":"core__search_tools","description":"stale","server":"core","serverDisplayName":"core","advertisedName":"core__search_tools","originalName":"search_tools"}],
    "source": "config",
    "cachedAt": "2026-04-04T00:00:00Z"
  }
}`
	if err := os.WriteFile(cachePath, []byte(staleCache), 0o644); err != nil {
		t.Fatalf("failed to seed stale cache: %v", err)
	}

	inventory, err := SyncInventoryCacheFromLiveSources(workspace, configDir, cachePath)
	if err != nil {
		t.Fatalf("SyncInventoryCacheFromLiveSources returned error: %v", err)
	}
	if inventory.Source != "empty" || len(inventory.Tools) != 0 || len(inventory.Servers) != 0 {
		t.Fatalf("expected empty live inventory after cache sync, got %#v", inventory)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("expected stale cache file to be removed, stat err=%v", err)
	}
}

func TestSyncRuntimeOverlayCachePersistsOverlayWithoutInventory(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "mcp_inventory_cache.json")
	overlay := []RuntimeOverlayServer{{
		Name:                "runtime-core",
		Command:             "node",
		Args:                []string{"runtime-server.js"},
		RuntimeConnected:    true,
		ToolCount:           1,
		ToolInventoryStatus: "live-probed",
		IntegrationLevel:    "runtime-added",
		Source:              "go-runtime-registry",
		LastCheckedAt:       "2026-04-04T00:00:05Z",
		Tools:               []MetadataTool{{Name: "runtime_search", Description: "Runtime search", InputSchema: map[string]any{"type": "object"}}},
	}}
	if err := SyncRuntimeOverlayCache(cachePath, overlay); err != nil {
		t.Fatalf("SyncRuntimeOverlayCache returned error: %v", err)
	}
	snapshot, err := LoadInventoryCacheSnapshot(cachePath)
	if err != nil {
		t.Fatalf("LoadInventoryCacheSnapshot returned error: %v", err)
	}
	if len(snapshot.RuntimeOverlay) != 1 || snapshot.RuntimeOverlay[0].Name != "runtime-core" {
		t.Fatalf("expected persisted runtime overlay, got %#v", snapshot.RuntimeOverlay)
	}
	if len(snapshot.Inventory.Tools) != 0 {
		t.Fatalf("expected no persisted base inventory tools, got %#v", snapshot.Inventory.Tools)
	}

	if err := SyncRuntimeOverlayCache(cachePath, nil); err != nil {
		t.Fatalf("SyncRuntimeOverlayCache clear returned error: %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("expected overlay-only cache file to be removed after clear, stat err=%v", err)
	}
}
