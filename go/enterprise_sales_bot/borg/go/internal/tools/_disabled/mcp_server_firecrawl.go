package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	apiKey, _ :=getString(args, "apiKey")

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.firecrawlcrawl.com/v1/scrape?url="+url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %s", resp.Status))
}

	return ok(string(body))
}
}