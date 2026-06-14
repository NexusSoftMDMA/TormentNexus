package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleFetchOpenAPISpec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var spec map[string]interface{}
	if e := json.Unmarshal(body, &spec); e != nil {
		return err("parse failed: " + e.Error())
}

	paths, found := spec["paths"]
	if !found {
		return err("no paths in spec")
}

	return success(fmt.Sprintf("Paths: %+v", paths))
}

func HandleServerInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("OpenMCP Server - converts OpenAPI specs to MCP tools")
}