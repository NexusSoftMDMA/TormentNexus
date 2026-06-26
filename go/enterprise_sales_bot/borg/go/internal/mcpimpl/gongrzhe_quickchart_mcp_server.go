package mcpimpl

import (
	"context"
	"fmt"
	"net/url"
)

func HandleGenerateChart_gongrzhe_quickchart_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chartConfig, _ :=getString(args, "chart")
	if chartConfig == "" {
		return err("chart configuration is required")
}

	width, _ :=getInt(args, "width")
	if width <= 0 {
		width = 500
	}
	height, _ :=getInt(args, "height")
	if height <= 0 {
		height = 300
	}
	format, _ :=getString(args, "format")
	if format == "" {
		format = "png"
	}
	u, e := url.Parse("https://quickchart.io/chart")
	if e != nil {
		return err("failed to parse base URL")
}

	q := u.Query()
	q.Set("c", chartConfig)
	q.Set("width", fmt.Sprintf("%d", width))
	q.Set("height", fmt.Sprintf("%d", height))
	q.Set("format", format)
	u.RawQuery = q.Encode()
	return ok(u.String())
}