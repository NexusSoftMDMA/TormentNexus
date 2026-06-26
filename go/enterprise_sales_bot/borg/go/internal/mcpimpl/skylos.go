package mcpimpl

import (
	"context"
)

func HandleGetSkyStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return success("Sky status: clear, visibility excellent")
}

	return success("Sky status at " + location + ": partly cloudy")
}

func HandleGetStarInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	star, _ :=getString(args, "star")
	if star == "" {
		return err("star name is required")
}

	return success("Star " + star + " is a main-sequence star in the Milky Way")
}