package mcpimpl

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func HandleListAssessments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    baseURL, _ :=getString(args, "baseUrl")
    if baseURL == "" {
        return err("baseUrl is required")
}

    url := baseURL + "/assessments"
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err(fmt.Sprintf("failed to create request: %v", e))
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err(fmt.Sprintf("request failed: %v", e))
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(fmt.Sprintf("failed to read response: %v", e))
}

    var data interface{}
    if e := json.Unmarshal(body, &data); e != nil {
        return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

    return ok(fmt.Sprintf("assessments: %v", data))
}

func HandleGetAssessment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    baseURL, _ :=getString(args, "baseUrl")
    id, _ :=getString(args, "id")
    if baseURL == "" || id == "" {
        return err("baseUrl and id are required")
}

    url := fmt.Sprintf("%s/assessments/%s", baseURL, id)
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err(fmt.Sprintf("failed to create request: %v", e))
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err(fmt.Sprintf("request failed: %v", e))
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(fmt.Sprintf("failed to read response: %v", e))
}

    var data interface{}
    if e := json.Unmarshal(body, &data); e != nil {
        return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

    return ok(fmt.Sprintf("assessment: %v", data))
}// touch 1781132142
