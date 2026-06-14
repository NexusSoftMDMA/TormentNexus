package tools

import (
	"context"
	"fmt"
)

func HandleGetCurrentSeason(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Current season in Japan: Spring")
}

func HandleGetSeasonInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	season, _ :=getString(args, "season")
	if season == "" {
		return err("season parameter is required")
}

	info := fmt.Sprintf("Season: %s – cherry blossoms bloom in spring, foliage in autumn", season)
	return ok(info)
}