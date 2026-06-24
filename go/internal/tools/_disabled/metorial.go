package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id <= 0 {
		return err("invalid id")
}

	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	title, found := data["title"].(string)
	if !found {
		return err("title not found")
}

	return success(title)
}

func HandleListPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://jsonplaceholder.typicode.com/posts"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var posts []map[string]interface{}
	if e := json.Unmarshal(body, &posts); e != nil {
		return err("parse failed: " + e.Error())
}

	count := len(posts)
	return ok(fmt.Sprintf("Found %d posts", count))
}