package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
)

// handleServiceConnectivity reports the health of connections between
// the Go sidecar and all upstream/downstream services.
func (s *Server) handleServiceConnectivity(w http.ResponseWriter, r *http.Request) {
	sd := config.DefaultServiceDiscovery()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := map[string]any{}

	// 1. Check tRPC upstream (TypeScript core)
	trpcStatus := checkTRPCUpstream(ctx, sd.TRPCUpstreamURLs)
	services["trpcUpstream"] = trpcStatus

	// 2. Check dashboard (Next.js)
	dashboardStatus := checkHTTPService(ctx, sd.DashboardBaseURL(), "dashboard")
	services["dashboard"] = dashboardStatus

	// 3. Check bridge (SSE/WebSocket)
	bridgeStatus := checkHTTPService(ctx, sd.BridgeBaseURL(), "bridge")
	services["bridge"] = bridgeStatus

	// 4. Go sidecar self-status
	services["goSidecar"] = map[string]any{
		"status":    "running",
		"port":      sd.GoSidecarPort,
		"baseURL":   sd.GoSidecarBaseURL(),
		"reachable": true,
	}

	// Overall summary
	allHealthy := true
	for _, svc := range services {
		if m, ok := svc.(map[string]any); ok {
			if reachable, ok := m["reachable"].(bool); ok && !reachable {
				allHealthy = false
				break
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"allHealthy": allHealthy,
		"services":   services,
		"discovery": map[string]any{
			"goSidecarPort":  sd.GoSidecarPort,
			"trpcUpstreams":  sd.TRPCUpstreamURLs,
			"bridgePort":     sd.BridgePort,
			"dashboardPort":  sd.DashboardPort,
			"dashboardHost":  sd.DashboardHost,
		},
	})
}

func checkTRPCUpstream(ctx context.Context, urls []string) map[string]any {
	for _, base := range urls {
		targetURL := strings.TrimRight(base, "/") + "/health"

		result := probeURL(ctx, targetURL, 3*time.Second)
		if result["status"] == "reachable" {
			result["activeBaseURL"] = base
			return result
		}
	}

	return map[string]any{
		"status":    "unreachable",
		"reachable": false,
		"tried":     urls,
		"error":     "all tRPC upstream candidates failed",
	}
}

func checkHTTPService(ctx context.Context, baseURL string, name string) map[string]any {
	healthURL := strings.TrimRight(baseURL, "/") + "/health"
	result := probeURL(ctx, healthURL, 3*time.Second)
	result["baseURL"] = baseURL
	result["name"] = name
	return result
}

func probeURL(ctx context.Context, targetURL string, timeout time.Duration) map[string]any {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return map[string]any{
			"status":    "error",
			"reachable": false,
			"url":       targetURL,
			"error":     err.Error(),
		}
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{
			"status":    "unreachable",
			"reachable": false,
			"url":       targetURL,
			"error":     err.Error(),
		}
	}
	defer resp.Body.Close()

	return map[string]any{
		"status":     "reachable",
		"reachable":  true,
		"url":        targetURL,
		"statusCode": resp.StatusCode,
	}
}

// handleMCPClientSync handles POST requests to sync MCP server configuration
// to supported IDE clients (Claude Desktop, Cursor, VS Code).
func (s *Server) handleMCPClientSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Client string `json:"client"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	sd := config.DefaultServiceDiscovery()
	sidecarBase := sd.GoSidecarBaseURL()

	// Build TormentNexus MCP server entries for the target client
	servers := map[string]any{
		"tormentnexus": map[string]any{
			"url":   fmt.Sprintf("%s/mcp", sidecarBase),
			"notes": "TormentNexus Go Control Plane — aggregated MCP router",
		},
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"client":        req.Client,
			"servers":       servers,
			"sidecarBase":   sidecarBase,
		},
	})
}
