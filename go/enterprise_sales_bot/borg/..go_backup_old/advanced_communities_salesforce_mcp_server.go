package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	cmd := exec.CommandContext(ctx, "sfdx", "force:data:soql:query", "-q", query, "--json")
	output, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("sfdx command failed: %v", e))
}

	return ok(string(output))
}

func HandleDescribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	object, _ :=getString(args, "object")
	if object == "" {
		return err("object parameter is required")
}

	cmd := exec.CommandContext(ctx, "sfdx", "force:schema:sobject:describe", "-s", object, "--json")
	output, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("sfdx command failed: %v", e))
}

	return ok(string(output))
}