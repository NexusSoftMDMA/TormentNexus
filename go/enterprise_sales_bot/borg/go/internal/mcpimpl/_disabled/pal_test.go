package mcpimpl

import (
	"context"
	"strings"
	"testing"
)

func TestPALTools(t *testing.T) {
	ctx := context.Background()

	// 1. Test PAL Chat
	t.Run("PalChat", func(t *testing.T) {
		args := map[string]interface{}{
			"prompt":                          "How do I optimize Go loops?",
			"working_directory_absolute_path": ".",
		}
		resp, err := HandlePalChat(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalChat failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandlePalChat returned error response: %v", resp.Content[0].Text)
		}
		if !strings.Contains(resp.Content[0].Text, "Collaborative Chat") && !strings.Contains(resp.Content[0].Text, "AGENT'S TURN") {
			t.Errorf("Unexpected Chat response content: %s", resp.Content[0].Text)
		}
	})

	// 2. Test PAL Think Deep
	t.Run("PalThinkDeep", func(t *testing.T) {
		args := map[string]interface{}{
			"step":               "Evaluate memory footprint",
			"step_number":        1.0,
			"total_steps":        5.0,
			"next_step_required": true,
			"findings":           "Memory consumption is stable",
			"confidence":         "high",
		}
		resp, err := HandlePalThinkDeep(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalThinkDeep failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandlePalThinkDeep returned error response")
		}
	})

	// 3. Test PAL Planner
	t.Run("PalPlanner", func(t *testing.T) {
		args := map[string]interface{}{
			"step":               "Bootstrap database schema",
			"step_number":        1.0,
			"total_steps":        3.0,
			"next_step_required": true,
		}
		resp, err := HandlePalPlanner(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalPlanner failed: %v", err)
		}
		if !strings.Contains(resp.Content[0].Text, "planning_in_progress") {
			t.Errorf("Planner response did not contain status: %s", resp.Content[0].Text)
		}
	})

	// 4. Test PAL Consensus
	t.Run("PalConsensus", func(t *testing.T) {
		args := map[string]interface{}{
			"step":               "Implement lock gates",
			"step_number":        1.0,
			"total_steps":        2.0,
			"next_step_required": true,
			"findings":           "Balanced approach needed",
		}
		resp, err := HandlePalConsensus(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalConsensus failed: %v", err)
		}
		if !strings.Contains(resp.Content[0].Text, "consulting_models") {
			t.Errorf("Consensus response did not contain consulting_models: %s", resp.Content[0].Text)
		}
	})

	// 5. Test PAL CodeReview
	t.Run("PalCodeReview", func(t *testing.T) {
		args := map[string]interface{}{
			"code": "package main\n\nfunc main() {}",
		}
		resp, err := HandlePalCodeReview(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalCodeReview failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandlePalCodeReview returned error")
		}
	})

	// 6. Test PAL Precommit
	t.Run("PalPrecommit", func(t *testing.T) {
		args := map[string]interface{}{
			"files": "main.go",
		}
		resp, err := HandlePalPrecommit(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalPrecommit failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandlePalPrecommit returned error")
		}
	})

	// 7. Test PAL Debug
	t.Run("PalDebug", func(t *testing.T) {
		args := map[string]interface{}{
			"error_log":    "panic: nil pointer dereference",
			"code_context": "func run() { var x *int; println(*x) }",
		}
		resp, err := HandlePalDebug(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalDebug failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandlePalDebug returned error")
		}
	})

	// 8. Test PAL Challenge
	t.Run("PalChallenge", func(t *testing.T) {
		args := map[string]interface{}{
			"proposal": "Store configuration in global variable",
		}
		resp, err := HandlePalChallenge(ctx, args)
		if err != nil {
			t.Fatalf("HandlePalChallenge failed: %v", err)
		}
		if resp.IsError {
			t.Fatalf("HandlePalChallenge returned error")
		}
	})
}
