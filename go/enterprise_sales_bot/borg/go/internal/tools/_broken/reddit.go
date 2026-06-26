package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ok(message string) ToolResponse {
	return ToolResponse{Success: true, Message: message}
}

func err(message string) ToolResponse {
	return ToolResponse{Success: false, Message: message}
}

func getString(args map[string]interface{}, key string) string {
	if value, found := args[key]; found {
		return value.(string)
}

	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if value, found := args[key]; found {
		return value.(int)
}

	return 0
}

func HandleGetHotPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subreddit, _ :=getString(args, "subreddit")
	if subreddit == "" {
		return err("subreddit is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	apiURL := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=%d", url.PathEscape(subreddit), limit)
	body, fetchErr := fetchRedditData(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	var listing RedditListingData
	if parseErr := json.Unmarshal(body, &listing); parseErr != nil {
		return err(parseErr.Error())
}

	var posts []RedditPost
	for _, child := range listing.Children {
		if post, found := child.Data.(map[string]interface{}); found {
			posts = append(posts, RedditPost{
				Title:     post["title"].(string),
				Author:    post["author"].(string),
				Score:     int(post["score"].(float64)),
				URL:       post["url"].(string),
				Permalink: "https://reddit.com" + post["permalink"].(string),
				CreatedAt: post["created_utc"].(float64),
			})

	}

	result := fmt.Sprintf("Hot posts in r/%s:\n\n", subreddit)
	for i, post := range posts {
		result += fmt.Sprintf("%d. %s (Score: %d, by u/%s)\n", i+1, post.Title, post.Score, post.Author)
		result += fmt.Sprintf("   URL: %s\n", post.URL)
		result += fmt.Sprintf("   Link: %s\n\n", post.Permalink)

	return ok(result)
}

}
}

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	apiURL := fmt.Sprintf("https://www.reddit.com/user/%s/about.json", url.PathEscape(username))
	body, fetchErr := fetchRedditData(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	var userData struct {
		Data RedditUser `json:"data"`
	}
	if parseErr := json.Unmarshal(body, &userData); parseErr != nil {
		return err(parseErr.Error())
}

	user := userData.Data
	createdAt := time.Unix(int64(user.CreatedAt), 0).UTC().Format(time.RFC1123)

	result := fmt.Sprintf("User: %s\n", user.Name)
	result += fmt.Sprintf("Link Karma: %d\n", user.LinkKarma)
	result += fmt.Sprintf("Comment Karma: %d\n", user.CommentKarma)
	result += fmt.Sprintf("Total Karma: %d\n", user.TotalKarma)
	result += fmt.Sprintf("Account Created: %s\n", createdAt)

	return ok(result)
}

func HandleGetPostComments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	postURL, _ :=getString(args, "post_url")
	if postURL == "" {
		return err("post_url is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	if !strings.HasPrefix(postURL, "https://") && !strings.HasPrefix(postURL, "http://") {
		postURL = "https://" + postURL
	}
	if !strings.HasSuffix(postURL, ".json") {
		if strings.Contains(postURL, "?") {
			postURL += "&raw_json=1"
		} else {
			postURL += "?raw_json=1"
		}
		postURL += "&limit=" + strconv.Itoa(limit)

	body, fetchErr := fetchRedditData(postURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	var responses []json.RawMessage
	if parseErr := json.Unmarshal(body, &responses); parseErr != nil {
		return err(parseErr.Error())
}

	if len(responses) < 2 {
		return err("No comments found")
}

	var listing RedditListingData
	if parseErr := json.Unmarshal(responses[1], &listing); parseErr != nil {
		return err(parseErr.Error())
}

	var comments []RedditComment
	for _, child := range listing.Children {
		if commentData, found := child.Data.(map[string]interface{}); found {
			comments = append(comments, RedditComment{
				Author: commentData["author"].(string),
				Body:   commentData["body"].(string),
				Score:  int(commentData["score"].(float64)),
			})

	}

	result := "Post Comments:\n\n"
	for i, comment := range comments {
		if i >= limit {
			break
		}
		result += fmt.Sprintf("%d. u/%s (Score: %d)\n", i+1, comment.Author, comment.Score)
		result += fmt.Sprintf("   %s\n\n", comment.Body)

	return ok(result)
}
}
}
}