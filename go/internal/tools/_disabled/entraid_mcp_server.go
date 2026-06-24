package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/users?$top=10", nil)
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
		return err("API returned status " + fmt.Sprint(resp.StatusCode) + ": " + string(body))
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success("Listed users successfully")
}