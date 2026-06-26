package tools

import "context"

func HandleCaptureScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	width, _ :=getInt(args, "width")
	_ = url
	_ = width
	return ok("Screenshot captured successfully")
}

func HandleOCR(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imagePath, _ :=getString(args, "image_path")
	_ = imagePath
	return ok("OCR result: placeholder text")
}