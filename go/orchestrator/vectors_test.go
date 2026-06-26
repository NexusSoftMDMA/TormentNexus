package orchestrator

import (
	"math"
	"os"
	"testing"

	"github.com/NexusSoftMDMA/TormentNexus/agents"
)

func TestVectorDatabaseSimilarityLoops(t *testing.T) {
	dbPath := "./.test_vector_rag.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	vdb, err := NewVectorDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to spin up native SQLite RAG container: %v", err)
	}

	// 1. Store synthetic tool records natively
	toolA := agents.Tool{Name: "execute_sql", Description: "SQL query parser"}
	err = vdb.StoreTool(toolA, []float64{1.0, 0.0, 0.5})
	if err != nil {
		t.Errorf("Failed JSON vector injection schema: %v", err)
	}

	toolB := agents.Tool{Name: "git_diff", Description: "Version control"}
	err = vdb.StoreTool(toolB, []float64{0.0, 1.0, 0.0})

	// 2. Perform native Cosine mathematical check
	intentVector := []float64{0.9, 0.1, 0.4} // Should strongly correlate to execute_sql

	results, err := vdb.Search(intentVector, 1)
	if err != nil {
		t.Fatalf("Vector similarity query failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("Similarity parser returned empty slice incorrectly.")
	}

	if results[0].ToolName != "execute_sql" {
		t.Errorf("Cosine logic misaligned. Expected execute_sql, got %s", results[0].ToolName)
	}

	// Hard mathematical test verifying pure Go float64 execution bounds
	valA := []float64{1.0, 2.0, 3.0}
	valB := []float64{1.0, 2.0, 3.0}
	sim := cosineSimilarity(valA, valB)
	if math.Abs(sim-1.0) > 0.001 {
		t.Errorf("L2 normative check failed perfect correlation. Expected 1.0, got %v", sim)
	}
}
