package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:11434/api/tags")
	if e != nil {
		return err("failed to fetch models: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse models: " + e.Error())
}

	msg := "Available models: "
	for i, m := range result.Models {
		if i > 0 {
			msg += ", "
		}
		msg += m.Name
	}
	return success(msg)
}

func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	if model == "" || prompt == "" {
		return err("model and prompt are required")
}

	body, _ := json.Marshal(map[string]string{"model": model, "prompt": prompt})
	resp, e := http.DefaultClient.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("generate request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Response string `json:"response"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(result.Response)
}