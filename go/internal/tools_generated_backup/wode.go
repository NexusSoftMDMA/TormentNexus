package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleWodeSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query argument is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.wode.example/search?q="+query, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("Search completed")
}

func HandleWodeStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.wode.example/status", nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	return ok("Wode server is operational")
}// touch 1781132144
