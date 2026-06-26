package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTormentNexusAdapterBuildsStatusWithoutPanicking(t *testing.T) {
	dir := t.TempDir()
	tormentnexusDir := filepath.Join(dir, "..", "tormentnexus")
	if err := os.MkdirAll(tormentnexusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tormentnexusDir, "README.md"), []byte("# TormentNexus"), 0o644); err != nil {
		t.Fatal(err)
	}
	adapter := NewTormentNexusAdapter(dir)
	status := adapter.Status()
	if !status.Assimilated {
		t.Fatal("expected assimilated tormentnexus adapter")
	}
	if status.MemoryContext == "" {
		t.Fatal("expected memory context")
	}
	if status.Provider.CurrentProvider == "" {
		t.Fatal("expected provider status")
	}
	if status.TormentNexusRepoPath == "" {
		t.Fatal("expected discovered tormentnexus repo path")
	}
	if adapter.RouteMCP("list tools") == "" {
		t.Fatal("expected routed MCP string")
	}
	if adapter.BuildSystemContext() == "" {
		t.Fatal("expected system context")
	}
}
