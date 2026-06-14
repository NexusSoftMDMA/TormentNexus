package tools

import (
	"context"
)

func HandleCreateJam(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	genre, _ :=getString(args, "genre")
	_ = genre
	return success("Jam created: " + name)
}

func HandleListJams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	page, _ :=getInt(args, "page")
	_ = page
	return success("List of jams (page " + string(rune(page+'0')) + ")")
}