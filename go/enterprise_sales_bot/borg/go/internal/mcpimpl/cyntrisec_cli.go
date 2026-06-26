package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleExecuteCommand_cyntrisec_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	url, _ :=getString(args, "url")
	if command == "" || url == "" {
		return err("command and url are required")
}

	payload, _ := json.Marshal(map[string]string{"command": command})
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleListCommands_cyntrisec_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var list []string
	if e := json.Unmarshal(body, &list); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("available commands: %v", list))
}