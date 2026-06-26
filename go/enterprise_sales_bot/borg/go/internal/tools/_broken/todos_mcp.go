package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListTodos_todos_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url+"/todos", nil)
	if e != nil {
		return err("failed to create request"), e
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed"), e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response"), e
	}
	return ok(string(body))
}

func HandleCreateTodo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080"
	}
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
	}
	payload := fmt.Sprintf(`{"title":"%s"}`, title)
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/todos", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request"), e
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed"), e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response"), e
	}
	return success(string(body))
}