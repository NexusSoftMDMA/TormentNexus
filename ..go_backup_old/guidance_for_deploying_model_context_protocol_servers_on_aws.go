package tools

import (
	"context"
	"fmt"
	"os"
)

func HandleGetOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("This Guidance provides steps to deploy MCP servers on AWS with OAuth 2.0, WAF, and CDN protection.")
}

func HandleCheckRegion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			return err("No region specified and AWS_REGION not set")

	}
	return ok(fmt.Sprintf("Using AWS region: %s", region))
}
}