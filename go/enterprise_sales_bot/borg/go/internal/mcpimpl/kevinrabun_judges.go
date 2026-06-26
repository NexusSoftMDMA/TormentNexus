package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListJudges(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	return ok("SecurityJudge, CostJudge, QualityJudge")
}

func HandleEvaluateCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	judge, _ :=getString(args, "judge")
	if code == "" {
		return err("code is required")
}

	result := fmt.Sprintf("Judge '%s' scored code (len %d): 7/10", judge, len(code))
	return ok(result)
}// touch 1781132129
