//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// callSlackAPI sends a request to the Slack Web API using the configured token.
func callSlackAPI(ctx context.Context, method string, urlPath string, queryParams map[string]string, body interface{}) ([]byte, error) {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		botToken = os.Getenv("SLACK_API_TOKEN")
	}
	if botToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN environment variable is not set")
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	baseUrl := os.Getenv("SLACK_API_URL")
	if baseUrl == "" {
		baseUrl = "https://slack.com/api/"
	}
	reqUrl := baseUrl + urlPath
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			if v != "" {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("slack API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// HandleSlackListChannels lists channels that the bot has access to.
func HandleSlackListChannels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limitVal := getInt(args, "limit")
	if limitVal <= 0 {
		limitVal = 100
	}
	if limitVal > 200 {
		limitVal = 200
	}

	cursor, _ := getString(args, "cursor")

	predefinedChannelIds := os.Getenv("SLACK_CHANNEL_IDS")
	if predefinedChannelIds != "" {
		// Fetch specific channel infos
		channelIDs := strings.Split(predefinedChannelIds, ",")
		var channels []interface{}

		for _, id := range channelIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}

			respData, err := callSlackAPI(ctx, "GET", "conversations.info", map[string]string{
				"channel": id,
			}, nil)
			if err != nil {
				continue
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(respData, &parsed); err != nil {
				continue
			}

			if okVal, exists := parsed["ok"].(bool); exists && okVal {
				if channel, ok := parsed["channel"].(map[string]interface{}); ok {
					if isArchived, okArch := channel["is_archived"].(bool); !okArch || !isArchived {
						channels = append(channels, channel)
					}
				}
			}
		}

		result := map[string]interface{}{
			"ok":       true,
			"channels": channels,
			"response_metadata": map[string]string{
				"next_cursor": "",
			},
		}

		b, err := json.Marshal(result)
		if err != nil {
			return errResponse(err)
		}
		return ok(string(b))
	}

	// Normal Conversations List
	queryParams := map[string]string{
		"types":            "public_channel,private_channel",
		"exclude_archived": "true",
		"limit":            strconv.Itoa(limitVal),
	}
	if teamID := os.Getenv("SLACK_TEAM_ID"); teamID != "" {
		queryParams["team_id"] = teamID
	}
	if cursor != "" {
		queryParams["cursor"] = cursor
	}

	respData, err := callSlackAPI(ctx, "GET", "conversations.list", queryParams, nil)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackPostMessage posts a message to a channel or direct message.
func HandleSlackPostMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channelID, _ := getString(args, "channel_id")
	if channelID == "" {
		return err("channel_id parameter is required")
	}

	text, _ := getString(args, "text")
	if text == "" {
		return err("text parameter is required")
	}

	body := map[string]interface{}{
		"channel": channelID,
		"text":    text,
	}

	respData, err := callSlackAPI(ctx, "POST", "chat.postMessage", nil, body)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackReplyToThread posts a reply to a thread.
func HandleSlackReplyToThread(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channelID, _ := getString(args, "channel_id")
	if channelID == "" {
		return err("channel_id parameter is required")
	}

	threadTs, _ := getString(args, "thread_ts")
	if threadTs == "" {
		return err("thread_ts parameter is required")
	}

	text, _ := getString(args, "text")
	if text == "" {
		return err("text parameter is required")
	}

	body := map[string]interface{}{
		"channel":   channelID,
		"thread_ts": threadTs,
		"text":      text,
	}

	respData, err := callSlackAPI(ctx, "POST", "chat.postMessage", nil, body)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackAddReaction adds an emoji reaction to a message.
func HandleSlackAddReaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channelID, _ := getString(args, "channel_id")
	if channelID == "" {
		return err("channel_id parameter is required")
	}

	timestamp, _ := getString(args, "timestamp")
	if timestamp == "" {
		return err("timestamp parameter is required")
	}

	reaction, _ := getString(args, "reaction")
	if reaction == "" {
		return err("reaction parameter is required")
	}

	body := map[string]interface{}{
		"channel":   channelID,
		"timestamp": timestamp,
		"name":      reaction,
	}

	respData, err := callSlackAPI(ctx, "POST", "reactions.add", nil, body)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackGetChannelHistory retrieves channel messages.
func HandleSlackGetChannelHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channelID, _ := getString(args, "channel_id")
	if channelID == "" {
		return err("channel_id parameter is required")
	}

	limitVal := getInt(args, "limit")
	if limitVal <= 0 {
		limitVal = 10
	}

	queryParams := map[string]string{
		"channel": channelID,
		"limit":   strconv.Itoa(limitVal),
	}

	respData, err := callSlackAPI(ctx, "GET", "conversations.history", queryParams, nil)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackGetThreadReplies retrieves all replies in a message thread.
func HandleSlackGetThreadReplies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channelID, _ := getString(args, "channel_id")
	if channelID == "" {
		return err("channel_id parameter is required")
	}

	threadTs, _ := getString(args, "thread_ts")
	if threadTs == "" {
		return err("thread_ts parameter is required")
	}

	queryParams := map[string]string{
		"channel": channelID,
		"ts":      threadTs,
	}

	respData, err := callSlackAPI(ctx, "GET", "conversations.replies", queryParams, nil)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackGetUsers lists all users in the workspace.
func HandleSlackGetUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limitVal := getInt(args, "limit")
	if limitVal <= 0 {
		limitVal = 100
	}
	if limitVal > 200 {
		limitVal = 200
	}

	cursor, _ := getString(args, "cursor")

	queryParams := map[string]string{
		"limit": strconv.Itoa(limitVal),
	}
	if teamID := os.Getenv("SLACK_TEAM_ID"); teamID != "" {
		queryParams["team_id"] = teamID
	}
	if cursor != "" {
		queryParams["cursor"] = cursor
	}

	respData, err := callSlackAPI(ctx, "GET", "users.list", queryParams, nil)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// HandleSlackGetUserProfile retrieves detailed profile information for a user.
func HandleSlackGetUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ := getString(args, "user_id")
	if userID == "" {
		return err("user_id parameter is required")
	}

	queryParams := map[string]string{
		"user":           userID,
		"include_labels": "true",
	}

	respData, err := callSlackAPI(ctx, "GET", "users.profile.get", queryParams, nil)
	if err != nil {
		return errResponse(err)
	}

	return ok(string(respData))
}

// errResponse is a helper that wraps Go errors in a ToolResponse
func errResponse(err error) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Slack API call failed: %v", err)}},
		IsError: true,
	}, nil
}
