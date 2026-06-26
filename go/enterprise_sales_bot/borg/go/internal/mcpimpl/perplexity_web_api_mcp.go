package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleSearch_perplexity_web_api_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	body, _ := json.Marshal(map[string]string{"query": query})
	resp, e := http.DefaultClient.Post("https://api.perplexity.ai/search", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(result)
}

func HandleResearch_perplexity_web_api_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	body, _ := json.Marshal(map[string]string{"query": query})
	resp, e := http.DefaultClient.Post("https://api.perplexity.ai/research", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(result)
}