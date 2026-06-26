package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleTrinoQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseUrl")
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := strings.TrimRight(baseURL, "/") + "/v1/statement"
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(query))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "text/plain")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Query result: %v", result))
}