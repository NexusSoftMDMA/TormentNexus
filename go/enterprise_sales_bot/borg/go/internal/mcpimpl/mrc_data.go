package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetData_mrc_data(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("https://api.example.com/mrc/data/%s", id)
	return fetchData(ctx, url)
}

func HandleSearch_mrc_data(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://api.example.com/mrc/search?q=%s", query)
	return fetchData(ctx, url)
}

func fetchData(ctx context.Context, url string) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(fmt.Sprintf("Data: %v", data))
}