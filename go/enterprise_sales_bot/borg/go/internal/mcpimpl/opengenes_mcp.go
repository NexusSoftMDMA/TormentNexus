package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearch_opengenes_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gene, _ :=getString(args, "gene")
	if gene == "" {
		return err("missing 'gene' parameter")
}

	return success(fmt.Sprintf("Opengenes Mcp: received gene query '%s'", gene))
}