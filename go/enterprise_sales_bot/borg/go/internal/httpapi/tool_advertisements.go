package httpapi

import (
	"net/http"
	"strings"
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
	searchPayload := map[string]any{
		"query": normalizedQuery,
	}
	if normalizedQuery != "" {
		searchPayload["profile"] = "repo-coding"
	}

	var recommendedTools any
	recommendedToolsBase, err := s.callUpstreamJSON(r.Context(), "mcp.searchTools", searchPayload, &recommendedTools)
	if err != nil {
		return ToolSuggestionSnapshot{}, err
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
		return ToolSuggestionSnapshot{}, err
	}

	return ToolSuggestionSnapshot{
		RecommendedTools: recommendedTools,
		RelatedTools:     relatedTools,
		Bridge: map[string]any{
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
		},
	}, nil
}
