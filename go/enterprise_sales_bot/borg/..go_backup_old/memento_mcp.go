package tools

import "context"

func HandleCaptureMemento(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	return success("Captured memento for " + url)
}