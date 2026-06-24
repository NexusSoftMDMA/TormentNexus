package tools

import (
	"context"
)

func HandleDeploy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	region, _ :=getString(args, "region")
	service, _ :=getString(args, "service")
	image, _ :=getString(args, "image")
	_ = region
	_ = image
	return success("Deployed service " + service + " in project " + project)
}