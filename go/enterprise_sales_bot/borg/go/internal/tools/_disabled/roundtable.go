package tools

import "context"

func HandleCreateRoundTable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	return ok("Roundtable discussion created: " + topic)
}

func HandleListParticipants(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	output := "Listing participants (limit: %d)"
	return ok(fmt.Sprintf(output, limit))
}