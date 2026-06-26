package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleRunDAG(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dagFile, _ :=getString(args, "dag_file")
	if dagFile == "" {
		return err("dag_file is required")
}

	cmd := exec.Command("dagu", "run", dagFile)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("failed to run DAG: %s", runErr.Error()))
}

	return ok(string(output))
}