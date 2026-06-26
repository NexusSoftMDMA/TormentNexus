package tools

import (
	"context"
)

func HandleGetEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Retrieved all planned events")
}

func HandleCreateEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	return ok("Event created: " + title)
}