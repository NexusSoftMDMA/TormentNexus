package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	body := map[string]string{"content": content}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.memoryplugin.com/memories", "application/json", bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("server returned " + resp.Status)
}

	return ok("memory created")
}

func HandleSearchMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.memoryplugin.com/memories?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("search failed")
}

	return success("search completed")
}