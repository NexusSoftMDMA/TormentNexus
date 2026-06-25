package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HandleRestheartHealth checks the health status of a RESTHeart server
func HandleRestheartHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", strings.TrimSuffix(baseURL, "/")+"/_health", nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return ok(fmt.Sprintf("Health check failed (status %d): %s", resp.StatusCode, string(body)))
}

	var health map[string]interface{}
	if parseErr := json.Unmarshal(body, &health); parseErr != nil {
		return ok(string(body))
}

	result, _ := json.MarshalIndent(health, "", "  ")
	return ok(string(result))
}

// HandleRestheartListDatabases lists all databases on the RESTHeart server
func HandleRestheartListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", strings.TrimSuffix(baseURL, "/")+"/_dbprops", nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to list databases (status %d): %s", resp.StatusCode, string(body)))
}

	var dbs []map[string]interface{}
	if parseErr := json.Unmarshal(body, &dbs); parseErr != nil {
		return ok(string(body))
}

	var names []string
	for _, db := range dbs {
		if name, found := db["name"].(string); found {
			names = append(names, name)

	}

	result, _ := json.MarshalIndent(names, "", "  ")
	return ok(string(result))
}

}

// HandleRestheartListCollections lists collections in a database
func HandleRestheartListCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	if baseURL == "" || dbName == "" {
		return err("base_url and db are required")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", strings.TrimSuffix(baseURL, "/")+"/"+url.PathEscape(dbName)+"/_collections", nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to list collections (status %d): %s", resp.StatusCode, string(body)))
}

	var collections []map[string]interface{}
	if parseErr := json.Unmarshal(body, &collections); parseErr != nil {
		return ok(string(body))
}

	var names []string
	for _, col := range collections {
		if name, found := col["name"].(string); found {
			names = append(names, name)

	}

	result, _ := json.MarshalIndent(names, "", "  ")
	return ok(string(result))
}

}

// HandleRestheartQueryDocuments queries documents from a collection
func HandleRestheartQueryDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	collection, _ :=getString(args, "collection")
	if baseURL == "" || dbName == "" || collection == "" {
		return err("base_url, db, and collection are required")
}

	client := http.DefaultClient
	targetURL := strings.TrimSuffix(baseURL, "/") + "/" + url.PathEscape(dbName) + "/" + url.PathEscape(collection)

	filter, _ :=getString(args, "filter")
	page, _ :=getInt(args, "page")
	pagesize, _ :=getInt(args, "pagesize")
	sort, _ :=getString(args, "sort")
	count, _ :=getBool(args, "count")

	if page == 0 {
		page = 1
	}
	if pagesize == 0 {
		pagesize = 100
	}

	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("pagesize", fmt.Sprintf("%d", pagesize))
	if filter != "" {
		params.Set("filter", filter)

	if sort != "" {
		params.Set("sort", sort)

	if count {
		params.Set("count", "true")

	targetURL += "?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to query documents (status %d): %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return ok(string(body))
}

	output, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(output))
}

}
}
}

// HandleRestheartGetDocument retrieves a specific document by ID
func HandleRestheartGetDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	collection, _ :=getString(args, "collection")
	documentID, _ :=getString(args, "document_id")
	if baseURL == "" || dbName == "" || collection == "" || documentID == "" {
		return err("base_url, db, collection, and document_id are required")
}

	client := http.DefaultClient
	targetURL := strings.TrimSuffix(baseURL, "/") + "/" + url.PathEscape(dbName) + "/" + url.PathEscape(collection) + "/" + url.PathEscape(documentID)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to get document (status %d): %s", resp.StatusCode, string(body)))
}

	var doc map[string]interface{}
	if parseErr := json.Unmarshal(body, &doc); parseErr != nil {
		return ok(string(body))
}

	output, _ := json.MarshalIndent(doc, "", "  ")
	return ok(string(output))
}

// HandleRestheartCreateDocument creates a new document in a collection
func HandleRestheartCreateDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	collection, _ :=getString(args, "collection")
	document, _ :=getString(args, "document")
	if baseURL == "" || dbName == "" || collection == "" || document == "" {
		return err("base_url, db, collection, and document are required")
}

	client := http.DefaultClient
	targetURL := strings.TrimSuffix(baseURL, "/") + "/" + url.PathEscape(dbName) + "/" + url.PathEscape(collection)

	req, reqErr := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBufferString(document))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to create document (status %d): %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return ok(string(body))
}

	output, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(output))
}

// HandleRestheartUpdateDocument updates an existing document
func HandleRestheartUpdateDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	collection, _ :=getString(args, "collection")
	documentID, _ :=getString(args, "document_id")
	document, _ :=getString(args, "document")
	if baseURL == "" || dbName == "" || collection == "" || documentID == "" || document == "" {
		return err("base_url, db, collection, document_id, and document are required")
}

	client := http.DefaultClient
	targetURL := strings.TrimSuffix(baseURL, "/") + "/" + url.PathEscape(dbName) + "/" + url.PathEscape(collection) + "/" + url.PathEscape(documentID)

	req, reqErr := http.NewRequestWithContext(ctx, "PUT", targetURL, bytes.NewBufferString(document))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to update document (status %d): %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return ok(string(body))
}

	output, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(output))
}

// HandleRestheartDeleteDocument deletes a document by ID
func HandleRestheartDeleteDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	collection, _ :=getString(args, "collection")
	documentID, _ :=getString(args, "document_id")
	if baseURL == "" || dbName == "" || collection == "" || documentID == "" {
		return err("base_url, db, collection, and document_id are required")
}

	client := http.DefaultClient
	targetURL := strings.TrimSuffix(baseURL, "/") + "/" + url.PathEscape(dbName) + "/" + url.PathEscape(collection) + "/" + url.PathEscape(documentID)

	req, reqErr := http.NewRequestWithContext(ctx, "DELETE", targetURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to delete document (status %d): %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Document '%s' deleted successfully", documentID))
}

// HandleRestheartAggregate executes an aggregation pipeline
func HandleRestheartAggregate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	dbName, _ :=getString(args, "db")
	collection, _ :=getString(args, "collection")
	pipeline, _ :=getString(args, "pipeline")
	if baseURL == "" || dbName == "" || collection == "" || pipeline == "" {
		return err("base_url, db, collection, and pipeline are required")
}

	client := http.DefaultClient
	targetURL := strings.TrimSuffix(baseURL, "/") + "/" + url.PathEscape(dbName) + "/" + url.PathEscape(collection) + "/_aggrs"

	req, reqErr := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBufferString(pipeline))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Failed to execute aggregation (status %d): %s", resp.StatusCode, string(body)))
}

	var result []interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return ok(string(body))
}

	output, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(output))
}