package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListBookmarks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("KARA_KEEP_BASE_URL")
	if baseURL == "" {
		return err("KARA_KEEP_BASE_URL not set")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 20
	}
	url := fmt.Sprintf("%s/api/bookmarks?limit=%d", baseURL, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse json: %v", e))
}

	return ok(fmt.Sprintf("Bookmarks: %+v", data))
}

func HandleAddBookmark(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("KARA_KEEP_BASE_URL")
	if baseURL == "" {
		return err("KARA_KEEP_BASE_URL not set")
}

	urlArg, _ :=getString(args, "url")
	if urlArg == "" {
		return err("url argument required")
}

	payload := map[string]string{"url": urlArg}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/bookmarks", baseURL), bytes.NewReader(bodyBytes))
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	return success("Bookmark added")
}