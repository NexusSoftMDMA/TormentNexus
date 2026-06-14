package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	doctype, _ :=getString(args, "doctype")
	apiURL, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	url := fmt.Sprintf("%s/api/resource/%s?fields=[\"name\",\"title\"]&limit_start=0&limit_page_length=10", apiURL, doctype)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleGetDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	doctype, _ :=getString(args, "doctype")
	name, _ :=getString(args, "name")
	apiURL, _ :=getString(args, "api_url")
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	url := fmt.Sprintf("%s/api/resource/%s/%s", apiURL, doctype, name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}