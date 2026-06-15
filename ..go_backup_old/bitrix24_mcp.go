package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	userID, _ :=getInt(args, "user_id")
	if apiURL == "" || userID == 0 {
		return err("api_url and user_id are required")
}

	url := fmt.Sprintf("%s/user.get?ID=%d", apiURL, userID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("User: %v", result))
}