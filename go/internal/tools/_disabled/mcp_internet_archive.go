package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleCheckAvailability(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	if target == "" {
		return err("Missing required parameter: url")
}

	apiURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s&output=json&limit=1", url.QueryEscape(target))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("Failed to query Wayback Machine: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
}

	var data [][]string
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("Failed to parse response: %v", e))
}

	if len(data) <= 1 {
		return ok(fmt.Sprintf("No archived snapshot found for %s", target))
}

	snapshot := data[1]
	return success(fmt.Sprintf("Found snapshot: %s - Timestamp: %s", target, snapshot[1]))
}

func HandleSaveURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	if target == "" {
		return err("Missing required parameter: url")
}

	apiURL := fmt.Sprintf("https://web.archive.org/save/%s", url.QueryEscape(target))
	resp, e := http.DefaultClient.Post(apiURL, "text/plain", strings.NewReader(""))
	if e != nil {
		return err(fmt.Sprintf("Failed to submit save request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
}

	return success(fmt.Sprintf("Save request submitted for %s. Response: %s", target, string(body)))
}// touch 1781132132
