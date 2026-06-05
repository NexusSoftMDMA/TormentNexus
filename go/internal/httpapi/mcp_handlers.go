package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
	"github.com/tormentnexushq/tormentnexus-go/internal/cache"
)

func (s *Server) handleMCPStatus(w http.ResponseWriter, r *http.Request) {
	// Cache MCP status for 10s to reduce upstream calls
	val, err := cache.Cached(s.cacheService, "mcp:status", func() (interface{}, error) {
		return s.buildMCPStatus(r.Context())
	}, 30000)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, val)
}

func (s *Server) buildMCPStatus(ctx context.Context) (map[string]any, error) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(ctx, "mcp.getStatus", nil, &result)
	if err == nil {
		return map[string]any{
			"success": true,
			"data": result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure": "mcp.getStatus",
			},
		}, nil
	}
	_, summary, localErr := s.localMCPSummary(ctx)
	if localErr != nil {
		return nil, localErr
	}
	return map[string]any{
		"success": true,
		"data": map[string]any{
			"initialized": true,
			"connected": summary.SourceBackedHarnessCount > 0,
			"toolCount": summary.SourceBackedToolCount,
			"serverCount": summary.InstalledHarnessCount,
			"connectedCount": summary.SourceBackedHarnessCount,
			"sourceBackedHarnessCount": summary.SourceBackedHarnessCount,
			"source": "source-backed-local-summary",
			"lazySessionMode": false,
			"singleActiveServerMode": false,
		},
		"bridge": map[string]any{
			"fallback": "go-local-mcp",
			"procedure": "mcp.getStatus",
			"reason": "upstream unavailable; using local MCP harness summary",
		},
	}, nil
}

func (s *Server) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.listTools", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.listTools",
			},
		})
		return
	}

	view, invErr := s.localMCPInventoryView()
	if invErr == nil && view != nil && len(view.Inventory.Tools) > 0 {
		bridge := map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.listTools",
			"reason":    "upstream unavailable; using local MCP inventory cache",
		}
		for key, value := range inventoryBridgeMeta(view) {
			bridge[key] = value
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    fallbackMCPInventoryTools(view),
			"bridge":  bridge,
		})
		return
	}

	_, summary, localErr := s.localMCPSummary(r.Context())
	if localErr != nil {
		if invErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error()})
			return
		}
		bridge := map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.listTools",
			"reason":    "upstream unavailable; local MCP inventory cache is empty",
		}
		for key, value := range inventoryBridgeMeta(view) {
			bridge[key] = value
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    []map[string]any{},
			"bridge":  bridge,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackMCPTools(summary.InstalledHarnesses),
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.listTools",
			"reason":    "upstream unavailable; using local MCP tool inventory",
		},
	})
}

func (s *Server) handleMCPSearchTools(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	payload := map[string]any{"query": query}
	if profile := strings.TrimSpace(r.URL.Query().Get("profile")); profile != "" {
		payload["profile"] = profile
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.searchTools", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.searchTools",
			},
		})
		return
	}

	view, invErr := s.localMCPInventoryView()
	if invErr == nil && view != nil && len(view.Inventory.Tools) > 0 {
		bridge := map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.searchTools",
			"reason":    "upstream unavailable; using local MCP inventory cache",
		}
		for key, value := range inventoryBridgeMeta(view) {
			bridge[key] = value
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    fallbackSearchMCPInventoryTools(query, view, 20),
			"bridge":  bridge,
		})
		return
	}

	_, summary, localErr := s.localMCPSummary(r.Context())
	if localErr != nil {
		if invErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error()})
			return
		}
		bridge := map[string]any{
			"fallback": "go-local-mcp",

			"procedure": "mcp.searchTools",
			"reason":    "upstream unavailable; local MCP inventory cache is empty",
		}
		for key, value := range inventoryBridgeMeta(view) {
			bridge[key] = value
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    []map[string]any{},
			"bridge":  bridge,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackSearchMCPTools(summary.InstalledHarnesses, query),
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.searchTools",
			"reason":    "upstream unavailable; using local MCP inventory cache",
		},
	})
}

func (s *Server) handleMCPRuntimeServers(w http.ResponseWriter, r *http.Request) {
	var result any
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	upstreamBase, err := s.callUpstreamJSON(ctx, "mcp.listServers", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.listServers",
			},
		})
		return
	}

	view, invErr := s.localMCPInventoryView()
	_, summary, localErr := s.localMCPSummary(r.Context())
	if localErr != nil {
		if invErr != nil || view == nil || (view.PersistedOverlayServerCount == 0 && view.RuntimeOverlayServerCount == 0) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error()})
			return
		}
		bridge := map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.listServers",
			"reason":    "upstream unavailable; using local MCP runtime overlay cache",
		}
		for key, value := range inventoryBridgeMeta(view) {
			bridge[key] = value
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    fallbackRuntimeServerListWithPrimaryProvenance(nil, view),
			"bridge":  bridge,
		})
		return
	}
	baseServers := fallbackRuntimeServerListWithPrimaryProvenance(summary.InstalledHarnesses, view)
	bridge := map[string]any{
		"fallback":  "go-local-mcp",
		"procedure": "mcp.listServers",
		"reason":    "upstream unavailable; using local MCP runtime server summary",
	}
	for key, value := range inventoryBridgeMeta(view) {
		bridge[key] = value
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    baseServers,
		"bridge":  bridge,
	})
}

