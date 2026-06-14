package tools

import (
	"context"
	"strings"
	"testing"
)

func TestASTGrepNotInstalled(t *testing.T) {
	// If sg is installed, this test will run and return either an error or output.
	// If sg is not installed, it should gracefully return a message indicating it wasn't found in PATH.
	ctx := context.Background()

	// Test syntax tree dump
	args := map[string]interface{}{
		"code":     "fn main() { println!(\"hello\"); }",
		"language": "rs",
	}
	resp, errVal := HandleDumpSyntaxTree(ctx, args)
	if errVal != nil {
		t.Fatalf("unexpected error: %v", errVal)
	}

	if resp.IsError {
		if !strings.Contains(resp.Content[0].Text, "not found in PATH") && !strings.Contains(resp.Content[0].Text, "ast-grep failed") {
			t.Errorf("unexpected error content: %s", resp.Content[0].Text)
		}
	} else {
		// If sg is installed, verify we got some response
		if len(resp.Content) == 0 || resp.Content[0].Text == "" {
			t.Errorf("empty syntax tree dump response")
		}
	}
}

func TestASTGrepFindCodeValidation(t *testing.T) {
	ctx := context.Background()
	args := map[string]interface{}{}
	resp, _ := HandleFindCode(ctx, args)
	if !resp.IsError || !strings.Contains(resp.Content[0].Text, "pattern parameter is required") {
		t.Errorf("expected missing pattern validation error, got: %v", resp)
	}
}
