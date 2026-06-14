package tools

import "context"

func HandleCsharpSdkInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("The official C# SDK for Model Context Protocol is available. Maintained with Microsoft.")
}

func HandleCsharpSdkVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "C# SDK"
	}
	return success("Version of " + name + " is 1.0.0")
}