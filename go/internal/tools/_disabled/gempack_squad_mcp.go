package tools

import (
    "context"
    "fmt"
)

func HandleClassify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    text, _ :=getString(args, "text")
    if text == "" {
        return err("text is required")
}

    result := fmt.Sprintf("Classified: %s as important", text)
    return ok(result)
}

func HandleScoreRisk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    project, _ :=getString(args, "project")
    budget, _ :=getInt(args, "budget")
    riskScore := 0.5
    if budget > 100000 {
        riskScore = 0.8
    }
    result := fmt.Sprintf("Risk score for %s: %.2f", project, riskScore)
    return ok(result)
}