package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleHydrolixQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url, _ :=getString(args, "url")
	if query == "" || url == "" {
		return err("query and url are required")
}

	body, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return success(fmt.Sprintf("result: %v", result))
}