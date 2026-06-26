package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

var droidMode = "normal"

func HandleSetDroidMode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mode, _ :=getString(args, "mode")
	if mode == "" {
		return err("mode is required")
}

	droidMode = mode
	return success("Droid mode set to " + mode)
}

func HandleDroidPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.factory.ai/droid/ping")
	if e != nil {
		return err("ping failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return success("Ping result: " + string(body))
}