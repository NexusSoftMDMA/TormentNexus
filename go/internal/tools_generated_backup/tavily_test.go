package tools

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestHandleTavilySearch(t *testing.T) {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping TestHandleTavilySearch: TAVILY_API_KEY environment variable not set")
	}

	ctx := context.Background()
	args := map[string]interface{}{
		"query":        "Model Context Protocol",
		"max_results": 3,
	}

	resp, err := HandleTavilySearch(ctx, args)
	if err != nil {
		t.Fatalf("HandleTavilySearch failed: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatalf("Expected content, got none")
	}

	text := resp.Content[0].Text
	if !strings.Contains(text, "results") {
		t.Errorf("Expected Tavily response keys or structure, got: %s", text)
	}
}
