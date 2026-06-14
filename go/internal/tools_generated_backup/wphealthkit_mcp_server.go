package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleAuditPlugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	slug, _ :=getString(args, "slug")
	if slug == "" {
		return err("slug is required")
}

	url := fmt.Sprintf("https://api.wordpress.org/plugins/info/1.0/%s.json", slug)
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch plugin info")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleGetGuide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	url := fmt.Sprintf("https://developer.wordpress.org/reference/%s/", topic)
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch guide")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}