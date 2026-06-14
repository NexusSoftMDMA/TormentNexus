package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	baseURL := os.Getenv("URDB_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	u, e := url.Parse(baseURL + "/query")
	if e != nil {
		return err("invalid base URL: " + e.Error())
}

	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("URDB_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	resp, e := http.DefaultClient.Get(baseURL + "/databases")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var databases []string
	if e := json.Unmarshal(body, &databases); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return success(fmt.Sprintf("Databases: %v", databases))
}