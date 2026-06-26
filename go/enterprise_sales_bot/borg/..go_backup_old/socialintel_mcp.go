package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("https://api.socialintel.example.com/posts?q=" + query)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("Found %d posts for %s", int(data["count"].(float64)), query))
}

func HandleGetUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	resp, e := http.DefaultClient.Get("https://api.socialintel.example.com/user/" + username)
	if e != nil {
		return err("failed to get user: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	name, _ := data["name"].(string)
	return ok(fmt.Sprintf("User %s has name: %s", username, name))
}