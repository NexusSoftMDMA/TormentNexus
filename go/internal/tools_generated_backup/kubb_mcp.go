package tools

import (
	"context"
)

func HandleGenerateCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	specUrl, _ :=getString(args, "specUrl")
	outputDir, _ :=getString(args, "outputDir")
	return success("Code generation started for spec at " + specUrl + " into " + outputDir)
}

func HandleListPlugins(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available plugins: typescript, javascript, zod, react-query")
}