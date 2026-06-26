package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/store"
)

// ArchiveConfig configures the archive middleware.
type ArchiveConfig struct {
	URL              string `json:"url"`                          // Target URL to POST archived data
	AuthType         string `json:"auth_type"`                    // "none", "bearer", "api_key"
	AuthValue        string `json:"auth_value"`                   // Token/key value
	APIKeyHeader     string `json:"api_key_header,omitempty"`     // Header name for api_key auth (default: X-API-Key)
	Include          string `json:"include"`                      // "request", "response", "both"
	NaClRecipientKey string `json:"nacl_recipient_key,omitempty"` // Base64-encoded Curve25519 public key for NaCl Box encryption
}

// DefaultArchiveConfig returns sensible defaults.
func DefaultArchiveConfig() ArchiveConfig {
	return ArchiveConfig{
		URL:     "",
		Include: "both",
	}
}

// archivePayload is the JSON envelope sent to the archive target.
type archivePayload struct {
	Version   string          `json:"version"`
	Source    string          `json:"source"`
	Phase     string          `json:"phase"`
	Timestamp string          `json:"timestamp"`
	Meta      archiveMeta     `json:"meta"`
	Request   json.RawMessage `json:"request,omitempty"`
	Response  json.RawMessage `json:"response,omitempty"`
}

type archiveMeta struct {
	ServerID   string `json:"server_id"`
	ServerName string `json:"server_name"`
	UserID     string `json:"user_id"`
	ClientIP   string `json:"client_ip"`
	Method     string `json:"method"`
	ToolName   string `json:"tool_name"`
	RequestID  string `json:"request_id"`
}

// Archive sends MCP request/response data to a configured HTTP endpoint
// via the shared ArchiveDispatcher. It is observe-only and never blocks MCP traffic.
type Archive struct {
	cfg          ArchiveConfig
	eventLogger  EventLogger
	dispatcher   *ArchiveDispatcher
	recipientKey *[32]byte // cached decoded NaCl key (nil if encryption disabled)
}

// ParseArchiveConfig unmarshals a JSON archive config into an ArchiveConfig
// struct, applying defaults. It does NOT validate semantic correctness -
// use ValidateArchiveConfig for that. Separating parse from validate lets
// the caller normalize a config before deciding whether to save it.
func ParseArchiveConfig(config json.RawMessage) (ArchiveConfig, error) {
	var cfg ArchiveConfig
	if len(config) > 0 && string(config) != "{}" {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return cfg, fmt.Errorf("archive: invalid config: %w", err)
		}
	} else {
		cfg = DefaultArchiveConfig()
	}
	if cfg.Include == "" {
		cfg.Include = "both"
	}
	if cfg.APIKeyHeader == "" {
		cfg.APIKeyHeader = "X-API-Key"
	}
	return cfg, nil
}

// ValidateArchiveConfig checks that an archive config is semantically
// well-formed. Called from both the save-time API handler and at
// middleware construction time so that bad config is rejected at the
// earliest possible moment. Returns the decoded NaCl recipient key (or
// nil if encryption is not configured) so callers that have already
// validated do not have to decode twice.
//
// Rules:
//   - URL is required.
//   - URL must parse and use http or https.
//   - Plain http is only allowed for localhost/loopback hosts so
//     operators developing locally are not forced into self-signed TLS.
//   - auth_type must be one of none/bearer/api_key (empty == none).
//   - include must be one of request/response/both (empty == both).
//   - nacl_recipient_key, when present, must decode to a 32-byte X25519
//     public key.
func ValidateArchiveConfig(cfg ArchiveConfig) (*[32]byte, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("archive: url is required")
	}
	parsed, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("archive: url is not parseable: %w", err)
	}
	switch parsed.Scheme {
	case "https":
		// always allowed
	case "http":
		if !isLoopbackHost(parsed.Hostname()) {
			return nil, fmt.Errorf("archive: url must use https (plain http allowed only for localhost)")
		}
	default:
		return nil, fmt.Errorf("archive: url must use http or https, got %q", parsed.Scheme)
	}
	switch cfg.AuthType {
	case "", "none":
		// ok
	case "bearer", "api_key":
		// Require a non-empty value so an accidental form clobber
		// cannot silently downgrade a configured bearer to
		// "authenticated with nothing" that still passes the save
		// handler. Callers can always set AuthType to none to
		// explicitly opt out of auth.
		if cfg.AuthValue == "" {
			return nil, fmt.Errorf("archive: auth_value is required when auth_type is %s", cfg.AuthType)
		}
	default:
		return nil, fmt.Errorf("archive: auth_type must be none/bearer/api_key, got %q", cfg.AuthType)
	}
	switch cfg.Include {
	case "", "request", "response", "both":
		// ok
	default:
		return nil, fmt.Errorf("archive: include must be request/response/both, got %q", cfg.Include)
	}
	if cfg.NaClRecipientKey == "" {
		return nil, nil
	}
	key, err := DecodeRecipientKey(cfg.NaClRecipientKey)
	if err != nil {
		return nil, fmt.Errorf("archive: invalid nacl_recipient_key: %w", err)
	}
	return &key, nil
}

