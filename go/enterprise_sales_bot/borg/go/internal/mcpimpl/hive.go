package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleQuery_hive(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	base := os.Getenv("HIVE_API_URL")
	if base == "" {
		base = "https://api.hive.com/v1"
	}
	u := fmt.Sprintf("%s/query?q=%s", base, url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	out, _ := json.Marshal(data)
	return ok(string(out))
}

func HandleList_hive(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("HIVE_API_URL")
	if base == "" {
		base = "https://api.hive.com/v1"
	}
	resp, e := http.DefaultClient.Get(base + "/databases")
	if e != nil {
		return err("list failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}