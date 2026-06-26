package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleCreateAssistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	instructions, _ :=getString(args, "instructions")
	model, _ :=getString(args, "model")
	if model == "" {
		model = "gpt-4o-mini"
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	body := map[string]interface{}{
		"name":         name,
		"instructions": instructions,
		"model":        model,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("Failed to marshal request body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/assistants", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(nil) // not used
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("Failed to parse response JSON: " + e.Error())
}

	if id, found := result["id"].(string); found {
		return success(fmt.Sprintf("Assistant created with ID: %s", id))
}

	if errorMsg, found := result["error"].(map[string]interface{}); found {
		msg, _ := errorMsg["message"].(string)
		return err("OpenAI error: " + msg)
}

	return err("Unknown response from OpenAI")
}