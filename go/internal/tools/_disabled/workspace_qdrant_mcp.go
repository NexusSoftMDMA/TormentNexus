package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	if collection == "" {
		return err("collection is required")
}

	vecStr, _ :=getString(args, "vector")
	if vecStr == "" {
		return err("vector is required")
}

	url := fmt.Sprintf("http://localhost:6333/collections/%s/points/search", collection)
	body := fmt.Sprintf(`{"vector": %s, "limit": 5}`, vecStr)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("search result: %v", result))
}