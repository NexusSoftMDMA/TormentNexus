package mcpimpl

import (
	"context"
	"fmt"
)

func HandleCreateDiagram_drawio_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success(fmt.Sprintf("Diagram '%s' created", name))
}

func HandleExportDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagramID, _ :=getString(args, "diagramId")
	format, _ :=getString(args, "format")
	if diagramID == "" {
		return err("diagramId is required")
}

	if format == "" {
		format = "png"
	}
	return ok(fmt.Sprintf("Exporting diagram %s as %s", diagramID, format))
}