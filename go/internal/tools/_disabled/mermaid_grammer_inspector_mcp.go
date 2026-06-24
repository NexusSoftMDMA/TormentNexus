package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleInspectMermaid(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("Missing 'code' argument")
}

	lines := strings.Split(code, "\n")
	info := fmt.Sprintf("Mermaid code has %d lines. Contains graph: %v, flowchart: %v", len(lines), strings.Contains(code, "graph"), strings.Contains(code, "flowchart"))
	return ok(info)
}