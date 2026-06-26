package tools

import (
	"context"
	"os/exec"
)

func HandleListResourceGroups(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "az", "group", "list", "--output", "json")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list resource groups")
}

	return ok(string(out))
}

func HandleListVirtualMachines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "az", "vm", "list", "--output", "json")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list virtual machines")
}

	return ok(string(out))
}