//go:build ignore
// +build ignore

package tools

/**
 * @file chroma.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of ChromaDB vector store operations.
 * Replaces `chroma-knowledge` (uvx chroma-mcp) STDIO entry in mcp.json.
 *
 * Connects to a local or remote ChromaDB HTTP server.
 * Improvements over original:
 *  - No uvx/Python dependency.
 *  - Supports: collection CRUD, document upsert, similarity query, filtering.
 *  - Context-aware with timeout.
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

func chromaBaseURL() string {
	if u := os.Getenv("CHROMA_URL"); u != "" {
		return u
	}
	host := os.Getenv("CHROMA_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("CHROMA_PORT")
	if port == "" {
		port = "8000"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

func chromaDo(ctx context.Context, method, path string, payload interface{}) (interface{}, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, chromaBaseURL()+"/api/v1"+path, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")

	if token := os.Getenv("CHROMA_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("ChromaDB connection failed: %v (is chroma running at %s?)", e, chromaBaseURL())
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ChromaDB API error (HTTP %d): %s", resp.StatusCode, string(body))
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

// HandleChromaListCollections lists all ChromaDB collections.
// Tool: chroma_list_collections
func HandleChromaListCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result, e := chromaDo(ctx, "GET", "/collections", nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleChromaCreateCollection creates a new ChromaDB collection.
// Tool: chroma_create_collection
func HandleChromaCreateCollection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name", "collection")
	if name == "" {
		return err("name parameter is required")
	}

	payload := map[string]interface{}{"name": name}
	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		payload["metadata"] = metadata
	}

	result, e := chromaDo(ctx, "POST", "/collections", payload)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Collection created:\n%s", string(out)))
}

// HandleChromaAddDocuments adds documents to a ChromaDB collection.
// Tool: chroma_add_documents
func HandleChromaAddDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ := getString(args, "collection", "collection_name")
	if collection == "" {
		return err("collection parameter is required")
	}

	documents, docsOk := args["documents"].([]interface{})
	if !docsOk || len(documents) == 0 {
		// Try single document
		if doc, _ := getString(args, "document", "text"); doc != "" {
			documents = []interface{}{doc}
		} else {
			return err("documents (array) or document parameter is required")
		}
	}

	ids, _ := args["ids"].([]interface{})
	if len(ids) != len(documents) {
		// Auto-generate IDs
		ids = make([]interface{}, len(documents))
		for i := range ids {
			ids[i] = fmt.Sprintf("doc_%d", i)
		}
	}

	payload := map[string]interface{}{
		"ids":       ids,
		"documents": documents,
	}

	if metadatas, mOk := args["metadatas"].([]interface{}); mOk {
		payload["metadatas"] = metadatas
	}

	result, e := chromaDo(ctx, "POST", "/collections/"+collection+"/add", payload)
	if e != nil {
		return err(e.Error())
	}
	outBytes, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Added %d document(s) to collection '%s':\n%s", len(documents), collection, string(outBytes)))
}

// HandleChromaQuery performs a similarity search in a ChromaDB collection.
// Tool: chroma_query
func HandleChromaQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ := getString(args, "collection", "collection_name")
	if collection == "" {
		return err("collection parameter is required")
	}

	queryText, _ := getString(args, "query", "query_text", "q")
	if queryText == "" {
		return err("query parameter is required")
	}

	nResults := getInt(args, "n_results", "k", "top_k")
	if nResults <= 0 {
		nResults = 5
	}

	payload := map[string]interface{}{
		"query_texts": []string{queryText},
		"n_results":   nResults,
	}

	if include, ok := args["include"].([]interface{}); ok {
		payload["include"] = include
	} else {
		payload["include"] = []string{"documents", "metadatas", "distances"}
	}

	if where, ok := args["where"].(map[string]interface{}); ok {
		payload["where"] = where
	}

	result, e := chromaDo(ctx, "POST", "/collections/"+collection+"/query", payload)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleChromaDeleteCollection deletes a ChromaDB collection.
// Tool: chroma_delete_collection
func HandleChromaDeleteCollection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name", "collection")
	if name == "" {
		return err("name parameter is required")
	}

	_, e := chromaDo(ctx, "DELETE", "/collections/"+name, nil)
	if e != nil {
		return err(e.Error())
	}
	return ok("Collection deleted: " + name)
}

// HandleChromaGetCollection retrieves documents from a collection.
// Tool: chroma_get_documents
func HandleChromaGetCollection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ := getString(args, "collection", "collection_name")
	if collection == "" {
		return err("collection parameter is required")
	}

	payload := map[string]interface{}{}
	if limit := getInt(args, "limit", "count"); limit > 0 {
		payload["limit"] = limit
	}
	if ids, ok := args["ids"].([]interface{}); ok {
		payload["ids"] = ids
	}

	result, e := chromaDo(ctx, "POST", "/collections/"+collection+"/get", payload)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

