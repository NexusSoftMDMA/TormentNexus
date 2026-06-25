package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
)

func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")

    apiKey := os.Getenv("ZHIPU_API_KEY")
    if apiKey == "" {
        return err("ZHIPU_API_KEY environment variable is not set")
}

    reqBody := map[string]interface{}{
        "tool": "web-search-pro",
        "messages": []map[string]string{
            {"role": "user", "content": query},
        },
        "stream": false,
    }

    jsonBody, e := json.Marshal(reqBody)
    if e != nil {
        return err(fmt.Sprintf("failed to marshal request: %v", e))
}

    req, e := http.NewRequestWithContext(ctx, "POST", "https://open.bigmodel.cn/api/paas/v4/tools", strings.NewReader(string(jsonBody)))
    if e != nil {
        return err(fmt.Sprintf("failed to create request: %v", e))
}

    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := http.DefaultClient
    resp, e := client.Do(req)
    if e != nil {
        return err(fmt.Sprintf("failed to send request: %v", e))
}

    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return err(fmt.Sprintf("API request failed with status: %d, body: %s", resp.StatusCode, string(body)))
}

    // Read response body
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(fmt.Sprintf("failed to read response body: %v", e))
}

    // Parse JSON response
    var apiResp struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    if e := json.Unmarshal(body, &apiResp); e != nil {
        return err(fmt.Sprintf("failed to parse response: %v", e))
}

    // Extract content
    if len(apiResp.Choices) == 0 {
        return err("no choices in response")
}

    result := apiResp.Choices[0].Message.Content

    return ok(result)
}