package tools

import (
	"context"
	"fmt"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s! From Agent Hellboy.", name))
}

func HandleGetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		return err("repo parameter is required")
}

	return success(fmt.Sprintf("Repository '%s' is a great choice!", repo))
}