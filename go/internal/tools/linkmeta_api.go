//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

func HandleGetLinkmeta(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response body: %v", e))
}

	title := ""
	titleRe := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := titleRe.FindStringSubmatch(string(body))
	if len(matches) >= 2 {

	}

	description := ""
	descRe := regexp.MustCompile(`<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	matches = descRe.FindStringSubmatch(string(body))
	if len(matches) >= 2 {

		return ok(fmt.Sprintf("Title: %s\nDescription: %s", title, description))
}
}