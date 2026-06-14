package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleYoutubeExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	videoURL, _ :=getString(args, "url")
	if videoURL == "" {
		return err("missing 'url' argument")
}

	query := url.Values{}
	query.Set("url", videoURL)
	query.Set("format", "json")
	apiURL := "https://www.youtube.com/oembed?" + query.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("oEmbed API returned status %d", resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	result := map[string]interface{}{
		"title":      data["title"],
		"author":     data["author_name"],
		"author_url": data["author_url"],
		"thumbnail":  data["thumbnail_url"],
		"html":       data["html"],
	}
	return success(fmt.Sprintf("Extracted YouTube info: %v", data["title"]), result)
}