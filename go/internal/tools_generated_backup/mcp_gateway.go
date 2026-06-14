package tools

import (
	"context"
	"encoding/json"
)

func HandleGatewayInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "McpGateway"
	}
	info := map[string]string{"gateway": name, "version": "1.0.0"}
	data, e := json.Marshal(info)
	if e != nil {
		return err("marshal error")
}

	return ok(string(data))
}

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tools := []string{"gateway_info", "list_tools"}
	data, e := json.Marshal(tools)
	if e != nil {
		return err("marshal error")
}

	return ok(string(data))
}