package tools

import "context"

func HandleCheckPermission(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	resource, _ :=getString(args, "resource")
	return ok("Permission granted for user " + user + " on " + resource)
}

func HandleGetUserRoles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	return ok("Roles for user " + user + ": admin, readonly")
}