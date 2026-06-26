package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SlackAPIResponse represents the standard Slack API response structure
type SlackAPIResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// SendMessageRequest represents the payload for sending a Slack message
type SendMessageRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// HandleSendMessage sends a message to a Slack channel
func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel")
	text, _ :=getString(args, "text")

	if channel == "" || text == "" {
		return err("channel and text are required parameters")
}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return err("SLACK_TOKEN environment variable is not set")
}

	reqBody := SendMessageRequest{
		Channel: channel,
		Text:    text,
	}

	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", marshalErr))
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage",
		strings.NewReader(string(jsonBody)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := http.DefaultClient
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", httpErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var apiResponse SlackAPIResponse
	if jsonErr := json.Unmarshal(body, &apiResponse); jsonErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", jsonErr))
}

	if !apiResponse.Ok {
		return err(fmt.Sprintf("Slack API error: %s", apiResponse.Error))
}

	return ok(fmt.Sprintf("Message sent successfully to channel %s", channel))
}

// HandleGetChannelHistory retrieves message history from a Slack channel
func HandleGetChannelHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel")
	if channel == "" {
		return err("channel is a required parameter")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return err("SLACK_TOKEN environment variable is not set")
}

	reqURL := fmt.Sprintf("https://slack.com/api/conversations.history?channel=%s&limit=%d",
		url.QueryEscape(channel), limit)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.DefaultClient
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", httpErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", jsonErr))
}

	if okVal, exists := result["ok"]; !exists || okVal != true {
		if errVal, exists := result["error"]; exists {
			return err(fmt.Sprintf("Slack API error: %v", errVal))
}

		return err("Unknown Slack API error")
}

	messages, exists := result["messages"]
	if !exists {
		return err("No messages found in response")
}

	jsonMessages, marshalErr := json.MarshalIndent(messages, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to format messages: %v", marshalErr))
}

	return ok(string(jsonMessages))
}

// HandleGetUserInfo retrieves information about a Slack user
func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	if user == "" {
		return err("user is a required parameter")
}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return err("SLACK_TOKEN environment variable is not set")
}

	reqURL := fmt.Sprintf("https://slack.com/api/users.info?user=%s", url.QueryEscape(user))

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.DefaultClient
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", httpErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", jsonErr))
}

	if okVal, exists := result["ok"]; !exists || okVal != true {
		if errVal, exists := result["error"]; exists {
			return err(fmt.Sprintf("Slack API error: %v", errVal))
}

		return err("Unknown Slack API error")
}

	userInfo, exists := result["user"]
	if !exists {
		return err("User information not found in response")
}

	jsonUser, marshalErr := json.MarshalIndent(userInfo, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to format user info: %v", marshalErr))
}

	return ok(string(jsonUser))
}

// HandleSetReminder creates a reminder for a Slack user
func HandleSetReminder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	text, _ :=getString(args, "text")
	timeStr, _ :=getString(args, "time")

	if user == "" || text == "" || timeStr == "" {
		return err("user, text, and time are required parameters")
}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return err("SLACK_TOKEN environment variable is not set")
}

	reqURL := fmt.Sprintf("https://slack.com/api/reminders.add?user=%s&text=%s&time=%s",
		url.QueryEscape(user), url.QueryEscape(text), url.QueryEscape(timeStr))

	req, reqErr := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.DefaultClient
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", httpErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var apiResponse SlackAPIResponse
	if jsonErr := json.Unmarshal(body, &apiResponse); jsonErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", jsonErr))
}

	if !apiResponse.Ok {
		return err(fmt.Sprintf("Slack API error: %s", apiResponse.Error))
}

	return ok(fmt.Sprintf("Reminder set successfully for user %s", user))
}