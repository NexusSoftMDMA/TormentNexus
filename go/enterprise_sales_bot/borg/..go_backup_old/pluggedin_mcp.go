package tools

import (
	"context"
	"time"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello from Pluggedin Mcp!")
}

	return ok("Hello " + name + ", welcome to Pluggedin Mcp!")
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	now := time.Now()
	if format == "unix" {
		return ok(success("Current unix timestamp: " + time.Unix(now.Unix(), 0).String()))
}

	return ok("Current time: " + now.Format(time.RFC3339))
}