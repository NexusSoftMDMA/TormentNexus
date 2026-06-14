package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleDoltQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	reqBody, e := json.Marshal(map[string]interface{}{
		"query": query,
	})
	if e != nil {
		return err("failed to encode request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:50051/query", strings.NewReader(string(reqBody)))
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
		return err("failed to decode response: " + e.Error())
}

	return success("Query result: " + toJSON(result))
}

func HandleDoltStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	if database == "" {
		database = "default"
	}

	url := "http://localhost:50051/status?database=" + database
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return err("failed to decode response: " + e.Error())
}

	return ok("Status: " + toJSON(result))
}

func toJSON(v interface{}) string {
	b, e := json.Marshal(v)
	if e != nil {
		return e.Error()
}

	return string(b)
}// touch 1781132125
