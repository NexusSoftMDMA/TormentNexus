package tools

import (
	"context"
)

func HandleGetUserAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userID")
	msg := "Analytics for user " + userID + ": 42 commits, 5 issues, 3 PRs"
	return ok(msg)
}

func HandleGetProjectAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "projectID")
	msg := "Project " + projectID + " metrics: stars: 10, forks: 2, open issues: 1"
	return ok(msg)
}