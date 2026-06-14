package tools

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

// HandleListSites returns all sites for the authenticated user.
func HandleListSites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    token := os.Getenv("NETLIFY_ACCESS_TOKEN")
    if token == "" {
        return err("NETLIFY_ACCESS_TOKEN not set")
}

    req, e := http.NewRequestWithContext(ctx, "GET", "https://api.netlify.com/api/v1/sites", nil)
    if e != nil {
        return err("failed to create request")
}

    req.Header.Set("Authorization", "Bearer "+token)
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to fetch sites")
}

    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response")
}

    var sites []map[string]interface{}
    if e := json.Unmarshal(body, &sites); e != nil {
        return err("failed to parse sites")
}

    return ok(fmt.Sprintf("Found %d sites", len(sites)))
}

// HandleCreateSite creates a new site from a repository.
func HandleCreateSite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    repoURL, _ :=getString(args, "repo_url")
    if name == "" || repoURL == "" {
        return err("name and repo_url are required")
}

    token := os.Getenv("NETLIFY_ACCESS_TOKEN")
    if token == "" {
        return err("NETLIFY_ACCESS_TOKEN not set")
}

    payload := map[string]string{"name": name, "repo": repoURL}
    body, e := json.Marshal(payload)
    if e != nil {
        return err("failed to marshal payload")
}

    req, e := http.NewRequestWithContext(ctx, "POST", "https://api.netlify.com/api/v1/sites", bytes.NewReader(body))
    if e != nil {
        return err("failed to create request")
}

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to create site")
}

    defer resp.Body.Close()
    if resp.StatusCode != 201 {
        return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

    return ok("Site created successfully")
}