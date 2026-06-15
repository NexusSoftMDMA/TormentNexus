package tools

import (
    "context"
    "io"
    "net/http"
)

func HandleGetScores(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    league, _ :=getString(args, "league")
    resp, e := http.DefaultClient.Get("https://api.example.com/scores?league=" + league)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return ok(string(body))
}