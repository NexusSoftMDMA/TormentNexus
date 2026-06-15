package tools

import (
	"context"
	"encoding/json"
)

func HandleGenerateFormSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	description, _ :=getString(args, "description")
	schema := map[string]interface{}{
		"type":       "object",
		"title":      "Generated Form",
		"properties": map[string]interface{}{},
	}
	if len(description) > 0 {
		schema["properties"] = map[string]interface{}{
			"name":  map[string]interface{}{"type": "string", "title": "Name"},
			"email": map[string]interface{}{"type": "string", "title": "Email"},
		}
	}
	b, e := json.Marshal(schema)
	if e != nil {
		return err("failed to marshal schema: " + e.Error())
}

	return success(string(b))
}

func HandleValidateFormSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	schemaStr, _ :=getString(args, "schema")
	if schemaStr == "" {
		return err("schema parameter is required")
}

	var js map[string]interface{}
	if e := json.Unmarshal([]byte(schemaStr), &js); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("schema is valid JSON")
}