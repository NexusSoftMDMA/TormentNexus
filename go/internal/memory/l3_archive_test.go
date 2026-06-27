package memory

import (
	"os"
	"testing"
)

func TestL3Archive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "l3archive_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	archive := NewL3Archive(tempDir)

	memories := []*Memory{
		{ID: "test-1", Content: "Hello World", Kind: KindFact, Tier: TierLongTerm},
		{ID: "test-2", Content: "Hello World 2", Kind: KindFact, Tier: TierLongTerm},
	}

	if err := archive.Archive(memories); err != nil {
		t.Fatalf("Failed to archive: %v", err)
	}

	loaded, err := archive.Unarchive()
	if err != nil {
		t.Fatalf("Failed to unarchive: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("Expected 2 unarchived memories, got %d", len(loaded))
	}

	if loaded[0].Content != "Hello World" {
		t.Fatalf("Content mismatch")
	}
}
