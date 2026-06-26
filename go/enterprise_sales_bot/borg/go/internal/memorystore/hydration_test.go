package memorystore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHydrationStoreNew(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	if len(hs.All()) != 0 {
		t.Error("new store should be empty")
	}
	if len(hs.SectionNames()) != 0 {
		t.Error("new store should have no sections")
	}
}

func TestHydrationStoreAddAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	hs.Add(HydrationEntry{
		Section: "project_context",
		Key:     "test-key",
		Content: "Test content for hydration",
		Source:  "test",
		Tags:    []string{"test"},
	})

	if len(hs.All()) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(hs.All()))
	}

	entries := hs.Get("project_context", "")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in project_context section, got %d", len(entries))
	}
	if entries[0].Key != "test-key" {
		t.Errorf("expected key 'test-key', got %q", entries[0].Key)
	}
	if entries[0].Content != "Test content for hydration" {
		t.Errorf("unexpected content: %q", entries[0].Content)
	}

	// Verify auto-generated fields
	if entries[0].ID == "" {
		t.Error("expected auto-generated ID")
	}
	if entries[0].CreatedAt == "" {
		t.Error("expected auto-generated createdAt")
	}
}

func TestHydrationStoreGetByKey(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	hs.Add(HydrationEntry{Section: "test", Key: "alpha", Content: "first"})
	hs.Add(HydrationEntry{Section: "test", Key: "beta", Content: "second"})
	hs.Add(HydrationEntry{Section: "test", Key: "alpha", Content: "duplicate"})

	entries := hs.Get("test", "alpha")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with key 'alpha', got %d", len(entries))
	}
}

func TestHydrationStoreQuery(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	hs.Add(HydrationEntry{Section: "architecture", Key: "go-pkgs", Content: "Go internal packages: ai, config, mcp"})
	hs.Add(HydrationEntry{Section: "project", Key: "version", Content: "Version: 1.0.0-alpha.49"})
	hs.Add(HydrationEntry{Section: "environment", Key: "runtime", Content: "OS: linux, Arch: amd64"})

	// Query by content
	results := hs.Query("mcp")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'mcp', got %d", len(results))
	}

	// Case-insensitive query
	results = hs.Query("ARCHITECTURE")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'ARCHITECTURE', got %d", len(results))
	}

	// Query by section name
	results = hs.Query("environment")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'environment', got %d", len(results))
	}
}

func TestHydrationStoreSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	hs.Add(HydrationEntry{Section: "test", Key: "persist", Content: "This should survive a reload"})
	if err := hs.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load from the same path
	hs2 := NewHydrationStore(tmpDir)
	all := hs2.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry after reload, got %d", len(all))
	}
	if all[0].Content != "This should survive a reload" {
		t.Errorf("content mismatch after reload: %q", all[0].Content)
	}
}

func TestHydrationStoreSectionCounts(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	hs.Add(HydrationEntry{Section: "a", Key: "1", Content: "x"})
	hs.Add(HydrationEntry{Section: "a", Key: "2", Content: "y"})
	hs.Add(HydrationEntry{Section: "b", Key: "3", Content: "z"})

	counts := hs.SectionCounts()
	if counts["a"] != 2 {
		t.Errorf("expected section 'a' count=2, got %d", counts["a"])
	}
	if counts["b"] != 1 {
		t.Errorf("expected section 'b' count=1, got %d", counts["b"])
	}
}

func TestHydrateFromWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Seed a go.mod
	goMod := `module github.com/test/project

go 1.22

require (
	github.com/some/dep v1.0.0
)
`
	if err := os.MkdirAll(filepath.Join(tmpDir, "go"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go", "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Seed AGENTS.md
	agentsMd := `# Agent Instructions\n- Follow TDD\n- Use Go 1.22\n`
	if err := os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsMd), 0644); err != nil {
		t.Fatal(err)
	}

	// Seed top-level dirs
	for _, dir := range []string{"src", "docs", "tests"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	hs := NewHydrationStore(tmpDir)
	report, err := hs.HydrateFromWorkspace(nil, tmpDir)
	if err != nil {
		t.Fatalf("HydrateFromWorkspace failed: %v", err)
	}

	if report.TotalEntries == 0 {
		t.Error("expected at least some hydration entries")
	}
	if report.ProjectContext == 0 {
		t.Error("expected project context entries from go.mod")
	}
	if report.AgentInstructions == 0 {
		t.Error("expected agent instruction entries from AGENTS.md")
	}
	if report.ArchitectureEntries == 0 {
		t.Error("expected architecture entries from directory scan")
	}

	// Verify the store was persisted
	data, err := os.ReadFile(filepath.Join(tmpDir, ".tormentnexus", "hydration", "context.json"))
	if err != nil {
		t.Fatalf("hydration store file should exist: %v", err)
	}

	var entries []HydrationEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("hydration store should be valid JSON: %v", err)
	}
	if len(entries) == 0 {
		t.Error("persisted entries should not be empty")
	}
}

func TestHydrationStoreNoDirtySave(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHydrationStore(tmpDir)

	// No changes — Save should be a no-op
	if err := hs.Save(); err != nil {
		t.Fatalf("Save on clean store should succeed: %v", err)
	}

	// File should not exist since no entries were added
	if _, err := os.Stat(hs.path); !os.IsNotExist(err) {
		t.Error("expected no file for clean store without entries")
	}
}
