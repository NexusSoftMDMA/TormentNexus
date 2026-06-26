package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleQueryData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", url+"?query="+query, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("Data retrieved")
}

func HandleListDataSources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", url+"/datasources", nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("Data sources listed")
}