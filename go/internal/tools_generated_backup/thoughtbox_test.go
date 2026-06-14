package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestThoughtboxSearch(t *testing.T) {
	ctx := context.Background()
	args := map[string]interface{}{
		"code": "async () => Object.keys(catalog.operations)",
	}

	resp, err := HandleThoughtboxSearch(ctx, args)
	if err != nil {
		t.Fatalf("HandleThoughtboxSearch failed: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatalf("Expected content, got none")
	}

	text := resp.Content[0].Text
	if !strings.Contains(text, "session") || !strings.Contains(text, "thought") {
		t.Errorf("Expected catalog operation keys, got: %s", text)
	}
}

func TestThoughtboxExecute(t *testing.T) {
	ctx := context.Background()
	args := map[string]interface{}{
		"code": "async () => { return await tb.session.list(); }",
	}

	resp, err := HandleThoughtboxExecute(ctx, args)
	if err != nil {
		t.Fatalf("HandleThoughtboxExecute failed: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatalf("Expected content, got none")
	}

	text := resp.Content[0].Text
	if !strings.Contains(text, "session_mock_123") {
		t.Errorf("Expected session list, got: %s", text)
	}
}

func TestThoughtboxPeerNotebook(t *testing.T) {
	ctx := context.Background()

	// 1. Seed artifact
	seedArgs := map[string]interface{}{
		"operation": "peer_artifact_seed",
		"text":      "The quick brown fox jumps over the lazy dog. A second sentence here.",
		"name":      "test.txt",
	}
	resp, err := HandleThoughtboxPeerNotebook(ctx, seedArgs)
	if err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	var seedResult struct {
		Artifact struct {
			ID string `json:"id"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal([]byte(resp.Content[0].Text), &seedResult); err != nil {
		t.Fatalf("Unmarshal seed response failed: %v, raw text: %s", err, resp.Content[0].Text)
	}

	artID := seedResult.Artifact.ID
	if artID == "" {
		t.Fatalf("Expected artifact ID, got empty")
	}

	// 2. Invoke peer tool
	invokeArgs := map[string]interface{}{
		"operation": "peer_invoke",
		"peerId":    "claim-extractor",
		"tool":      "extract_claims",
		"args": map[string]interface{}{
			"textArtifactId": artID,
		},
	}
	respInvoke, err := HandleThoughtboxPeerNotebook(ctx, invokeArgs)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	var invokeResult struct {
		InvocationID string `json:"invocationId"`
		Result       struct {
			ClaimCount int    `json:"claimCount"`
			ClaimsArt  string `json:"claimsArtifactId"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(respInvoke.Content[0].Text), &invokeResult); err != nil {
		t.Fatalf("Unmarshal invoke response failed: %v, raw text: %s", err, respInvoke.Content[0].Text)
	}

	if invokeResult.Result.ClaimCount != 2 {
		t.Errorf("Expected 2 claims, got %d", invokeResult.Result.ClaimCount)
	}

	// 3. Get trace events
	traceArgs := map[string]interface{}{
		"operation":    "peer_list_trace_events",
		"invocationId": invokeResult.InvocationID,
	}
	respTrace, err := HandleThoughtboxPeerNotebook(ctx, traceArgs)
	if err != nil {
		t.Fatalf("Get trace events failed: %v", err)
	}

	var traceResult struct {
		Events []interface{} `json:"events"`
	}
	if err := json.Unmarshal([]byte(respTrace.Content[0].Text), &traceResult); err != nil {
		t.Fatalf("Unmarshal trace response failed: %v", err)
	}

	if len(traceResult.Events) < 3 {
		t.Errorf("Expected at least 3 trace events, got %d", len(traceResult.Events))
	}
}
