package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreatePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	body := map[string]string{"title": title, "content": content}
	b, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post("https://api.blogburst.com/posts", "application/json", io.NopCloser(io.Reader(nil)))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("post created successfully")
}

func HandleListPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.blogburst.com/posts")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	raw, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	return success(string(raw))
}