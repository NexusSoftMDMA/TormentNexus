package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	if target == "" {
		return err("missing required parameter: url")
}

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}
	req, e := http.NewRequestWithContext(ctx, "GET", target, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("scraped %d bytes from %s", len(body), target))
}

func HandleListScrapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	scrapers := []string{"html", "json", "text", "markdown"}
	data, e := json.Marshal(scrapers)
	if e != nil {
		return err("failed to marshal scrapers: " + e.Error())
}

	return success(string(data))
}