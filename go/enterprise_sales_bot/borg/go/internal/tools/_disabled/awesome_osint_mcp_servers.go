package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

// HandleIpInfo looks up location info for an IP address using ip-api.com
func HandleIpInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")

	resp, e := http.DefaultClient.Get("http://ip-api.com/json/" + ip)
	if e != nil {
		return err("Failed to query ip-api: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to decode JSON: " + e.Error())
}

	return success(fmt.Sprintf("IP %s: %v", ip, result))
}

// HandleDnsLookup resolves a domain name to its IP addresses
func HandleDnsLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")

	ips, e := net.LookupHost(domain)
	if e != nil {
		return err("DNS lookup failed: " + e.Error())
}

	return success(fmt.Sprintf("Domain %s resolves to: %v", domain, ips))
}