package mcpimpl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleOllamaTools(t *testing.T) {
	// Start a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		path := r.URL.Path
		switch path {
		case "/":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ollama is running"))
		case "/api/tags":
			w.Write([]byte(`{
				"models": [
					{
						"name": "llama3.2:latest",
						"size": 2019488102,
						"modified_at": "2026-06-04T12:00:00Z"
					}
				]
			}`))
		case "/api/chat":
			w.Write([]byte(`{
				"model": "llama3.2:latest",
				"message": {
					"role": "assistant",
					"content": "Hello! I am a local LLM."
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Redirect getOllamaHost to use our local mock server
	os.Setenv("OLLAMA_HOST", server.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	ctx := context.Background()

	// Test 1: HandleListLocalModels
	resp, err := HandleListLocalModels(ctx, nil)
	if err != nil {
		t.Fatalf("HandleListLocalModels failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "llama3.2:latest") {
		t.Errorf("Expected model list to contain llama3.2:latest, got: %s", resp.Content[0].Text)
	}

	// Test 2: HandleLocalLLMChat
	resp, err = HandleLocalLLMChat(ctx, map[string]interface{}{
		"message": "Hello",
	})
	if err != nil {
		t.Fatalf("HandleLocalLLMChat failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "local LLM") {
		t.Errorf("Expected chat response to contain 'local LLM', got: %s", resp.Content[0].Text)
	}

	// Test 3: HandleOllamaHealthCheck
	resp, err = HandleOllamaHealthCheck(ctx, nil)
	if err != nil {
		t.Fatalf("HandleOllamaHealthCheck failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "HEALTHY") {
		t.Errorf("Expected health status HEALTHY, got: %s", resp.Content[0].Text)
	}

	// Test 4: HandleSystemResourceCheck
	resp, err = HandleSystemResourceCheck(ctx, nil)
	if err != nil {
		t.Fatalf("HandleSystemResourceCheck failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "cpus") {
		t.Errorf("Expected system check to contain cpus field, got: %s", resp.Content[0].Text)
	}
}
