package tools

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func HandleGetDeepWiki(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	text := removeHTMLTags(string(body))
	return ok(text)
}

func removeHTMLTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")
	return strings.TrimSpace(text)
}