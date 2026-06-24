package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTabs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getString(args, "port")
	if port == "" {
		port = "9222"
	}
	url := fmt.Sprintf("http://%s:%s/json", host, port)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get tabs: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var tabs []map[string]interface{}
	e = json.Unmarshal(body, &tabs)
	if e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Found %d tabs", len(tabs)))
}

func HandleGetTabInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getString(args, "port")
	if port == "" {
		port = "9222"
	}
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter required")
}

	url := fmt.Sprintf("http://%s:%s/json/version/%s", host, port, id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get tab info: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var info map[string]interface{}
	e = json.Unmarshal(body, &info)
	if e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(fmt.Sprintf("Tab info: %v", info))
}