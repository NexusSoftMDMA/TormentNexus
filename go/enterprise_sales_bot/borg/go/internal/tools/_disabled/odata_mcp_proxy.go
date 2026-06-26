package tools

import (
    "context"
    "io"
    "net/http"
    "strings"
)

func HandleODataQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    serviceURL, _ :=getString(args, "serviceUrl")
    entitySet, _ :=getString(args, "entitySet")
    filter, _ :=getString(args, "filter")
    url := serviceURL + "/" + entitySet
    if filter != "" {
        url += "?$filter=" + filter
    }
    req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    res, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer res.Body.Close()
    body, e := io.ReadAll(res.Body)
    if e != nil {
        return err("failed to read body: " + e.Error())
}

    return success(string(body))
}

func HandleRestApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    method, _ :=getString(args, "method")
    if method == "" {
        method = http.MethodGet
    }
    bodyStr, _ :=getString(args, "body")
    var bodyReader io.Reader
    if bodyStr != "" {
        bodyReader = strings.NewReader(bodyStr)

    req, e := http.NewRequestWithContext(ctx, method, url, bodyReader)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    res, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer res.Body.Close()
    body, e := io.ReadAll(res.Body)
    if e != nil {
        return err("failed to read body: " + e.Error())
}

    return success(string(body))
}
}