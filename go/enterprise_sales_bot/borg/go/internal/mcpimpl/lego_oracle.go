package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleGetLegoSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	setNumber, _ :=getString(args, "set_number")
	if setNumber == "" {
		return err("set_number is required")
}

	data := map[string]interface{}{
		"set_number": setNumber,
		"name":       "Millennium Falcon",
		"pieces":     7541,
		"price":      "$799.99",
	}
	jsonBytes, e := json.Marshal(data)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	return success(string(jsonBytes))
}

func HandlePing_lego_oracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}