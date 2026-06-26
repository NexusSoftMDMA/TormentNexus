package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func HandleCalculateCarbonFootprint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	activity, _ :=getString(args, "activity")
	activityType, _ :=getString(args, "activity_type")
	country, _ :=getString(args, "country")
	apiKey := os.Getenv("CLIMATIQ_API_KEY")
	if apiKey == "" {
		return err("CLIMATIQ_API_KEY not set")
}

	u, _ := url.Parse("https://api.climatiq.io/data/v1/estimate")
	q := url.Values{}
	q.Set("activity", activity)
	q.Set("activity_type", activityType)
	if country != "" {
		q.Set("country", country)

	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(result)
}
}