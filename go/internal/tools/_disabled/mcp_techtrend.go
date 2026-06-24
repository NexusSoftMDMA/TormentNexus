package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTrending(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	url := fmt.Sprintf("https://gh-trending-api.herokuapp.com/repositories?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch trending: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("trending repos: %v", data))
}

func HandleGetTrendingByLanguage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lang, _ :=getString(args, "language")
	if lang == "" {
		return err("language is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	url := fmt.Sprintf("https://gh-trending-api.herokuapp.com/repositories?language=%s&limit=%d", lang, limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch trending: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("trending repos in %s: %v", lang, data))
}