package tools

import (
	"context"
	"io"
	"net/http"
	"regexp"
)

func HandleReadWebsiteFast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("URL is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	// crude HTML-to-Markdown: remove tags, collapse whitespace
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(string(body), "")
	return success(text)
}// touch 1781132128
