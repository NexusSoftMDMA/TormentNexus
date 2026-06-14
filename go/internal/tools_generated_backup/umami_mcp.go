package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetPageViews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	startAt, _ :=getString(args, "startAt")
	endAt, _ :=getString(args, "endAt")
	apiKey := os.Getenv("UMAMI_API_KEY")
	apiBase := os.Getenv("UMAMI_BASE_URL")
	if apiKey == "" || apiBase == "" {
		return err("UMAMI_API_KEY or UMAMI_BASE_URL not set")
}

	reqURL := fmt.Sprintf("%s/api/websites/%s/pageviews?startAt=%s&endAt=%s", apiBase, url, startAt, endAt)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err("failed creating request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("reading body failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Page views data: %v", data))
}