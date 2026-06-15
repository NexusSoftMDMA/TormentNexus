package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://api.gqai.example/info?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch info: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Result: %v", result))
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	resp, e := http.DefaultClient.Get("https://api.gqai.example/query?prompt=" + prompt)
	if e != nil {
		return err("query failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return success(string(data))
}