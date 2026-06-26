package tools

import (
	"context"
	"fmt"
)

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		return err("missing server argument")
}

	return success(fmt.Sprintf("connected to %s", server))
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("missing query argument")
}

	return ok(fmt.Sprintf("processed query: %s", q))
}