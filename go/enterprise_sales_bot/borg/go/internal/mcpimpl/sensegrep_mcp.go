package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch_sensegrep_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 {
		maxResults = 10
	}
	url := fmt.Sprintf("https://api.sensegrep.dev/search?q=%s&max=%d", query, maxResults)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleGetContext_sensegrep_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	if filePath == "" {
		return err("filePath is required")
}

	line, _ :=getInt(args, "line")
	if line <= 0 {
		line = 1
	}
	url := fmt.Sprintf("https://api.sensegrep.dev/context?file=%s&line=%d", filePath, line)
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