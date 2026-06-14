package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetChainlinkFeeds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://reference-data-directory.vercel.app/feeds.json")
	if e != nil {
		return err("failed to get feeds: " + e.Error())
}

	defer resp.Body.Close()
	var feeds []interface{}
	e = json.NewDecoder(resp.Body).Decode(&feeds)
	if e != nil {
		return err("failed to decode: " + e.Error())
}

	data, e := json.Marshal(feeds)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	return ok(string(data))
}