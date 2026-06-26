package mcpimpl

import (
	"context"
)

func HandleListDatabases_iotdb_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Listed IoTDB databases: root.db1, root.db2")
}

func HandleExecuteQuery_iotdb_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Executed query: " + query)
}