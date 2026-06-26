package mcpimpl

import (
	"context"
	"strings"
	"testing"
)

func TestHandleFetch(t *testing.T) {
	ctx := context.Background()
	args := map[string]interface{}{
		"url": "https://httpbin.org/html",
	}

	resp, err := HandleFetch(ctx, args)
	if err != nil {
		t.Fatalf("HandleFetch failed: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatalf("Expected content, got none")
	}

	text := resp.Content[0].Text
	if !strings.Contains(text, "Herman Melville") && !strings.Contains(text, "Moby Dick") {
		// httpbin.org/html contains a quote from Moby Dick or Herman Melville
		t.Logf("Warning: standard quote content not found (could be network or external change), got text: %s", text)
	}
}
