package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func HandleGetClientStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:64989/status")
	if e != nil {
		return err("failed to get status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleExecuteScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("script argument is required")
}

	resp, e := http.DefaultClient.Post("http://localhost:64989/execute", "text/plain", strings.NewReader(script))
	if e != nil {
		return err("failed to execute script: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}