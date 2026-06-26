package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceID, _ :=getString(args, "service_id")
	if serviceID == "" {
		return err("service_id is required")
}

	url := fmt.Sprintf("http://localhost:8080/api/services/%s", serviceID)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceID, _ :=getString(args, "service_id")
	metricName, _ :=getString(args, "metric")
	if serviceID == "" || metricName == "" {
		return err("service_id and metric are required")
}

	url := fmt.Sprintf("http://localhost:8080/api/metrics/%s/%s", serviceID, metricName)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}