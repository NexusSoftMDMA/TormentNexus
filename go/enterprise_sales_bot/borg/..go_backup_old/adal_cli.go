package tools

import (
	"context"
	"fmt"
)

func HandleAdalLogin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tenant, _ :=getString(args, "tenant")
	clientID, _ :=getString(args, "clientId")
	clientSecret, _ :=getString(args, "clientSecret")
	if tenant == "" || clientID == "" || clientSecret == "" {
		return err("tenant, clientId, and clientSecret are required")
	}
	return ok(fmt.Sprintf("Logged in to tenant %s with client %s", tenant, clientID))
}

func HandleAdalGetToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resource, _ :=getString(args, "resource")
	if resource == "" {
		return err("resource is required")
	}
	return success(fmt.Sprintf("Access token obtained for resource: %s (simulated)", resource))
}