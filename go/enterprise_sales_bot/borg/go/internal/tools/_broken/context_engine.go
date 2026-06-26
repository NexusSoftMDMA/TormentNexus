package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	contextEngineBaseURL = "http://localhost:8080/api"
)

// HandleRepoSearch performs semantic code search within a repository
func HandleRepoSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	collection, _ :=getString(args, "collection")
	limit, _ :=getInt(args, "limit", 5)
	includeSnippet, _ :=getBool(args, "include_snippet", false)
	contextLines, _ :=getInt(args, "context_lines", 3)
	language, _ :=getString(args, "language")
	under, _ :=getString(args, "under")
	compact, _ :=getBool(args, "compact", false)
	outputFormat, _ :=getString(args, "output_format", "default")

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("include_snippet", fmt.Sprintf("%v", includeSnippet))
	params.Set("context_lines", fmt.Sprintf("%d", contextLines))
	params.Set("compact", fmt.Sprintf("%v", compact))
	params.Set("output_format", outputFormat)

	if collection != "" {
		params.Set("collection", collection)

	if language != "" {
		params.Set("language", language)

	if under != "" {
		params.Set("under", under)

	// Handle path_glob and not_glob arrays
	if pathGlobs, found := args["path_glob"].([]interface{}); found {
		for _, g := range pathGlobs {
			if s, found := g.(string); found {
				params.Add("path_glob", s)

		}
	}
	if notGlobs, found := args["not_glob"].([]interface{}); found {
		for _, g := range notGlobs {
			if s, found := g.(string); found {
				params.Add("not_glob", s)

		}
	}

	apiURL := fmt.Sprintf("%s/repo_search?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}
}
}
}
}

// HandleCrossRepoSearch searches across multiple repositories
func HandleCrossRepoSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	discover, _ :=getString(args, "discover", "auto")
	traceBoundary, _ :=getBool(args, "trace_boundary", false)
	limit, _ :=getInt(args, "limit", 5)
	boundaryKey, _ :=getString(args, "boundary_key")

	params := url.Values{}
	params.Set("query", query)
	params.Set("discover", discover)
	params.Set("trace_boundary", fmt.Sprintf("%v", traceBoundary))
	params.Set("limit", fmt.Sprintf("%d", limit))

	if boundaryKey != "" {
		params.Set("boundary_key", boundaryKey)

	// Handle target_repos array
	if targetRepos, found := args["target_repos"].([]interface{}); found {
		for _, r := range targetRepos {
			if s, found := r.(string); found {
				params.Add("target_repos", s)

		}
	}

	apiURL := fmt.Sprintf("%s/cross_repo_search?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}
}

// HandleSymbolGraph queries symbol relationships (callers, callees, definitions, importers)
func HandleSymbolGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	queryType, _ :=getString(args, "query_type", "callers")
	collection, _ :=getString(args, "collection")
	limit, _ :=getInt(args, "limit", 10)
	depth, _ :=getInt(args, "depth", 1)

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("query_type", queryType)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("depth", fmt.Sprintf("%d", depth))

	if collection != "" {
		params.Set("collection", collection)

	apiURL := fmt.Sprintf("%s/symbol_graph?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}

// HandleContextAnswer provides natural language answers with code citations
func HandleContextAnswer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	collection, _ :=getString(args, "collection")
	limit, _ :=getInt(args, "limit", 5)

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))

	if collection != "" {
		params.Set("collection", collection)

	apiURL := fmt.Sprintf("%s/context_answer?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}

// HandleMemoryStore stores knowledge/information in the memory system
func HandleMemoryStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	information, _ :=getString(args, "information")
	if information == "" {
		return err("information is required")
}

	topic, _ :=getString(args, "topic")
	kind, _ :=getString(args, "kind")

	params := url.Values{}
	params.Set("information", information)

	if topic != "" {
		params.Set("topic", topic)

	if kind != "" {
		params.Set("kind", kind)

	apiURL := fmt.Sprintf("%s/memory_store", contextEngineBaseURL)
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(params.Encode()))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}
}

// HandleMemoryFind retrieves stored knowledge from the memory system
func HandleMemoryFind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	limit, _ :=getInt(args, "limit", 5)
	topic, _ :=getString(args, "topic")

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))

	if topic != "" {
		params.Set("topic", topic)

	apiURL := fmt.Sprintf("%s/memory_find?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}

// HandleQdrantStatus returns the status of the Qdrant search backend
func HandleQdrantStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	listAll, _ :=getBool(args, "list_all", false)

	params := url.Values{}
	params.Set("list_all", fmt.Sprintf("%v", listAll))

	apiURL := fmt.Sprintf("%s/qdrant_status?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

// HandleSetSessionDefaults sets default parameters for the session
func HandleSetSessionDefaults(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	outputFormat, _ :=getString(args, "output_format", "toon")
	compact, _ :=getBool(args, "compact", true)
	limit, _ :=getInt(args, "limit", 5)
	collection, _ :=getString(args, "collection")

	params := url.Values{}
	params.Set("output_format", outputFormat)
	params.Set("compact", fmt.Sprintf("%v", compact))
	params.Set("limit", fmt.Sprintf("%d", limit))

	if collection != "" {
		params.Set("collection", collection)

	apiURL := fmt.Sprintf("%s/set_session_defaults", contextEngineBaseURL)
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(params.Encode()))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}

// HandleBatchSearch performs multiple independent searches at once
func HandleBatchSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queries, found := args["queries"].([]interface{})
	if !ok || len(queries) == 0 {
		return err("queries array is required")
}

	limit, _ :=getInt(args, "limit", 5)
	collection, _ :=getString(args, "collection")

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))

	if collection != "" {
		params.Set("collection", collection)

	// Add each query as a separate parameter
	for _, q := range queries {
		if s, found := q.(string); found {
			params.Add("queries", s)

	}

	apiURL := fmt.Sprintf("%s/batch_search?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}
}

// HandleInfoRequest performs a lightweight natural language lookup
func HandleInfoRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	collection, _ :=getString(args, "collection")
	includeExplanation, _ :=getBool(args, "include_explanation", true)

	params := url.Values{}
	params.Set("query", query)
	params.Set("include_explanation", fmt.Sprintf("%v", includeExplanation))

	if collection != "" {
		params.Set("collection", collection)

	apiURL := fmt.Sprintf("%s/info_request?%s", contextEngineBaseURL, params.Encode())
	
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	output, outErr := json.MarshalIndent(result, "", "  ")
	if outErr != nil {
		return err(outErr.Error())
}

	return ok(string(output))
}

}

// HandlePatternSearch searches for structurally similar code patterns
func HandlePatternSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	limit, _ :=getInt(args, "limit", 10)
	queryMode, _ :=getString(args, "query_mode", "natural")
	collection, _ :=getString(args, "collection")

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
}