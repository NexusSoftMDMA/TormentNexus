package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleListModules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modules := []string{"core", "utils", "network", "data"}
	return ok(strings.Join(modules, ", "))
}

func HandleDescribeModule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	module, _ :=getString(args, "module")
	descriptions := map[string]string{
		"core":    "Core utilities",
		"utils":   "General utilities",
		"network": "Network operations",
		"data":    "Data processing",
	}
	desc, found := descriptions[module]
	if found {
		return ok(desc)
}

	return err(fmt.Sprintf("Module '%s' not found", module))
}