package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	query, _ :=getString(args, "query")
	if collection == "" || query == "" {
		return err("collection and query are required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	body := map[string]interface{}{
		"query":     query,
		"n_results": limit,
	}
	data, e := json.Marshal(body)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	url := "http://localhost:8000/api/v1/collections/" + collection + "/query"
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return success(result)
}