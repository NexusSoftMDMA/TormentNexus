// Package middleware implements a bidirectional processing pipeline for MCP
// JSON-RPC messages flowing through the proxy. Each middleware can inspect and
// modify both requests (before the backend) and responses (before the client).
package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/store"
)

// Middleware processes MCP messages flowing through the proxy.
type Middleware interface {
	// Name returns the unique identifier for this middleware.
	Name() string

	// ProcessRequest is called before the request reaches the backend.
	// Return modified request, or error to block the request entirely.
	ProcessRequest(ctx context.Context, req *mcp.Request, meta *RequestMeta) (*mcp.Request, error)

	// ProcessResponse is called before the response reaches the client.
	// Return modified response, or error to inject an error response.
	ProcessResponse(ctx context.Context, req *mcp.Request, resp *mcp.Response, meta *RequestMeta) (*mcp.Response, error)
}

// RequestMeta carries context about the current request for middleware decisions.
type RequestMeta struct {
	UserID     string
	ServerID   string
	ServerName string
	Method     string // e.g. "tools/call", "tools/list"
	ToolName   string // for tools/call: which tool
	ClientIP   string
	RequestID  string // JSON-RPC id as string
}

// Pipeline holds an ordered list of middleware and executes them in sequence.
type Pipeline struct {
	middlewares []Middleware
}

// NewPipeline creates an empty pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Add appends a middleware to the pipeline.
func (p *Pipeline) Add(m Middleware) {
	p.middlewares = append(p.middlewares, m)
}

// Len returns the number of middleware in the pipeline.
func (p *Pipeline) Len() int {
	return len(p.middlewares)
}

// ProcessRequest runs all middleware on the request in order.
// If any middleware returns an error, processing stops and the error is returned.
func (p *Pipeline) ProcessRequest(ctx context.Context, req *mcp.Request, meta *RequestMeta) (*mcp.Request, error) {
	var err error
	for _, m := range p.middlewares {
		req, err = m.ProcessRequest(ctx, req, meta)
		if err != nil {
			return nil, err
		}
	}
	return req, nil
}

