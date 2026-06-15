package tools

import (
	"context"
)

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Listed IoTDB databases: root.db1, root.db2")
}

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Executed query: " + query)
}