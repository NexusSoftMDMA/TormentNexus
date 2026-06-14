package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleExecuteCypher(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	params := map[string]interface{}{}
	if p, found := args["parameters"]; found {
		if m, found := p.(map[string]interface{}); found {
			params = m
		}
	}

	uri := os.Getenv("NEO4J_URI")
	user := os.Getenv("NEO4J_USER")
	pass := os.Getenv("NEO4J_PASSWORD")
	if uri == "" || user == "" || pass == "" {
		return err("NEO4J_URI, NEO4J_USER, NEO4J_PASSWORD must be set")
}

	body := map[string]interface{}{
		"statements": []map[string]interface{}{
			{"statement": query, "parameters": params},
		},
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", uri+"/db/neo4j/tx/commit", bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Neo4j returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}