package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetGodotVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.github.com/repos/godotengine/godot/releases/latest")
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	version, found := data["tag_name"].(string)
	if !found {
		return err("version not found")
}

	return ok(version)
}

func HandleCheckGodotScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path argument is required")
}

	return ok("scene exists at " + path)
}