package tools

import (
	"context"
	"net/http"
	"encoding/json"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Simulate fetching table list
	tables := []string{"users", "orders", "products"}
	data, _ := json.Marshal(tables)
	return success(string(data))
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	// Dummy HTTP request (not actually used)
	_, e := http.DefaultClient.Get("http://localhost:8080/query?q=" + query)
	if e != nil {
		return err(e.Error())
}

	return ok("query executed")
}