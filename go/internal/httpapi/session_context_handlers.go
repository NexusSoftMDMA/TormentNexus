package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/interop"
)

type SessionBootstrapPayload struct {
	Goal                   string `json:"goal,omitempty"`
	Objective              string `json:"objective,omitempty"`
	SummaryCount           int    `json:"summaryCount"`
	ObservationCount       int    `json:"observationCount"`
	ToolAdvertisementCount int    `json:"toolAdvertisementCount"`
	Prompt                 string `json:"prompt"`
}

type SessionContext struct {
	ActiveGoal       string                  `json:"activeGoal,omitempty"`
	LastObjective    string                  `json:"lastObjective,omitempty"`
	Startup          StartupStatus           `json:"startup"`
	Bootstrap        SessionBootstrapPayload `json:"bootstrap"`
	RecommendedTools any                     `json:"recommendedTools"`
	ToolAds          any                     `json:"toolAds"`
	Bridge           map[string]any          `json:"bridge"`
}

func (s *Server) handleSessionContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	activeGoal := strings.TrimSpace(r.URL.Query().Get("activeGoal"))
	lastObjective := strings.TrimSpace(r.URL.Query().Get("lastObjective"))

	startup, err := s.buildStartupStatus(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	bootstrapPayload := map[string]any{}
	if activeGoal != "" {
		bootstrapPayload["activeGoal"] = activeGoal
	}
	if lastObjective != "" {
		bootstrapPayload["lastObjective"] = lastObjective
	}

	var bootstrap SessionBootstrapPayload
	bootstrapBase, err := s.callUpstreamJSON(r.Context(), "memory.getSessionBootstrap", bootstrapPayload, &bootstrap)
	if err != nil {
		promptParts := make([]string, 0, 3)
		promptParts = append(promptParts, "Memory bootstrap:")
		if strings.TrimSpace(activeGoal) != "" {
			promptParts = append(promptParts, "Current goal: "+activeGoal)
		}
		if strings.TrimSpace(lastObjective) != "" {
			promptParts = append(promptParts, "Last objective: "+lastObjective)
		}
		promptParts = append(promptParts, "No relevant prior memory was found.")
		bootstrap = SessionBootstrapPayload{
			Goal:                   activeGoal,
			Objective:              lastObjective,
			SummaryCount:           0,
			ObservationCount:       0,
			ToolAdvertisementCount: 0,
			Prompt:                 strings.Join(promptParts, "\n"),
		}
		bootstrapBase = ""
	}

	query := strings.TrimSpace(strings.Join([]string{lastObjective, activeGoal}, " "))
	toolSuggestions, err := s.buildToolSuggestionSnapshot(r, query)
	if err != nil {
		_, summary, fallbackErr := s.localMCPSummary(r.Context())
		if fallbackErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error(), "detail": fallbackErr.Error()})
			return
		}
		searchResults := fallbackSearchMCPTools(summary.InstalledHarnesses, query)
		toolSuggestions = ToolSuggestionSnapshot{
			RecommendedTools: searchResults,
			RelatedTools: map[string]any{
				"toolName": "list_all_tools",
				"args": map[string]any{
					"query": query,
					"limit": 8,
				},
				"preview": map[string]any{
					"ok": true,
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": "list_all_tools"}},
					},
				},
			},
			Bridge: map[string]any{
				"recommendedTools": map[string]any{
					"fallback":  "go-local-mcp",
					"procedure": "mcp.searchTools",
					"reason":    err.Error(),
				},
				"toolAds": map[string]any{
					"fallback":  "go-local-mcp",
					"procedure": "mcp.callTool",
					"toolName":  "list_all_tools",
					"reason":    err.Error(),
				},
			},
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": SessionContext{
			ActiveGoal:       activeGoal,
			LastObjective:    lastObjective,
			Startup:          startup,
			Bootstrap:        bootstrap,
			RecommendedTools: toolSuggestions.RecommendedTools,
			ToolAds:          toolSuggestions.RelatedTools,
			Bridge: map[string]any{
				"bootstrap": map[string]any{
					"upstreamBase": bootstrapBase,
					"procedure":    "memory.getSessionBootstrap",
				},
				"recommendedTools": toolSuggestions.Bridge["recommendedTools"],
				"toolAds":          toolSuggestions.Bridge["relatedTools"],
			},
		},
	})
}

func (s *Server) callUpstreamJSON(ctx context.Context, procedure string, payload any, target any) (string, error) {
	result, err := interop.CallTRPCProcedure(ctx, s.cfg.MainLockPath(), procedure, payload)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(result.Data, target); err != nil {
		return "", err
	}
	return result.BaseURL, nil
}
