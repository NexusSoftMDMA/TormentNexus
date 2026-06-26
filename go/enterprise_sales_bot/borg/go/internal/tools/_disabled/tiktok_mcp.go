package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := "https://api.tiktok.com/user/info/?username=" + username
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call TikTok API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(string(body))
}

func HandleSearchVideos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	url := "https://api.tiktok.com/search/video/?keyword=" + keyword
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}