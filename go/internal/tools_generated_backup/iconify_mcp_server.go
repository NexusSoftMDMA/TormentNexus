package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleIconSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://api.iconify.design/search?query=%s&limit=5", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Icons []string `json:"icons"`
		Total int      `json:"total"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	out := fmt.Sprintf("Found %d icons.\n", result.Total)
	for _, icon := range result.Icons {
		out += "  " + icon + "\n"
	}
	return ok(out)
}

func HandleGetIcon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prefix, _ :=getString(args, "prefix")
	name, _ :=getString(args, "name")
	if prefix == "" || name == "" {
		return err("prefix and name parameters are required")
}

	url := fmt.Sprintf("https://api.iconify.design/%s/%s.svg", prefix, name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("icon fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("icon not found (status %d)", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read icon: " + e.Error())
}

	return ok(string(body))
}