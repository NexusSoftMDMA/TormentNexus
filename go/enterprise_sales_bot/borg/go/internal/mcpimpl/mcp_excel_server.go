package mcpimpl

import (
	"context"
)

func HandleReadExcel_mcp_excel_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return success("Read Excel file: " + path)
}

func HandleWriteExcel_mcp_excel_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	data, _ :=getString(args, "data")
	return success("Write Excel file: " + path + " with data length " + string(len(data)))
}