package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGmailManager(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "accessToken")
	if accessToken == "" {
		return err("accessToken is required")
}

	u, e := url.Parse("https://gmail.googleapis.com/gmail/v1/users/me/labels")
	if e != nil {
		return err("failed to parse URL")
}

	q := u.Query()
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	labels, found := result["labels"].([]interface{})
	if !found {
		return err("no labels returned")
}

	return ok(fmt.Sprintf("Found %d labels", len(labels)))
}