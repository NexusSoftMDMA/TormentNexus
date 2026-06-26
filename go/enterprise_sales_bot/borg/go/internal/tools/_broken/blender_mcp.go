package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandlePing_blender_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleCreateCube(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	size, _ :=getInt(args, "size")
	payload := map[string]interface{}{"name": name, "size": size}
	body, e0 := json.Marshal(payload)
	if e0 != nil {
		return err("failed to marshal: " + e0.Error())
}

	resp, e1 := http.DefaultClient.Post("http://localhost:8080/blender/create_cube", "application/json", bytes.NewReader(body))
	if e1 != nil {
		return err("request failed: " + e1.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e2 := json.NewDecoder(resp.Body).Decode(&result); e2 != nil {
		return err("decode failed: " + e2.Error())
}

	return success("cube created", result)
}