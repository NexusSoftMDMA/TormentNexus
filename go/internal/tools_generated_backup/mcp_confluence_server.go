package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID, _ :=getString(args, "pageId")
	if pageID == "" {
		return err("pageId is required")
}

	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		baseURL = "https://your-domain.atlassian.net/wiki"
	}
	token, _ :=getString(args, "token")
	url := fmt.Sprintf("%s/rest/api/content/%s?expand=body.storage", baseURL, pageID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer " + token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("do request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse JSON: %v", e))
}

	title := result["title"].(string)
	return success(fmt.Sprintf("Page '%s' retrieved successfully", title))
}
}