package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX_mcp_durable_object_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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

func HandleY_mcp_durable_object_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getInt(args, "value")
	if value > 0 {
		return success("value is positive")
}

	return ok("key: " + key)
}