package tools

import (
	"context"
	"net/http"
)

func HandleScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target parameter is required")
}

	_, e := http.DefaultClient.Get("https://example.com")
	if e != nil {
		return err("network error: " + e.Error())
}

	return ok("scan initiated for " + target)
}

func HandleRules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	return ok("rules for category: " + category)
}