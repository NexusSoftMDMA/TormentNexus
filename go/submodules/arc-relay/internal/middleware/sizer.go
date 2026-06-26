package middleware

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/store"
)

// SizerConfig configures the content sizer middleware.
type SizerConfig struct {
	MaxResponseBytes int    `json:"max_response_bytes"` // max response size in bytes (0 = unlimited)
	Action           string `json:"action"`             // "truncate", "warn", "block"
}

// DefaultSizerConfig returns sensible defaults.
func DefaultSizerConfig() SizerConfig {
	return SizerConfig{
		MaxResponseBytes: 500000, // ~500KB
		Action:           "truncate",
	}
}

// Sizer measures and optionally truncates large responses.
type Sizer struct {
	maxBytes    int
	action      string
	eventLogger EventLogger
}

// NewSizerFromConfig creates a Sizer from JSON config.
func NewSizerFromConfig(config json.RawMessage, logger EventLogger) (Middleware, error) {
	var cfg SizerConfig
	if len(config) > 0 && string(config) != "{}" {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("sizer: invalid config: %w", err)
		}
	} else {
		cfg = DefaultSizerConfig()
	}
	return NewSizer(cfg, logger), nil
}

// NewSizer creates a content sizer middleware.
func NewSizer(cfg SizerConfig, logger EventLogger) *Sizer {
	if cfg.MaxResponseBytes <= 0 {
		cfg.MaxResponseBytes = 500000
	}
	if cfg.Action == "" {
		cfg.Action = "truncate"
	}
	return &Sizer{
		maxBytes:    cfg.MaxResponseBytes,
		action:      cfg.Action,
		eventLogger: logger,
	}
}

func (s *Sizer) Name() string { return "sizer" }

func (s *Sizer) ProcessRequest(ctx context.Context, req *mcp.Request, meta *RequestMeta) (*mcp.Request, error) {
	return req, nil
}

func (s *Sizer) ProcessResponse(ctx context.Context, req *mcp.Request, resp *mcp.Response, meta *RequestMeta) (*mcp.Response, error) {
	if resp.Error != nil || resp.Result == nil {
		return resp, nil
	}

	size := len(resp.Result)
	if size <= s.maxBytes {
		return resp, nil
	}

	summary := fmt.Sprintf("Response size %d bytes exceeds limit %d bytes", size, s.maxBytes)

	switch s.action {
	case "block":
		s.logEvent(meta, "blocked", summary)
		return mcp.NewErrorResponse(resp.ID, mcp.ErrCodeInternal,
			fmt.Sprintf("response too large (%d bytes, limit %d)", size, s.maxBytes)), nil

	case "truncate":
		s.logEvent(meta, "truncated", summary)
		truncated := make(json.RawMessage, s.maxBytes)
		copy(truncated, resp.Result[:s.maxBytes])

		// Try to produce valid JSON by wrapping in a content structure
		truncMsg := fmt.Sprintf(`{"content":[{"type":"text","text":"[Response truncated: %d bytes exceeded %d byte limit. First %d bytes shown.]\n%s"}]}`,
			size, s.maxBytes, s.maxBytes,
			jsonEscape(string(resp.Result[:s.maxBytes])))

		return &mcp.Response{
			JSONRPC: resp.JSONRPC,
			ID:      resp.ID,
			Result:  json.RawMessage(truncMsg),
		}, nil

	case "warn":
		// Pass through but log
		s.logEvent(meta, "alert", summary)
		return resp, nil

	default:
		return resp, nil
	}
}

func (s *Sizer) logEvent(meta *RequestMeta, eventType, summary string) {
	if s.eventLogger == nil {
		return
	}
	s.eventLogger(&store.MiddlewareEvent{
		Middleware:    "sizer",
		EventType:     eventType,
		Summary:       summary,
		RequestMethod: meta.Method,
		EndpointName:  meta.ToolName,
		UserID:        meta.UserID,
	})
}

// jsonEscape escapes a string for safe embedding in JSON.
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	// Strip surrounding quotes
	if len(b) >= 2 {
		return string(b[1 : len(b)-1])
	}
	return s
}
