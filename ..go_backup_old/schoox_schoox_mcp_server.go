package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListCourses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.schoox.com/v1/courses", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}