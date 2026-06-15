package tools

import (
	"context"
)

func HandleGetDecision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nodeID, _ :=getString(args, "node_id")
	if nodeID == "" {
		nodeID = "root"
	}
	decision := "approve"
	if nodeID == "root" {
		decision = "review"
	}
	return success("Decision for node " + nodeID + ": " + decision)
}