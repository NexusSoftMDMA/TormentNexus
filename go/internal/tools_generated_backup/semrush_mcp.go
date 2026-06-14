package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleDomainAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	domain, _ :=getString(args, "domain")
	if apiKey == "" || domain == "" {
		return err("api_key and domain are required")
}

	u := fmt.Sprintf("https://api.semrush.com/analytics/v1/?key=%s&type=domain_organic&domain=%s", url.QueryEscape(apiKey), url.QueryEscape(domain))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return success(string(body))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func HandleKeywordOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	keyword, _ :=getString(args, "keyword")
	if apiKey == "" || keyword == "" {
		return err("api_key and keyword are required")
}

	u := fmt.Sprintf("https://api.semrush.com/analytics/v1/?key=%s&type=phrase_all&phrase=%s&export_columns=Ph,Nq,Cp,Co,Nr,Td", url.QueryEscape(apiKey), url.QueryEscape(keyword))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}