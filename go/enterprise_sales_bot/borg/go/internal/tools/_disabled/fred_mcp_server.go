package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleFredGetObservations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	seriesID, _ :=getString(args, "series_id")
	if seriesID == "" {
		return err("series_id is required")
}

	apiKey := os.Getenv("FRED_API_KEY")
	if apiKey == "" {
		return err("FRED_API_KEY environment variable not set")
}

	url := fmt.Sprintf("https://api.stlouisfed.org/fred/series/observations?series_id=%s&api_key=%s&file_type=json", seriesID, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP status %d", resp.StatusCode))
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	observations, found := result["observations"]
	if !found {
		return err("no observations in response")
}

	return ok(fmt.Sprintf("Got %d observations", len(observations.([]interface{}))))
}