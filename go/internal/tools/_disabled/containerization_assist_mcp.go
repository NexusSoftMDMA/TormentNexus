package tools

import (
	"context"
	"fmt"
)

func HandleGenerateDockerfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	app, _ :=getString(args, "app")
	msg := fmt.Sprintf("Generated Dockerfile using base %s for app %s", base, app)
	return ok(msg)
}

func HandleAnalyzeContainer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	image, _ :=getString(args, "image")
	msg := fmt.Sprintf("Analyzed container image: %s", image)
	return ok(msg)
}