package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	status, _ :=getString(args, "status")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 20
	}
	url := fmt.Sprintf("https://api.opsgenie.com/v2/alerts?limit=%d&status=%s", limit, status)
	if status == "" {
		url = fmt.Sprintf("https://api.opsgenie.com/v2/alerts?limit=%d", limit)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "GenieKey "+os.Getenv("OPSGENIE_API_KEY"))
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
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d alerts", len(result["data"].([]interface{}))))
}

}

func HandleGetAlert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required 'id' argument")
}

	url := "https://api.opsgenie.com/v2/alerts/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "GenieKey "+os.Getenv("OPSGENIE_API_KEY"))
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
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Alert details: %s", string(body)))
}