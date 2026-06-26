package mcpimpl

import (
	"context"
)

// HandleCheckFirewall checks if an IP is allowed by the firewall.
func HandleCheckFirewall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	if ip == "" {
		return err("missing ip")
}

	// In a real implementation, query the firewall. Here we just allow all.
	return ok("IP " + ip + " is allowed")
}

// HandleBlockFirewall blocks an IP in the firewall.
func HandleBlockFirewall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	if ip == "" {
		return err("missing ip")
}

	// In a real implementation, add a rule to block. Here we simulate.
	return ok("IP " + ip + " has been blocked")
}