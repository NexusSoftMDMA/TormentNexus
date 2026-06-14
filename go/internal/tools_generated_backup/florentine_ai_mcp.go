package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// HandleQueryMongo queries MongoDB using natural language
func HandleQueryMongo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	payload, e := json.Marshal(map[string]string{"query": query, "database": "mongodb"})
	if e != nil {
		return err("failed to marshal payload")
}

	resp, e := http.DefaultClient.Post("https://api.florentine.ai/query", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to make request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result struct {
		Result string `json:"result"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return ok(result.Result)
}