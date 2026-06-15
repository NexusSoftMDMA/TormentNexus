package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetServerStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		return err("server is required")
}

	status := map[string]interface{}{"name": server, "status": "running"}
	data, e := json.Marshal(status)
	if e != nil {
		return err("failed to marshal status")
}

	return ok(string(data))
}