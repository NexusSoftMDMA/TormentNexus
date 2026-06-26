package mcpimpl

import (
	"context"
	"fmt"
	"net"
)

func HandleLookupDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain parameter is required")
}

	ips, e := net.LookupHost(domain)
	if e != nil {
		return err("lookup failed: " + e.Error())
}

	return ok(fmt.Sprintf("IP addresses for %s: %v", domain, ips))
}