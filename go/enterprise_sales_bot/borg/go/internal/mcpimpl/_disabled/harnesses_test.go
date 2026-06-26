package mcpimpl

import (
	"context"
	"strings"
	"testing"
)

func TestHarnessHandlers(t *testing.T) {
	ctx := context.Background()

	t.Run("HandleTabby", func(t *testing.T) {
		// We expect this to fail in the sandbox if 'tabby' isn't in PATH,
		// but we want to check it doesn't crash and returns an error response if missing.
		resp, err := HandleTabby(ctx, nil)
		if err != nil {
			t.Fatalf("HandleTabby crashed: %v", err)
		}
		if resp.IsError {
			if !strings.Contains(resp.Content[0].Text, "failed to start") && !strings.Contains(resp.Content[0].Text, "not found") {
				t.Errorf("HandleTabby returned unexpected error: %s", resp.Content[0].Text)
			}
		} else {
			if !strings.Contains(resp.Content[0].Text, "launched") {
				t.Errorf("HandleTabby success msg missing: %s", resp.Content[0].Text)
			}
		}
	})

	t.Run("HandleHermesAgent", func(t *testing.T) {
		resp, err := HandleHermesAgent(ctx, map[string]interface{}{"task": "test task"})
		if err != nil {
			t.Fatalf("HandleHermesAgent crashed: %v", err)
		}
		if resp.IsError {
			t.Errorf("HandleHermesAgent error: %s", resp.Content[0].Text)
		}
		if !strings.Contains(resp.Content[0].Text, "test task") {
			t.Errorf("HandleHermesAgent response missing task name: %s", resp.Content[0].Text)
		}
	})

	t.Run("HandlePiMono", func(t *testing.T) {
		resp, err := HandlePiMono(ctx, map[string]interface{}{"task": "pi task"})
		if err != nil {
			t.Fatalf("HandlePiMono crashed: %v", err)
		}
		if resp.IsError {
			t.Errorf("HandlePiMono error: %s", resp.Content[0].Text)
		}
		if !strings.Contains(resp.Content[0].Text, "pi task") {
			t.Errorf("HandlePiMono response missing task name: %s", resp.Content[0].Text)
		}
	})
}
