package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleAddCompetitor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("competitor name is required")
}

	url, _ :=getString(args, "url")
	payload := fmt.Sprintf(`{"name":"%s","url":"%s"}`, name, url)
	resp, e := http.DefaultClient.Post("https://api.example.com/competitors", "application/json", strings.NewReader(payload))
	if e != nil {
		return err("failed to add competitor: " + e.Error())
}

	defer resp.Body.Close()
	return ok("competitor " + name + " added successfully")
}

func HandleGetCompetitorInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("competitor name is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/competitors/" + name)
	if e != nil {
		return err("failed to get competitor info: " + e.Error())
}

	defer resp.Body.Close()
	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	return ok(fmt.Sprintf("competitor info: %v", info))
}

func HandleTrackMention(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	mention, _ :=getString(args, "mention")
	payload := fmt.Sprintf(`{"keyword":"%s","mention":"%s"}`, keyword, mention)
	resp, e := http.DefaultClient.Post("https://api.example.com/mentions", "application/json", strings.NewReader(payload))
	if e != nil {
		return err("failed to track mention: " + e.Error())
}

	defer resp.Body.Close()
	return ok("mention tracked for keyword: " + keyword)
}