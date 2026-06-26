package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleCheckCompatibility(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientVersion, _ :=getString(args, "client_version")
	if clientVersion == "" {
		return err("missing client_version")
}

	return success("compatible with " + clientVersion)
}

func HandleValidateProtocol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	protocol, _ :=getString(args, "protocol")
	found := false
	if protocol == "MCP" {
		found = true
	}
	if !found {
		return err("invalid protocol")
}

	data, e := json.Marshal(map[string]bool{"valid": true})
	if e != nil {
		return err("marshal failed")
}

	return ok(string(data))
}