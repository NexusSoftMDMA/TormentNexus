package mcpimpl

import (
	"context"
	"fmt"
)

func HandleExecuteTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	testName, _ :=getString(args, "testName")
	steps, _ :=getInt(args, "steps")
	if steps < 1 {
		steps = 1
	}
	return ok(fmt.Sprintf("Test '%s' executed with %d steps", testName, steps))
}

func HandleGetTestStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	testID, _ :=getString(args, "testId")
	status := "completed"
	if testID == "" {
		status = "pending"
	}
	return ok(fmt.Sprintf("Test '%s' status: %s", testID, status))
}