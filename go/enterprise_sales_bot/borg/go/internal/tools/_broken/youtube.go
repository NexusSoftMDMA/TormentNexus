package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	youtubeVideoRegex = regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})`)")
)

func HandleYoutubeSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&key=AIzaSyD3JQ5J5J5J5J5J5J5J5J5J5J5J5J5J5J5", url.QueryEscape(query))
	resp, fetchErr := client.Get(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
		} `json:"items"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Items) == 0 {
		return err("no results found")
}

	var videos []string
	for _, item := range result.Items {
		videos = append(videos, fmt.Sprintf("%s - %s", item.Snippet.Title, item.ID.VideoID))

	return ok(strings.Join(videos, "\n"))
}

}

func HandleYoutubeInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videoURL, _ :=getString(args, "url")
	if videoURL == "" {
		return err("url parameter is required")
}

	videoID := extractYoutubeID(videoURL)
	if videoID == "" {
		return err("invalid YouTube URL")
}

	client := http.Client{Timeout: 30 * time.Second}
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=%s&key=AIzaSyD3JQ5J5J5J5J5J5J5J5J5J5J5J5J5J5J5", videoID)
	resp, fetchErr := client.Get(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var result struct {
		Items []struct {
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount    string `json:"viewCount"`
				LikeCount    string `json:"likeCount"`
				CommentCount string `json:"commentCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Items) == 0 {
		return err("video not found")
}

	item := result.Items[0]
	info := fmt.Sprintf("Title: %s\nDescription: %s\nViews: %s\nLikes: %s\nComments: %s",
		item.Snippet.Title,
		item.Snippet.Description,
		item.Statistics.ViewCount,
		item.Statistics.LikeCount,
		item.Statistics.CommentCount)

	return ok(info)
}

func extractYoutubeID(videoURL string) string {
	matches := youtubeVideoRegex.FindStringSubmatch(videoURL)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func HandleYoutubeTranscript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videoURL, _ :=getString(args, "url")
	if videoURL == "" {
		return err("url parameter is required")
}

	videoID := extractYoutubeID(videoURL)
	if videoID == "" {
		return err("invalid YouTube URL")
}

	client := http.Client{Timeout: 30 * time.Second}
	apiURL := fmt.Sprintf("https://www.youtube.com/api/timedtext?lang=en&type=caption&v=%s", videoID)
	resp, fetchErr := client.Get(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}