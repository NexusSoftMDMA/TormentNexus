package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleAskLm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	body := fmt.Sprintf(`{"prompt":"%s"}`, prompt)
	resp, e := http.DefaultClient.Post("https://api.houtini-lm.example.com/ask", "application/json", strings.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	answer, found := result["answer"].(string)
	if !found {
		return err("no answer in response")
}

	return ok("Answer: " + answer)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}