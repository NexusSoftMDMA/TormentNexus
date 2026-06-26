package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGetAnalyticsReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	propertyID, _ :=getString(args, "property_id")
	accessToken, _ :=getString(args, "access_token")
	startDate, _ :=getString(args, "start_date")
	endDate, _ :=getString(args, "end_date")
	if propertyID == "" || accessToken == "" {
		return err("property_id and access_token are required")
}

	if startDate == "" {
		startDate = "7daysAgo"
	}
	if endDate == "" {
		endDate = "today"
	}
	url := fmt.Sprintf("https://analyticsdata.googleapis.com/v1beta/properties/%s:runReport", propertyID)
	body := fmt.Sprintf(`{"dateRanges":[{"startDate":"%s","endDate":"%s"}],"metrics":[{"name":"activeUsers"}]}`, startDate, endDate)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Report: %v", result))
}