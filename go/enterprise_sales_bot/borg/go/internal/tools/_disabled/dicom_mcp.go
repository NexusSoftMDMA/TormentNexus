package tools

import (
	"context"
	"fmt"
)

func HandleSearchPatients(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok(fmt.Sprintf("Found patients matching '%s': Patient A, Patient B", query))
}

func HandleGetStudy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	studyUID, _ :=getString(args, "studyUID")
	return ok(fmt.Sprintf("Study %s: Modality=CT, Date=2025-01-01", studyUID))
}