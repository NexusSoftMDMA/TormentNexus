package sessionimport

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func strPtr(value string) *string {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

func TestImportedSessionStoreUpsertListGetAndDedup(t *testing.T) {
	workspace := t.TempDir()
	store := NewImportedSessionStore(workspace)
	ctx := context.Background()

	created, err := store.UpsertSession(ctx, ImportedSessionRecordInput{
		SourceTool:        "antigravity",
		SourcePath:        "C:/tmp/session-1.jsonl",
		ExternalSessionID: strPtr("session-1"),
		Title:             strPtr("Imported Session"),
		SessionFormat:     "jsonl",
		Transcript:        "User: keep durable defaults\nAssistant: Always prefer port 4100.",
		Excerpt:           strPtr("Always prefer port 4100."),
		WorkingDirectory:  strPtr("C:/tmp"),
		TranscriptHash:    "abcd1234ef567890abcd1234ef567890",
		NormalizedSession: map[string]any{
			"sourceTool": "antigravity",
		},
		Metadata: map[string]any{
			"retentionSummary": map[string]any{"archiveDisposition": "archive_only"},
		},
		DiscoveredAt:   1712000000000,
		ImportedAt:     1712000001000,
		LastModifiedAt: int64Ptr(1712000002000),
		ParsedMemories: []ImportedSessionMemoryInput{
			{Kind: ImportedSessionMemoryKindInstruction, Content: "Always prefer port 4100.", Tags: []string{"ports", "defaults"}, Source: ImportedSessionMemorySourceHeuristic, Metadata: map[string]any{"sourceTool": "antigravity", "path": "C:/tmp/session-1.jsonl"}},
			{Kind: ImportedSessionMemoryKindMemory, Content: "The operator prefers Go-first migration.", Tags: []string{"go", "migration"}, Source: ImportedSessionMemorySourceLLM},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created record, got %#v", created)
	}
	if created.Transcript != "User: keep durable defaults\nAssistant: Always prefer port 4100." {
		t.Fatalf("unexpected transcript: %#v", created)
	}
	if len(created.ParsedMemories) != 2 {
		t.Fatalf("expected parsed memories, got %#v", created.ParsedMemories)
	}

	listed, err := store.ListImportedSessions(ctx, 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one imported session, got %#v", listed)
	}

	fetched, err := store.GetImportedSession(ctx, created.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if fetched == nil || fetched.TranscriptHash != "abcd1234ef567890abcd1234ef567890" {
		t.Fatalf("unexpected fetched record: %#v", fetched)
	}
	if fetched.Metadata["archiveFormat"] != "gzip-text-v1" {
		t.Fatalf("expected archive metadata, got %#v", fetched.Metadata)
	}

	updated, err := store.UpsertSession(ctx, ImportedSessionRecordInput{
		SourceTool:        "antigravity",
		SourcePath:        "C:/tmp/session-1.jsonl",
		ExternalSessionID: strPtr("session-1"),
		Title:             strPtr("Imported Session Updated"),
		SessionFormat:     "jsonl",
		Transcript:        "User: keep durable defaults\nAssistant: Use port 4100 for compatibility.",
		Excerpt:           strPtr("Use port 4100 for compatibility."),
		WorkingDirectory:  strPtr("C:/tmp"),
		TranscriptHash:    "abcd1234ef567890abcd1234ef567890",
		NormalizedSession: map[string]any{"sourceTool": "antigravity", "updated": true},
		Metadata:          map[string]any{},
		DiscoveredAt:      1712000000000,
		ImportedAt:        1712000003000,
		ParsedMemories: []ImportedSessionMemoryInput{
			{Kind: ImportedSessionMemoryKindInstruction, Content: "Use port 4100 for compatibility.", Tags: []string{"ports"}, Source: ImportedSessionMemorySourceHeuristic, Metadata: map[string]any{"sourceTool": "antigravity", "path": "C:/tmp/session-1.jsonl"}},
		},
	})
	if err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("expected deduped session id reuse, got %q vs %q", updated.ID, created.ID)
	}
	if len(updated.ParsedMemories) != 1 || updated.ParsedMemories[0].Content != "Use port 4100 for compatibility." {
		t.Fatalf("expected replacement memories, got %#v", updated.ParsedMemories)
	}
}

func TestImportedSessionStoreInstructionDocsAndStats(t *testing.T) {
	workspace := t.TempDir()
	store := NewImportedSessionStore(workspace)
	ctx := context.Background()

	_, err := store.UpsertSession(ctx, ImportedSessionRecordInput{
		SourceTool:        "claude-code",
		SourcePath:        "C:/tmp/claude/session-2.jsonl",
		Title:             strPtr("Claude Session"),
		SessionFormat:     "jsonl",
		Transcript:        "Assistant: Always prefer PowerShell on Windows.",
		TranscriptHash:    "dcba4321ef567890dcba4321ef567890",
		NormalizedSession: map[string]any{"sourceTool": "claude-code"},
		Metadata:          map[string]any{},
		DiscoveredAt:      1712000010000,
		ImportedAt:        1712000011000,
		ParsedMemories: []ImportedSessionMemoryInput{
			{Kind: ImportedSessionMemoryKindInstruction, Content: "Always prefer PowerShell on Windows.", Tags: []string{"shell", "windows"}, Source: ImportedSessionMemorySourceHeuristic, Metadata: map[string]any{"sourceTool": "claude-code", "path": "C:/tmp/claude/session-2.jsonl"}},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	instructions, err := store.ListInstructionMemories(ctx, 10)
	if err != nil {
		t.Fatalf("list instruction memories failed: %v", err)
	}
	if len(instructions) != 1 {
		t.Fatalf("expected one instruction memory, got %#v", instructions)
	}

	doc, err := store.WriteInstructionDoc(ctx, 10)
	if err != nil {
		t.Fatalf("write instruction doc failed: %v", err)
	}
	if doc == nil {
		t.Fatalf("expected generated instruction doc")
	}
	content, err := os.ReadFile(doc.Path)
	if err != nil {
		t.Fatalf("failed reading generated doc: %v", err)
	}
	if !strings.Contains(string(content), "Always prefer PowerShell on Windows.") {
		t.Fatalf("expected generated doc content, got %s", string(content))
	}

	docs, err := store.ListInstructionDocs()
	if err != nil {
		t.Fatalf("list instruction docs failed: %v", err)
	}
	if len(docs) != 1 || filepath.Base(docs[0].Path) != "auto-imported-agent-instructions.md" {
		t.Fatalf("unexpected docs: %#v", docs)
	}

	stats, err := store.GetMaintenanceStats(ctx)
	if err != nil {
		t.Fatalf("maintenance stats failed: %v", err)
	}
	if stats.TotalSessions != 1 || stats.ArchivedTranscriptCount != 1 {
		t.Fatalf("unexpected maintenance stats: %#v", stats)
	}
	if stats.MissingRetentionSummaryCount != 1 {
		t.Fatalf("expected missing retention summary count 1, got %#v", stats)
	}
}
