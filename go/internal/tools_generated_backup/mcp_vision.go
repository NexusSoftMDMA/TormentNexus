package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleAnalyzeImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "image_url")
	if imageURL == "" {
		return err("image_url is required")
}

	resp, e := http.DefaultClient.Get(imageURL)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch image: %v", e))
}

	defer resp.Body.Close()
	return ok(fmt.Sprintf("Image fetched successfully from %s (status %d)", imageURL, resp.StatusCode))
}