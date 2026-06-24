package tools

import (
	"context"
	"net/http"
)

func HandlePowerdrillStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch status")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status")
}

	return ok("Powerdrill MCP server is running")
}

func HandlePowerdrillAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action required")
}

	return success("executed: " + action)
}// touch 1781132138
