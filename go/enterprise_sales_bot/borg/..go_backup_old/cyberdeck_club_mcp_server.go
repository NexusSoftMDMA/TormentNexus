package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchBuilds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://cyberdeck.club/api/builds?search=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch builds")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response")
	}
	return success(fmt.Sprintf("Found %d builds", int(result["count"].(float64))))
}

func HandleGetWiki(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	page, _ :=getString(args, "page")
	url := fmt.Sprintf("https://cyberdeck.club/api/wiki/%s", page)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch wiki page")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	return ok(string(body))
}