package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

func HandleNxListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("npx", "nx", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err("failed to list projects: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(out.Bytes(), &result); e != nil {
		return err("failed to parse output: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d projects", len(result)))
}

func HandleNxRunTarget(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	target, _ :=getString(args, "target")
	if project == "" || target == "" {
		return err("project and target are required")
}

	cmd := exec.Command("npx", "nx", "run", project+":"+target)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("failed to run target: %s output: %s", e.Error(), string(output)))
}

	return success(fmt.Sprintf("Target output: %s", string(output)))
}