package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

func HandleMoonPhase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    dateStr, _ :=getString(args, "date")
    if dateStr == "" {
        dateStr = time.Now().Format("2006-01-02")

    url := fmt.Sprintf("https://api.lunar.com/phase?date=%s", dateStr)
    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch moon phase: " + e.Error())
}

    defer resp.Body.Close()

    var result struct {
        Phase string `json:"phase"`
    }
    if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
        return err("failed to parse response: " + e.Error())
}

    return ok(fmt.Sprintf("Moon phase on %s: %s", dateStr, result.Phase))
}
}