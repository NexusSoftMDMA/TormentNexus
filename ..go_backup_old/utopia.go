package tools

import (
	"context"
)

func HandleGetFeatures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	features := []map[string]interface{}{
		{"id": "1", "name": "Universal Basic Income"},
		{"id": "2", "name": "Automated Governance"},
	}
	return ok(map[string]interface{}{"features": features})
}

func HandleGetFeature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
	}
	if id == "1" {
		return ok(map[string]interface{}{"id": "1", "name": "Universal Basic Income", "description": "A fully automated UBI system."})
}

	if id == "2" {
		return ok(map[string]interface{}{"id": "2", "name": "Automated Governance", "description": "AI-driven decision making."})
}

	return err("feature not found")
}