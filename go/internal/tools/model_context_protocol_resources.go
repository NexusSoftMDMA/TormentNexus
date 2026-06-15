//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"strings"
)

var resourceMap = map[string]string{

func HandleListMCPResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var sb strings.Builder
	sb.WriteString("Available MCP resources:\n")
	for name, desc := range resourceMap {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", name, desc))

	return ok(sb.String())
}

func HandleGetMCPResource(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, found := resourceMap[name]
	if !found {
		return err("resource not found: " + name)
	return ok(fmt.Sprintf("Resource: %s\nDescription: %s", name, desc))
}
}
}
}