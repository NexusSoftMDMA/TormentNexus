package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetProjectInfo_unity_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getString(args, "port")
	if port == "" {
		port = "8080"
	}
	url := fmt.Sprintf("http://%s:%s/api/project", host, port)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http get: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json decode: %v", e))
	}
	raw, _ := json.Marshal(result)
	return success(string(raw))
}

func HandleListScenes_unity_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getString(args, "port")
	if port == "" {
		port = "8080"
	}
	url := fmt.Sprintf("http://%s:%s/api/scenes", host, port)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http get: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
	}
	var result []string
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json decode: " + e.Error())
	}
	raw, _ := json.Marshal(result)
	return success(string(raw))
}