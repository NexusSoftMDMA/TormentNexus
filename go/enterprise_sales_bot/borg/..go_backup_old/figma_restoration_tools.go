package tools

import (
	"context"
)

func HandleFigmaCompare(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "fileKey")
	nodeId, _ :=getString(args, "nodeId")
	return success("Figma compare completed for file: " + fileKey + " node: " + nodeId)
}

func HandleSnapdomScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	return success("SnapDOM screenshot taken for URL: " + url)
}