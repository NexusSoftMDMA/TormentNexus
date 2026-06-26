package tools

import "context"

func HandleSshRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	cmd, _ :=getString(args, "command")
	return ok("Ran: " + cmd + " on " + host)
}

func HandleSshCheckHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	return ok("Health check passed for " + host)
}