package tools

import "context"

func HandleGetAgenda(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Current agenda: 1. Review project, 2. Team meeting")
}

func HandleAddAgendaItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	item, _ :=getString(args, "item")
	return success("Added agenda item: " + item)
}