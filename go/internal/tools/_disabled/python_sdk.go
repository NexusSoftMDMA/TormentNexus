package tools

import (
	"context"
)

func HandleGetPythonVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version := "3.12.0"
	return success(version)
}

func HandleRunPythonCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
	}
	result := "Executed: " + code
	return success(result)
}