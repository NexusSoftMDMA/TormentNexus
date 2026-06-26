package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleExecuteCommand_unreal_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	form := url.Values{}
	form.Set("command", cmd)
	resp, e := http.DefaultClient.PostForm("http://localhost:30010/api/v1/execute", form)
	if e != nil {
		return err("execute failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	output, found := result["output"].(string)
	if !found {
		output = "ok"
	}
	return ok(output)
}

func HandleGetLevels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:30010/api/v1/levels")
	if e != nil {
		return err("get levels failed: " + e.Error())
}

	defer resp.Body.Close()
	var levels []string
	if e := json.NewDecoder(resp.Body).Decode(&levels); e != nil {
		return err("decode levels failed: " + e.Error())
}

	return success(levels)
}