package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	bearerToken, _ :=getString(args, "bearer_token")
	if username == "" || bearerToken == "" {
		return err("username and bearer_token are required")
}

	url := fmt.Sprintf("https://api.twitter.com/2/users/by/username/%s/tweets", username)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+bearerToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("json decode error: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}