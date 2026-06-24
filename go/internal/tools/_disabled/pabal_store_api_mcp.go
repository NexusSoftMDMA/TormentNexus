package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetAppReviews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appId, _ :=getString(args, "appId")
	url := fmt.Sprintf("https://itunes.apple.com/rss/customerreviews/id=%s/sortBy=mostRecent/json", appId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch reviews: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetAppRatings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appId, _ :=getString(args, "appId")
	url := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s", appId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch ratings: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}