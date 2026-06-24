package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetProjectInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "default"
	}
	info := map[string]string{"project": path, "engine": "Unity"}
	data, e := json.Marshal(info)
	if e != nil {
		return err("marshal failed")
}

	return success(string(data))
}

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/execute", nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("execute: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode: " + e.Error())
}

	return ok("command executed")
}