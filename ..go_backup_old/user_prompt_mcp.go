package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetUserPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/prompts/" + name)
	if e != nil {
		return err("failed to fetch prompt: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleSetUserPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	if name == "" || content == "" {
		return err("name and content are required")
}

	return ok(fmt.Sprintf("prompt set for %s", name))
}