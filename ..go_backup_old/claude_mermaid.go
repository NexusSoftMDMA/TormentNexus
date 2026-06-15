package tools

import (
	"context"
	"fmt"
	"net/url"
)

func HandleRenderMermaid(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagram, _ :=getString(args, "diagram")
	if diagram == "" {
		return err("diagram is required")
}

	encoded := url.QueryEscape(diagram)
	renderURL := fmt.Sprintf("https://mermaid.ink/img/%s", encoded)
	return success(fmt.Sprintf("View diagram: %s", renderURL))
}