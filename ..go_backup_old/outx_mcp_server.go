package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
}

	base := "https://api.outx.io/v1/listen"
	reqURL := fmt.Sprintf("%s?q=%s", base, url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

	return success(fmt.Sprintf("OutX data retrieved for: %s", query))
}

func HandleProfileX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("missing required argument: username")
}

	base := "https://api.outx.io/v1/profile"
	reqURL := fmt.Sprintf("%s/%s", base, url.PathEscape(username))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

	return success(fmt.Sprintf("OutX profile retrieved for: %s", username))
}