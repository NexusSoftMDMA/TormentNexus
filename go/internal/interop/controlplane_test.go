package interop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tormentnexushq/tormentnexus-go/internal/lockfile"
)

func TestDiscoverControlPlanes(t *testing.T) {
	tempDir := t.TempDir()
	mainLockPath := filepath.Join(tempDir, "main-lock.json")
	goLockPath := filepath.Join(tempDir, "go-lock.json")

	if err := lockfile.Write(mainLockPath, lockfile.Record{
		Host:      "127.0.0.1",
		Port:      4100,
		Version:   "0.99.1",
		StartedAt: "2026-03-28T00:00:00Z",
	}); err != nil {
		t.Fatalf("failed to write main lock: %v", err)
	}

	statuses := DiscoverControlPlanes(mainLockPath, goLockPath)
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}

	if !statuses[0].Running || statuses[0].Port != 4100 {
		t.Fatalf("expected running node control plane, got %+v", statuses[0])
	}

	if statuses[1].Running {
		t.Fatalf("expected go control plane to be absent, got %+v", statuses[1])
	}
}

func TestReadImportedInstructions(t *testing.T) {
	tempDir := t.TempDir()
	docPath := filepath.Join(tempDir, "auto-imported-agent-instructions.md")
	content := "# Auto-imported Agent Instructions\n\n- Prefer safe ports.\n"

	if err := os.WriteFile(docPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write instructions doc: %v", err)
	}

	result := ReadImportedInstructions(docPath)
	if !result.Available {
		t.Fatalf("expected instructions to be available, got %+v", result)
	}

	if result.Content != content {
		t.Fatalf("expected content %q, got %q", content, result.Content)
	}
}
