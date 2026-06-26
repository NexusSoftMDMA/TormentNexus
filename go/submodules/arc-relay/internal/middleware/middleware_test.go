package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/store"
)

func TestPipeline_ProcessRequest(t *testing.T) {
	pipeline := NewPipeline()
	meta := &RequestMeta{Method: "tools/call", ToolName: "test"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/call"}

	// Empty pipeline passes through
	result, err := pipeline.ProcessRequest(context.Background(), req, meta)
	if err != nil {
		t.Fatalf("empty pipeline: %v", err)
	}
	if result != req {
		t.Fatal("empty pipeline should return same request")
	}
}

func TestPipeline_ProcessResponse(t *testing.T) {
	pipeline := NewPipeline()
	meta := &RequestMeta{Method: "tools/call", ToolName: "test"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/call"}
	resp := &mcp.Response{JSONRPC: "2.0", ID: json.RawMessage(`1`), Result: json.RawMessage(`{"ok":true}`)}

	result, err := pipeline.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatalf("empty pipeline: %v", err)
	}
	if result != resp {
		t.Fatal("empty pipeline should return same response")
	}
}

func TestSanitizer_Redact(t *testing.T) {
	cfg := SanitizerConfig{
		Patterns: []SanitizePattern{
			{Name: "api_key", Regex: `(?i)api_key\s*=\s*\S+`, Action: "redact"},
		},
	}
	var events []*store.MiddlewareEvent
	logger := func(evt *store.MiddlewareEvent) { events = append(events, evt) }

	s, err := NewSanitizer(cfg, logger)
	if err != nil {
		t.Fatal(err)
	}

	meta := &RequestMeta{Method: "tools/call", ToolName: "get_config"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/call"}
	resp := &mcp.Response{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Result:  json.RawMessage(`{"text":"api_key = secret123"}`),
	}

	result, err := s.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}

	resultStr := string(result.Result)
	if resultStr == string(resp.Result) {
		t.Fatal("sanitizer should have modified the response")
	}
	if !contains(resultStr, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] in result, got: %s", resultStr)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "redacted" {
		t.Fatalf("expected event type 'redacted', got %q", events[0].EventType)
	}
}

func TestSanitizer_Block(t *testing.T) {
	cfg := SanitizerConfig{
		Patterns: []SanitizePattern{
			{Name: "credit_card", Regex: `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`, Action: "block"},
		},
	}
	s, err := NewSanitizer(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	meta := &RequestMeta{Method: "tools/call"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/call"}
	resp := &mcp.Response{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Result:  json.RawMessage(`{"text":"Card: 4111-1111-1111-1111"}`),
	}

	result, err := s.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}
	if result.Error == nil {
		t.Fatal("expected blocked response to have error")
	}
	if !contains(result.Error.Message, "blocked") {
		t.Fatalf("expected 'blocked' in error, got: %s", result.Error.Message)
	}
}

func TestSanitizer_NoMatch(t *testing.T) {
	cfg := SanitizerConfig{
		Patterns: []SanitizePattern{
			{Name: "ssn", Regex: `\b\d{3}-\d{2}-\d{4}\b`, Action: "redact"},
		},
	}
	s, err := NewSanitizer(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	meta := &RequestMeta{Method: "tools/call"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`)}
	resp := &mcp.Response{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Result:  json.RawMessage(`{"text":"hello world"}`),
	}

	result, err := s.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}
	if string(result.Result) != string(resp.Result) {
		t.Fatal("response should not have been modified")
	}
}

func TestSizer_Truncate(t *testing.T) {
	cfg := SizerConfig{MaxResponseBytes: 50, Action: "truncate"}
	var events []*store.MiddlewareEvent
	logger := func(evt *store.MiddlewareEvent) { events = append(events, evt) }

	s := NewSizer(cfg, logger)
	meta := &RequestMeta{Method: "tools/call", ToolName: "big_query"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`)}

	// Create a large response
	largeResult := make([]byte, 200)
	for i := range largeResult {
		largeResult[i] = 'x'
	}
	resp := &mcp.Response{JSONRPC: "2.0", ID: json.RawMessage(`1`), Result: json.RawMessage(largeResult)}

	result, err := s.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}
	if result.Error != nil {
		t.Fatal("truncate should not produce error response")
	}
	if len(events) != 1 || events[0].EventType != "truncated" {
		t.Fatal("expected truncated event")
	}
}

