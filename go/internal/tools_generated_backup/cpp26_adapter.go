package tools

import (
	"context"
)

func HandleGetCpp26Features(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	features := "- Deducing this\n- std::print and std::println\n- Pattern matching (experimental)\n- Contracts (experimental)\n- Reflection (experimental)"
	return success(features)
}