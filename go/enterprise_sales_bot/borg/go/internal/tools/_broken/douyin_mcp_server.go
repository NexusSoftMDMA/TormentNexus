package tools

import "context"

func HandleGetDouyinVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	videoURL := "https://example.com/no_watermark_video.mp4"
	description := "示例视频文案"
	result := "VideoURL: " + videoURL + "\nDescription: " + description
	return ok(result)
}

func HandleLobster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("龙虾已支持")
}