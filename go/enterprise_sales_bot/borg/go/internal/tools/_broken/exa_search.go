package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"
)

type ExaSearchResult struct {
    Title   string `json:"title"`
    URL     string `json:"url"`
    Snippet string `json:"snippet"`
    Score   string `json:"score,omitempty"`
}

type ExaSearchResponse struct {
    Results    []ExaSearchResult `json:"results"`
    APIMetadata struct {
        CreditsUsed int `json:"credits_used,omitempty"`
    } `json:"api_metadata,omitempty"`
}

type ExaContentsResponse struct {
    Contents []struct {
        URL     string `json:"url"`
        Text    string `json:"text"`
        Title   string `json:"title,omitempty"`
    } `json:"contents"`
}

func HandleExaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query parameter is required")
}

    numResults := 10
    if n := getInt(args, "num_results"); n > 0 {
        numResults = n
    }

    typeFilter, _ :=getString(args, "type_filter")
    startDate, _ :=getString(args, "start_date")
    endDate, _ :=getString(args, "end_date")

    apiKey := os.Getenv("EXA_API_KEY")
    if apiKey == "" {
        return err("EXA_API_KEY environment variable not set")
}

    searchURL := "https://api.exa.ai/search"
    params := url.Values{}
    params.Set("query", query)
    params.Set("num_results", strconv.Itoa(numResults))

    if typeFilter != "" {
        params.Set("type", typeFilter)

    if startDate != "" {
        params.Set("start_date", startDate)

    if endDate != "" {
        params.Set("end_date", endDate)

    req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL+"?"+params.Encode(), nil)
    if reqErr != nil {
        return err(reqErr.Error())
}

    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Accept", "application/json")

    client := http.DefaultClient
    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fetchErr.Error())
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(readErr.Error())
}

    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("Exa API error: status %d, body: %s", resp.StatusCode, string(body)))
}

    var result ExaSearchResponse
    if parseErr := json.Unmarshal(body, &result); parseErr != nil {
        return err(parseErr.Error())
}

    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("Found %d results for: %s\n\n", len(result.Results), query))

    for i, r := range result.Results {
        builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
        builder.WriteString(fmt.Sprintf("    URL: %s\n", r.URL))
        if r.Snippet != "" {
            builder.WriteString(fmt.Sprintf("    Snippet: %s\n", r.Snippet))

        if r.Score != "" {
            builder.WriteString(fmt.Sprintf("    Score: %s\n", r.Score))

        builder.WriteString("\n")

    return ok(builder.String())
}

}
}
}
}
}
}

func HandleExaContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    targetURL, _ :=getString(args, "url")
    if targetURL == "" {
        return err("url parameter is required")
}

    apiKey := os.Getenv("EXA_API_KEY")
    if apiKey == "" {
        return err("EXA_API_KEY environment variable not set")
}

    contentsURL := "https://api.exa.ai/contents"
    params := url.Values{}
    params.Set("urls", targetURL)

    req, reqErr := http.NewRequestWithContext(ctx, "GET", contentsURL+"?"+params.Encode(), nil)
    if reqErr != nil {
        return err(reqErr.Error())
}

    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Accept", "application/json")

    client := http.DefaultClient
    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fetchErr.Error())
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(readErr.Error())
}

    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("Exa API error: status %d, body: %s", resp.StatusCode, string(body)))
}

    var result ExaContentsResponse
    if parseErr := json.Unmarshal(body, &result); parseErr != nil {
        return err(parseErr.Error())
}

    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("Contents from: %s\n\n", targetURL))

    for _, c := range result.Contents {
        if c.Title != "" {
            builder.WriteString(fmt.Sprintf("Title: %s\n", c.Title))

        builder.WriteString(fmt.Sprintf("Text:\n%s\n", c.Text))

    return ok(builder.String())
}

}
}

func HandleExaFindSimilar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    urlToFind, _ :=getString(args, "url")
    if urlToFind == "" {
        return err("url parameter is required")
}

    numResults := 10
    if n := getInt(args, "num_results"); n > 0 {
        numResults = n
    }

    apiKey := os.Getenv("EXA_API_KEY")
    if apiKey == "" {
        return err("EXA_API_KEY environment variable not set")
}

    similarURL := "https://api.exa.ai/find-similar"
    params := url.Values{}
    params.Set("url", urlToFind)
    params.Set("num_results", strconv.Itoa(numResults))

    req, reqErr := http.NewRequestWithContext(ctx, "GET", similarURL+"?"+params.Encode(), nil)
    if reqErr != nil {
        return err(reqErr.Error())
}

    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Accept", "application/json")

    client := http.DefaultClient
    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fetchErr.Error())
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(readErr.Error())
}

    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("Exa API error: status %d, body: %s", resp.StatusCode, string(body)))
}

    var result ExaSearchResponse
    if parseErr := json.Unmarshal(body, &result); parseErr != nil {
        return err(parseErr.Error())
}

    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("Found %d similar pages to: %s\n\n", len(result.Results), urlToFind))

    for i, r := range result.Results {
        builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
        builder.WriteString(fmt.Sprintf("    URL: %s\n", r.URL))
        if r.Snippet != "" {
            builder.WriteString(fmt.Sprintf("    Snippet: %s\n", r.Snippet))

        builder.WriteString("\n")

    return ok(builder.String())
}
}
}