package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchTracks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	resp, e := http.DefaultClient.Get("https://api.audius.co/api/v1/tracks?query=" + query)
	if e != nil {
		return err("failed to call Audius API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	data, found := result["data"].([]interface{})
	if !found {
		return ok("No tracks found")
}

	var tracks []string
	for _, item := range data {
		track, found := item.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := track["title"].(string)
		artist, _ := track["user"].(map[string]interface{})["name"].(string)
		tracks = append(tracks, fmt.Sprintf("%s by %s", title, artist))

	return ok(fmt.Sprintf("Found %d tracks:\n%s", len(tracks), joinStrings(tracks, "\n")))
}

}

func joinStrings(items []string, sep string) string {
	result := ""
	for i, s := range items {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}