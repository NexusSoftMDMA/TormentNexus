package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetDependencyGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "projectPath")
	if projectPath == "" {
		return err("projectPath is required")
	}
	depth, _ :=getInt(args, "depth")
	url := fmt.Sprintf("http://localhost:8080/graph?project=%s&depth=%d", projectPath, depth)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	return ok(string(body))
}