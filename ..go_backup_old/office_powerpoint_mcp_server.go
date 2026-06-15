package tools

import (
	"context"
)

func HandleCreatePresentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "file_path")
	return ok("Presentation created at " + path)
}

func HandleAddSlide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "file_path")
	title, _ :=getString(args, "title")
	return ok("Slide added to " + path + " with title: " + title)
}