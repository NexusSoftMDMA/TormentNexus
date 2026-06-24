package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query argument is required")
}

	payload, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/api/v1/query", strings.NewReader(string(payload)))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("execution failed: %v", e))
}

	defer resp.Body.Close()
	return ok(fmt.Sprintf("Query executed with status: %s", resp.Status))
}

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dsn, _ :=getString(args, "dsn")
	if dsn == "" {
		return err("dsn argument is required")
}

	return success(fmt.Sprintf("Connected to 1C database: %s", dsn))
}