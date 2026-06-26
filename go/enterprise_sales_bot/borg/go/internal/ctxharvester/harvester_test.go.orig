package ctxharvester

import (
	"strings"
	"testing"
	"time"
)

func TestHarvestAndRetrieve(t *testing.T) {
	cfg := DefaultHarvestConfig()
	h := NewContextHarvester(&cfg)

	chunks := h.Harvest(SourceActiveFile, "This is a test file. It contains some test content. The content is about testing.", nil)
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if chunks[0].Source != SourceActiveFile {
		t.Errorf("expected source=active-file, got %s", chunks[0].Source)
	}
	if chunks[0].TokenCount == 0 {
		t.Error("expected nonzero token count")
	}

	results := h.Retrieve("test", 0)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestPrune(t *testing.T) {
	cfg := HarvestConfig{
		PruneThreshold:    0.9,
		DecayRatePerHour:  1.0, // Fast decay
		MaxTokenBudget:    50000,
		CompactAfterMs:    30000,
		MaxChunksPerSource: 50,
	}
	h := NewContextHarvester(&cfg)
	h.Harvest(SourceMemory, "short content here.", nil)

	// Simulate aging by manually adjusting createdAt
	for _, c := range h.chunks {
		c.CreatedAt -= 2 * 60 * 60 * 1000 // 2 hours ago
	}

	pruned := h.Prune()
	if pruned == 0 {
		t.Error("expected some chunks to be pruned with high decay + low relevance")
	}
}

func TestCompact(t *testing.T) {
	cfg := HarvestConfig{
		CompactAfterMs:    100, // Compact after 100ms
		MaxTokenBudget:    50000,
		MaxChunksPerSource: 50,
		PruneThreshold:    0.01,
		DecayRatePerHour:  0.0,
	}
	h := NewContextHarvester(&cfg)
	h.Harvest(SourceGitDiff, "diff chunk one. ", nil)
	time.Sleep(10 * time.Millisecond)
	h.Harvest(SourceGitDiff, "diff chunk two. ", nil)

	// Age them past CompactAfterMs
	for _, c := range h.chunks {
		c.CreatedAt -= 200 // 200ms ago
	}

	compacted := h.Compact()
	if compacted < 1 {
		t.Errorf("expected at least 1 compaction, got %d", compacted)
	}
}

func TestGetReport(t *testing.T) {
	h := NewContextHarvester(nil)
	h.Harvest(SourceActiveFile, "content one two three four five six seven eight.", nil)
	h.Harvest(SourceGitDiff, "diff content four five six seven eight nine ten.", nil)

	report := h.GetReport()
	if report.TotalChunks < 2 {
		t.Errorf("expected at least 2 chunks, got %d", report.TotalChunks)
	}
	// Verify source breakdown has entries
	if len(report.SourceBreakdown) == 0 {
		t.Error("expected non-empty source breakdown")
	}
}

func TestSemanticChunk(t *testing.T) {
	text := "First sentence here with more words to exceed. Second sentence follows with even more words to pack. Third sentence ends with enough words already. Fourth sentence begins here now with more text. Fifth sentence concludes this test with final words."
	chunks := semanticChunk(text, 8, 2) // 8 words per chunk, 2 overlap
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
}

func TestEstimateTokens(t *testing.T) {
	n := estimateTokens("hello world")
	if n <= 0 {
		t.Error("expected positive token estimate")
	}
}

func TestClear(t *testing.T) {
	h := NewContextHarvester(nil)
	h.Harvest(SourceConversation, "test", nil)
	h.Clear()
	report := h.GetReport()
	if report.TotalChunks != 0 {
		t.Error("expected 0 chunks after clear")
	}
}

// Ensure strings import is used
var _ = strings.TrimSpace
