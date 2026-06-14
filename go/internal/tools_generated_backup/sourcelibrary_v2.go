package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearchSources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query := strings.TrimSpace(getString(args, "query"))
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://api.sourcelibrary.com/v2/sources?name=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var sources []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&sources); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	count := len(sources)
	msg := fmt.Sprintf("Found %d source(s) matching '%s'", count, query)
	return success(msg)
}