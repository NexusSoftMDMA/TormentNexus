package mcpimpl

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleAnalyzeUrbanData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataPath, _ :=getString(args, "data_path")
	if dataPath == "" {
		return err("data_path is required")
}

	cmd := exec.CommandContext(ctx, "python3", "-m", "pymdu", "analyze", dataPath)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("execution failed: %s", e.Error()))
}

	return ok(string(output))
}

func HandleProcessUrbanData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataPath, _ :=getString(args, "data_path")
	param, _ :=getString(args, "param")
	argsList := []string{"-m", "pymdu", "process", dataPath}
	if param != "" {
		argsList = append(argsList, "--param", param)

	cmd := exec.CommandContext(ctx, "python3", argsList...)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("execution failed: %s", e.Error()))
}

	return success(string(output))
}
}