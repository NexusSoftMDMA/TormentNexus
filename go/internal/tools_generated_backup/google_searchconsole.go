package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListSites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/webmasters/v3/sites", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result struct {
		SiteEntry []struct {
			SiteURL string `json:"siteUrl"`
		} `json:"siteEntry"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	urls := make([]string, len(result.SiteEntry))
	for i, s := range result.SiteEntry {
		urls[i] = s.SiteURL
	}
	return ok(fmt.Sprintf("Sites: %v", urls))
}