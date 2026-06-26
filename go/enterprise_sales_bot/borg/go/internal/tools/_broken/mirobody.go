package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

func ok(text string) (ToolResponse, error) {
	return ToolResponse{Text: text, Ok: true}, nil
}

func err(msg string) (ToolResponse, error) {
	return ToolResponse{Text: msg, Ok: false}, fmt.Errorf(msg)
}

func getString(args map[string]interface{}, key string) string {
	if val, found := args[key]; found {
		if str, found := val.(string); found {
			return str
		}
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if val, found := args[key]; found {
		if num, found := val.(float64); found {
			return int(num)
}

		if num, found := val.(int); found {
			return num
		}
	}
	return 0
}

func getBool(args map[string]interface{}, key string) bool {
	if val, found := args[key]; found {
		if b, found := val.(bool); found {
			return b
		}
	}
	return false
}

// shared HTTP client with timeout
var http.DefaultClient = http.DefaultClient

// HandleEcho returns the provided message unchanged.
// Expected args: {"message": "<text>"}
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("missing required argument: message")
}

	return ok(msg)
}

// HandleAdd adds two integers and returns the sum as text.
// Expected args: {"a": <int>, "b": <int>}
func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return ok(fmt.Sprintf("%d", sum))
}

// HandleFetchURL performs an HTTP GET request to the supplied URL
// and returns the first 1000 characters of the response body.
// Expected args: {"url": "<http://...>"}
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rawURL, _ :=getString(args, "url")
	if rawURL == "" {
		return err("missing required argument: url")
}

	parsedURL, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return err(parseErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	content := string(body)
	if len(content) > 1000 {
		content = content[:1000] + "...(truncated)"
	}
	return ok(content)
}

// HandleRunCommand executes a shell command and returns its combined stdout+stderr.
// Expected args: {"command": "<executable> [args...]"}
func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdLine, _ :=getString(args, "command")
	if cmdLine == "" {
		return err("missing required argument: command")
}

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return err("empty command")
}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		// include command output for debugging
		return err(fmt.Sprintf("%s: %s", execErr.Error(), string(output)))
}

	return ok(string(output))
}