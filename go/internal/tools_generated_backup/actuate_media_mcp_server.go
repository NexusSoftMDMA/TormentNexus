package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleCreateContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contentType, _ :=getString(args, "type")
	if contentType == "" {
		return err("missing 'type' parameter")
}

	title, _ :=getString(args, "title")
	body, _ :=getString(args, "body")
	payload := map[string]string{"type": contentType, "title": title, "body": body}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/content", strings.NewReader(string(data)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(bodyBytes)))
}

	return success("content created")
}

func HandleListContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contentType, _ :=getString(args, "type")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("http://localhost:3000/api/content?type=%s&limit=%d", contentType, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}