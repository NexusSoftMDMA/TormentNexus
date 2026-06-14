package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetWorkouts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	page, _ :=getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	url := fmt.Sprintf("https://api.hevyapp.com/v1/workouts?limit=%d&page=%d", limit, page)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Key "+apiKey)
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
		return err("API returned status " + resp.Status + ": " + string(body))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	jsonStr, _ := json.MarshalIndent(data, "", "  ")
	return success(string(jsonStr))
}