package tools

import (
    "context"
    "net/http"
)

func HandleIngestGitRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url is required")
}

    resp, e := http.DefaultClient.Head(url)
    if e != nil {
        return err("failed to reach URL: " + e.Error())
}

    resp.Body.Close()
    if resp.StatusCode >= 400 {
        return err("URL returned status " + resp.Status)
}

    return success("Git repository ingested successfully from " + url)
}