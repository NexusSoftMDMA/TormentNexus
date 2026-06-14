package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListQueries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.metoro.io/v1/queries", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to query Metoro: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleRunQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	reqData := map[string]string{"query": query}
	jsonData, e := json.Marshal(reqData)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.metoro.io/v1/queries", bytes.NewReader(jsonData))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to query Metoro: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}