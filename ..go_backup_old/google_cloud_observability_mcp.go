package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleListMetricDescriptors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ :=getString(args, "projectId")
	reqURL := fmt.Sprintf("https://monitoring.googleapis.com/v3/projects/%s/metricDescriptors", projectId)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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

	return success(string(body))
}

func HandleQueryTimeSeries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ :=getString(args, "projectId")
	metricType, _ :=getString(args, "metricType")
	filter := "metric.type=" + url.QueryEscape(metricType)
	reqURL := fmt.Sprintf("https://monitoring.googleapis.com/v3/projects/%s/timeSeries?filter=%s", projectId, filter)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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

	return success(string(body))
}