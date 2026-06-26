package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleCreateWorldBook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	wb := map[string]interface{}{
		"name":        name,
		"description": desc,
		"scan_until":  "before",
		"keys":        []string{},
		"content":     "New entry",
		"extensions":  map[string]interface{}{},
	}
	b, e := json.MarshalIndent(wb, "", "  ")
	if e != nil {
		return err("failed to marshal world book: " + e.Error())
}

	return ok(fmt.Sprintf("```json\n%s\n```", string(b)))
}

func HandleValidateWorldBook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	raw, _ :=getString(args, "json")
	if raw == "" {
		return err("missing 'json' argument")
}

	var wb map[string]interface{}
	if e := json.Unmarshal([]byte(raw), &wb); e != nil {
		return err("invalid JSON: " + e.Error())
}

	if _, found := wb["name"]; !found {
		return err("missing required field 'name'")
}

	if _, found := wb["content"]; !found {
		return err("missing required field 'content'")
}

	return success("valid world book JSON")
}