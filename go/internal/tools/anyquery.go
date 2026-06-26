package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}

	cmd := exec.CommandContext(ctx, "anyquery", "-q", query)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to execute query: %v\nOutput: %s", execErr, string(output)))
	}

	return ok(string(output))
}