package tools

import (
	"context"
	"encoding/json"
)

func ListBridges(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bridgeNames := []string{"eth-bridge", "sol-bridge", "btc-bridge"}
	data, e := json.Marshal(bridgeNames)
	if e != nil {
		return err("failed to marshal bridge list")
}

	return success(string(data))
}

func GetBridge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	bridge := map[string]interface{}{
		"name":   name,
		"status": "active",
		"token":  "HASH",
	}
	data, e := json.Marshal(bridge)
	if e != nil {
		return err("failed to marshal bridge")
}

	return success(string(data))
}