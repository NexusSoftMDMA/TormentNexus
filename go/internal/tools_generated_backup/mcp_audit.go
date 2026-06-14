package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAuditSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	limit, _ :=getInt(args, "limit")
	if limit > 0 {
		url = fmt.Sprintf("%s?limit=%d", url, limit)

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return success(fmt.Sprintf("audit search results: %v", result))
}

}

func HandleAuditGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	id, _ :=getString(args, "id")
	if url == "" || id == "" {
		return err("url and id are required")
}

	fullURL := fmt.Sprintf("%s/%s", url, id)
	resp, e := http.DefaultClient.Get(fullURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return success(fmt.Sprintf("audit entry: %v", result))
}