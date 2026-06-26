package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleGetUser_cinderwright_api(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	if userID == "" {
		return err("user_id is required")
}

	url := fmt.Sprintf("https://api.cinderwright.com/users/%s", userID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON")
}

	return ok(fmt.Sprintf("User data: %v", data))
}

func HandleListPosts_cinderwright_api(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.cinderwright.com/posts?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var data []interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON")
}

	return ok(fmt.Sprintf("Posts: %v", data))
}