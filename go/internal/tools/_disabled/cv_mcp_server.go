package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

func HandleGetCv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    url := fmt.Sprintf("https://api.example.com/cv/%s", name)
    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch CV: " + e.Error())
}

    defer resp.Body.Close()
    var cv interface{}
    if e := json.NewDecoder(resp.Body).Decode(&cv); e != nil {
        return err("parse error: " + e.Error())
}

    return ok(fmt.Sprintf("CV for %s: %v", name, cv))
}

func HandleUpdateCv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    data, _ :=getString(args, "data")
    if name == "" || data == "" {
        return err("name and data required")
}

    url := "https://api.example.com/cv/update"
    payload := map[string]string{"name": name, "data": data}
    body, _ := json.Marshal(payload)
    resp, e := http.DefaultClient.Post(url, "application/json", nil)
    if e != nil {
        return err("update failed: " + e.Error())
}

    defer resp.Body.Close()
    return ok(fmt.Sprintf("Updated CV for %s", name))
}