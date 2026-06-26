package mcp

import (
	"context"
	"fmt"
)

// ToolSchema represents a simplified MCP tool schema
type ToolSchema struct {
	Name        string
	Description string
	InputSchema interface{}
}

// SemanticToolIndex defines Layer 1: Semantic Search
type SemanticToolIndex interface {
	SearchTools(ctx context.Context, query string, limit int) ([]ToolSchema, error)
}

// ProgressiveRouter defines Layer 2: The Router
type ProgressiveRouter struct {
	index SemanticToolIndex
}

func NewProgressiveRouter(index SemanticToolIndex) *ProgressiveRouter {
	return &ProgressiveRouter{index: index}
}

// FilterContext retrieves the top N relevant tools to prevent context bloat
func (pr *ProgressiveRouter) FilterContext(ctx context.Context, prompt string) ([]ToolSchema, error) {
	fmt.Println("[ProgressiveRouter] Running Layer 1 Semantic Search...")
	// We retrieve only the top 5 to 10 relevant tool schemas
	tools, err := pr.index.SearchTools(ctx, prompt, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search tools: %w", err)
	}

	fmt.Printf("[ProgressiveRouter] Layer 1 returned %d relevant tools.\n", len(tools))
	return tools, nil
}
