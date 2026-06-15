package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	sql, _ :=getString(args, "sql")
	if url == "" {
		return err("url is required")
}

	if sql == "" {
		return err("sql is required")
}

	fullURL := url + "/influxdb/v2/query?q=" + sql
	resp, e := http.DefaultClient.Get(fullURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleListStorageGroups(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	fullURL := url + "/v1/storageGroups"
	resp, e := http.DefaultClient.Get(fullURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var groups []string
	if e = json.Unmarshal(body, &groups); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return ok(fmt.Sprintf("Storage groups: %v", groups))
}