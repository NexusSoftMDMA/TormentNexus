package tools

import (
	"context"
	"fmt"
)

func HandleSearchMCEHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	scope, _ :=getString(args, "scope")
	if scope == "" {
		scope = "MCE"
	}
	result := fmt.Sprintf("Search results for '%s' (scope: %s):\n- AMPscript Functions\n- SSJS Methods\n- GTL Syntax", query, scope)
	return success(result)
}

func HandleRunAmpScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("AMPscript code is required")
}

	result := fmt.Sprintf("Executed AMPscript:\n%s\n\nOutput: (simulated) Hello, World!", code)
	return ok(result)
}