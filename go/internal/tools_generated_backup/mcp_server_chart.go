package tools

import (
    "context"
    "fmt"
)

func HandleGenerateChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    chartType, _ :=getString(args, "type")
    data, _ :=getString(args, "data")
    if chartType == "" {
        return err("chart type is required")
}

    msg := fmt.Sprintf("Generated chart of type '%s' with data: %s", chartType, data)
    return ok(msg)
}

func HandleListChartTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    types := []string{"bar", "line", "pie", "scatter", "area", "radar", "heatmap", "treemap", "box", "histogram", "funnel", "gauge", "map", "sankey", "wordcloud", "rose", "sunburst", "parallel", "radial", "chord", "force", "graph", "tree", "circle", "chord"}
    msg := fmt.Sprintf("Available chart types: %v", types)
    return ok(msg)
}