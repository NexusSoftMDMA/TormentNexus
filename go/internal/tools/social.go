//go:build ignore
// +build ignore

package tools

/**
 * @file social.go
 * @module go/internal/tools
 *
 * WHAT: Social media integration — Twitter/X and Reddit API.
 * Replaces: various social MCP servers
 *
 * Provides social media search and content access.
 * Configurable via TWITTER_BEARER_TOKEN and REDDIT_CLIENT_ID/CLIENT_SECRET.
 *
 * Tools:
 *  - twitter_search — search recent tweets
 *  - twitter_user_timeline — get user's recent tweets
 *  - reddit_search — search Reddit
 *  - reddit_get_posts — get top/new/hot posts from a subreddit
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ---- Twitter ----

func twitterBearerToken() string {
	return os.Getenv("TWITTER_BEARER_TOKEN")
}

func twitterGet(ctx context.Context, path string, params map[string]string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	u, _ := url.Parse("https://api.twitter.com/2" + path)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("Authorization", "Bearer "+twitterBearerToken())

	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("Twitter API error: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Twitter error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandleTwitterSearch searches recent tweets.
func HandleTwitterSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "text")
	if query == "" {
		return err("query is required")
	}
	if token := twitterBearerToken(); token == "" {
		return err("TWITTER_BEARER_TOKEN not set")
	}

	maxResults := getInt(args, "max_results", "limit")
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 10
	}

	params := map[string]string{
		"query":       query,
		"max_results": fmt.Sprintf("%d", maxResults),
		"tweet.fields": "created_at,public_metrics",
	}

	result, e := twitterGet(ctx, "/tweets/search/recent", params)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	return ok(result)
}

// HandleTwitterUserTimeline gets a user's recent tweets.
func HandleTwitterUserTimeline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ := getString(args, "username", "user", "screen_name")
	if username == "" {
		return err("username is required")
	}
	if token := twitterBearerToken(); token == "" {
		return err("TWITTER_BEARER_TOKEN not set")
	}

	maxResults := getInt(args, "max_results", "limit")
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 10
	}

	params := map[string]string{
		"max_results":  fmt.Sprintf("%d", maxResults),
		"tweet.fields": "created_at,public_metrics",
	}

	result, e := twitterGet(ctx, fmt.Sprintf("/users/by/username/%s", username), params)
	if e != nil {
		return err(fmt.Sprintf("user lookup failed: %v", e))
	}

	var userResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(result), &userResp); e != nil || userResp.Data.ID == "" {
		return ok(result)
	}

	timeline, e := twitterGet(ctx, fmt.Sprintf("/users/%s/tweets", userResp.Data.ID), params)
	if e != nil {
		return err(fmt.Sprintf("timeline failed: %v", e))
	}
	return ok(timeline)
}

// ---- Reddit ----

func redditUserAgent() string {
	return "TormentNexus/1.0"
}

func redditGet(ctx context.Context, path string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.reddit.com"+path, nil)
	if e != nil {
		return "", fmt.Errorf("request error: %v", e)
	}
	req.Header.Set("User-Agent", redditUserAgent())

	resp, e := client.Do(req)
	if e != nil {
		return "", fmt.Errorf("Reddit API error: %v", e)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Reddit error (%d): %s", resp.StatusCode, string(data))
	}
	return string(data), nil
}

// HandleRedditSearch searches Reddit.
func HandleRedditSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "text")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	result, e := redditGet(ctx, fmt.Sprintf("/search.json?q=%s&limit=%d&sort=relevance", url.QueryEscape(query), limit))
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	return ok(result)
}

// HandleRedditGetPosts gets posts from a subreddit.
func HandleRedditGetPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subreddit, _ := getString(args, "subreddit", "sub", "r")
	if subreddit == "" {
		return err("subreddit is required")
	}
	sort, _ := getString(args, "sort", "order")
	if sort == "" {
		sort = "hot"
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	result, e := redditGet(ctx, fmt.Sprintf("/r/%s/%s.json?limit=%d", subreddit, sort, limit))
	if e != nil {
		return err(fmt.Sprintf("get posts failed: %v", e))
	}
	return ok(result)
}
