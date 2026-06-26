package mcp

import (
	"context"
	"testing"
	"time"
)

func TestNewNativeMCPRouter(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	if router == nil {
		t.Fatal("expected non-nil router")
	}
	if router.workingSet == nil {
		t.Error("expected non-nil working set")
	}
	if router.catalog == nil {
		t.Error("expected non-nil catalog")
	}
}

func TestNativeMCPRouterCatalog(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	router.catalog.Add(CatalogEntry{
		Name:         "test__bash",
		OriginalName: "bash",
		Server:       "test",
		Description:  "Run bash commands",
		AlwaysOn:     true,
	})

	if router.catalog.Size() != 1 {
		t.Errorf("expected catalog size 1, got %d", router.catalog.Size())
	}

	entry, ok := router.catalog.Get("test__bash")
	if !ok {
		t.Fatal("expected to find test__bash in catalog")
	}
	if entry.OriginalName != "bash" {
		t.Errorf("expected OriginalName=bash, got %s", entry.OriginalName)
	}
}

func TestNativeMCPRouterWorkingSet(t *testing.T) {
	ws := NewWorkingSet(3, 2, []string{"always"})

	// Load always-on tool
	evicted := ws.Load(WorkingSetEntry{
		Name:     "always",
		AlwaysOn: true,
	})
	if len(evicted) != 0 {
		t.Error("should not evict for first load")
	}

	// Load regular tools
	ws.Load(WorkingSetEntry{Name: "tool1"})
	ws.Load(WorkingSetEntry{Name: "tool2"})

	// Working set should have 3 entries
	if len(ws.List()) != 3 {
		t.Errorf("expected 3 entries, got %d", len(ws.List()))
	}

	// Load one more — should evict tool1 or tool2 (not always)
	evicted = ws.Load(WorkingSetEntry{Name: "tool3"})
	if len(evicted) == 0 {
		t.Error("expected eviction when over capacity")
	}

	// Always-on should still be there
	if _, ok := ws.Get("always"); !ok {
		t.Error("always-on tool should not be evicted")
	}
}

func TestNativeMCPRouterWorkingSetUnload(t *testing.T) {
	ws := NewWorkingSet(10, 5, []string{"always_on"})

	ws.Load(WorkingSetEntry{Name: "always_on", AlwaysOn: true})
	ws.Load(WorkingSetEntry{Name: "removable"})

	// Cannot unload always-on
	if ws.Unload("always_on") {
		t.Error("should not be able to unload always-on tool")
	}

	// Can unload regular tool
	if !ws.Unload("removable") {
		t.Error("should be able to unload regular tool")
	}

	// Unloading non-existent returns false
	if ws.Unload("nonexistent") {
		t.Error("should return false for non-existent tool")
	}
}

func TestNativeMCPRouterWorkingSetTouch(t *testing.T) {
	ws := NewWorkingSet(10, 5, nil)

	ws.Load(WorkingSetEntry{Name: "tool1"})
	time.Sleep(1 * time.Millisecond)
	ws.Load(WorkingSetEntry{Name: "tool2"})

	ws.Touch("tool1")

	entries := ws.List()
	if len(entries) < 2 {
		t.Fatal("expected at least 2 entries")
	}

	// After touching tool1, it should have a higher use count
	entry, _ := ws.Get("tool1")
	if entry.UseCount != 1 {
		t.Errorf("expected UseCount=1, got %d", entry.UseCount)
	}
}

func TestNativeMCPRouterLoadTool(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	router.catalog.Add(CatalogEntry{
		Name:         "srv__echo",
		OriginalName: "echo",
		Server:       "srv",
		Description:  "Echo input back",
	})

	_, err := router.LoadTool(context.Background(), "srv__echo")
	if err != nil {
		t.Fatalf("LoadTool failed: %v", err)
	}

	entry, ok := router.workingSet.Get("srv__echo")
	if !ok {
		t.Fatal("expected tool to be in working set")
	}
	if entry.OriginalName != "echo" {
		t.Errorf("expected OriginalName=echo, got %s", entry.OriginalName)
	}
}

