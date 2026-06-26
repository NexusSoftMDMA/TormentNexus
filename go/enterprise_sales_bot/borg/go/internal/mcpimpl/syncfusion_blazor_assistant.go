package mcpimpl

import "context"

func HandleGetComponentGuidance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	if component == "" {
		return err("component is required")
}

	guidance := "Component: " + component + ". For Blazor, use <SyncfusionBlazor." + component + "> with appropriate parameters."
	return ok(guidance)
}

func HandleGenerateComponentCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	props, _ :=getString(args, "properties")
	if component == "" {
		return err("component is required")
}

	code := "<SyncfusionBlazor." + component + " " + props + "></SyncfusionBlazor." + component + ">"
	return ok(code)
}