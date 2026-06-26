package httpapi

import (
	"net/http"
	"strings"
)

type ToolContextPayload struct {
	ToolName         string   `json:"toolName"`
	Query            string   `json:"query"`
	MatchedPaths     []string `json:"matchedPaths,omitempty"`
	ObservationCount int      `json:"observationCount"`
	SummaryCount     int      `json:"summaryCount"`
	Prompt           string   `json:"prompt"`
}

type ToolsContext struct {
	ToolName         string             `json:"toolName"`
	ActiveGoal       string             `json:"activeGoal,omitempty"`
	LastObjective    string             `json:"lastObjective,omitempty"`
	Startup          StartupStatus      `json:"startup"`
	ToolContext      ToolContextPayload `json:"toolContext"`
	RecommendedTools any                `json:"recommendedTools"`
	RelatedTools     any                `json:"relatedTools"`
	Bridge           map[string]any     `json:"bridge"`
}

func (s *Server) handleToolsContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	toolName := strings.TrimSpace(r.URL.Query().Get("toolName"))
	if toolName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing toolName query parameter"})
		return
	}

	activeGoal := strings.TrimSpace(r.URL.Query().Get("activeGoal"))
	lastObjective := strings.TrimSpace(r.URL.Query().Get("lastObjective"))

	startup, err := s.buildStartupStatus(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	toolContextPayload := map[string]any{"toolName": toolName}
	if activeGoal != "" {
		toolContextPayload["activeGoal"] = activeGoal
	}
	if lastObjective != "" {
		toolContextPayload["lastObjective"] = lastObjective
	}

	var toolContext ToolContextPayload
	toolContextBase, err := s.callUpstreamJSON(r.Context(), "memory.getToolContext", toolContextPayload, &toolContext)
	if err != nil {
		query := strings.TrimSpace(strings.Join([]string{toolName, lastObjective, activeGoal}, " "))
		if query == "" {
			query = toolName
		}
		toolContext = ToolContextPayload{
			ToolName:         toolName,
			Query:            query,
			MatchedPaths:     []string{},
			ObservationCount: 0,
			SummaryCount:     0,
			Prompt:           "JIT tool context for " + toolName + ":\nNo relevant prior memory was found.",
		}
		toolContextBase = ""
	}

	toolAdsQuery := strings.TrimSpace(toolContext.Query)
	if toolAdsQuery == "" {
		toolAdsQuery = strings.TrimSpace(strings.Join([]string{toolName, lastObjective, activeGoal}, " "))
	}

	toolSuggestions, err := s.buildToolSuggestionSnapshot(r, toolAdsQuery)
	if err != nil {
		_, summary, fallbackErr := s.localMCPSummary(r.Context())
		if fallbackErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error(), "detail": fallbackErr.Error()})
			return
		}
		searchResults := fallbackSearchMCPTools(summary.InstalledHarnesses, toolAdsQuery)
		related := map[string]any{
			"toolName": "list_all_tools",
			"args": map[string]any{
				"query": toolAdsQuery,
				"limit": 8,
			},
			"preview": map[string]any{
				"ok": true,
				"result": map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": "list_all_tools",
						},
					},
				},
			},
		}
		toolSuggestions = ToolSuggestionSnapshot{
			RecommendedTools: searchResults,
			RelatedTools:     related,
			Bridge: map[string]any{
				"recommendedTools": map[string]any{
					"fallback":  "go-local-mcp",
					"procedure": "mcp.searchTools",
					"reason":    err.Error(),
				},
				"relatedTools": map[string]any{
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
		"data": ToolsContext{
			ToolName:         toolName,
			ActiveGoal:       activeGoal,
			LastObjective:    lastObjective,
			Startup:          startup,
			ToolContext:      toolContext,
			RecommendedTools: toolSuggestions.RecommendedTools,
			RelatedTools:     toolSuggestions.RelatedTools,
			Bridge: map[string]any{
				"toolContext": map[string]any{
					"upstreamBase": toolContextBase,
					"procedure":    "memory.getToolContext",
				},
				"recommendedTools": toolSuggestions.Bridge["recommendedTools"],
				"relatedTools":     toolSuggestions.Bridge["relatedTools"],
			},
		},
	})
}
