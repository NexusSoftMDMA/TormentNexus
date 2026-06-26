package mcpimpl

import "context"

var tokens = map[string]string{}

func HandleSaveToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	token, _ :=getString(args, "token")
	if key == "" || token == "" {
		return err("key and token are required")
}

	tokens[key] = token
	return ok("token saved")
}

func HandleGetToken_token_saver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	val, found := tokens[key]
	if !found {
		return err("token not found")
}

	return success(val)
}