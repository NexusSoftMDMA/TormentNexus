//go:build ignore
// +build ignore

package tools

/**
 * @file automem.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of AutoMem MCP — graph-vector memory for AI agents.
 * Replaces: @verygoodplugins/mcp-automem (npm)
 *
 * AutoMem provides persistent memory via a graph-vector architecture.
 * Features: add, search, get, delete, update, associate memories.
 *
 * Connects to the AutoMem HTTP API (defaults to http://localhost:3000).
 * Set AUTOMEM_BASE_URL to override.
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

const automemDefaultBaseURL = "http://localhost:3000"

func automemBaseURL() string {
	if u := os.Getenv("AUTOMEM_BASE_URL"); u != "" {
		return u
	}
	return automemDefaultBaseURL
}

// HandleAutoMemAdd stores a new memory.
func HandleAutoMemAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ := getString(args, "content", "text", "memory")
	if content == "" {
		return err("content is required")
	}
	sessionID, _ := getString(args, "session_id", "sessionId")

	payload := map[string]interface{}{"content": content}
	if sessionID != "" {
		payload["session_id"] = sessionID
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "POST",
		automemBaseURL()+"/api/v1/memories", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("AutoMem error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// HandleAutoMemSearch searches memories.
func HandleAutoMemSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	resp, e := http.Get(fmt.Sprintf("%s/api/v1/memories/search?q=%s&limit=%d",
		automemBaseURL(), query, limit))
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("AutoMem search error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// HandleAutoMemGet retrieves a memory by ID.
func HandleAutoMemGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "memory_id", "memoryId")
	if id == "" {
		return err("id is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET",
		automemBaseURL()+"/api/v1/memories/"+id, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("get failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("AutoMem get error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// HandleAutoMemDelete deletes a memory by ID.
func HandleAutoMemDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "memory_id", "memoryId")
	if id == "" {
		return err("id is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "DELETE",
		automemBaseURL()+"/api/v1/memories/"+id, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("delete failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("AutoMem delete error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// HandleAutoMemList lists all memories.
func HandleAutoMemList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ := getString(args, "session_id", "sessionId")
	url := automemBaseURL() + "/api/v1/memories"
	if sessionID != "" {
		url += "?session_id=" + sessionID
	}

	resp, e := http.Get(url)
	if e != nil {
		return err(fmt.Sprintf("list failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("AutoMem list error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// HandleAutoMemAssociate links two memories.
func HandleAutoMemAssociate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sourceID, _ := getString(args, "source_id", "sourceId")
	targetID, _ := getString(args, "target_id", "targetId")
	relation, _ := getString(args, "relation", "relationship")

	if sourceID == "" || targetID == "" {
		return err("source_id and target_id are required")
	}
	if relation == "" {
		relation = "related"
	}

	payload := map[string]string{
		"source_id": sourceID,
		"target_id": targetID,
		"relation":  relation,
	}
	body, _ := json.Marshal(payload)

	resp, e := http.Post(automemBaseURL()+"/api/v1/memories/associate",
		"application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("associate failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("AutoMem associate error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}
