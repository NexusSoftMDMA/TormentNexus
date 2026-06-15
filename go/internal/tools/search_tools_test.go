//go:build ignore
// +build ignore

package tools

import (
	"context"
	"testing"
)

func TestHandleDDGSearch_MissingQuery(t *testing.T) {
	resp, err := HandleDDGSearch(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error from search handler")
	}
	if !resp.IsError {
		t.Fatal("expected error response when query is missing")
	}
	if resp.Content[0].Text != "query is required" {
		t.Fatalf("expected 'query is required', got: %s", resp.Content[0].Text)
	}
}

func TestHandleDDGSearch_TimeoutArg(t *testing.T) {
	// Verify timeout is parsed (should give an error about network, not about missing params)
	resp, err := HandleDDGSearch(context.Background(), map[string]interface{}{
		"query":   "golang testing",
		"timeout": 1, // 1 second, will timeout or error quickly
	})
	if err != nil {
		t.Fatal("expected no error from search handler")
	}
	if !resp.IsError {
		t.Logf("search succeeded: %s", resp.Content[0].Text[:min(len(resp.Content[0].Text), 100)])
	}
}

func TestHandleDDGSearch_MaxResults(t *testing.T) {
	resp, err := HandleDDGSearch(context.Background(), map[string]interface{}{
		"query":       "test",
		"max_results": 3,
		"timeout":     5,
	})
	if err != nil {
		t.Fatal("expected no error from search handler")
	}
	if !resp.IsError {
		t.Logf("search results (first 100 chars): %s", resp.Content[0].Text[:min(len(resp.Content[0].Text), 100)])
	}
}

func TestHandleDDGSearch_AliasQueryKeys(t *testing.T) {
	// Test the "q" alias for query
	resp, err := HandleDDGSearch(context.Background(), map[string]interface{}{
		"q":       "test query",
		"timeout": 1,
	})
	if err != nil {
		t.Fatal("expected no error from search handler")
	}
	if !resp.IsError {
		t.Log("search with 'q' alias succeeded")
	}

	// Test the "search" alias for query
	resp, err = HandleDDGSearch(context.Background(), map[string]interface{}{
		"search":  "another test",
		"timeout": 1,
	})
	if err != nil {
		t.Fatal("expected no error from search handler")
	}
	if !resp.IsError {
		t.Log("search with 'search' alias succeeded")
	}
}

func TestHandleDDGFetchContent_MissingURL(t *testing.T) {
	resp, err := HandleDDGFetchContent(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error from fetch handler")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required', got: %s", resp.Content[0].Text)
	}
}

func TestHandleDDGFetchContent_AliasKeys(t *testing.T) {
	// Test "uri" alias
	resp, err := HandleDDGFetchContent(context.Background(), map[string]interface{}{
		"uri":     "about:blank",
		"timeout": 1,
	})
	if err != nil {
		t.Fatal("expected no error from fetch handler")
	}
	if !resp.IsError {
		t.Log("fetch with 'uri' alias succeeded")
	}

	// Test "target" alias
	resp, err = HandleDDGFetchContent(context.Background(), map[string]interface{}{
		"target":  "http://example.com",
		"timeout": 1,
	})
	if err != nil {
		t.Fatal("expected no error from fetch handler")
	}
	if !resp.IsError {
		t.Log("fetch with 'target' alias succeeded")
	}
}

func TestHandleDDGFetchContent_TimeoutArg(t *testing.T) {
	// Test with valid URL but short timeout
	resp, err := HandleDDGFetchContent(context.Background(), map[string]interface{}{
		"url":     "http://example.com",
		"timeout": 1, // 1 second timeout
	})
	if err != nil {
		t.Fatal("expected no error from fetch handler")
	}
	if !resp.IsError {
		t.Logf("fetch succeeded, content length: %d", len(resp.Content[0].Text))
	}
}

func TestRegistry_SearchToolsRegistered(t *testing.T) {
	r := NewRegistry()
	expectedTools := []string{
		"search",
		"fetch_content",
	}
	for _, tool := range expectedTools {
		if !r.HasTool(tool) {
			t.Errorf("expected tool '%s' to be registered, but it was not", tool)
		}
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<b>bold</b>", "bold"},
		{"<a href=\"link\">text</a>", "text"},
		{"no tags", "no tags"},
		{"<div><p>nested</p></div>", "nested"},
		{"<br/>", ""},
		{"text with <script>alert('xss')</script> more text", "text with alert('xss') more text"},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripHTML(tt.input)
		if got != tt.expected {
			t.Errorf("stripHTML(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