func TestNativeMCPRouterUnloadTool(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	router.catalog.Add(CatalogEntry{Name: "srv__echo", OriginalName: "echo", Server: "srv"})
	router.LoadTool(context.Background(), "srv__echo")

	if err := router.UnloadTool("srv__echo"); err != nil {
		t.Fatalf("UnloadTool failed: %v", err)
	}

	if _, ok := router.workingSet.Get("srv__echo"); ok {
		t.Error("tool should not be in working set after unload")
	}
}

func TestNativeMCPRouterCallToolNotLoaded(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	_, err := router.CallTool(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error when calling tool not in working set")
	}
}

func TestNativeMCPRouterRefreshCatalog(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	inventory := &Inventory{
		Tools: []ToolEntry{
			{Name: "srv__bash", OriginalName: "bash", Server: "srv", AlwaysOn: true, Description: "Run commands"},
			{Name: "srv__read", OriginalName: "read", Server: "srv", Description: "Read files"},
		},
	}

	count := router.RefreshCatalog(inventory)
	if count != 2 {
		t.Errorf("expected 2 catalog entries, got %d", count)
	}

	// Always-on tool should be in working set
	if _, ok := router.workingSet.Get("srv__bash"); !ok {
		t.Error("always-on tool should be auto-loaded into working set")
	}
}

func TestNativeMCPRouterEvents(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	router.catalog.Add(CatalogEntry{Name: "srv__test", OriginalName: "test", Server: "srv"})
	router.LoadTool(context.Background(), "srv__test")

	events := router.GetEvents(10)
	if len(events) == 0 {
		t.Error("expected at least one event after loading a tool")
	}

	found := false
	for _, e := range events {
		if e.Type == "load" && e.ToolName == "srv__test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a 'load' event for srv__test")
	}
}

func TestNativeMCPRouterMarshalState(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewNativeMCPRouter(nil, nil, cfg)

	router.catalog.Add(CatalogEntry{Name: "srv__test", OriginalName: "test", Server: "srv"})

	state := router.MarshalState()
	if len(state) == 0 {
		t.Error("expected non-empty state JSON")
	}
}

func TestNativeMCPRouterAutoLoad(t *testing.T) {
	cfg := DefaultRouterConfig()
	cfg.AutoLoadMinConfidence = 0.7
	router := NewNativeMCPRouter(nil, nil, cfg)

	ranked := []NativeRankedTool{
		{Name: "high_conf", Score: 0.95, AutoLoadable: true, Server: "srv", OriginalName: "high"},
		{Name: "low_conf", Score: 0.3, AutoLoadable: false, Server: "srv", OriginalName: "low"},
	}

	loaded := router.AutoLoadTools(context.Background(), ranked)
	if len(loaded) != 1 {
		t.Fatalf("expected 1 auto-loaded tool, got %d", len(loaded))
	}
	if loaded[0] != "high_conf" {
		t.Errorf("expected 'high_conf' to be auto-loaded, got %s", loaded[0])
	}

	if _, ok := router.workingSet.Get("high_conf"); !ok {
		t.Error("high confidence tool should be in working set")
	}
	if _, ok := router.workingSet.Get("low_conf"); ok {
		t.Error("low confidence tool should not be in working set")
	}
}

func TestWorkingSetLRUEviction(t *testing.T) {
	ws := NewWorkingSet(3, 10, nil)

	ws.Load(WorkingSetEntry{Name: "a", LastUsedAt: time.Now().Add(-3 * time.Second)})
	ws.Load(WorkingSetEntry{Name: "b", LastUsedAt: time.Now().Add(-2 * time.Second)})
	ws.Load(WorkingSetEntry{Name: "c", LastUsedAt: time.Now().Add(-1 * time.Second)})

	// Load a 4th entry — should evict 'a' (oldest)
	evicted := ws.Load(WorkingSetEntry{Name: "d"})
	if len(evicted) == 0 {
		t.Error("expected eviction")
	}
	if len(evicted) > 0 && evicted[0] != "a" {
		t.Errorf("expected 'a' to be evicted (oldest), got %q", evicted[0])
	}

	// 'd' should be present (just loaded)
	if _, ok := ws.Get("d"); !ok {
		t.Error("newly loaded tool 'd' should be present")
	}

	// Working set should have exactly 3 entries
	if len(ws.List()) != 3 {
		t.Errorf("expected 3 entries after eviction, got %d", len(ws.List()))
	}
}
