package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects_ghidra_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	if baseURL == "" {
		baseURL = "http://localhost:1313"
	}
	resp, e := http.DefaultClient.Get(baseURL + "/projects")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var projects []string
	e = json.Unmarshal(body, &projects)
	if e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Projects: %v", projects))
}

func HandleAnalyzeBinary_ghidra_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	if baseURL == "" {
		baseURL = "http://localhost:1313"
	}
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path argument is required")
}

	resp, e := http.DefaultClient.Post(baseURL+"/analyze", "application/json", nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(fmt.Sprintf("Analysis result: %s", string(body)))
}