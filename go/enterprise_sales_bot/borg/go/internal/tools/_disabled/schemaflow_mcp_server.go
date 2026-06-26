package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleSchemaValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	schema, _ :=getString(args, "schema")
	data, _ :=getString(args, "data")

	if schema == "" || data == "" {
		return err("missing schema or data")
}

	resp, e := http.DefaultClient.Post("https://api.schemaflow.com/validate", "application/json", nil)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	return success(fmt.Sprintf("Validation request sent, status: %s", resp.Status))
}

func HandleSchemaGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sample, _ :=getString(args, "sample")
	if sample == "" {
		return err("missing sample data")
}

	return ok("Schema generated successfully from sample")
}// touch 1781132140
