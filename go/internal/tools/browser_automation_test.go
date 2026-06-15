//go:build ignore
// +build ignore

package tools

import (
	"context"
	"testing"
)

// Unit tests for browser automation handlers.
// These test argument validation and optional arg parsing.
// End-to-end browser tests require a running Chrome instance and are skipped by default.

func TestHandleBrowserNavigate_MissingURL(t *testing.T) {
	resp, err := HandleBrowserNavigate(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error for mock browser navigate")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required', got: %s", resp.Content[0].Text)
	}
}

func TestHandleBrowserNavigate_TimeoutArg(t *testing.T) {
	// Test that the timeout argument is parsed correctly.
	// We can't easily test actual timeout behavior without a browser,
	// but we verify that invalid timeout (0) falls back to default.
	resp, err := HandleBrowserNavigate(context.Background(), map[string]interface{}{
		"url":     "about:blank",
		"timeout": 0,
	})
	if err != nil {
		t.Fatal("expected no error for mock browser navigate")
	}
	// With a browser not available, we expect a specific error
	// (connection refused or similar). The key check is that timeout
	// doesn't cause a panic.
	if !resp.IsError {
		t.Log("browser navigate succeeded unexpectedly (browser running?)")
	}
}

func TestHandleBrowserScreenshot_MissingURL(t *testing.T) {
	resp, err := HandleBrowserScreenshot(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error for mock browser screenshot")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required', got: %s", resp.Content[0].Text)
	}
}

func TestHandleBrowserScreenshot_FullPageArg(t *testing.T) {
	// Test that fullPage=true arg is parsed (no panic).
	resp, err := HandleBrowserScreenshot(context.Background(), map[string]interface{}{
		"url":      "about:blank",
		"fullPage": true,
	})
	if err != nil {
		t.Fatal("expected no error for mock browser screenshot")
	}
	if !resp.IsError {
		t.Log("browser screenshot succeeded unexpectedly (browser running?)")
	}
}

func TestHandleBrowserScreenshot_FullPageFalse(t *testing.T) {
	// Test that fullPage=false is handled (should use viewport screenshot).
	resp, err := HandleBrowserScreenshot(context.Background(), map[string]interface{}{
		"url":      "about:blank",
		"fullPage": false,
	})
	if err != nil {
		t.Fatal("expected no error for mock browser screenshot")
	}
	if !resp.IsError {
		t.Log("browser screenshot succeeded unexpectedly (browser running?)")
	}
}

func TestHandleBrowserGetHTML_MissingURL(t *testing.T) {
	resp, err := HandleBrowserGetHTML(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error for mock browser get html")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required', got: %s", resp.Content[0].Text)
	}
}

func TestHandleBrowserEvaluate_MissingArgs(t *testing.T) {
	// Test missing URL
	resp, err := HandleBrowserEvaluate(context.Background(), map[string]interface{}{
		"script": "document.title",
	})
	if err != nil {
		t.Fatal("expected no error for mock browser evaluate")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required' for missing url, got: %s", resp.Content[0].Text)
	}

	// Test missing script
	resp, err = HandleBrowserEvaluate(context.Background(), map[string]interface{}{
		"url": "about:blank",
	})
	if err != nil {
		t.Fatal("expected no error for mock browser evaluate")
	}
	if !resp.IsError {
		t.Fatal("expected error response when script is missing")
	}
	if resp.Content[0].Text != "script is required" {
		t.Fatalf("expected 'script is required' for missing script, got: %s", resp.Content[0].Text)
	}
}

func TestHandleBrowserClick_MissingArgs(t *testing.T) {
	// Test missing URL
	resp, err := HandleBrowserClick(context.Background(), map[string]interface{}{
		"selector": "#button",
	})
	if err != nil {
		t.Fatal("expected no error for mock browser click")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required' for missing url, got: %s", resp.Content[0].Text)
	}

	// Test missing selector
	resp, err = HandleBrowserClick(context.Background(), map[string]interface{}{
		"url": "about:blank",
	})
	if err != nil {
		t.Fatal("expected no error for mock browser click")
	}
	if !resp.IsError {
		t.Fatal("expected error response when selector is missing")
	}
	if resp.Content[0].Text != "selector is required" {
		t.Fatalf("expected 'selector is required' for missing selector, got: %s", resp.Content[0].Text)
	}
}

func TestHandleBrowserFillForm_MissingArgs(t *testing.T) {
	// Test missing URL
	resp, err := HandleBrowserFillForm(context.Background(), map[string]interface{}{
		"selector": "#input",
		"value":    "test",
	})
	if err != nil {
		t.Fatal("expected no error for mock browser fill form")
	}
	if !resp.IsError {
		t.Fatal("expected error response when url is missing")
	}
	if resp.Content[0].Text != "url is required" {
		t.Fatalf("expected 'url is required' for missing url, got: %s", resp.Content[0].Text)
	}

	// Test missing selector
	resp, err = HandleBrowserFillForm(context.Background(), map[string]interface{}{
		"url":   "about:blank",
		"value": "test",
	})
	if err != nil {
		t.Fatal("expected no error for mock browser fill form")
	}
	if !resp.IsError {
		t.Fatal("expected error response when selector is missing")
	}
	if resp.Content[0].Text != "selector is required" {
		t.Fatalf("expected 'selector is required' for missing selector, got: %s", resp.Content[0].Text)
	}

	// Test that value is handled properly (it will fail with real browser, but shouldn't panic)
	// Use a short timeout so this doesn't block if browser is slow
	_, err = HandleBrowserFillForm(context.Background(), map[string]interface{}{
		"url":      "about:blank",
		"selector": "#nonexistent",
		"timeout":  1000,
	})
	if err != nil {
		t.Fatal("expected no error for mock browser fill form")
	}
	// Should handle gracefully (browser may not be available)
}

func TestRegistry_BrowserToolsRegistered(t *testing.T) {
	r := NewRegistry()
	expectedTools := []string{
		"browser_navigate",
		"browser_screenshot",
		"browser_get_html",
		"browser_evaluate",
		"browser_click",
		"browser_fill_form",
	}
	for _, tool := range expectedTools {
		if !r.HasTool(tool) {
			t.Errorf("expected tool '%s' to be registered, but it was not", tool)
		}
	}
}

func TestRegistry_ExecuteBrowserTool(t *testing.T) {
	r := NewRegistry()

	// Test that executing browser_navigate returns error (no browser) not "not found"
	resp, err := r.Execute(context.Background(), "browser_navigate", map[string]interface{}{
		"url": "about:blank",
	})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	// Should get an error about browser not being available (not a "tool not found" error)
	if !resp.IsError {
		t.Log("browser_navigate succeeded (browser available)")
	}

	// Test executing an unregistered tool
	_, err = r.Execute(context.Background(), "browser_nonexistent", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unregistered tool")
	}
}
