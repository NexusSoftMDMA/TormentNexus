package tools

import "context"

func HandleListEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	return ok(`{"emails":[{"id":"1","subject":"Hello"}]}`)
}

func HandleGetEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return ok(`{"id":"` + id + `","subject":"Email ` + id + `"}`)
}