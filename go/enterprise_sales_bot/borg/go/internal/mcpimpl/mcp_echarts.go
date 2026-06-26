package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleGenerateEchart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chartType, _ :=getString(args, "type")
	title, _ :=getString(args, "title")
	config := map[string]interface{}{
		"type":  chartType,
		"title": title,
		"series": []map[string]interface{}{
			{"data": []int{10, 20, 30}},
		},
	}
	jsonBytes, e := json.Marshal(config)
	if e != nil {
		return err("failed to marshal config: " + e.Error())
}

	return ok("ECharts config: " + string(jsonBytes))
}