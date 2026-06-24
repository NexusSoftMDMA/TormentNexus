package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleListDatabases lists YDB databases
func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		return err("endpoint is required")
}

	resp, e := http.DefaultClient.Get("http://" + endpoint + "/list_databases")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data []string
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	return ok(fmt.Sprintf("Databases: %v", data))
}

// HandleQuery executes a YDB SQL query
func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	query, _ :=getString(args, "query")
	if endpoint == "" || query == "" {
		return err("endpoint and query are required")
}

	resp, e := http.DefaultClient.Get("http://" + endpoint + "/query?sql=" + query)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result []map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	return ok(fmt.Sprintf("Result: %v", result))
}