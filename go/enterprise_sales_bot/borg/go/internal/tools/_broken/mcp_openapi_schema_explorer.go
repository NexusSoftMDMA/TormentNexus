package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleX_mcp_openapi_schema_explorer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url parameter")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("fetch failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var schema map[string]interface{}
	e = json.Unmarshal(body, &schema)
	if e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	info, found := schema["info"].(map[string]interface{})
	title := ""
	version := ""
	if found {
		title, _ = info["title"].(string)
		version, _ = info["version"].(string)

	pathsCount := 0
	if paths, found := schema["paths"].(map[string]interface{}); found {
		pathsCount = len(paths)

	return success(fmt.Sprintf("title: %s, version: %s, paths: %d", title, version, pathsCount))
}
}
}