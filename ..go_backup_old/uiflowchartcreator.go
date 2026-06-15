package tools

import (
	"context"
	"net/http"
)

func HandleCreateFlowchart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	description, _ :=getString(args, "description")
	if name == "" {
		return err("name is required")
}

	// Simulate creation
	_ = description
	_ = http.DefaultClient
	return success("flowchart created: " + name)
}

func HandleAddNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flowchartID, _ :=getString(args, "flowchartId")
	label, _ :=getString(args, "label")
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	if flowchartID == "" {
		return err("flowchartId is required")
}

	if label == "" {
		return err("label is required")
}

	_ = x
	_ = y
	return ok("node added to flowchart " + flowchartID)
}