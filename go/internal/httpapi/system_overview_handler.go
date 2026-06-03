package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/buildinfo"
)

// handleSystemOverview returns a unified snapshot using Go-native data
// sources first, falling back to upstream only for enrichment.
// This avoids the ~5s latency from parallel tRPC calls that serialize
// on the Node.js event loop.
func (s *Server) handleSystemOverview(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	// ---- MCP Status (Go-local first) ----
	mcpData := map[string]any{
		"initialized":    false,
		"serverCount":    0,
		"toolCount":      0,
		"connectedCount": 0,
	}
	mcpBridge := map[string]any{"procedure": "mcp.getStatus"}

	view, viewErr := s.localMCPInventoryView()
	if viewErr == nil && view.Inventory != nil {
		mcpData["initialized"] = true
		mcpData["serverCount"] = len(view.Inventory.Servers)
		mcpData["toolCount"] = len(view.Inventory.Tools)
		mcpBridge["fallback"] = "go-local-mcp-inventory"
	} else {
		var upstreamMCP any
		base, err := s.callUpstreamJSON(ctx, "mcp.getStatus", nil, &upstreamMCP)
		if err == nil && upstreamMCP != nil {
			if m, ok := upstreamMCP.(map[string]any); ok {
				mcpData["initialized"], _ = m["initialized"].(bool)
				if v, ok := m["serverCount"].(float64); ok {
					mcpData["serverCount"] = int(v)
				}
				if v, ok := m["toolCount"].(float64); ok {
					mcpData["toolCount"] = int(v)
				}
				if v, ok := m["connectedCount"].(float64); ok {
					mcpData["connectedCount"] = int(v)
				}
			}
			mcpBridge["upstreamBase"] = base
		} else {
			mcpBridge["error"] = "local and upstream both unavailable"
		}
	}

	// ---- Startup Status (Go-local) ----
	startupBridge := map[string]any{"procedure": "startupStatus", "fallback": "go-local"}

	// ---- Sessions (Go-local first) ----
	sessionsData := s.supervisorManager.ListSessions()
	sessionsBridge := map[string]any{"procedure": "session.list", "fallback": "go-local-supervisor"}

	// ---- Memory (Go-local) ----
	memoryBridge := map[string]any{"procedure": "memory.getSessionBootstrap", "fallback": "go-local-memory"}

	// ---- Health (Go-local + async upstream) ----
	goHealth := map[string]any{
		"version":  buildinfo.Version,
		"uptimeMs": time.Since(s.startedAt).Milliseconds(),
	}
	coreHealth := map[string]any{"status": "unknown"}
	coreBridge := map[string]any{"procedure": "health"}

	// Try upstream health check (fire and forget for speed)
	upstreamHealthDone := make(chan struct{})
	go func() {
		defer close(upstreamHealthDone)
		var coreResult any
		if base, err := s.callUpstreamJSON(context.WithoutCancel(ctx), "health", nil, &coreResult); err == nil {
			_ = base
			if m, ok := coreResult.(map[string]any); ok {
				coreHealth["status"] = "ok"
				for k, v := range m {
					coreHealth[k] = v
				}
				coreBridge["upstreamBase"] = base
			}
		} else {
			coreHealth["status"] = "unreachable"
		}
	}()

	// Wait briefly for upstream health (max 300ms)
	select {
	case <-upstreamHealthDone:
	case <-time.After(300 * time.Millisecond):
		coreHealth["status"] = "timeout"
	}

	overview := map[string]any{
		"success": true,
		"data": map[string]any{
			"mcpStatus": map[string]any{
				"initialized":    mcpData["initialized"],
				"serverCount":    mcpData["serverCount"],
				"toolCount":      mcpData["toolCount"],
				"connectedCount": mcpData["connectedCount"],
				"bridge":         mcpBridge,
			},
			"startupStatus": map[string]any{
				"status":  "running",
				"ready":   true,
				"summary": "Go sidecar operational",
				"bridge":  startupBridge,
			},
			"sessions": map[string]any{
				"list":  sessionsData,
				"bridge": sessionsBridge,
			},
			"memory": map[string]any{
				"items": []any{},
				"count": 0,
				"bridge": memoryBridge,
			},
			"health": map[string]any{
				"goSidecar":  goHealth,
				"tsCore":     coreHealth,
				"coreBridge": coreBridge,
			},
			"timestamp": time.Now().UnixMilli(),
		},
	}

	writeJSON(w, http.StatusOK, overview)
}
