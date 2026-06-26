package mcpimpl

import "context"

func HandleAnalyzeLWC(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "componentName")
	if componentName == "" {
		return err("componentName is required")
}

	return ok("Analyzed LWC component: " + componentName)
}

func HandleImproveLWC(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "componentName")
	issue, _ :=getString(args, "issue")
	if componentName == "" {
		return err("componentName is required")
}

	return ok("Improvement suggestions for " + componentName + ": " + issue)
}