package tools

import (
	"context"
	"os/exec"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	if project == "" {
		return err("project path required")
}

	cmd := exec.CommandContext(ctx, "xcodebuild", "-list", "-project", project)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("failed to list schemes: " + string(output) + ": " + e.Error())
}

	return ok("Schemes:\n" + string(output))
}