package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListModels_ollama_mcp_bridge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:11434/api/tags")
	if e != nil {
		return err("failed to fetch models: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	models, found := result["models"].([]interface{})
	if !found {
		return err("no models field in response")
}

	var names []string
	for _, m := range models {
		mm, _ := m.(map[string]interface{})
		if name, found := mm["name"].(string); found {
			names = append(names, name)

	}
	out, _ := json.Marshal(names)
	return success(string(out))
}

}

func HandleGenerate_ollama_mcp_bridge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	if model == "" || prompt == "" {
		return err("model and prompt are required")
}

	payload := map[string]string{"model": model, "prompt": prompt}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(data))
	if e != nil {
		return err("failed to contact Ollama: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	text, _ := result["response"].(string)
	return ok(text)
}