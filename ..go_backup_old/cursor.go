package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.example.com/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("parse error: " + e.Error())
}

	return ok(fmt.Sprintf("Found %v", result["count"]))
}

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://api.example.com/info/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("info failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return success("Info: " + string(data))
}