package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSearchDashboards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	grafanaURL := os.Getenv("GRAFANA_URL")
	apiKey := os.Getenv("GRAFANA_API_KEY")
	if grafanaURL == "" || apiKey == "" {
		return err("GRAFANA_URL and GRAFANA_API_KEY must be set")
}

	url := grafanaURL + "/api/search?type=dash-db"
	if query != "" {
		url += "&query=" + query
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return err("API error: " + resp.Status + " - " + string(body))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok("dashboards: " + fmt.Sprintf("%v", result))
}