package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleOdaOperation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := "https://api.oda.com/operations?query=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, found := json.Marshal(result)
	if found != nil {
		return err("marshal failed")
}

	return ok(string(data))
}