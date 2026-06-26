package mcpimpl

import (
    "context"
    "fmt"
    "io"
    "net/http"
)

func HandleGetChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    chartId, _ :=getString(args, "chartId")
    if chartId == "" {
        return err("chartId is required")
}

    url := "https://api.datawrapper.de/v3/charts/" + chartId
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to fetch chart: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return success(string(body))
}

func HandleListCharts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    limit, _ :=getInt(args, "limit")
    if limit == 0 {
        limit = 10
    }
    url := fmt.Sprintf("https://api.datawrapper.de/v3/charts?limit=%d", limit)
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to list charts: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return success(string(body))
}