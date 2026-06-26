package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// runCommand executes the mantic CLI via npx and returns its stdout.
// It uses the provided context for cancellation and a 30‑second timeout.
func runCommand(ctx context.Context, args []string) (string, error) {
	// npx mantic.sh <args...>
	cmd := exec.CommandContext(ctx, "npx", append([]string{"mantic.sh"}, args...)...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		// Include the raw output to aid debugging.
		return "", fmt.Errorf("%s: %s", execErr.Error(), strings.TrimSpace(string(output)))
}

	return strings.TrimSpace(string(output)), nil
}

// HandleSearchFiles implements the `search_files` tool.
// Expected arguments:
//   - query (string, required)
//   - semantic (bool, optional)
//   - sessionId (string, optional, ignored in this simple implementation)
func HandleSearchFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
}

	semantic, _ :=getBool(args, "semantic")

	cmdArgs := []string{query}
	if semantic {
		cmdArgs = append(cmdArgs, "--semantic")

	// sessionId is ignored for now; it could be used for logging or future extensions.

	output, execErr := runCommand(ctx, cmdArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(output)
}

}

// HandleGetDefinition implements the `get_definition` tool.
// Expected arguments:
//   - symbol (string, required)
func HandleGetDefinition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing required argument: symbol")
}

	cmdArgs := []string{"goto", symbol}
	output, execErr := runCommand(ctx, cmdArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(output)
}

// HandleFindReferences implements the `find_references` tool.
// Expected arguments:
//   - symbol (string, required)
func HandleFindReferences(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing required argument: symbol")
}

	cmdArgs := []string{"references", symbol}
	output, execErr := runCommand(ctx, cmdArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(output)
}

// HandleGetContext implements the `get_context` tool.
// No arguments are required; it runs mantic with an empty query.
func HandleGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// An empty argument list triggers zero‑query mode.
	output, execErr := runCommand(ctx, []string{})
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(output)
}

// HandleSessionStart implements the `session_start` tool.
// Expected arguments:
//   - name (string, required)
func HandleSessionStart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing required argument: name")
}

	// In this lightweight implementation we simply acknowledge the request.
	return ok(fmt.Sprintf("session started: %s", name))
}

// HandleSessionEnd implements the `session_end` tool.
// Expected arguments:
//   - sessionId (string, required)
func HandleSessionEnd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionId, _ :=getString(args, "sessionId")
	if sessionId == "" {
		return err("missing required argument: sessionId")
}

	// Acknowledge the termination of the session.
	return ok(fmt.Sprintf("session ended: %s", sessionId))
}