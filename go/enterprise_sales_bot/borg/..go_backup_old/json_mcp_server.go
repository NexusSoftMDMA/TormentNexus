package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandlePrettify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "json")
	if input == "" {
		return err("missing 'json' argument")
}

	var v interface{}
	e := json.Unmarshal([]byte(input), &v)
	if e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	pretty, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	return ok(string(pretty))
}

func HandleValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "json")
	if input == "" {
		return err("missing 'json' argument")
}

	var v interface{}
	e := json.Unmarshal([]byte(input), &v)
	if e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	return ok("valid JSON")
}