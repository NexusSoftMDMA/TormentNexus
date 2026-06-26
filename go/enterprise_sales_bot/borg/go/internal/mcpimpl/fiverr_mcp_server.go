package mcpimpl

import (
	"context"
	"net/http"
	"encoding/json"
	"io"
)

func HandleSearchGigs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := "https://api.fiverr.com/v1/gigs/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response")
}

	return success(string(body))
}

func HandleGetGig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("gig id is required")
}

	url := "https://api.fiverr.com/v1/gigs/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch gig: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}