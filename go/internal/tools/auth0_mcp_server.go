//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	accessToken, _ :=getString(args, "access_token")
	userID, _ :=getString(args, "user_id")
	if domain == "" || accessToken == "" || userID == "" {
		return err("missing required args: domain, access_token, user_id")
}

	url := fmt.Sprintf("https://%s/api/v2/users/%s", domain, userID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	return ok("User info retrieved successfully: " + string(pretty))
}
