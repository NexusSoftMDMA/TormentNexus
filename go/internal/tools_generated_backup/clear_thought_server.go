package tools

import (
	"context"
	"fmt"
)

func HandleThink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	model, _ :=getString(args, "model")
	if model == "" {
		model = "first principles"
	}
	return ok(fmt.Sprintf("Applying %s thinking to: %s", model, query))
}

func HandleDebug(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issue, _ :=getString(args, "issue")
	steps, _ :=getString(args, "steps")
	if steps == "" {
		steps = "1. Reproduce 2. Isolate 3. Analyze 4. Fix 5. Verify"
	}
	return ok(fmt.Sprintf("Debugging: %s\nSteps: %s", issue, steps))
}