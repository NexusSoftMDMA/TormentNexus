package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListNodes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "/root"
	}
	url := "http://localhost:8090/api/nodes?path=" + path
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get nodes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleListScenes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "project_path")
	if path == "" {
		path = "."
	}
	url := "http://localhost:8090/api/scenes?path=" + path
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list scenes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}