package mcpimpl

import (
	"context"
	"net/http"
)

func HandleProcessImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	image, _ :=getString(args, "image")
	action, _ :=getString(args, "action")
	if image == "" {
		return err("image is required")
}

	if action == "" {
		return err("action is required")
}

	resp, e := http.DefaultClient.Get("https://www.photopea.com/api/process?image=" + image + "&action=" + action)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("processing completed, status: " + resp.Status)
}

func HandleGetImageInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	image, _ :=getString(args, "image")
	if image == "" {
		return err("image is required")
}

	resp, e := http.DefaultClient.Get("https://www.photopea.com/api/info?image=" + image)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("image info retrieved, status: " + resp.Status)
}