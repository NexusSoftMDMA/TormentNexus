package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func HandleDNSLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hostname, _ :=getString(args, "hostname")
	if hostname == "" {
		return err("hostname is required")
}

	ips, e := net.LookupHost(hostname)
	if e != nil {
		return err(fmt.Sprintf("DNS lookup failed: %v", e))
}

	data, _ := json.Marshal(ips)
	return ok(string(data))
}

func HandleHTTPGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var body map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&body); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	data, _ := json.Marshal(body)
	return ok(string(data))
}