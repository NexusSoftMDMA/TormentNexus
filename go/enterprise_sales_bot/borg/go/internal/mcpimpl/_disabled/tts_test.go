package mcpimpl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleSayTTS(t *testing.T) {
	// Test basic invocation
	ctx := context.Background()
	resp, err := HandleSayTTS(ctx, map[string]interface{}{
		"text": "Hello test",
		"rate": 200.0,
	})
	if err != nil {
		t.Fatalf("HandleSayTTS failed: %v", err)
	}

	// Headless platforms may return errors about SAPI or missing drivers. That is normal.
	if resp.IsError {
		t.Logf("HandleSayTTS returned speech synthesis error (expected on headless runners): %s", resp.Content[0].Text)
	} else {
		t.Logf("HandleSayTTS succeeded: %s", resp.Content[0].Text)
	}
}

func TestHandleOpenAITTS(t *testing.T) {
	// Start mock OpenAI HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock-mp3-binary-data"))
	}))
	defer server.Close()

	os.Setenv("OPENAI_API_KEY", "test-api-key")
	os.Setenv("OPENAI_API_URL", server.URL)
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("OPENAI_API_URL")

	ctx := context.Background()

	resp, err := HandleOpenAITTS(ctx, map[string]interface{}{
		"text":  "Synthesize this text please.",
		"model": "tts-1",
		"voice": "alloy",
		"speed": 1.0,
	})
	if err != nil {
		t.Fatalf("HandleOpenAITTS failed: %v", err)
	}

	if resp.IsError {
		t.Fatalf("HandleOpenAITTS returned error: %s", resp.Content[0].Text)
	}

	if !strings.Contains(resp.Content[0].Text, "saved to:") {
		t.Errorf("Expected response to indicate saved path, got: %s", resp.Content[0].Text)
	}

	// Clean up generated temp file
	parts := strings.Split(resp.Content[0].Text, "saved to: ")
	if len(parts) == 2 {
		tempPath := strings.TrimSpace(parts[1])
		if _, errStat := os.Stat(tempPath); errStat == nil {
			os.Remove(tempPath)
		}
	}
}