// isLoopbackHost returns true for hostnames that point at the local
// machine. Used to whitelist plain http for developer ergonomics
// without opening up unencrypted traffic in production deployments.
// Note: 0.0.0.0 is intentionally NOT treated as loopback - it is the
// wildcard bind address, not a local address, so a server listening
// there is reachable from the network.
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// NewArchiveFromConfig creates an Archive from JSON config.
func NewArchiveFromConfig(config json.RawMessage, logger EventLogger, dispatcher *ArchiveDispatcher) (Middleware, error) {
	cfg, err := ParseArchiveConfig(config)
	if err != nil {
		return nil, err
	}
	recipientKey, err := ValidateArchiveConfig(cfg)
	if err != nil {
		return nil, err
	}
	if dispatcher == nil {
		return nil, fmt.Errorf("archive: dispatcher not available")
	}
	return &Archive{
		cfg:          cfg,
		eventLogger:  logger,
		dispatcher:   dispatcher,
		recipientKey: recipientKey,
	}, nil
}

func (a *Archive) Name() string { return "archive" }

func (a *Archive) ProcessRequest(ctx context.Context, req *mcp.Request, meta *RequestMeta) (*mcp.Request, error) {
	if a.cfg.Include != "request" {
		return req, nil // will archive in ProcessResponse with both request+response
	}

	reqJSON, _ := json.Marshal(req)
	payload := a.buildPayload("request", meta, reqJSON, nil)
	a.enqueue(payload, meta)
	return req, nil
}

func (a *Archive) ProcessResponse(ctx context.Context, req *mcp.Request, resp *mcp.Response, meta *RequestMeta) (*mcp.Response, error) {
	if a.cfg.Include == "request" {
		return resp, nil // already archived in ProcessRequest
	}

	var reqJSON json.RawMessage
	var respJSON json.RawMessage

	if a.cfg.Include == "both" {
		reqJSON, _ = json.Marshal(req)
	}
	respJSON, _ = json.Marshal(resp)

	phase := "response"
	if a.cfg.Include == "both" {
		phase = "exchange"
	}

	payload := a.buildPayload(phase, meta, reqJSON, respJSON)
	a.enqueue(payload, meta)
	return resp, nil
}

func (a *Archive) buildPayload(phase string, meta *RequestMeta, reqJSON, respJSON json.RawMessage) []byte {
	p := archivePayload{
		Version:   "v1",
		Source:    "arc_relay",
		Phase:     phase,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Meta: archiveMeta{
			ServerID:   meta.ServerID,
			ServerName: meta.ServerName,
			UserID:     meta.UserID,
			ClientIP:   meta.ClientIP,
			Method:     meta.Method,
			ToolName:   meta.ToolName,
			RequestID:  meta.RequestID,
		},
		Request:  reqJSON,
		Response: respJSON,
	}
	body, _ := json.Marshal(p)
	return body
}

func (a *Archive) enqueue(body []byte, meta *RequestMeta) {
	payload, err := sealArchivePayload(body, a.recipientKey)
	if err != nil {
		// Do not include the underlying error text in the operator-facing
		// event. sealArchivePayload errors come from rand.Read or box.Seal
		// and never contain plaintext, but we keep the summary terse to
		// avoid ever echoing back secret material if the error surface
		// expands in the future.
		slog.Error("archive: encryption failed", "error", err)
		if a.eventLogger != nil {
			a.eventLogger(&store.MiddlewareEvent{
				Middleware: "archive",
				EventType:  "error",
				Summary:    "archive payload dropped: envelope encryption failed",
			})
		}
		return
	}
	if err := a.dispatcher.EnqueueWithServer(payload, a.cfg, meta.ServerID); err != nil {
		slog.Error("archive: failed to enqueue", "error", err)
		if a.eventLogger != nil {
			a.eventLogger(&store.MiddlewareEvent{
				Middleware: "archive",
				EventType:  "error",
				Summary:    "failed to enqueue archive payload: " + err.Error(),
			})
		}
	}
}
