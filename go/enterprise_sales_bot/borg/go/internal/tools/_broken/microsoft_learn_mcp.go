package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"
)

type SearchResult struct {
    URL   string `json:"url"`
    Title string `json:"title"`
    // ...
}

var http.DefaultClient = http.DefaultClient

func HandleSearchLearn(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    maxResultsStr := "5"
    if maxResultsVal, found := args["max_results"]; found {
        if s, ok2 := maxResultsVal.(string); ok2 {
            maxResultsStr = s
        }
    }
    maxResults, parseErr := strconv.Atoi(maxResultsStr)
    if parseErr != nil {
        return err("invalid max_results")
}

    if maxResults < 1 {
        maxResults = 5
    }

    apiURL := fmt.Sprintf("https://learn.microsoft.com/api/search?search=%s&count=%d", url.QueryEscape(query), maxResults)
    req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if reqErr != nil {
        return err("failed to create request: " + reqErr.Error())
}

    req.Header.Set("User-Agent", "MCP-Go-Tool/1.0")

    resp, fetchErr := http.DefaultClient.Do(req)
    if fetchErr != nil {
        return err("failed to fetch search results: " + fetchErr.Error())
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err("failed to read response: " + readErr.Error())
}

    var results []SearchResult
    if parseErr := json.Unmarshal(body, &results); parseErr != nil {
        // maybe response structure is different
        return err("failed to parse results: " + parseErr.Error())
}

    if len(results) == 0 {
        return ok("No results found for '" + query + "'")
}

    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("Search results for '%s':\n\n", query))
    for i, r := range results {
        sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n\n", i+1, r.Title, r.URL))

    return ok(sb.String())
}

}

func HandleGetModule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    moduleID, _ :=getString(args, "module_id")
    if moduleID == "" {
        return err("module_id is required")
}

    apiURL := fmt.Sprintf("https://learn.microsoft.com/api/module?moduleId=%s", url.QueryEscape(moduleID))
    req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if reqErr != nil {
        return err("failed to create request: " + reqErr.Error())
}

    resp, fetchErr := http.DefaultClient.Do(req)
    if fetchErr != nil {
        return err("failed to fetch module: " + fetchErr.Error())
}

    defer resp.Body.Close()
    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err("failed to read response: " + readErr.Error())
}

    // For simplicity, return raw JSON as string
    return ok(fmt.Sprintf("Module details for %s:\n%s", moduleID, string(body)))
}

func HandleGetLearningPath(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    pathID, _ :=getString(args, "path_id")
    if pathID == "" {
        return err("path_id is required")
}

    apiURL := fmt.Sprintf("https://learn.microsoft.com/api/learning-path?pathId=%s", url.QueryEscape(pathID))
    req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if reqErr != nil {
        return err("failed to create request: " + reqErr.Error())
}

    resp, fetchErr := http.DefaultClient.Do(req)
    if fetchErr != nil {
        return err("failed to fetch learning path: " + fetchErr.Error())
}

    defer resp.Body.Close()
    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err("failed to read response: " + readErr.Error())
}

    return ok(fmt.Sprintf("Learning path details for %s:\n%s", pathID, string(body)))
}