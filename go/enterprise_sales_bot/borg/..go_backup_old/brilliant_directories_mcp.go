package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGetMember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteURL, _ :=getString(args, "site_url")
	apiKey, _ :=getString(args, "api_key")
	memberID, _ :=getString(args, "member_id")
	if siteURL == "" || apiKey == "" || memberID == "" {
		return err("Missing required parameters: site_url, api_key, member_id")
}

	url := fmt.Sprintf("%s/api/member/%s?api_key=%s", strings.TrimRight(siteURL, "/"), memberID, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("Decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Member data: %v", result))
}

func HandleCreatePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteURL, _ :=getString(args, "site_url")
	apiKey, _ :=getString(args, "api_key")
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if siteURL == "" || apiKey == "" || title == "" || content == "" {
		return err("Missing required parameters: site_url, api_key, title, content")
}

	payload := map[string]string{"title": title, "content": content}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/posts?api_key=%s", strings.TrimRight(siteURL, "/"), apiKey)
	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("Decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Post created: %v", result))
}