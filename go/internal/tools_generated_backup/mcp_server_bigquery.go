package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleRunQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ :=getString(args, "projectId")
	query, _ :=getString(args, "query")
	if projectId == "" || query == "" {
		return err("projectId and query are required")
}

	bodyMap := map[string]string{"query": query}
	bodyBytes, e := json.Marshal(bodyMap)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://bigquery.googleapis.com/bigquery/v2/projects/%s/queries", projectId), bytes.NewReader(bodyBytes))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleListDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ :=getString(args, "projectId")
	if projectId == "" {
		return err("projectId is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://bigquery.googleapis.com/bigquery/v2/projects/%s/datasets", projectId), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Datasets: %v", result))
}