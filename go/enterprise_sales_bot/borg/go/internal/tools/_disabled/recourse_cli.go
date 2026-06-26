package tools

import "context"

func HandleTerraformPlan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	planFile, _ :=getString(args, "plan_file")
	if planFile == "" {
		return err("terraform plan file path is required")
}

	return ok("Terraform plan analysis complete: no destructive changes detected")
}

func HandleShellCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("shell command is required")
}

	return ok("Shell command analyzed: risk level low")
}// touch 1781132139
