package tools

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"
)

// HandlePing performs a basic ICMP ping using the system ping command.
func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host parameter is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 4
	}
	cmd := exec.CommandContext(ctx, "ping", "-c", fmt.Sprintf("%d", count), host)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("ping failed: %v", e))
}

	return ok(fmt.Sprintf("Ping result for %s:\n%s", host, string(out)))
}

// HandleDNSlookup resolves a hostname to IP addresses.
func HandleDNSlookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host parameter is required")
}

	ips, e := net.LookupHost(host)
	if e != nil {
		return err(fmt.Sprintf("DNS lookup failed: %v", e))
}

	if len(ips) == 0 {
		return ok(fmt.Sprintf("No IP addresses found for %s", host))
}

	out := fmt.Sprintf("IP addresses for %s:\n", host)
	for _, ip := range ips {
		out += fmt.Sprintf("  %s\n", ip)

	return ok(out)
}
}