package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleShodanSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing api_key")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
}

	u := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=%s", apiKey, url.QueryEscape(query))
	resp, e := http.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("%v", result))
}

func HandleShodanHost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing api_key")
}

	ip, _ :=getString(args, "ip")
	if ip == "" {
		return err("missing ip")
}

	u := fmt.Sprintf("https://api.shodan.io/shodan/host/%s?key=%s", ip, apiKey)
	resp, e := http.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("%v", result))
}