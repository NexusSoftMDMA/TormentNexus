package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	body, _ := json.Marshal(map[string]interface{}{"statements": []map[string]string{{"statement": query}}})
	resp, e := http.Post(url+"/db/neo4j/tx/commit", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(result)
}

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.Get(url + "/db/neo4j/tx/commit?statements=" + "SHOW DATABASES")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var databases []string
	json.NewDecoder(resp.Body).Decode(&databases)
	return ok(databases)
}