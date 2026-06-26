package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleGenerateBarChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	xField, _ :=getString(args, "xField")
	yField, _ :=getString(args, "yField")
	dataUrl, _ :=getString(args, "dataUrl")
	spec := map[string]interface{}{
		"$schema": "https://vega.github.io/schema/vega-lite/v5.json",
		"title":   title,
		"data":    map[string]string{"url": dataUrl},
		"mark":    "bar",
		"encoding": map[string]interface{}{
			"x": map[string]string{"field": xField, "type": "nominal"},
			"y": map[string]string{"field": yField, "type": "quantitative"},
		},
	}
	b, e := json.Marshal(spec)
	if e != nil {
		return err("failed to marshal spec")
}

	return success(string(b))
}

func HandleGenerateLineChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	xField, _ :=getString(args, "xField")
	yField, _ :=getString(args, "yField")
	dataUrl, _ :=getString(args, "dataUrl")
	spec := map[string]interface{}{
		"$schema": "https://vega.github.io/schema/vega-lite/v5.json",
		"title":   title,
		"data":    map[string]string{"url": dataUrl},
		"mark":    "line",
		"encoding": map[string]interface{}{
			"x": map[string]string{"field": xField, "type": "temporal"},
			"y": map[string]string{"field": yField, "type": "quantitative"},
		},
	}
	b, e := json.Marshal(spec)
	if e != nil {
		return err("failed to marshal spec")
}

	return success(string(b))
}