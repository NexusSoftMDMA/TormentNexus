package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
)

func HandleCreateZettel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    title, _ :=getString(args, "title")
    content, _ :=getString(args, "content")
    url, _ :=getString(args, "url")
    body := map[string]string{"title": title, "content": content}
    jsonBytes, _ := json.Marshal(body)
    req, e := http.NewRequestWithContext(ctx, "POST", url+"/zettel", strings.NewReader(string(jsonBytes)))
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    respBytes, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return ok("created zettel: " + string(respBytes))
}

func HandleSearchZettel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    url, _ :=getString(args, "url")
    req, e := http.NewRequestWithContext(ctx, "GET", url+"/zettel?q="+query, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    respBytes, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    return ok("search results: " + string(respBytes))
}