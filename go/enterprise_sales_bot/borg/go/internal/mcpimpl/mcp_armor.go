package mcpimpl

import (
	"context"
	"strings"
)

func HandleStaticScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	payload, _ :=getString(args, "payload")
	if payload == "" {
		return err("missing payload")
}

	if strings.Contains(payload, "dangerous") {
		return ok("Risk detected: found dangerous keyword")
}

	return ok("Static scan passed")
}

func HandleDynamicScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	payload, _ :=getString(args, "payload")
	if payload == "" {
		return err("missing payload")
}

	if len(payload) > 100 {
		return ok("Dynamic scan warning: payload too large")
}

	return ok("Dynamic scan passed")
}