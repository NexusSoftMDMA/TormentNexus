package tools

import (
	"context"
	"fmt"
)

func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	var resources []string
	if filter == "" {
		resources = []string{
			"OpenClaw GitHub - https://github.com/openclaw/openclaw",
			"OpenClaw Docs - https://docs.openclaw.ai",
			"Tutorial: Getting Started - https://example.com/openclaw-tutorial",
			"Article: Self-hosted AI agents - https://example.com/ai-agents",
		}
	} else {
		resources = []string{"Filtered resource for " + filter}
	}
	msg := "Curated OpenClaw resources:\n"
	for _, r := range resources {
		msg += "- " + r + "\n"
	}
	return ok(msg)
}