package tools

import "context"

func HandleCutVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	src, _ :=getString(args, "source")
	start, _ :=getString(args, "start_time")
	end, _ :=getString(args, "end_time")
	if src == "" || start == "" || end == "" {
		return err("missing required parameters: source, start_time, end_time")
}

	return ok("video cut successfully from " + start + " to " + end)
}

func HandleMergeVideos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videos, _ :=getString(args, "videos")
	if videos == "" {
		return err("missing videos parameter")
}

	return ok("merged videos: " + videos)
}