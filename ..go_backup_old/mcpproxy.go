package tools

import (
	"context"
	"fmt"
)

func HandleListProxies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	if filter != "" {
		return ok(fmt.Sprintf("Filtered proxies: %s", filter))
}

	return ok("Available proxies: proxy1, proxy2, proxy3")
}

func HandleGetProxy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("proxy name is required")
}

	return ok(fmt.Sprintf("Proxy '%s' is active", name))
}