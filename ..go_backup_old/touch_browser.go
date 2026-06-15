package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	u, _ :=getString(args, "url")
	if u == "" {
		return err("url is required")
}

	parsed, e := url.Parse(u)
	if e != nil {
		return err("invalid URL")
}

	resp, e := http.DefaultClient.Get(parsed.String())
	if e != nil {
		return err("failed to navigate: " + e.Error())
}

	defer resp.Body.Close()
	return success(fmt.Sprintf("Navigated to %s, status %d", u, resp.StatusCode))
}

func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	if selector == "" {
		return err("selector is required")
}

	return success(fmt.Sprintf("Clicked element: %s (simulated)", selector))
}