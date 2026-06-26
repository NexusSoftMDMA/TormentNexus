package mcpimpl

import "context"

func HandleIceberg(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Hello, " + name + "!")
}

func HandleIcebergTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Tables: [table1, table2, table3]")
}