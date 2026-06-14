package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchGovData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	apiURL := fmt.Sprintf("https://api.govdata.gov/search?q=%s", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
	}
	return ok("found " + fmt.Sprint(len(result)) + " results")
}

func HandleGetGovRider(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	state, _ :=getString(args, "state")
	if state == "" {
		return err("state is required")
	}
	apiURL := fmt.Sprintf("https://api.govrider.gov/data?state=%s", url.QueryEscape(state))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
	}
	return success(fmt.Sprintf("Got data for %s: %v", state, data))
}