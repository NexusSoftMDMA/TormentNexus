package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSplitScenes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videoURL, _ :=getString(args, "video_url")
	if videoURL == "" {
		return err("missing video_url")
}

	resp, e := http.DefaultClient.Get(videoURL)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch video: %s", e.Error()))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result struct {
		Scenes []string `json:"scenes"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response")
}

	return success(fmt.Sprintf("split into %d scenes", len(result.Scenes)))
}

func HandleGenerateSubtitles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videoURL, _ :=getString(args, "video_url")
	if videoURL == "" {
		return err("missing video_url")
}

	lang, _ :=getString(args, "language")
	if lang == "" {
		lang = "en"
	}
	resp, e := http.DefaultClient.Get(videoURL + "/subtitles?lang=" + lang)
	if e != nil {
		return err(fmt.Sprintf("failed to generate subtitles: %s", e.Error()))
}

	defer resp.Body.Close()
	return ok("subtitles generated")
}