package tools

import "context"

func HandleGenerateDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagramType, _ :=getString(args, "type")
	name, _ :=getString(args, "name")
	return ok("Generated " + diagramType + " diagram: " + name)
}

func HandleGetDiagrams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	return ok("Retrieved diagrams with filter: " + filter)
}