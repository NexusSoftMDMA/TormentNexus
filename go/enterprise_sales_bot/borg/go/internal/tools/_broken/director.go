package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// runDirector executes the director CLI command with the given arguments.
func runDirector(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "director", args...)
	outputBytes, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return "", fmt.Errorf("failed to execute director command: %w, output: %s", execErr, string(outputBytes))
}

	return string(outputBytes), nil
}

// HandleListPlaybooks lists all available playbooks.
func HandleListPlaybooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	output, runErr := runDirector(ctx, "ls")
	if runErr != nil {
		return err(runErr.Error())
}

	return ok(output)
}

// HandleGetPlaybook retrieves details for a specific playbook.
func HandleGetPlaybook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	playbookID, _ :=getString(args, "playbookId")
	if playbookID == "" {
		return err("playbookId is required")
}

	output, runErr := runDirector(ctx, "get", playbookID)
	if runErr != nil {
		return err(runErr.Error())
}

	return ok(output)
}

// HandleListTools lists the tools available within a specific playbook.
func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	playbookID, _ :=getString(args, "playbookId")
	if playbookID == "" {
		return err("playbookId is required")
}

	output, runErr := runDirector(ctx, "mcp", "list-tools", playbookID)
	if runErr != nil {
		return err(runErr.Error())
}

	return ok(output)
}

// HandleCallTool executes a specific tool within a playbook.
func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	playbookID, _ :=getString(args, "playbookId")
	toolName, _ :=getString(args, "toolName")

	if playbookID == "" {
		return err("playbookId is required")
}

	if toolName == "" {
		return err("toolName is required")
}

	cmdArgs := []string{"mcp", "call-tool", playbookID, toolName}

	// Check for tool arguments
	if rawArgs, exists := args["arguments"]; exists {
		jsonBytes, jsonErr := json.Marshal(rawArgs)
		if jsonErr != nil {
			return err("failed to marshal tool arguments: " + jsonErr.Error())
}

		cmdArgs = append(cmdArgs, "--args", string(jsonBytes))

	output, runErr := runDirector(ctx, cmdArgs...)
	if runErr != nil {
		return err(runErr.Error())
}

	return ok(output)
}
}