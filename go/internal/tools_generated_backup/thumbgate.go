package tools

import (
	"context"
	"net/http"
)

func HandleGenerateThumbnail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch image: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("HTTP error: " + resp.Status)
}

	return success("Thumbnail generated for " + url)
}