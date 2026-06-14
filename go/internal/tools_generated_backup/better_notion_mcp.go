package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey, found := ctx.Value("notion_api_key").(string)
	if !found {
		return err("missing notion_api_key in context")
}

	body, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/search", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("json decode: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal result: " + e.Error())
}

	return ok(string(data))
}