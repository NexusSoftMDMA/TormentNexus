//go:build ignore
// +build ignore

package tools

/**
 * @file supermemory.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of SuperMemory MCP tools.
 * Replaces `mcp-supermemory-ai` (npx mcp-remote@latest https://mcp.supermemory.ai/mcp) entry in mcp.json.
 *
 * SuperMemory provides cloud-based semantic memory storage and retrieval.
 *
 * Improvements over original:
 * - No npx/mcp-remote dependency.
 * - Direct REST API integration.
 * - Supports: add_memory, search_memory, delete_memory, list_memories.
 * - Context-aware with timeout; uses SUPERMEMORY_API_KEY for auth.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const supermemoryBaseURL = "https://api.supermemory.ai/v1"

func supermemoryAPIKey() string {
	if k := os.Getenv("SUPERMEMORY_API_KEY"); k != "" {
		return k
	}
	return os.Getenv("SUPERMEMORY_KEY")
}

func supermemoryDo(ctx context.Context, method, urlPath string, payload interface{}) (interface{}, error) {
	apiKey := supermemoryAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("SUPERMEMORY_API_KEY environment variable is not set")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, supermemoryBaseURL+urlPath, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("SuperMemory API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// HandleSuperMemoryAdd adds a memory to SuperMemory.
// Tool: supermemory_add
func HandleSuperMemoryAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ := getString(args, "content", "text", "message")
	if content == "" {
		return err("content parameter is required")
	}

	payload := map[string]interface{}{
		"content": content,
	}

	if title, _ := getString(args, "title"); title != "" {
		payload["title"] = title
	}
	if tags, ok := args["tags"].([]interface{}); ok {
		payload["tags"] = tags
	}
	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		payload["metadata"] = metadata
	}

	result, e := supermemoryDo(ctx, "POST", "/memories", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleSuperMemorySearch searches memories semantically.
// Tool: supermemory_search
func HandleSuperMemorySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	payload := map[string]interface{}{
		"query": query,
	}

	if limit := getInt(args, "limit", "count"); limit > 0 {
		payload["limit"] = limit
	} else {
		payload["limit"] = 10
	}

	if tags, ok := args["tags"].([]interface{}); ok {
		payload["tags"] = tags
	}

	result, e := supermemoryDo(ctx, "POST", "/memories/search", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleSuperMemoryDelete deletes a memory by ID.
// Tool: supermemory_delete
func HandleSuperMemoryDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memoryID, _ := getString(args, "memory_id", "id")
	if memoryID == "" {
		return err("memory_id parameter is required")
	}

	_, e := supermemoryDo(ctx, "DELETE", "/memories/"+memoryID, nil)
	if e != nil {
		return err(e.Error())
	}

	return ok("Memory deleted: " + memoryID)
}

// HandleSuperMemoryList lists all stored memories.
// Tool: supermemory_list
func HandleSuperMemoryList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 50
	}

	path := fmt.Sprintf("/memories?limit=%d", limit)
	result, e := supermemoryDo(ctx, "GET", path, nil)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
