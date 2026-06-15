//go:build ignore
// +build ignore

package tools

/**
 * @file openmemory.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of OpenMemory — a local persistent memory store for LLM agents.
 * Replaces the Python/Node OpenMemory SDK dependency with direct HTTP calls.
 *
 * OpenMemory features:
 *  - add(text, user_id) — store a memory
 *  - search(query, user_id) — search memories
 *  - get(memory_id) — retrieve a specific memory
 *  - delete(memory_id) — delete a memory
 *  - list(user_id) — list all memories for a user
 *
 * Connects to the OpenMemory HTTP API (defaults to http://localhost:8000).
 * Set OPENMEMORY_BASE_URL to override.
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

// Default OpenMemory server URL
const openMemoryDefaultBaseURL = "http://localhost:8000"

func openMemoryBaseURL() string {
	if u := os.Getenv("OPENMEMORY_BASE_URL"); u != "" {
		return u
	}
	return openMemoryDefaultBaseURL
}

// ---------------------------------------------------------------------------
// HandleOpenMemoryAdd — store a memory
// ---------------------------------------------------------------------------

// HandleOpenMemoryAdd stores a new memory entry.
// Required: text (string) — the memory content to store
// Optional: user_id (string) — owner of the memory
func HandleOpenMemoryAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ := getString(args, "text", "memory", "content")
	if text == "" {
		return err("text is required")
	}
	userID, _ := getString(args, "user_id", "userId", "user")

	payload := map[string]interface{}{"text": text}
	if userID != "" {
		payload["user_id"] = userID
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "POST",
		openMemoryBaseURL()+"/api/v1/memories", bytes.NewReader(body))
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
		return err(fmt.Sprintf("OpenMemory API error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// ---------------------------------------------------------------------------
// HandleOpenMemorySearch — search memories
// ---------------------------------------------------------------------------

// HandleOpenMemorySearch searches memories by query text.
// Required: query (string) — the search query
// Optional: user_id (string) — scope search to a specific user
func HandleOpenMemorySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "text")
	if query == "" {
		return err("query is required")
	}
	userID, _ := getString(args, "user_id", "userId", "user")

	url := openMemoryBaseURL() + "/api/v1/memories/search?q=" + query
	if userID != "" {
		url += "&user_id=" + userID
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("OpenMemory search error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// ---------------------------------------------------------------------------
// HandleOpenMemoryGet — retrieve a specific memory
// ---------------------------------------------------------------------------

// HandleOpenMemoryGet retrieves a memory by its ID.
// Required: id (string) — memory ID
func HandleOpenMemoryGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "memory_id", "memoryId")
	if id == "" {
		return err("id is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET",
		openMemoryBaseURL()+"/api/v1/memories/"+id, nil)
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
		return err(fmt.Sprintf("OpenMemory get error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// ---------------------------------------------------------------------------
// HandleOpenMemoryDelete — delete a memory
// ---------------------------------------------------------------------------

// HandleOpenMemoryDelete deletes a memory by its ID.
// Required: id (string) — memory ID
func HandleOpenMemoryDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ := getString(args, "id", "memory_id", "memoryId")
	if id == "" {
		return err("id is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "DELETE",
		openMemoryBaseURL()+"/api/v1/memories/"+id, nil)
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
		return err(fmt.Sprintf("OpenMemory delete error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}

// ---------------------------------------------------------------------------
// HandleOpenMemoryList — list all memories for a user
// ---------------------------------------------------------------------------

// HandleOpenMemoryList lists all memories, optionally filtered by user_id.
// Optional: user_id (string) — scope to a specific user
func HandleOpenMemoryList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ := getString(args, "user_id", "userId", "user")

	url := openMemoryBaseURL() + "/api/v1/memories"
	if userID != "" {
		url += "?user_id=" + userID
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("list failed: %v", e))
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("OpenMemory list error (%d): %s", resp.StatusCode, string(data)))
	}
	return ok(string(data))
}
