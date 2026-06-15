package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func HandleGetUserRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    username, _ :=getString(args, "username")
    if username == "" {
        return err("username is required")
}

    url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return err("GitHub API returned status " + resp.Status)
}

    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    var repos []map[string]interface{}
    if e := json.Unmarshal(body, &repos); e != nil {
        return err("failed to parse response: " + e.Error())
}

    names := []string{}
    for _, repo := range repos {
        name, found := repo["name"].(string)
        if found {
            names = append(names, name)

    }
    return ok(fmt.Sprintf("Repos: %v", names))
}
}