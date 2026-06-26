package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/store"
)

// SanitizerConfig configures the sanitizer middleware.
type SanitizerConfig struct {
	Patterns []SanitizePattern `json:"patterns"`
}

// SanitizePattern is a named regex pattern with an action.
type SanitizePattern struct {
	Name   string `json:"name"`
	Regex  string `json:"regex"`
	Action string `json:"action"` // "redact" or "block"
}

// DefaultSanitizerConfig returns sensible default patterns.
func DefaultSanitizerConfig() SanitizerConfig {
	return SanitizerConfig{
		Patterns: []SanitizePattern{
			{Name: "api_key", Regex: `(?i)(api[_-]?key|secret[_-]?key|access[_-]?token|private[_-]?key)\s*[=:]\s*\S+`, Action: "redact"},
			{Name: "password", Regex: `(?i)(password|passwd|pwd)\s*[=:]\s*\S+`, Action: "redact"},
			{Name: "ssn", Regex: `\b\d{3}-\d{2}-\d{4}\b`, Action: "redact"},
			{Name: "credit_card", Regex: `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`, Action: "block"},
		},
	}
}

type compiledPattern struct {
	name   string
	re     *regexp.Regexp
	action string // "redact" or "block"
}

// Sanitizer scans tool responses for sensitive patterns and redacts or blocks them.
type Sanitizer struct {
	patterns    []compiledPattern
	eventLogger EventLogger
}

// NewSanitizerFromConfig creates a Sanitizer from JSON config.
func NewSanitizerFromConfig(config json.RawMessage, logger EventLogger) (Middleware, error) {
	var cfg SanitizerConfig
	if len(config) > 0 && string(config) != "{}" {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("sanitizer: invalid config: %w", err)
		}
	} else {
		cfg = DefaultSanitizerConfig()
	}
	return NewSanitizer(cfg, logger)
}

// NewSanitizer creates a Sanitizer with pre-compiled regex patterns.
func NewSanitizer(cfg SanitizerConfig, logger EventLogger) (*Sanitizer, error) {
	patterns := make([]compiledPattern, 0, len(cfg.Patterns))
	for _, p := range cfg.Patterns {
		re, err := regexp.Compile(p.Regex)
		if err != nil {
			return nil, fmt.Errorf("sanitizer: invalid regex for pattern %q: %w", p.Name, err)
		}
		action := p.Action
		if action == "" {
			action = "redact"
		}
		patterns = append(patterns, compiledPattern{name: p.Name, re: re, action: action})
	}
	return &Sanitizer{patterns: patterns, eventLogger: logger}, nil
}

func (s *Sanitizer) Name() string { return "sanitizer" }

func (s *Sanitizer) ProcessRequest(ctx context.Context, req *mcp.Request, meta *RequestMeta) (*mcp.Request, error) {
	// Sanitizer only processes responses (tool output going back to AI)
	return req, nil
}

func (s *Sanitizer) ProcessResponse(ctx context.Context, req *mcp.Request, resp *mcp.Response, meta *RequestMeta) (*mcp.Response, error) {
	if resp.Error != nil || resp.Result == nil {
		return resp, nil
	}

	resultStr := string(resp.Result)
	modified := false

	for _, p := range s.patterns {
		if !p.re.MatchString(resultStr) {
			continue
		}

		if p.action == "block" {
			s.logEvent(meta, "blocked", fmt.Sprintf("Response blocked: pattern %q matched", p.name))
			return mcp.NewErrorResponse(resp.ID, mcp.ErrCodeInternal,
				fmt.Sprintf("response blocked by sanitizer: %s pattern detected", p.name)), nil
		}

		// Redact
		before := resultStr
		resultStr = p.re.ReplaceAllString(resultStr, "[REDACTED]")
		if resultStr != before {
			modified = true
			s.logEvent(meta, "redacted", fmt.Sprintf("Pattern %q redacted from response", p.name))
		}
	}

	if modified {
		resp = &mcp.Response{
			JSONRPC: resp.JSONRPC,
			ID:      resp.ID,
			Result:  json.RawMessage(resultStr),
		}
	}
	return resp, nil
}

func (s *Sanitizer) logEvent(meta *RequestMeta, eventType, summary string) {
	if s.eventLogger == nil {
		return
	}
	s.eventLogger(&store.MiddlewareEvent{
		Middleware:    "sanitizer",
		EventType:     eventType,
		Summary:       summary,
		RequestMethod: meta.Method,
		EndpointName:  meta.ToolName,
		UserID:        meta.UserID,
	})
}

// ScanText checks text against all sanitizer patterns and returns matches.
// Useful for testing/preview in the UI.
func (s *Sanitizer) ScanText(text string) []string {
	var matches []string
	for _, p := range s.patterns {
		if p.re.MatchString(text) {
			matches = append(matches, p.name+" ("+p.action+")")
		}
	}
	return matches
}

// RedactText applies all redaction patterns to text (for preview).
func (s *Sanitizer) RedactText(text string) string {
	for _, p := range s.patterns {
		if p.action == "redact" {
			text = p.re.ReplaceAllString(text, "[REDACTED]")
		}
	}
	return text
}
