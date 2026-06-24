package tools

import (
	"context"
	"fmt"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gene, _ :=getString(args, "gene")
	if gene == "" {
		return err("missing 'gene' parameter")
}

	return success(fmt.Sprintf("Opengenes Mcp: received gene query '%s'", gene))
}