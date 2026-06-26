package mcpimpl

import (
    "context"
)

func HandleCheckDomain_muumuu_domain_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    domain, _ :=getString(args, "domain")
    return ok("Domain checked: " + domain)
}

func HandleGetDomainInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    domain, _ :=getString(args, "domain")
    return ok("Domain info for: " + domain)
}