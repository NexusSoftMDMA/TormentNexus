package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX_browser_use_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	response, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to make request")
}

	defer response.Body.Close()

	var result map[string]interface{}
	e = json.NewDecoder(response.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return success("request successful")
}

func HandleY_browser_use_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "data")
	if data == "" {
		return err("data is required")
}

	return success("data received")
}