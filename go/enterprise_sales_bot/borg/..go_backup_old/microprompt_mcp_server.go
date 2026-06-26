package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListPrompts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/prompts")
	if e != nil {
		return err("failed to list prompts: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON")
}

	return success(fmt.Sprintf("%v", data))
}

func HandleGetPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "prompt_id")
	if id == "" {
		return err("prompt_id is required")
}

	url := fmt.Sprintf("http://localhost:8080/prompts/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get prompt: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON")
}

	return success(fmt.Sprintf("%v", data))
}