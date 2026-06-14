package tools

import (
    "context"
    "net"
    "strings"
)

func HandleRecon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    domain, _ :=getString(args, "domain")
    if domain == "" {
        return err("domain is required")
    }
    ips, e := net.LookupHost(domain)
    if e != nil {
        return err("lookup failed: " + e.Error())
    }
    return ok("Resolved " + domain + " to " + strings.Join(ips, ", "))
}