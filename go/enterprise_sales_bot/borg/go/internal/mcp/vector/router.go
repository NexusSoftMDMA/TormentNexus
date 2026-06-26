package vector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/llm"
)

type ToolRouter struct {
	store *VectorStore
}

func NewToolRouter(store *VectorStore) *ToolRouter {
	return &ToolRouter{store: store}
}

func (r *ToolRouter) RouteForPrompt(queryText string, queryVec []float32, topK int) ([]llm.ToolSchema, []SearchResult, error) {
	if topK <= 0 {
		topK = 10
	}
	results, err := r.store.Search(SearchQuery{QueryText: queryText, QueryVec: queryVec, TopK: topK, MinScore: 0.3})
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}
	schemas := make([]llm.ToolSchema, 0, len(results))
	for _, res := range results {
		var params map[string]interface{}
		if json.Unmarshal([]byte(res.Tool.SchemaJSON), &params) != nil {
			params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
		}
		schemas = append(schemas, llm.ToolSchema{Type: "function", Function: llm.FunctionDef{Name: res.Tool.ToolName, Description: res.Tool.Description, Parameters: params}})
	}
	return schemas, results, nil
}

func (r *ToolRouter) SelectTool(toolName string, candidates []SearchResult) (*ToolRecord, error) {
	for _, c := range candidates {
		if c.Tool.ToolName == toolName {
			go func(id string) { _ = r.store.RecordUsage(id, true) }(c.Tool.ID)
			return &c.Tool, nil
		}
	}
	tool, err := r.store.GetTool(toolName)
	if err != nil || tool == nil {
		return nil, fmt.Errorf("tool %q not found", toolName)
	}
	return tool, nil
}

func (r *ToolRouter) FormatToolSummary(results []SearchResult) string {
	if len(results) == 0 {
		return "No tools matched."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Routed %d tools:\n", len(results)))
	for _, res := range results {
		boost := ""
		if res.Boosted {
			boost = " [boosted]"
		}
		desc := res.Tool.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %d. %s/%s (score=%.3f%s) - %s\n", res.Rank, res.Tool.ServerName, res.Tool.ToolName, res.Score, boost, desc))
	}
	return sb.String()
}
