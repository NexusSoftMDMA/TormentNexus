package tools

import (
    "context"
)

func HandleCheckDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    domain, _ :=getString(args, "domain")
    return ok("Domain checked: " + domain)
}

func HandleGetDomainInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    domain, _ :=getString(args, "domain")
    return ok("Domain info for: " + domain)
}