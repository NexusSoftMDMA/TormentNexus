//go:build ignore
// +build ignore

package tools

/**
 * @file mem0.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of mem0 memory system.
 * Replaces `@mem0/mcp-server@latest` STDIO entry in mcp.json.
 *
 * Uses the mem0 REST API (https://api.mem0.ai).
 * Improvements over original:
 *  - No npx/Node dependency.
 *  - Full CRUD on memories: add, search, get, update, delete.
 *  - Context-aware with timeout; works with MEM0_API_KEY.
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

const mem0BaseURL = "https://api.mem0.ai/v1"

func mem0APIKey() string {
	return os.Getenv("MEM0_API_KEY")
}

func mem0Do(ctx context.Context, method, path string, payload interface{}) (interface{}, error) {
	apiKey := mem0APIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("MEM0_API_KEY environment variable is not set")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, mem0BaseURL+path, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Token "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("mem0 API error (HTTP %d): %s", resp.StatusCode, string(body))
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

// HandleMem0AddMemory adds a memory to mem0.
// Tool: mem0_add_memory
func HandleMem0AddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ := getString(args, "content", "text", "message")
	if content == "" {
		return err("content parameter is required")
	}

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": content},
		},
	}

	if userID, _ := getString(args, "user_id", "userId"); userID != "" {
		payload["user_id"] = userID
	}
	if agentID, _ := getString(args, "agent_id", "agentId"); agentID != "" {
		payload["agent_id"] = agentID
	}
	if appID, _ := getString(args, "app_id", "appId"); appID != "" {
		payload["app_id"] = appID
	}

	result, e := mem0Do(ctx, "POST", "/memories/", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Memory added successfully:\n%s", string(out)))
}

// HandleMem0SearchMemory searches memories via semantic query.
// Tool: mem0_search_memory
func HandleMem0SearchMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	payload := map[string]interface{}{
		"query": query,
	}

	if userID, _ := getString(args, "user_id", "userId"); userID != "" {
		payload["user_id"] = userID
	}
	if limit := getInt(args, "limit", "count"); limit > 0 {
		payload["limit"] = limit
	} else {
		payload["limit"] = 10
	}

	result, e := mem0Do(ctx, "POST", "/memories/search/", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleMem0GetMemories retrieves all memories for a user/agent.
// Tool: mem0_get_memories
func HandleMem0GetMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path := "/memories/"
	params := []string{}

	if userID, _ := getString(args, "user_id", "userId"); userID != "" {
		params = append(params, "user_id="+userID)
	}
	if agentID, _ := getString(args, "agent_id", "agentId"); agentID != "" {
		params = append(params, "agent_id="+agentID)
	}
	if limit := getInt(args, "limit", "count"); limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}

	if len(params) > 0 {
		path += "?"
		for i, p := range params {
			if i > 0 {
				path += "&"
			}
			path += p
		}
	}

	result, e := mem0Do(ctx, "GET", path, nil)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleMem0DeleteMemory deletes a specific memory by ID.
// Tool: mem0_delete_memory
func HandleMem0DeleteMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memoryID, _ := getString(args, "memory_id", "id")
	if memoryID == "" {
		return err("memory_id parameter is required")
	}

	_, e := mem0Do(ctx, "DELETE", "/memories/"+memoryID+"/", nil)
	if e != nil {
		return err(e.Error())
	}

	return ok("Memory deleted: " + memoryID)
}

// HandleMem0UpdateMemory updates an existing memory by ID.
// Tool: mem0_update_memory
func HandleMem0UpdateMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memoryID, _ := getString(args, "memory_id", "id")
	if memoryID == "" {
		return err("memory_id parameter is required")
	}

	content, _ := getString(args, "content", "text", "memory")
	if content == "" {
		return err("content parameter is required")
	}

	payload := map[string]interface{}{"memory": content}
	result, e := mem0Do(ctx, "PUT", "/memories/"+memoryID+"/", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Memory updated:\n%s", string(out)))
}
