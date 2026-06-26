package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var http.DefaultClient = http.DefaultClient

// HandleChat handles chat message interactions
func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	model, _ :=getString(args, "model")
	if model == "" {
		model = "gpt-4"
	}

	payload := map[string]interface{}{
		"message": message,
		"model":   model,
	}
	payloadBytes, parseErr := json.Marshal(payload)
	if parseErr != nil {
		return err(parseErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/chat", strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("chat API error: %s", string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return err(jsonErr.Error())
}

	responseVal, found := result["response"].(string)
	if !found {
		return err("invalid response format")
}

	return ok(responseVal)
}

// HandleGenerateImage handles image generation requests
func HandleGenerateImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	payload := map[string]interface{}{
		"prompt": prompt,
	}
	payloadBytes, parseErr := json.Marshal(payload)
	if parseErr != nil {
		return err(parseErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/generate-image", strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("image generation error: %s", string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return err(jsonErr.Error())
}

	imageURL, found := result["url"].(string)
	if !found {
		return err("invalid response format")
}

	return ok(fmt.Sprintf("Image generated successfully: %s", imageURL))
}

// HandleWebSearch handles web search requests
func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	searchURL, urlErr := url.Parse("http://localhost:3000/api/search")
	if urlErr != nil {
		return err(urlErr.Error())
}

	q := searchURL.Query()
	q.Set("q", query)
	searchURL.RawQuery = q.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("search error: %s", string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return err(jsonErr.Error())
}

	results, found := result["results"].([]interface{})
	if !found {
		return err("invalid response format")
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search results for '%s':\n\n", query))
	for i, r := range results {
		if i >= 10 {
			break
		}
		if item, found := r.(map[string]interface{}); found {
			title, _ := item["title"].(string)
			link, _ := item["url"].(string)
			snippet, _ := item["snippet"].(string)
			output.WriteString(fmt.Sprintf("%d. %s\n   %s\n   %s\n\n", i+1, title, link, snippet))

	}

	return ok(output.String())
}

}

// HandleListAgents lists available AI agents
func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, reqErr := http.NewRequestWithContext(ctx, "GET", "http://localhost:3000/api/agents", nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("list agents error: %s", string(body)))
}

	var agents []map[string]interface{}
	if jsonErr := json.Unmarshal(body, &agents); jsonErr != nil {
		return err(jsonErr.Error())
}

	var output strings.Builder
	output.WriteString("Available Agents:\n\n")
	for _, agent := range agents {
		name, _ := agent["name"].(string)
		description, _ := agent["description"].(string)
		id, _ := agent["id"].(string)
		output.WriteString(fmt.Sprintf("- %s (ID: %s)\n  %s\n\n", name, id, description))

	return ok(output.String())
}

}

// HandleGetHistory retrieves conversation history
func HandleGetHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	threadID, _ :=getString(args, "thread_id")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 50
	}

	historyURL, urlErr := url.Parse(fmt.Sprintf("http://localhost:3000/api/history/%s", threadID))
	if urlErr != nil {
		return err(urlErr.Error())
}

	q := historyURL.Query()
	q.Set("limit", fmt.Sprintf("%d", limit))
	historyURL.RawQuery = q.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", historyURL.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("history error: %s", string(body)))
}

	var messages []map[string]interface{}
	if jsonErr := json.Unmarshal(body, &messages); jsonErr != nil {
		return err(jsonErr.Error())
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Conversation History (Thread: %s):\n\n", threadID))
	for _, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		timestamp, _ := msg["timestamp"].(string)
		output.WriteString(fmt.Sprintf("[%s] %s: %s\n", timestamp, role, content))

	return ok(output.String())
}
}