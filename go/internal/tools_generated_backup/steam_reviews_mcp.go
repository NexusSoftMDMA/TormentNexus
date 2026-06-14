package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetSteamReviews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appid, _ :=getInt(args, "appid")
	if appid <= 0 {
		return err("appid is required and must be positive")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
	url := fmt.Sprintf("https://store.steampowered.com/appreviews/%d?json=1&num_per_page=%d&language=all", appid, count)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch reviews: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	raw, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return ok(fmt.Sprintf("Steam reviews for appid %d:\n%s", appid, string(raw)))
}