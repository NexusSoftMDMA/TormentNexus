package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("Missing required argument: query")
}

	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		return err("MINIMAX_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.minimax.com/v1/web/search?q="+query, nil)
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("Failed to parse JSON: %v", e))
}

	text, found := result["answer"].(string)
	if !found {
		return err("Unexpected response format")
}

	return success(text)
}

func HandleImageUnderstanding(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "image_url")
	question, _ :=getString(args, "question")
	if imageURL == "" || question == "" {
		return err("Missing required arguments: image_url and question")
}

	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		return err("MINIMAX_API_KEY not set")
}

	payload := map[string]string{"image_url": imageURL, "question": question}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("Failed to marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.minimax.com/v1/image/understand", nil)
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("Failed to parse JSON: %v", e))
}

	answer, found := result["answer"].(string)
	if !found {
		return err("Unexpected response format")
}

	return success(answer)
}