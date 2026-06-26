package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToolResponse, ok, e, getString, getInt, getBool, TextContent は parity.go で定義されていると仮定

// HandleFetch handles web fetching requests
func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	method, _ :=getString(args, "method")
	headers := args["headers"].(map[string]interface{})
	body, _ :=getString(args, "body")

	req, e := http.NewRequest(method, urlStr, strings.NewReader(body))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	for key, value := range headers {
		req.Header.Set(key, fmt.Sprintf("%v", value))

	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return err("Failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response body: " + e.Error())
}

	return ok(fmt.Sprintf("Status: %d\nHeaders: %v\nBody: %s", resp.StatusCode, resp.Header, string(respBody)))
}

}

// HandleFilesystem handles file system operations
func HandleFilesystem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	operation, _ :=getString(args, "operation")
	content, _ :=getString(args, "content")

	switch operation {
	case "read":
		data, e := os.ReadFile(path)
		if e != nil {
			return err("Failed to read file: " + e.Error())
}

		return ok(string(data))
}

	case "write":
		dir := filepath.Dir(path)
		if e := os.MkdirAll(dir, 0755); e != nil {
			return err("Failed to create directory: " + e.Error())
}

		if e := os.WriteFile(path, []byte(content), 0644); e != nil {
			return err("Failed to write file: " + e.Error())
}

		return ok("File written successfully")
}

	case "list":
		files, e := os.ReadDir(path)
		if e != nil {
			return err("Failed to list directory: " + e.Error())
}

		var result strings.Builder
		for _, file := range files {
			result.WriteString(file.Name())
			if !file.IsDir() {
				result.WriteString(" (file)")
			} else {
				result.WriteString(" (dir)")

			result.WriteString("\n")

		return ok(result.String())
}

	default:
		return err("Unsupported operation: " + operation)

}

// HandleBash handles bash command execution
func HandleBash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	timeout, _ :=getInt(args, "timeout")
	if timeout == 0 {
		timeout = 30
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := cmd.Run()
	if e != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return err("Command timed out after " + fmt.Sprintf("%d", timeout) + " seconds")
}

		return err("Command failed: " + e.Error() + "\nStderr: " + stderr.String())
}

	result := "Exit Code: " + fmt.Sprintf("%d", cmd.ProcessState.ExitCode()) + "\n"
	if stdout.Len() > 0 {
		result += "Stdout:\n" + stdout.String() + "\n"
	}
	if stderr.Len() > 0 {
		result += "Stderr:\n" + stderr.String() + "\n"
	}

	return ok(result)
}