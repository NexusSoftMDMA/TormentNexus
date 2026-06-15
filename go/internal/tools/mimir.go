//go:build ignore
// +build ignore

package tools

/**
 * @file mimir.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Mimir — Neo4j-backed persistent memory for AI agents.
 * Replaces: Mimir MCP server (Node.js + Neo4j)
 *
 * Provides graph-based memory storage using Neo4j's HTTP Cypher API.
 * Configurable via MIMIR_NEO4J_URL (default http://localhost:7474).
 *
 * Tools:
 *  - mimir_store — store a memory with context
 *  - mimir_search — search memories by query
 *  - mimir_retrieve — retrieve a specific memory
 *  - mimir_connect — create relationships between memories
 *  - mimir_forget — delete a memory
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

const mimirDefaultNeo4jURL = "http://localhost:7474"

func mimirNeo4jURL() string {
	if u := os.Getenv("MIMIR_NEO4J_URL"); u != "" {
		return u
	}
	return mimirDefaultNeo4jURL
}

func mimirCypherAuth() (string, string) {
	return os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASS")
}

type cypherRequest struct {
	Statements []cypherStatement `json:"statements"`
}

type cypherStatement struct {
	Statement  string          `json:"statement"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

func runCypher(ctx context.Context, query string, params map[string]interface{}) (string, error) {
	body, _ := json.Marshal(cypherRequest{
		Statements: []cypherStatement{{
			Statement:  query,
			Parameters: mustMarshalJSON(params),
		}},
	})

	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "POST",
		mimirNeo4jURL()+"/db/neo4j/tx/commit", bytes.NewReader(body))
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Content-Type", "application/json")
	if u, p := mimirCypherAuth(); u != "" {
		req.SetBasicAuth(u, p)
	}

	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("cypher request failed: %v", e)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Neo4j error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

// HandleMimirStore stores a memory node in the graph.
func HandleMimirStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ := getString(args, "content", "text", "memory")
	if content == "" {
		return err("content is required")
	}
	tags, _ := getString(args, "tags", "category")

	query := "CREATE (m:Memory {content: $content, created: datetime()}) RETURN id(m) as id"
	params := map[string]interface{}{"content": content}
	if tags != "" {
		query = "CREATE (m:Memory {content: $content, tags: $tags, created: datetime()}) RETURN id(m) as id"
		params["tags"] = tags
	}

	result, e := runCypher(ctx, query, params)
	if e != nil {
		return err(fmt.Sprintf("store failed: %v", e))
	}
	return ok(result)
}

// HandleMimirSearch searches memories by content similarity.
func HandleMimirSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "text")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	cql := fmt.Sprintf(
		"MATCH (m:Memory) WHERE m.content CONTAINS $query RETURN id(m) as id, m.content as content, m.tags as tags LIMIT %d",
		limit)
	result, e := runCypher(ctx, cql, map[string]interface{}{"query": query})
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	return ok(result)
}

// HandleMimirRetrieve retrieves a memory by ID.
func HandleMimirRetrieve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id", "memory_id", "node_id")
	if id <= 0 {
		return err("valid id is required")
	}

	result, e := runCypher(ctx,
		"MATCH (m:Memory) WHERE id(m) = $id RETURN id(m) as id, m.content as content, m.tags as tags, m.created as created",
		map[string]interface{}{"id": id})
	if e != nil {
		return err(fmt.Sprintf("retrieve failed: %v", e))
	}
	return ok(result)
}

// HandleMimirConnect creates a relationship between two memory nodes.
func HandleMimirConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sourceID := getInt(args, "source_id", "from")
	targetID := getInt(args, "target_id", "to")
	relType, _ := getString(args, "relationship", "rel", "type")

	if sourceID <= 0 || targetID <= 0 {
		return err("source_id and target_id are required")
	}
	if relType == "" {
		relType = "RELATED_TO"
	}

	result, e := runCypher(ctx,
		fmt.Sprintf("MATCH (a:Memory), (b:Memory) WHERE id(a)=$sid AND id(b)=$tid CREATE (a)-[:%s]->(b) RETURN id(a) as source, id(b) as target", relType),
		map[string]interface{}{"sid": sourceID, "tid": targetID})
	if e != nil {
		return err(fmt.Sprintf("connect failed: %v", e))
	}
	return ok(result)
}

// HandleMimirForget deletes a memory node.
func HandleMimirForget(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id", "memory_id", "node_id")
	if id <= 0 {
		return err("valid id is required")
	}

	result, e := runCypher(ctx,
		"MATCH (m:Memory) WHERE id(m)=$id DETACH DELETE m RETURN count(*) as deleted",
		map[string]interface{}{"id": id})
	if e != nil {
		return err(fmt.Sprintf("forget failed: %v", e))
	}
	return ok(result)
}
