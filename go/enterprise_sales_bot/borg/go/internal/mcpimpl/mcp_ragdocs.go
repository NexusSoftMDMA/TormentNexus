package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleQuery_mcp_ragdocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	topK, _ :=getInt(args, "top_k")
	if topK <= 0 {
		topK = 5
	}
	body, _ := json.Marshal(map[string]interface{}{"query": query, "top_k": topK})
	resp, e := http.DefaultClient.Post("http://localhost:8020/query", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("query failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(result)
}

func HandleIngest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	source, _ :=getString(args, "source")
	if source == "" {
		source = "default"
	}
	body, _ := json.Marshal(map[string]interface{}{"text": text, "source": source})
	resp, e := http.DefaultClient.Post("http://localhost:8020/ingest", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("ingest failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(result)
}