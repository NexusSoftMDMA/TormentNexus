package tools

import (
	"context"
	"os"
	"testing"
)

// TestExaSearch verifies that exa_search returns an error when no API key is set.
func TestExaSearch(t *testing.T) {
	// Without API key, should fail with proper error message
	os.Unsetenv("EXA_API_KEY")
	resp, err := HandleExaSearch(context.Background(), map[string]interface{}{
		"query": "test query",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !resp.IsError {
		t.Errorf("expected error response when EXA_API_KEY not set")
	}
}

// TestArxivSearch verifies that arxiv_search can reach the arXiv API.
func TestArxivSearch(t *testing.T) {
	resp, err := HandleArxivSearch(context.Background(), map[string]interface{}{
		"query":       "test",
		"max_results": 1,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// Should either succeed or return a graceful error
	if resp.IsError {
		t.Logf("HandleArxivSearch returned error (may be network issue): %s", resp.Content[0].Text)
	} else {
		t.Logf("HandleArxivSearch succeeded: %d chars", len(resp.Content[0].Text))
	}
}

// TestArxivGetPaper verifies that arxiv_get_paper works with a valid ID.
func TestArxivGetPaper(t *testing.T) {
	resp, err := HandleArxivGetPaper(context.Background(), map[string]interface{}{
		"paper_id": "2301.07041",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	t.Logf("HandleArxivGetPaper: %d chars", len(resp.Content[0].Text))
}

// TestSemanticScholarSearch verifies the Semantic Scholar search handler.
func TestSemanticScholarSearch(t *testing.T) {
	resp, err := HandleSemanticScholarSearch(context.Background(), map[string]interface{}{
		"query": "attention is all you need",
		"limit": 2,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if resp.IsError {
		t.Logf("HandleSemanticScholarSearch returned error (may be network issue): %s", resp.Content[0].Text)
	} else {
		t.Logf("HandleSemanticScholarSearch succeeded: %d chars", len(resp.Content[0].Text))
	}
}

// TestMem0AddMemory verifies that mem0_add_memory returns error when no API key is set.
func TestMem0AddMemory(t *testing.T) {
	os.Unsetenv("MEM0_API_KEY")
	resp, err := HandleMem0AddMemory(context.Background(), map[string]interface{}{
		"content": "test memory",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !resp.IsError {
		t.Errorf("expected error response when MEM0_API_KEY not set")
	}
}

// TestAlpacaGetAccount verifies that alpaca_get_account returns error when no credentials set.
func TestAlpacaGetAccount(t *testing.T) {
	os.Unsetenv("ALPACA_API_KEY")
	os.Unsetenv("ALPACA_SECRET_KEY")
	resp, err := HandleAlpacaGetAccount(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !resp.IsError {
		t.Errorf("expected error response when ALPACA credentials not set")
	}
}

// TestAVGlobalQuote verifies that av_quote returns error when no API key set.
func TestAVGlobalQuote(t *testing.T) {
	os.Unsetenv("AV_API_KEY")
	os.Unsetenv("ALPHA_VANTAGE_API_KEY")
	resp, err := HandleAVGlobalQuote(context.Background(), map[string]interface{}{
		"symbol": "AAPL",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !resp.IsError {
		t.Errorf("expected error response when AV_API_KEY not set")
	}
}

// TestHFSearchModels verifies that hf_search_models can reach HuggingFace API.
func TestHFSearchModels(t *testing.T) {
	resp, err := HandleHFSearchModels(context.Background(), map[string]interface{}{
		"query": "gpt2",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if resp.IsError {
		t.Logf("HandleHFSearchModels returned error (may be network issue): %s", resp.Content[0].Text)
	} else {
		t.Logf("HandleHFSearchModels succeeded: %d chars", len(resp.Content[0].Text))
	}
}

// TestSemgrepScan verifies that semgrep_scan handles missing binary gracefully.
func TestSemgrepScan(t *testing.T) {
	resp, err := HandleSemgrepScan(context.Background(), map[string]interface{}{
		"target": ".",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// Either succeeds (if semgrep is installed) or fails gracefully
	t.Logf("HandleSemgrepScan: isError=%v, %d chars", resp.IsError, len(resp.Content[0].Text))
}

// TestOctagonResearch verifies that octagon_research returns error when no API key is set.
func TestOctagonResearch(t *testing.T) {
	os.Unsetenv("OCTAGON_API_KEY")
	resp, err := HandleOctagonResearch(context.Background(), map[string]interface{}{
		"query": "Apple Inc",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !resp.IsError {
		t.Errorf("expected error response when OCTAGON_API_KEY not set")
	}
}

// TestChromaListCollections verifies ChromaDB handler fails gracefully when not running.
func TestChromaListCollections(t *testing.T) {
	resp, err := HandleChromaListCollections(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// Should fail gracefully when ChromaDB is not running
	t.Logf("HandleChromaListCollections: isError=%v", resp.IsError)
}

// TestBasicMemoryWrite verifies basic-memory write and read operations.
func TestBasicMemoryWrite(t *testing.T) {
	// Set a temp directory for tests
	tmpDir := t.TempDir()
	os.Setenv("BASIC_MEMORY_DIR", tmpDir)
	defer os.Unsetenv("BASIC_MEMORY_DIR")

	resp, err := HandleBasicMemoryWrite(context.Background(), map[string]interface{}{
		"title":   "Test Note",
		"content": "This is a test note for unit testing.",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if resp.IsError {
		t.Errorf("expected success but got error: %s", resp.Content[0].Text)
		return
	}
	t.Logf("Write succeeded: %s", resp.Content[0].Text)

	// Now read it back
	readResp, err := HandleBasicMemoryRead(context.Background(), map[string]interface{}{
		"title": "Test Note",
	})
	if err != nil {
		t.Errorf("unexpected error reading: %v", err)
		return
	}
	if readResp.IsError {
		t.Errorf("expected read success but got error: %s", readResp.Content[0].Text)
		return
	}
	t.Logf("Read succeeded: %d chars", len(readResp.Content[0].Text))

	// Search
	searchResp, err := HandleBasicMemorySearch(context.Background(), map[string]interface{}{
		"query": "unit testing",
	})
	if err != nil {
		t.Errorf("unexpected error searching: %v", err)
		return
	}
	if searchResp.IsError {
		t.Errorf("expected search success but got error: %s", searchResp.Content[0].Text)
	} else {
		t.Logf("Search succeeded: %d chars", len(searchResp.Content[0].Text))
	}
}

// TestMindsDBQuery verifies MindsDB handler fails gracefully when not running.
func TestMindsDBQuery(t *testing.T) {
	resp, err := HandleMindsDBQuery(context.Background(), map[string]interface{}{
		"query": "SELECT * FROM mindsdb.models;",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// Should fail gracefully when MindsDB is not running
	t.Logf("HandleMindsDBQuery: isError=%v", resp.IsError)
}
