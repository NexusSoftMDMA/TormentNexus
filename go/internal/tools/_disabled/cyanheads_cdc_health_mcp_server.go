package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchCDC(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query parameter")
	}

	base := "https://data.cdc.gov/resource/nn67-46kw.json"
	params := url.Values{}
	params.Set("$search", query)
	params.Set("$limit", "10")

	reqURL := base + "?" + params.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}

	return ok(fmt.Sprintf("Found CDC data for: %s", query))
}

func HandleGetVaccinations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	state, _ :=getString(args, "state")
	if state == "" {
		return err("missing state parameter")
	}

	base := "https://data.cdc.gov/resource/unsk-b7fc.json"
	params := url.Values{}
	params.Set("state", state)
	params.Set("$limit", "5")

	reqURL := base + "?" + params.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}

	return ok(fmt.Sprintf("Retrieved vaccination data for %s", state))
}// touch 1781132124
