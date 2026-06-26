package tools

import "context"

func HandleOpencvProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imagePath, _ :=getString(args, "image_path")
	if imagePath == "" {
		return err("No image path provided")
}

	result := "Processed image: " + imagePath
	return ok(result)
}

func HandleOpencvInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version := "4.5.5"
	return ok("OpenCV version: " + version)
}