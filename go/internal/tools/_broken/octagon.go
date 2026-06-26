package tools

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// HandlePing returns a simple pong response.
func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

// HandleEcho returns the provided text argument.
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	return ok(text)
}

// HandleFetchURL retrieves the content of the given URL.
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	client := http.DefaultClient
	resp, fetchErr := client.Get(urlStr)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleListDir lists files in the specified directory.
func HandleListDir(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	entries, readErr := os.ReadDir(dirPath)
	if readErr != nil {
		return err(readErr.Error())
}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(strings.Join(names, "\n"))
}

}

// HandleRunCommand executes a shell command and returns its combined output.
func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdLine, _ :=getString(args, "cmd")
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return err("empty command")
}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(string(output))
}