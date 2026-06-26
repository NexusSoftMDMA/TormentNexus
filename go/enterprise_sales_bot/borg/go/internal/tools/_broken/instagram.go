package tools'.package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// -------------------------------------------------------------------
// Helper: safe extraction from nested map[string]interface{}
// -------------------------------------------------------------------

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, found := m[key]; found {
		if s, found := v.(string); found {
			return s
		}
	}
	return ""
}

func getIntFromNested(m map[string]interface{}, keys ...string) int {
	current := m
	for i, k := range keys {
		if i == len(keys)-1 {
			if v, found := current[k]; found {
				switch n := v.(type) {
				case float64:
					return int(n)
}
				case int:
					return n
}
				case int64:
					return int(n)
}
				case json.Number:
					if i, e := n.Int64(); e == nil {
						return int(i)

				}
			}
			return 0
		}
		if next, found := current[k].(map[string]interface{}); found {
			current = next
		} else {
			return 0
		}
	}
	return 0
}

// -------------------------------------------------------------------
// HandleGetUserInfo – fetch public Instagram user info via __a=1 API
// -------------------------------------------------------------------
func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")

	apiURL := fmt.Sprintf("https://www.instagram.com/%s/?__a=1", username)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %s", reqErr.Error()))
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Instagram returned HTTP %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %s", readErr.Error()))
}

	var raw map[string]interface{}
	if parseErr := json.Unmarshal(body, &raw); parseErr != nil {
		return err(fmt.Sprintf("JSON parse error: %s", parseErr.Error()))
}

	graphql, found := raw["graphql"].(map[string]interface{})
	if !found {
		return err("unexpected response: missing 'graphql'")
}

	user, found := graphql["user"].(map[string]interface{})
	if !found {
		return err("unexpected response: missing 'user'")
}

	userName := getStringFromMap(user, "username")
	fullName := getStringFromMap(user, "full_name")
	bio := getStringFromMap(user, "biography")
	pic := getStringFromMap(user, "profile_pic_url_hd")
	followers := getIntFromNested(user, "edge_followed_by", "count")
	following := getIntFromNested(user, "edge_follow", "count")

	result := fmt.Sprintf("Username: %s\nFull Name: %s\nBio: %s\nProfile Pic: %s\nFollowers: %d\nFollowing: %d",
		userName, fullName, bio, pic, followers, following)

	return ok(result)
}