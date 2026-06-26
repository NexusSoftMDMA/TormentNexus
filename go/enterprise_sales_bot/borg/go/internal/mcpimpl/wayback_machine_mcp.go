package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleCheckArchived(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	if target == "" {
		return err("Missing URL")
}

	apiURL := fmt.Sprintf("https://archive.org/wayback/available?url=%s", url.QueryEscape(target))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("Parse failed: %v", e))
}

	archived := false
	if results, found := result["results"].([]interface{}); found && len(results) > 0 {
		archived = true
	}
	if archived {
		return ok(fmt.Sprintf("URL is archived: %s", target))
}

	return ok(fmt.Sprintf("URL not archived: %s", target))
}// touch 1781132144
