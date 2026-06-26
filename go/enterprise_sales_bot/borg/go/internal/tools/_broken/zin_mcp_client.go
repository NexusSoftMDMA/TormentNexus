package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGenerate_zin_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	if model == "" || prompt == "" {
		return err("model and prompt are required")
}

	body := fmt.Sprintf(`{"model":"%s","prompt":"%s"}`, model, prompt)
	resp, e := http.DefaultClient.Post("http://localhost:11434/api/generate", "application/json", strings.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Response string `json:"response"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(result.Response)
}

func HandleListModels_zin_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:11434/api/tags")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	raw, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var list struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if e := json.Unmarshal(raw, &list); e != nil {
		return err("decode failed: " + e.Error())
}

	var names []string
	for _, m := range list.Models {
		names = append(names, m.Name)

	return success(fmt.Sprintf("Available models: %v", names))
}
}