package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleSemanticSearch_obsidian_mcp_tools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("http://localhost:27123/search?q=" + query)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok("Search completed. Results: " + string(body))
}

func HandleTemplaterPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	template, _ :=getString(args, "template")
	payload := map[string]string{"template": template}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal payload failed: " + e.Error())
}

	resp, e := http.DefaultClient.Post("http://localhost:27123/templater", "application/json", io.NopCloser(nil))
	if e != nil {
		return err("execute templater failed: " + e.Error())
}

	defer resp.Body.Close()
	output, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return ok("Templater executed. Output: " + string(output))
}