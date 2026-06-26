package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleGetVideoDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videoId, _ :=getString(args, "videoId")
	apiKey, _ :=getString(args, "apiKey")
	if videoId == "" || apiKey == "" {
		return err("missing required parameters")
}

	u, _ := url.Parse("https://www.googleapis.com/youtube/v3/videos")
	u.RawQuery = url.Values{"id": {videoId}, "part": {"snippet"}, "key": {apiKey}}.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Items []struct {
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"snippet"`
		} `json:"items"`
	}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed: " + e.Error())
}

	if len(result.Items) == 0 {
		return err("video not found")
}

	return ok("Title: " + result.Items[0].Snippet.Title + ", Description: " + result.Items[0].Snippet.Description)
}

func HandleGetChannelStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channelId, _ :=getString(args, "channelId")
	apiKey, _ :=getString(args, "apiKey")
	if channelId == "" || apiKey == "" {
		return err("missing required parameters")
}

	u, _ := url.Parse("https://www.googleapis.com/youtube/v3/channels")
	u.RawQuery = url.Values{"id": {channelId}, "part": {"statistics"}, "key": {apiKey}}.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Items []struct {
			Statistics struct {
				SubscriberCount string `json:"subscriberCount"`
				ViewCount       string `json:"viewCount"`
				VideoCount      string `json:"videoCount"`
			} `json:"statistics"`
		} `json:"items"`
	}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed: " + e.Error())
}

	if len(result.Items) == 0 {
		return err("channel not found")
}

	stats := result.Items[0].Statistics
	return ok("Subscribers: " + stats.SubscriberCount + ", Views: " + stats.ViewCount + ", Videos: " + stats.VideoCount)
}