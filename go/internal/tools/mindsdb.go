//go:build ignore
// +build ignore

package tools

/**
 * @file mindsdb.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of MindsDB ML/AI database queries.
 * Replaces `mindsdb` (SSE: http://localhost:47334/mcp/sse) in mcp.json.
 *
 * Uses the MindsDB REST API (local instance).
 * Improvements over original:
 *  - No SSE connection overhead.
 *  - Supports: SQL queries, model listing, model training, predictions.
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

func mindsdbBaseURL() string {
	if u := os.Getenv("MINDSDB_URL"); u != "" {
		return u
	}
	return "http://localhost:47334"
}

func mindsdbDo(ctx context.Context, method, path string, payload interface{}) (interface{}, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, mindsdbBaseURL()+path, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")

	if token := os.Getenv("MINDSDB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("MindsDB connection failed: %v (is MindsDB running at %s?)", e, mindsdbBaseURL())
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("MindsDB API error (HTTP %d): %s", resp.StatusCode, string(body))
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

// HandleMindsDBQuery executes a SQL query against MindsDB.
// Tool: mindsdb_query
func HandleMindsDBQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "sql", "q")
	if query == "" {
		return err("query parameter is required")
	}

	result, e := mindsdbDo(ctx, "POST", "/api/sql/query", map[string]string{"query": query})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleMindsDBListModels lists all ML models in MindsDB.
// Tool: mindsdb_list_models
func HandleMindsDBListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result, e := mindsdbDo(ctx, "GET", "/api/projects/mindsdb/models", nil)
	if e != nil {
		return err(e.Error())
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleMindsDBPredict uses a MindsDB model to make a prediction.
// Tool: mindsdb_predict
func HandleMindsDBPredict(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ := getString(args, "model", "model_name")
	if model == "" {
		return err("model parameter is required")
	}

	// Build prediction query
	target, _ := getString(args, "target")
	if target == "" {
		target = model
	}

	// Build WHERE clause from input args
	inputData, dataOk := args["data"].(map[string]interface{})
	if !dataOk {
		return err("data parameter is required (object with feature values)")
	}

	whereParts := []string{}
	for k, v := range inputData {
		switch val := v.(type) {
		case string:
			whereParts = append(whereParts, fmt.Sprintf("%s = '%s'", k, val))
		case float64:
			whereParts = append(whereParts, fmt.Sprintf("%s = %v", k, val))
		case bool:
			whereParts = append(whereParts, fmt.Sprintf("%s = %v", k, val))
		default:
			whereParts = append(whereParts, fmt.Sprintf("%s = '%v'", k, val))
		}
	}

	whereClause := ""
	for i, p := range whereParts {
		if i == 0 {
			whereClause = "WHERE " + p
		} else {
			whereClause += " AND " + p
		}
	}

	query := fmt.Sprintf("SELECT * FROM mindsdb.%s %s;", model, whereClause)
	result, e := mindsdbDo(ctx, "POST", "/api/sql/query", map[string]string{"query": query})
	if e != nil {
		return err(e.Error())
	}

	outBytes, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(outBytes))
}
