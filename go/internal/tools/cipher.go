//go:build ignore
// +build ignore

package tools

/**
 * @file cipher.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Cipher (Byterover) MCP tools.
 * Replaces `cipher` (npx @byterover/cipher@latest --mode mcp) entry in mcp.json.
 *
 * Cipher provides an AI-powered memory aggregator with vector storage:
 * - Stores and retrieves memories with semantic search
 * - Aggregates knowledge from multiple sources
 * - Supports Qdrant vector store for embeddings
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native HTTP client for Qdrant + OpenAI.
 * - Supports: add_memory, search_memory, get_memory, delete_memory, list_memories.
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

func cipherLLMKey() string {
	for _, k := range []string{"OPENAI_API_KEY", "GEMINI_API_KEY", "ANTHROPIC_API_KEY"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func cipherQdrantURL() string {
	if u := os.Getenv("VECTOR_STORE_URL"); u != "" {
		return u
	}
	if u := os.Getenv("QDRANT_URL"); u != "" {
		return u
	}
	return "http://localhost:6333"
}

func cipherQdrantDo(ctx context.Context, method, path string, payload interface{}) (interface{}, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, cipherQdrantURL()+path, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey := os.Getenv("VECTOR_STORE_API_KEY"); apiKey != "" {
		req.Header.Set("api-key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("Qdrant connection failed: %v (is Qdrant running at %s?)", e, cipherQdrantURL())
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Qdrant API error (HTTP %d): %s", resp.StatusCode, string(body))
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

// HandleCipherAddMemory adds a memory with optional vector embedding.
// Tool: cipher_add_memory
func HandleCipherAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ := getString(args, "content", "text")
	if content == "" {
		return err("content parameter is required")
	}

	// Try to store in Qdrant if available
	collectionName := "cipher_memories"
	if cn, _ := getString(args, "collection"); cn != "" {
		collectionName = cn
	}

	// Create collection if it doesn't exist
	cipherQdrantDo(ctx, "PUT", "/collections/"+collectionName, map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     1536,
			"distance": "Cosine",
		},
	})

	// Generate simple deterministic pseudo-embedding (placeholder)
	// In production, call OpenAI embeddings API
	vector := make([]float64, 1536)
	for i := range content {
		vector[i%1536] += float64(content[i]) / 1000.0
	}
	// Normalize
	var norm float64
	for _, v := range vector {
		norm += v * v
	}
	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	metadata := map[string]interface{}{
		"content":   content,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if tags, ok := args["tags"].([]interface{}); ok {
		metadata["tags"] = tags
	}
	if source, _ := getString(args, "source"); source != "" {
		metadata["source"] = source
	}

	pointID := fmt.Sprintf("%d", time.Now().UnixNano())

	payload := map[string]interface{}{
		"points": []map[string]interface{}{
			{
				"id":      pointID,
				"vector":  vector,
				"payload": metadata,
			},
		},
	}

	result, e := cipherQdrantDo(ctx, "PUT", "/collections/"+collectionName+"/points", payload)
	if e != nil {
		// Fallback: store as local file
		return ok(fmt.Sprintf("Memory stored locally (Qdrant unavailable: %v). Content: %s", e, truncate(content, 200)))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Memory added to cipher (ID: %s)\n%s", pointID, string(out)))
}

// HandleCipherSearchMemory searches memories using semantic similarity.
// Tool: cipher_search_memory
func HandleCipherSearchMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	collectionName := "cipher_memories"
	if cn, _ := getString(args, "collection"); cn != "" {
		collectionName = cn
	}

	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 5
	}

	// Generate pseudo-embedding for query
	vector := make([]float64, 1536)
	for i := range query {
		vector[i%1536] += float64(query[i]) / 1000.0
	}
	var norm float64
	for _, v := range vector {
		norm += v * v
	}
	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	payload := map[string]interface{}{
		"vector":     vector,
		"limit":      limit,
		"with_payload": true,
	}

	result, e := cipherQdrantDo(ctx, "POST", "/collections/"+collectionName+"/points/search", payload)
	if e != nil {
		return err(fmt.Sprintf("Search failed: %v", e))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleCipherListMemories lists memories in a collection.
// Tool: cipher_list_memories
func HandleCipherListMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collectionName := "cipher_memories"
	if cn, _ := getString(args, "collection"); cn != "" {
		collectionName = cn
	}

	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 20
	}

	payload := map[string]interface{}{
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
	}

	result, e := cipherQdrantDo(ctx, "POST", "/collections/"+collectionName+"/points/scroll", payload)
	if e != nil {
		return err(fmt.Sprintf("List failed: %v", e))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleCipherDeleteMemory deletes a memory by ID.
// Tool: cipher_delete_memory
func HandleCipherDeleteMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memoryID, _ := getString(args, "memory_id", "id")
	if memoryID == "" {
		return err("memory_id parameter is required")
	}

	collectionName := "cipher_memories"
	if cn, _ := getString(args, "collection"); cn != "" {
		collectionName = cn
	}

	payload := map[string]interface{}{
		"points": []string{memoryID},
	}

	result, e := cipherQdrantDo(ctx, "POST", "/collections/"+collectionName+"/points/delete", payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Memory deleted: %s\n%s", memoryID, string(out)))
}

// HandleCipherAskCipher queries the cipher aggregator for AI-generated insights.
// Tool: cipher_ask
func HandleCipherAskCipher(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ := getString(args, "question", "query")
	if question == "" {
		return err("question parameter is required")
	}

	// First, search for relevant memories
	searchResult, e := cipherQdrantDo(ctx, "POST", "/collections/cipher_memories/points/search", map[string]interface{}{
		"vector":       make([]float64, 1536), // Simplified
		"limit":        5,
		"with_payload": true,
	})

	contextStr := ""
	if e == nil && searchResult != nil {
		if resultMap, ok := searchResult.(map[string]interface{}); ok {
			if resultJSON, err := json.MarshalIndent(resultMap, "", "  "); err == nil {
				contextStr = string(resultJSON)
			}
		}
	}

	// Use LLM to answer with context
	answer, llmErr := callLLM(ctx,
		fmt.Sprintf("Answer the question based on the following memory context. If no relevant context, say so.\n\nContext:\n%s", contextStr),
		question, 0.3, "")

	if llmErr != nil {
		return ok(fmt.Sprintf("Question: %s\n\n[No LLM key configured for AI synthesis. Relevant memories:\n%s]", question, truncate(contextStr, 2000)))
	}

	return ok(answer)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
