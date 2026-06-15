package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleGetEntities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "projectPath")
	if projectPath == "" {
		return err("projectPath is required")
}

	entities := []string{"Books", "Authors", "Orders"}
	data, e := json.Marshal(entities)
	if e != nil {
		return err("failed to marshal entities")
}

	return ok(fmt.Sprintf("Entities: %s", string(data)))
}

func HandleGetEntityFields(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityName, _ :=getString(args, "entityName")
	if entityName == "" {
		return err("entityName is required")
}

	fields := map[string]string{"ID": "Integer", "Title": "String", "AuthorID": "Association"}
	data, e := json.Marshal(fields)
	if e != nil {
		return err("failed to marshal fields")
}

	return ok(fmt.Sprintf("Fields for %s: %s", entityName, string(data)))
}// touch 1781132121
