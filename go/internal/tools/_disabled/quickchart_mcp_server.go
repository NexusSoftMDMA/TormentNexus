package tools

import (
	"context"
	"net/url"
)

func HandleGenerateChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spec, _ :=getString(args, "spec")
	if spec == "" {
		return err("spec parameter is required")
}

	chartURL := "https://quickchart.io/chart?c=" + url.QueryEscape(spec)
	return ok(chartURL)
}