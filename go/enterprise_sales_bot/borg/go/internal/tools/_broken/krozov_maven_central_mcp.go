package mcpimpl

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

func HandleSearchMaven_krozov_maven_central_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query parameter is required")
}

    u := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=%s&rows=5&wt=json", url.QueryEscape(query))
    resp, e := http.DefaultClient.Get(u)
    if e != nil {
        return err(fmt.Sprintf("request failed: %v", e))
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(fmt.Sprintf("read failed: %v", e))
}

    var result struct {
        Response struct {
            Docs []struct {
                ID      string `json:"id"`
                LatestVersion string `json:"latestVersion"`
            } `json:"docs"`
        } `json:"response"`
    }
    if e := json.Unmarshal(body, &result); e != nil {
        return err(fmt.Sprintf("parse failed: %v", e))
}

    if len(result.Response.Docs) == 0 {
        return success("No results found")
}

    out := "Maven Central search results:\n"
    for _, doc := range result.Response.Docs {
        out += fmt.Sprintf("- %s (latest: %s)\n", doc.ID, doc.LatestVersion)

    return success(out)
}
}