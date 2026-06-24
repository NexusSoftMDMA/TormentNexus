package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	db, _ :=getString(args, "database")
	if db == "" {
		db = "public"
	}
	url := fmt.Sprintf("http://localhost:4000/v1/sql?db=%s", db)
	resp, e := http.DefaultClient.Post(url, "text/plain", nil)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return ok(string(body))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, _ :=getString(args, "database")
	if db == "" {
		db = "public"
	}
	url := fmt.Sprintf("http://localhost:4000/v1/sql?db=%s&sql=SHOW TABLES", db)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return ok(string(body))
}