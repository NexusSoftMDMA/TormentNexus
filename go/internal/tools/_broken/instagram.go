package tools

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

func fetchOembed(ctx context.Context, instagramURL string) (string, error) {
    base := "https://api.instagram.com/oembed"
    params := url.Values{}
    params.Set("url", instagramURL)
    reqURL := base + "?" + params.Encode()
    req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
    if reqErr != nil {
        return "", reqErr
    }
    client := http.Client{Timeout: 30 * time.Second}
    resp, doErr := client.Do(req)
    if doErr != nil {
        return "", doErr
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("oembed request failed with status %d", resp.StatusCode)
}

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return "", readErr
    }
    return string(body), nil
}

func HandleInstagramOembed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    urlStr, _ :=getString(args, "url")
    if urlStr == "" {
        return err("url argument is required")
}

    result, fetchErr := fetchOembed(ctx, urlStr)
    if fetchErr != nil {
        return err(fetchErr.Error())
}

    return ok(result)
}

func HandleInstagramUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    username, _ :=getString(args, "username")
    if username == "" {
        return err("username argument is required")
}

    profileURL := "https://www.instagram.com/" + username + "/"
    result, fetchErr := fetchOembed(ctx, profileURL)
    if fetchErr != nil {
        return err(fetchErr.Error())
}

    return ok(result)
}

func HandleInstagramMediaInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    shortcode, _ :=getString(args, "shortcode")
    if shortcode == "" {
        return err("shortcode argument is required")
}

    mediaURL := "https://www.instagram.com/p/" + shortcode + "/"
    result, fetchErr := fetchOembed(ctx, mediaURL)
    if fetchErr != nil {
        return err(fetchErr.Error())
}

    return ok(result)
}