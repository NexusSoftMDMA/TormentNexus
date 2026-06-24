package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchDomains(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.ctscout.dev/v1/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	return ok(fmt.Sprintf("Found %v", result))
}

func HandleGetDomainDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	url := fmt.Sprintf("https://api.ctscout.dev/v1/domain/%s", domain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	return ok(fmt.Sprintf("Details: %v", result))
}