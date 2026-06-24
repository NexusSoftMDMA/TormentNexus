package tools

import "context"

func HandleSshList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	return ok("SSH list for host: " + host)
}

func HandleSshExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	cmd, _ :=getString(args, "command")
	return success("Executing on " + host + ": " + cmd)
}