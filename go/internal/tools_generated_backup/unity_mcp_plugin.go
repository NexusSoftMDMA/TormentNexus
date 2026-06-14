package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListScenes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8888/api/scenes")
	if e != nil {
		return err("failed to list scenes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}

func HandleListGameObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	scene, _ :=getString(args, "scene")
	url := "http://localhost:8888/api/gameobjects"
	if scene != "" {
		url += "?scene=" + scene
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list game objects: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}