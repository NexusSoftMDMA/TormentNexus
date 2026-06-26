package tools

import (
    "context"
    "encoding/json"
    "net/http"
    "os"
)

func HandleListUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    baseURL, _ :=getString(args, "base_url")
    if baseURL == "" {
        baseURL = os.Getenv("ROADRECON_BASE_URL")

    if baseURL == "" {
        baseURL = "http://localhost:8000"
    }
    req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/users", nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to execute request: " + e.Error())
}

    defer resp.Body.Close()
    var result interface{}
    if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
        return err("failed to decode response: " + e.Error())
}

    data, _ := json.MarshalIndent(result, "", "  ")
    return success(string(data))
}

}

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    userID, _ :=getString(args, "user_id")
    if userID == "" {
        return err("user_id is required")
}

    baseURL, _ :=getString(args, "base_url")
    if baseURL == "" {
        baseURL = os.Getenv("ROADRECON_BASE_URL")

    if baseURL == "" {
        baseURL = "http://localhost:8000"
    }
    req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/users/"+userID, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to execute request: " + e.Error())
}

    defer resp.Body.Close()
    var result interface{}
    if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
        return err("failed to decode response: " + e.Error())
}

    data, _ := json.MarshalIndent(result, "", "  ")
    return success(string(data))
}
}