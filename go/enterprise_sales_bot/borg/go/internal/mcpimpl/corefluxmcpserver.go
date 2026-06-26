package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetStatus_corefluxmcpserver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.coreflux.io/status")
	if e != nil {
		return err("failed to fetch status")
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return ok("status fetched")
}

func HandlePing_corefluxmcpserver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}