// ProcessResponse runs all middleware on the response in reverse order.
// If any middleware returns an error, processing stops and the error is returned.
func (p *Pipeline) ProcessResponse(ctx context.Context, req *mcp.Request, resp *mcp.Response, meta *RequestMeta) (*mcp.Response, error) {
	var err error
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		resp, err = p.middlewares[i].ProcessResponse(ctx, req, resp, meta)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// Descriptor describes a registered middleware for the UI and handler layer.
// It is provided at Register() time, not on the Middleware interface itself,
// so that processing-only middleware stay decoupled from presentation.
type Descriptor struct {
	Name            string   // registry key: "archive"
	DisplayName     string   // "Compliance Archive"
	Description     string   // one-liner for toggle card
	DefaultPriority int      // 40
	DisplayOrder    int      // UI ordering, separate from pipeline priority
	Scope           string   // "server", "global", "both"
	TemplateName    string   // "middleware/archive" - empty = toggle-only
	Actions         []string // ["test", "retry", "clear"] - whitelisted action names
}

// Registry holds middleware factories and builds pipelines from DB configs.
type Registry struct {
	factories         map[string]Factory
	descriptors       map[string]Descriptor
	store             *store.MiddlewareStore
	archiveDispatcher *ArchiveDispatcher
}

// Factory creates a middleware instance from a JSON config.
type Factory func(config json.RawMessage, eventLogger EventLogger) (Middleware, error)

// EventLogger is a callback for middleware to log events.
type EventLogger func(evt *store.MiddlewareEvent)

// NewRegistry creates a registry with the built-in middleware factories.
func NewRegistry(mwStore *store.MiddlewareStore, archiveDispatcher *ArchiveDispatcher) *Registry {
	r := &Registry{
		factories:         make(map[string]Factory),
		descriptors:       make(map[string]Descriptor),
		store:             mwStore,
		archiveDispatcher: archiveDispatcher,
	}
	// Register built-in middleware
	r.Register(Descriptor{
		Name:            "sanitizer",
		DisplayName:     "Sanitizer",
		Description:     "PII & secret redaction",
		DefaultPriority: 10,
		DisplayOrder:    10,
		Scope:           "server",
	}, NewSanitizerFromConfig)

	r.Register(Descriptor{
		Name:            "sizer",
		DisplayName:     "Content Sizer",
		Description:     "Response size limits",
		DefaultPriority: 20,
		DisplayOrder:    20,
		Scope:           "server",
	}, NewSizerFromConfig)

	r.Register(Descriptor{
		Name:            "alerter",
		DisplayName:     "Alerter",
		Description:     "Pattern monitoring",
		DefaultPriority: 30,
		DisplayOrder:    30,
		Scope:           "server",
	}, NewAlerterFromConfig)

	// Archive uses a closure to capture the shared dispatcher
	r.Register(Descriptor{
		Name:            "archive",
		DisplayName:     "Compliance Archive",
		Description:     "Audit trail",
		DefaultPriority: 40,
		DisplayOrder:    40,
		Scope:           "both",
		TemplateName:    "middleware_archive",
		Actions:         []string{"test", "retry", "clear", "status"},
	}, func(config json.RawMessage, logger EventLogger) (Middleware, error) {
		return NewArchiveFromConfig(config, logger, archiveDispatcher)
	})
	return r
}

// ArchiveDispatcher returns the shared archive dispatcher, or nil if not configured.
func (r *Registry) ArchiveDispatcher() *ArchiveDispatcher {
	return r.archiveDispatcher
}

// Register adds a middleware factory with its descriptor.
func (r *Registry) Register(desc Descriptor, factory Factory) {
	r.factories[desc.Name] = factory
	r.descriptors[desc.Name] = desc
}

// IsRegistered returns true if a middleware with the given name is registered.
func (r *Registry) IsRegistered(name string) bool {
	_, ok := r.factories[name]
	return ok
}

// Descriptor returns the descriptor for a named middleware, or ok=false.
func (r *Registry) Descriptor(name string) (Descriptor, bool) {
	d, ok := r.descriptors[name]
	return d, ok
}

// Descriptors returns all registered descriptors sorted by DisplayOrder.
func (r *Registry) Descriptors() []Descriptor {
	descs := make([]Descriptor, 0, len(r.descriptors))
	for _, d := range r.descriptors {
		descs = append(descs, d)
	}
	sort.Slice(descs, func(i, j int) bool {
		return descs[i].DisplayOrder < descs[j].DisplayOrder
	})
	return descs
}

// BuildPipeline creates a pipeline for a specific server by loading configs from the DB.
func (r *Registry) BuildPipeline(serverID string) *Pipeline {
	configs, err := r.store.GetForServer(serverID)
	if err != nil {
		slog.Error("middleware: failed to load configs for server", "server_id", serverID, "error", err)
		return NewPipeline()
	}

	pipeline := NewPipeline()
	for _, mc := range configs {
		if !mc.Enabled {
			continue
		}
		factory, ok := r.factories[mc.Middleware]
		if !ok {
			slog.Error("middleware: unknown middleware", "middleware", mc.Middleware, "server_id", serverID)
			continue
		}

		logger := r.makeEventLogger(serverID)
		m, err := factory(mc.Config, logger)
		if err != nil {
			slog.Error("middleware: failed to create middleware", "middleware", mc.Middleware, "server_id", serverID, "error", err)
			continue
		}
		pipeline.Add(m)
	}
	return pipeline
}

func (r *Registry) makeEventLogger(serverID string) EventLogger {
	return func(evt *store.MiddlewareEvent) {
		evt.ServerID = serverID
		if err := r.store.LogEvent(evt); err != nil {
			slog.Error("middleware: failed to log event", "error", err)
		}
	}
}
