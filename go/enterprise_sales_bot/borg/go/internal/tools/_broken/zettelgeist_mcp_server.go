package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateZettel_zettelgeist_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "api_url")
	if baseURL == "" {
		baseURL = "https://api.zettelgeist.com"
	}
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" {
		return err("title is required")
}

	body, _ := json.Marshal(map[string]string{"title": title, "content": content})
	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/zettel", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Body = io.NopCloser(nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to send request: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("zettel created")
}

func HandleGetZettel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "api_url")
	if baseURL == "" {
		baseURL = "https://api.zettelgeist.com"
	}
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/zettel/"+id, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to send request: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return success(fmt.Sprintf("got zettel: %v", result))
}