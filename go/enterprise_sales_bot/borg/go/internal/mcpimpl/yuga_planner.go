package mcpimpl

import (
	"context"
)

func HandleGetEvents_yuga_planner(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Retrieved all planned events")
}

func HandleCreateEvent_yuga_planner(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	return ok("Event created: " + title)
}