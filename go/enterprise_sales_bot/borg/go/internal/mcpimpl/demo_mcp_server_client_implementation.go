package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX_demo_mcp_server_client_implementation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response")
}

	return success("data received")
}

func HandleY_demo_mcp_server_client_implementation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id <= 0 {
		return err("invalid id")
}

	return success("valid id")
}