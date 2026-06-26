package mcpimpl

import "context"

func HandleGetProfile_mcp_linkedin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("missing username")
}

	return ok(`{"name":"John Doe","headline":"Software Engineer"}`)
}

func HandlePostShare(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("missing content")
}

	return success("Share posted")
}