func TestSizer_Block(t *testing.T) {
	cfg := SizerConfig{MaxResponseBytes: 50, Action: "block"}
	s := NewSizer(cfg, nil)

	meta := &RequestMeta{Method: "tools/call"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`)}
	largeResult := make([]byte, 200)
	for i := range largeResult {
		largeResult[i] = 'y'
	}
	resp := &mcp.Response{JSONRPC: "2.0", ID: json.RawMessage(`1`), Result: json.RawMessage(largeResult)}

	result, err := s.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}
	if result.Error == nil {
		t.Fatal("block action should produce error response")
	}
}

func TestSizer_UnderLimit(t *testing.T) {
	cfg := SizerConfig{MaxResponseBytes: 1000, Action: "truncate"}
	s := NewSizer(cfg, nil)

	meta := &RequestMeta{Method: "tools/call"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`)}
	resp := &mcp.Response{JSONRPC: "2.0", ID: json.RawMessage(`1`), Result: json.RawMessage(`{"ok":true}`)}

	result, err := s.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}
	if result != resp {
		t.Fatal("under-limit response should pass through unchanged")
	}
}

func TestAlerter_PatternMatch(t *testing.T) {
	cfg := AlerterConfig{
		Rules: []AlertRule{
			{Name: "prod_access", Match: `(?i)production`, Direction: "request", Action: "log"},
		},
	}
	var events []*store.MiddlewareEvent
	logger := func(evt *store.MiddlewareEvent) { events = append(events, evt) }

	a, err := NewAlerter(cfg, logger)
	if err != nil {
		t.Fatal(err)
	}

	meta := &RequestMeta{Method: "tools/call", ToolName: "query_db"}
	req := &mcp.Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"query_db","arguments":{"query":"SELECT * FROM production.users"}}`),
	}

	result, err := a.ProcessRequest(context.Background(), req, meta)
	if err != nil {
		t.Fatal(err)
	}
	// Alerter should never modify the request
	if result != req {
		t.Fatal("alerter should pass request through unchanged")
	}
	if len(events) != 1 || events[0].EventType != "alert" {
		t.Fatalf("expected 1 alert event, got %d", len(events))
	}
}

func TestFullPipeline(t *testing.T) {
	var events []*store.MiddlewareEvent
	logger := func(evt *store.MiddlewareEvent) { events = append(events, evt) }

	san, _ := NewSanitizer(SanitizerConfig{
		Patterns: []SanitizePattern{
			{Name: "api_key", Regex: `api_key=\S+`, Action: "redact"},
		},
	}, logger)

	sizer := NewSizer(SizerConfig{MaxResponseBytes: 10000, Action: "warn"}, logger)

	alerter, _ := NewAlerter(AlerterConfig{
		Rules: []AlertRule{
			{Name: "test", Match: `sensitive`, Direction: "response", Action: "log"},
		},
	}, logger)

	pipeline := NewPipeline()
	pipeline.Add(san)
	pipeline.Add(sizer)
	pipeline.Add(alerter)

	meta := &RequestMeta{Method: "tools/call", ToolName: "get_info"}
	req := &mcp.Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/call"}
	resp := &mcp.Response{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Result:  json.RawMessage(`{"text":"sensitive data with api_key=mysecret123"}`),
	}

	// Process response
	result, err := pipeline.ProcessResponse(context.Background(), req, resp, meta)
	if err != nil {
		t.Fatal(err)
	}

	// api_key should be redacted
	if !contains(string(result.Result), "[REDACTED]") {
		t.Fatalf("expected redaction, got: %s", string(result.Result))
	}
	// Should have events from sanitizer and alerter
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
