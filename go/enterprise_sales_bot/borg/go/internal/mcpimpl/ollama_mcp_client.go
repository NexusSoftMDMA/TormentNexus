package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleX_ollama_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	input, _ :=getString(args, "input")
	if model == "" {
		return err("model is required")
}

	if input == "" {
		return err("input is required")
}

	response, e := http.DefaultClient.Post("http://localhost:8080/ollama", "application/json", bytes.NewBuffer([]byte(`{"model":"`+model+`","input":"`+input+`"}`)))
	if e != nil {
		return err("failed to call ollama")
}

	defer response.Body.Close()

	var result map[string]interface{}
	e = json.NewDecoder(response.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return success(result["output"].(string))
}

func HandleY_ollama_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Additional handler can be implemented here
	return ok("Handler Y not implemented")
}