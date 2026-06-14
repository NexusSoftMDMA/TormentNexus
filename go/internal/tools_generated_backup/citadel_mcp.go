package tools

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
)

func HandleListStacks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    base := os.Getenv("DOCS_BASE_URL")
    if base == "" {
        return err("DOCS_BASE_URL not set")
}

    resp, e := http.DefaultClient.Get(base + "/stacks")
    if e != nil {
        return err(fmt.Sprintf("request failed: %v", e))
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(fmt.Sprintf("read failed: %v", e))
}

    return ok(string(body))
}

func HandleGetStackDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    stack, _ :=getString(args, "stack")
    if stack == "" {
        return err("stack argument required")
}

    base := os.Getenv("DOCS_BASE_URL")
    if base == "" {
        return err("DOCS_BASE_URL not set")
}

    resp, e := http.DefaultClient.Get(base + "/docs/" + stack)
    if e != nil {
        return err(fmt.Sprintf("request failed: %v", e))
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err(fmt.Sprintf("read failed: %v", e))
}

    return ok(string(body))
}