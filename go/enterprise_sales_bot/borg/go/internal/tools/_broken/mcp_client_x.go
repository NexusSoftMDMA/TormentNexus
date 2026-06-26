package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX_mcp_client_x(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	response, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to make request"), e
	}
	defer response.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(response.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response"), e
	}

	return success("data received")
}

func HandleY_mcp_client_x(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("HandleY executed")
}