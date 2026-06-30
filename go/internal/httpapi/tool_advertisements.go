package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type ToolSuggestionSnapshot struct {
	RecommendedTools any            `json:"recommendedTools"`
	RelatedTools     any            `json:"relatedTools"`
	Bridge           map[string]any `json:"bridge"`
}

func (s *Server) buildToolSuggestionSnapshot(r *http.Request, query string) (ToolSuggestionSnapshot, error) {
	return s.buildToolSuggestionSnapshotWithLimit(r, query, 8)
}

func (s *Server) buildToolSuggestionSnapshotWithLimit(r *http.Request, query string, limit int) (ToolSuggestionSnapshot, error) {
	normalizedQuery := strings.TrimSpace(query)

	// Inject Predictive tool recommendations using local FreeLLM
	var predicted []string
	var predictErr error
	if normalizedQuery != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		predicted, predictErr = ai.PredictTools(ctx, normalizedQuery, s.toolsRegistry.List())
		cancel()
	}

	searchPayload := map[string]any{
		"query": normalizedQuery,
	}
	if normalizedQuery != "" {
		searchPayload["profile"] = "repo-coding"
	}

	var recommendedTools any
	recommendedToolsBase, err := s.callUpstreamJSON(r.Context(), "mcp.searchTools", searchPayload, &recommendedTools)
	if err != nil {
		// If upstream fails, use the local predicted tools or empty list
		if len(predicted) > 0 {
			recommendedTools = map[string]any{
				"tools": predicted,
			}
		} else {
			recommendedTools = map[string]any{
				"tools": []string{},
			}
		}
	} else if len(predicted) > 0 {
		// Merge predicted tools into recommended tools if upstream succeeded
		if recMap, ok := recommendedTools.(map[string]any); ok {
			recMap["predictedTools"] = predicted
			recommendedTools = recMap
		}
	}

	var relatedTools any
	relatedToolsBase, err := s.callUpstreamJSON(r.Context(), "mcp.callTool", map[string]any{
		"name": "list_all_tools",
		"args": map[string]any{
			"query": normalizedQuery,
			"limit": limit,
		},
	}, &relatedTools)
	if err != nil {
		// Return static empty results for related tools on error
		relatedTools = map[string]any{}
	}

	bridgeInfo := map[string]any{
		"recommendedTools": map[string]any{
			"upstreamBase": recommendedToolsBase,
			"procedure":    "mcp.searchTools",
		},
		"relatedTools": map[string]any{
			"upstreamBase": relatedToolsBase,
			"procedure":    "mcp.callTool",
			"toolName":     "list_all_tools",
			"limit":        limit,
		},
	}
	if predictErr != nil {
		bridgeInfo["predictiveError"] = predictErr.Error()
	}

	return ToolSuggestionSnapshot{
		RecommendedTools: recommendedTools,
		RelatedTools:     relatedTools,
		Bridge:           bridgeInfo,
	}, nil
}
