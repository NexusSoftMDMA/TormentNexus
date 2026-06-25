package tools

import (
    "context"
    "encoding/json"
    "encoding/base64"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"
)

func ok(text string) ToolResponse {
    return ToolResponse{Text: text}
}

func err(e error) ToolResponse {
    return ToolResponse{Text: e.Error()}
}

func getString(args map[string]interface{}, key string) string {
    if val, found := args[key]; found {
        return val.(string)
}

    return ""
}

func getInt(args map[string]interface{}, key string) int {
    if val, found := args[key]; found {
        return val.(int)
}

    return 0
}

func getBool(args map[string]interface{}, key string) bool {
    if val, found := args[key]; found {
        return val.(bool)
}

    return false
}

func getProfileEmail(ctx context.Context) (string, error) {
    // Implement fetching user email from profile
    // For simplicity, return a placeholder email
    return "user@example.com", nil
}

func callGmailAPI(ctx context.Context, method, path string, body io.Reader) (map[string]interface{}, error) {
    token, apiErr := getAccessToken()
    if apiErr != nil {
        return nil, apiErr
    }
    urlStr := "https://gmail.googleapis.com/gmail/v1/users/me" + path
    req, reqErr := http.NewRequestWithContext(ctx, method, urlStr, body)
    if reqErr != nil {
        return nil, reqErr
    }
    req.Header.Set("Authorization", "Bearer "+token)
    res, resErr := http.DefaultClient.Do(req)
    if resErr != nil {
        return nil, resErr
    }
    defer res.Body.Close()
    var result map[string]interface{}
    decodeErr := json.NewDecoder(res.Body).Decode(&result)
    if decodeErr != nil {
        return nil, decodeErr
    }
    if res.StatusCode >= 400 {
        return nil, fmt.Errorf("Gmail API error: %s", result["error"])
}

    return result, nil
}

func HandleListMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    maxResults, _ :=getInt(args, "maxResults")
    if maxResults <= 0 {
        maxResults = 10
    }
    params := url.Values{}
    if query != "" {
        params.Set("q", query)

    if maxResults > 0 {
        params.Set("maxResults", strconv.Itoa(maxResults))

    path := "/messages?" + params.Encode()
    data, apiErr := callGmailAPI(ctx, "GET", path, nil)
    if apiErr != nil {
        return err(apiErr.Error())
}

    messages, found := data["messages"].([]interface{})
    if !found {
        return ok("No messages found.")
}

    var b strings.Builder
    for i, msg := range messages {
        m, _ := msg.(map[string]interface{})
        id, _ := m["id"].(string)
        threadId, _ := m["threadId"].(string)
        b.WriteString(fmt.Sprintf("%d. ID: %s (Thread: %s)\n", i+1, id, threadId))

    return ok(b.String())
}

}
}
}

func HandleGetMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    if id == "" {
        return err("id is required")
}

    data, apiErr := callGmailAPI(ctx, "GET", "/messages/"+id, nil)
    if apiErr != nil {
        return err(apiErr.Error())
}

    var b strings.Builder
    b.WriteString(fmt.Sprintf("ID: %s\n", data["id"]))
    b.WriteString(fmt.Sprintf("Thread: %s\n", data["threadId"]))
    if snippet, found := data["snippet"].(string); found {
        b.WriteString(fmt.Sprintf("Snippet: %s\n", snippet))

    if headers, found := data["payload"].(map[string]interface{}); found {
        if headerList, found := headers["headers"].([]interface{}); found {
            for _, h := range headerList {
                header, _ := h.(map[string]interface{})
                name, _ := header["name"].(string)
                value, _ := header["value"].(string)
                if strings.EqualFold(name, "From") || strings.EqualFold(name, "To") || strings.EqualFold(name, "Subject") || strings.EqualFold(name, "Date") {
                    b.WriteString(fmt.Sprintf("%s: %s\n", name, value))

            }
        }
    }
    return ok(b.String())
}

}
}

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    to, _ :=getString(args, "to")
    subject, _ :=getString(args, "subject")
    body, _ :=getString(args, "body")
    if to == "" || subject == "" || body == "" {
        return err("to, subject, and body are required")
}

    userEmail, emailErr := getProfileEmail(ctx)
    if emailErr != nil {
        return err(emailErr.Error())
}

    message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s", userEmail, to, subject, body)
    encoded := base64.URLEncoding.EncodeToString([]byte(message))
    payload := map[string]string{
        "raw": encoded,
    }
    jsonBytes, jsonErr := json.Marshal(payload)
    if jsonErr != nil {
        return err(jsonErr.Error())
}

    data, apiErr := callGmailAPI(ctx, "POST", "/messages/send", strings.NewReader(string(jsonBytes)))
    if apiErr != nil {
        return err(apiErr.Error())
}

    id, _ := data["id"].(string)
    return ok(fmt.Sprintf("Email sent successfully. Message ID: %s", id))
}