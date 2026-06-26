package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

// HandleRun executes semgrep with a given pattern against a target path.
// Expected arguments:
//   - "pattern": the semgrep pattern to search for (string, required)
//   - "target": the file or directory to scan (string, required)
func HandleRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	target, _ :=getString(args, "target")

	if pattern == "" || target == "" {
		return err("both 'pattern' and 'target' arguments must be provided")
}

	cmd := exec.CommandContext(ctx, "semgrep", "--json", "-e", pattern, target)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("semgrep execution failed: %s", execErr.Error()))
}

	return ok(string(output))
}

// HandleListRules returns the list of available semgrep rules.
// No arguments are required.
func HandleListRules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "semgrep", "--list-rules")
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("semgrep list-rules failed: %s", execErr.Error()))
}

	return ok(string(output))
}

// HandleFetchRule downloads a semgrep rule file from a remote URL.
// Expected arguments:
//   - "url": the URL to fetch the rule from (string, required)
func HandleFetchRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ruleURL, _ :=getString(args, "url")
	if ruleURL == "" {
		return err("argument 'url' must be provided")
}

	client := http.DefaultClient
	resp, fetchErr := client.Get(ruleURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch rule: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read rule content: %s", readErr.Error()))
}

	return ok(string(body))
}

// HandleHealth provides a simple health check for the semgrepstream tool.
func HandleHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("semgrepstream healthy")
}