package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// HandleEcho returns the provided message unchanged.
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok(message)
}

// HandleFetch retrieves the content of a URL and returns it as text.
func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rawURL, _ :=getString(args, "url")
	if strings.TrimSpace(rawURL) == "" {
		return err("url parameter is required")
}

	parsedURL, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid url: %v", parseErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch url: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	return ok(string(body))
}

// HandleRun executes a command with optional arguments and returns its stdout.
func HandleRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "cmd")
	if strings.TrimSpace(cmdStr) == "" {
		return err("cmd parameter is required")
}

	argStr, _ :=getString(args, "args")
	var cmdArgs []string
	if strings.TrimSpace(argStr) != "" {
		cmdArgs = strings.Fields(argStr)

	cmd := exec.CommandContext(ctx, cmdStr, cmdArgs...)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("command execution failed: %v, output: %s", runErr, string(output)))
}

	return ok(string(output))
}

}

// HandleListFiles lists files in a directory, optionally including hidden files.
func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	if strings.TrimSpace(dirPath) == "" {
		return err("path parameter is required")
}

	includeHidden, _ :=getBool(args, "include_hidden")

	absPath, absErr := filepath.Abs(dirPath)
	if absErr != nil {
		return err(fmt.Sprintf("failed to resolve absolute path: %v", absErr))
}

	entries, readErr := os.ReadDir(absPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read directory: %v", readErr))
}

	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if !includeHidden && strings.HasPrefix(name, ".") {
			continue
		}
		names = append(names, name)

	jsonBytes, jsonErr := json.Marshal(names)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", jsonErr))
}

	return ok(string(jsonBytes))
}
}