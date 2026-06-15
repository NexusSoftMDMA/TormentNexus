//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	endpoint := os.Getenv("HOLOGRES_ENDPOINT")
	if endpoint == "" {
		return err("HOLOGRES_ENDPOINT not set")
}

	req, e := http.NewRequestWithContext(ctx, "POST", endpoint+"/query", nil)
	if e != nil {
		return err(e.Error())
}

	req.Body = io.NopCloser(io.NopCloser(nil)) // dummy
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint := os.Getenv("HOLOGRES_ENDPOINT")
	if endpoint == "" {
		return err("HOLOGRES_ENDPOINT not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", endpoint+"/tables", nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var tables []string
	if e = json.Unmarshal(body, &tables); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Tables: %v", tables))
}
