package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleCounsel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}

	url := "https://api.example.com/counsel?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach counsel service: " + e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}

	return ok("counsel response: " + formatResult(result))
}

func formatResult(data map[string]interface{}) string {
	b, _ := json.Marshal(data)
	return string(b)
}