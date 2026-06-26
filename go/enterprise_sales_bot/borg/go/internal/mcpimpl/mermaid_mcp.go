package mcpimpl

import (
	"context"
	"fmt"
	"net/url"
)

func HandleGenerateDiagram_mermaid_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	desc, _ :=getString(args, "description")
	if desc == "" {
		return err("description is required")
}

	code := "graph TD\n    A[" + desc + "] --> B[End]"
	return ok("```mermaid\n" + code + "\n```")
}

func HandleRenderDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	encoded := url.QueryEscape(code)
	link := fmt.Sprintf("https://mermaid.ink/img/%s", encoded)
	return success(link)
}