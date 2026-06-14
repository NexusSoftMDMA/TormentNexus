package tools

import (
	"context"
)

func HandleSQLMap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
	}
	return success("sqlmap scan started for " + url)
}

func HandleNmap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("missing target")
	}
	return success("nmap scan started for " + target)
}