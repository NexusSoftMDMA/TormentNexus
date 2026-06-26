package mcpimpl

import (
	"context"
)

// HandleGetMeme returns a sample meme response.
func HandleGetMeme_mengram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("You requested meme: " + query)
}