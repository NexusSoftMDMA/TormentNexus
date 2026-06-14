package tools

import (
	"context"
)

var approvals map[string]bool

func init() {
	approvals = make(map[string]bool)

}

func HandleRequestApproval(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	reason, _ :=getString(args, "reason")
	if action == "" {
		return err("action is required")
}

	approvals[action] = false
	_ = reason
	return ok("Approval requested for " + action)
}

func HandleCheckApproval(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action is required")
}

	approved, found := approvals[action]
	if !found {
		return err("no approval request for " + action)
}

	if approved {
		return ok("Approved")
}

	return ok("Pending")
}