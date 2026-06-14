package tools

import (
	"context"
	"fmt"
)

func HandleSearchMedical(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	term, _ :=getString(args, "term")
	if term == "" {
		return err("term is required")
}

	return ok(fmt.Sprintf("Search results for '%s' would be returned here", term))
}

func HandleDrugInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	drug, _ :=getString(args, "drug")
	if drug == "" {
		return err("drug is required")
}

	return success(fmt.Sprintf("Drug info for '%s' would be returned here", drug))
}