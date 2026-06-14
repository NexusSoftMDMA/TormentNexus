package tools

import (
	"context"
	"strings"
	"testing"
)

func TestHandleChromeDevTools(t *testing.T) {
	ctx := context.Background()

	// 1. Test Navigate
	navArgs := map[string]interface{}{
		"action": "navigate",
		"url":    "https://httpbin.org/status/200",
	}
	respNav, err := HandleChromeDevTools(ctx, navArgs)
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}
	if !strings.Contains(respNav.Content[0].Text, "Successfully navigated") && !strings.Contains(respNav.Content[0].Text, "Navigated to") {
		t.Errorf("Unexpected navigate output: %s", respNav.Content[0].Text)
	}

	// 2. Test Evaluate
	evalArgs := map[string]interface{}{
		"action": "evaluate",
		"script": "console.log(1 + 2)",
	}
	respEval, err := HandleChromeDevTools(ctx, evalArgs)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if strings.TrimSpace(respEval.Content[0].Text) != "3" {
		t.Errorf("Expected 3, got: %s", respEval.Content[0].Text)
	}

	// 3. Test Click
	clickArgs := map[string]interface{}{
		"action":   "click",
		"selector": "#submit-btn",
	}
	respClick, err := HandleChromeDevTools(ctx, clickArgs)
	if err != nil {
		t.Fatalf("Click failed: %v", err)
	}
	if !strings.Contains(respClick.Content[0].Text, "#submit-btn") {
		t.Errorf("Unexpected click output: %s", respClick.Content[0].Text)
	}
}
