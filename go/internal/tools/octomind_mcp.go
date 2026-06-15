//go:build ignore
// +build ignore

package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func HandleListTestSuites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    apiURL, _ :=getString(args, "api_url")
    if apiURL == "" {
        return err("api_url is required")
}

    apiKey, _ :=getString(args, "api_key")
    req, e := http.NewRequestWithContext(ctx, "GET")


-reasoner (deepseek)*
}