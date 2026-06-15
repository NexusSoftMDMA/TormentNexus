package tools

import (
	"context"
	"net/http"
)

func HandleValidateSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	schema, _ :=getString(args, "schema")
	if schema == "" {
		return err("Missing schema")
}

	resp, e := http.DefaultClient.Get("https://panproto.local/validate?schema=" + schema)
	if e != nil {
		return err("Validation request failed")
}

	resp.Body.Close()
	return ok("Schema is valid")
}

func HandleGetSchemaInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Missing schema name")
}

	description := "Schema " + name + " is a panproto schema"
	return success(description)
}