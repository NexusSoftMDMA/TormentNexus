package mcpimpl

import (
	"context"
)

func HandleGetApplicationStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID, _ :=getString(args, "applicationId")
	return ok("Application " + appID + " status: pending")
}

func HandleSubmitApplication(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	email, _ :=getString(args, "email")
	return success("Application submitted for " + name + " (" + email + ")")
}