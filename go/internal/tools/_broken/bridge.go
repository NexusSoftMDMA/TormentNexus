package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

// HandleEcho returns the provided message unchanged.
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ := getString(args, "message")
	if message == "" {
		return err("message argument is required")
	}
	return ok(message)
}

// HandleFetchURL performs a GET request to the supplied URL and returns the response body.
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rawURL, _ := getString(args, "url")
	if rawURL == "" {
		return err("url argument is required")
	}

	parsed, parseErr := url.ParseRequestURI(rawURL)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid url: %v", parseErr))
	}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
	}

	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("http request failed: %v", respErr))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
	}
	return ok(string(body))
}

// HandleRunCommand executes a command with optional arguments and returns its stdout.
func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ := getString(args, "command")
	if cmdStr == "" {
		return err("command argument is required")
	}

	var cmdArgs []string
	if rawArgs, found := args["args"]; found {
		if slice, isSlice := rawArgs.([]interface{}); isSlice {
			for _, a := range slice {
				if s, isStr := a.(string); isStr {
					cmdArgs = append(cmdArgs, s)
				}
			}
		}
	}

	cmd := exec.CommandContext(ctx, cmdStr, cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("command execution failed: %v, output: %s", execErr, string(output)))
	}
	return ok(string(output))
}

// HandleEnv returns the value of an environment variable.
func HandleEnv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	varName, _ := getString(args, "name")
	if varName == "" {
		return err("name argument is required")
	}

	val, exists := os.LookupEnv(varName)
	if !exists {
		return err(fmt.Sprintf("environment variable %s not set", varName))
	}
	return ok(val)
}

// HandleTimeNow returns the current server time in RFC3339 format.
func HandleTimeNow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	return ok(now)
}