func (s *Server) handleMCPPredictTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		ChatHistory string `json:"chatHistory"`
		ActiveGoal  string `json:"activeGoal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try native Go prediction first
	predicted, err := s.mcpPredictor.PredictAndPreload(r.Context(), payload.ChatHistory, payload.ActiveGoal)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"predictedTools": predicted,
				"reasoning":      "Predicted via Go native predictor",
			},
			"bridge": map[string]any{
				"source": "go-native-prediction",
			},
		})
		return
	}

	// Fallback to upstream
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.predictTools", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.predictTools",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   err.Error(),
	})
}

func (s *Server) handleMCPSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	homeDir, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	appData := os.Getenv("APPDATA")

	targets := mcp.ResolveClientTargets(homeDir, appData, cwd)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"targets": targets,
		},
	})
}

// handleMCPServersList returns a combined view of runtime + configured servers.
func (s *Server) handleMCPServersList(w http.ResponseWriter, r *http.Request) {
	// Cache MCP server list for 10s to reduce upstream calls
	val, err := cache.Cached(s.cacheService, "mcp:servers", func() (interface{}, error) {
		return s.buildMCPServersList(r.Context())
	}, 30000)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, val)
}

func (s *Server) buildMCPServersList(ctx context.Context) (map[string]any, error) {
	var result any
	upstreamCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	upstreamBase, err := s.callUpstreamJSON(upstreamCtx, "mcp.listServers", nil, &result)
	if err == nil {
		return map[string]any{
			"success": true,
			"data": result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure": "mcp.listServers",
			},
		}, nil
	}

	view, _ := s.localMCPInventoryView()
	_, cliSummary, _ := s.localMCPSummary(ctx)

	type serverEntry struct {
		Name string `json:"name"`
		Status string `json:"status"`
		ToolCount int `json:"toolCount"`
	}

	var servers []serverEntry
	seen := make(map[string]bool)

	if view != nil {
		for name, srv := range view.PersistedOverlayRecords {
			if seen[name] {
				continue
			}
			seen[name] = true
			status := "configured"
			if srv.RuntimeConnected {
				status = "connected"
			}
			servers = append(servers, serverEntry{
				Name: srv.Name,
				Status: status,
				ToolCount: srv.ToolCount,
			})
		}
		for name, srv := range view.LiveOverlayRecords {
			if seen[name] {
				continue
			}
			seen[name] = true
			status := "configured"
			if srv.RuntimeConnected {
				status = "connected"
			}
			servers = append(servers, serverEntry{
				Name: srv.Name,
				Status: status,
				ToolCount: srv.ToolCount,
			})
		}
	}

	for _, h := range cliSummary.InstalledHarnesses {
		if seen[h.ID] {
			continue
		}
		seen[h.ID] = true
		servers = append(servers, serverEntry{
			Name: h.ID,
			Status: "available",
		})
	}

	return map[string]any{
		"success": true,
		"data": servers,
		"bridge": map[string]any{
			"fallback": "go-local-mcp",
			"procedure": "mcp.listServers",
			"reason": "upstream unavailable; using local MCP inventory",
		},
	}, nil
}

// handleMCPPredictConversational is the primary sidecar endpoint called by the
// TypeScript ConversationalToolInjector before falling back to cloud LLMs.
//
// Request:  POST /api/mcp/tools/predict-conversational
//           { "prompt": "...", "systemPrompt": "..." }
//
// Response: { "success": true, "data": { "tools": ["tool_name_1", ...] } }
func (s *Server) handleMCPPredictConversational(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Prompt       string `json:"prompt"`
		SystemPrompt string `json:"systemPrompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	if strings.TrimSpace(payload.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "prompt is required"})
		return
	}

	tools, err := s.conversationalPredictor.PredictFromPrompt(r.Context(), payload.SystemPrompt, payload.Prompt)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   err.Error(),
			"bridge":  map[string]any{"source": "go-native-ollama", "reason": "prediction failed"},
		})
		return
	}

	if tools == nil {
		tools = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"tools": tools,
		},
		"bridge": map[string]any{
			"source": "go-native-ollama",
		},
	})
}

// handleMCPConversationAppend receives an explicit conversation turn pushed by
// the TypeScript tRPC bridge or dashboard. This allows the Go-native predictor
// to build its own window independently for preemptive advertisement.
//
// Request:  POST /api/mcp/conversation/append
//           { "role": "user"|"assistant"|"tool", "text": "..." }
//
// Response: { "success": true }
func (s *Server) handleMCPConversationAppend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Role string `json:"role"`
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	role := strings.ToLower(strings.TrimSpace(payload.Role))
	if role != "user" && role != "assistant" && role != "tool" {
		role = "user"
	}
	text := strings.TrimSpace(payload.Text)
	if text == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "text is required"})
		return
	}

	s.conversationalPredictor.AppendTurn(role, text)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleMCPConversationWindow returns a debug snapshot of the current
// conversational predictor sliding window.
//
// Request:  GET /api/mcp/conversation/window
//
// Response: { "success": true, "data": { "turns": [...], "tokenCount": N } }
func (s *Server) handleMCPConversationWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	turns := s.conversationalPredictor.WindowSnapshot()
	tokenCount := s.conversationalPredictor.WindowTokenCount()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"turns":      turns,
			"tokenCount": tokenCount,
		},
	})
}

