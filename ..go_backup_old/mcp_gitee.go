package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func HandleGetGiteeUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://gitee.com/api/v5/users/%s", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("json decode error: %v", e))
}

	login, _ := data["login"].(string)
	name, found := data["name"].(string)
	if !found || name == "" {
		name = login
	}
	return ok(fmt.Sprintf("User: %s (%s)", login, name))
}

func HandleSearchGiteeRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "q")
	if q == "" {
		return err("query 'q' is required")
}

	page, _ :=getInt(args, "page")
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("https://gitee.com/api/v5/search/repositories?q=%s&page=%d", q, page)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("json decode error: %v", e))
}

	itemsRaw, found := result["items"].([]interface{})
	if !found || len(itemsRaw) == 0 {
		return ok("No repositories found")
}

	names := make([]string, 0, len(itemsRaw))
	for _, item := range itemsRaw {
		repo, _ := item.(map[string]interface{})
		if fullName, found := repo["full_name"].(string); found {
			names = append(names, fullName)

	}
	if len(names) == 0 {
		return ok("No repositories found")
}

	total, _ := strconv.Atoi(fmt.Sprintf("%v", result["total_count"]))
	msg := fmt.Sprintf("Found %d repositories (page %d): %s", total, page, names[0:min(5, len(names))])
	return ok(msg)
}

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}