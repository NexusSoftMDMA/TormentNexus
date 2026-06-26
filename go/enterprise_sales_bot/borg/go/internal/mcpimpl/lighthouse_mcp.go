package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleRunLighthouse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	if target == "" {
		return err("missing required argument: url")
}

	strategy, _ :=getString(args, "strategy")
	if strategy == "" {
		strategy = "mobile"
	}
	apiURL := fmt.Sprintf("https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=%s&strategy=%s", url.QueryEscape(target), strategy)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("failed to call PageSpeed API: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result struct {
		LighthouseResult struct {
			Categories struct {
				Performance struct {
					Score float64 `json:"score"`
				} `json:"performance"`
			} `json:"categories"`
		} `json:"lighthouseResult"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	score := result.LighthouseResult.Categories.Performance.Score * 100
	msg := fmt.Sprintf("Lighthouse performance score for %s (%s): %.0f/100", target, strategy, score)
	return success(msg)
}