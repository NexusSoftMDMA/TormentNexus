//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	token := os.Getenv("AZURE_ACCESS_TOKEN")
	if token == "" {
		return err("AZURE_ACCESS_TOKEN not set")
}

	body := map[string]interface{}{
		"query":   query,
		"options": map[string]string{"resultFormat": "objectArray"},
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://management.azure.com/providers/Microsoft.ResourceGraph/resources?api-version=2021-03-01", bytes.NewReader(b))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	r, _ := json.Marshal(result)
	return success(string(r))
}
