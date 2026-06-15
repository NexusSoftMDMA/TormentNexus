//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	port := os.Getenv("UNITY_MCP_PORT")
	if port == "" {
		port = "8080"
	}
	url := "http://127.0.0.1:" + port + "/execute"
	body := `{"command":"` + cmd + `"}`
	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(body))
	if e != nil {
		return err("failed to send command: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

func HandleGetGameObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port := os.Getenv("UNITY_MCP_PORT")
	if port == "" {
		port = "8080"
	}
	resp, e := http.DefaultClient.Get("http://127.0.0.1:" + port + "/gameobjects")
	if e != nil {
		return err("failed to get game objects: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}
