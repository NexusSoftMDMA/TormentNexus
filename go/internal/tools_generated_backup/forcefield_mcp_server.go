package tools

import (
	"context"
)

func AnalyzeCompliance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	if len(text) > 100 {
		return success("Non-compliant: text too long")
}

	return ok("Compliant")
}

func GetRiskLevel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entity, _ :=getString(args, "entity")
	if entity == "" {
		return err("entity is required")
}

	risk := "low"
	if entity == "high_risk" {
		risk = "high"
	}
	return ok("Risk level: " + risk)
}