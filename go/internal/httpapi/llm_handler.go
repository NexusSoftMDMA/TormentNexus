package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

func (s *Server) handleLLMGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		TaskType string       `json:"taskType"`
		Model    string       `json:"model,omitempty"`
		Messages []ai.Message `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	if len(payload.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "messages array is required"})
		return
	}
	if strings.TrimSpace(payload.TaskType) == "" {
		payload.TaskType = "chat"
	}

	if s.waterfallClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "LLM provider routing is not initialized",
		})
		return
	}

	startedAt := time.Now()

	// If a specific model is requested, try direct provider call
	if strings.TrimSpace(payload.Model) != "" {
		resp, err := s.callModelDirectly(r.Context(), payload.Model, payload.Messages)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data": map[string]any{
					"content":   resp.Content,
					"provider":  resp.Provider,
					"model":     resp.Model,
					"usage":     resp.Usage,
					"latencyMs": time.Since(startedAt).Milliseconds(),
				},
				"bridge": map[string]any{"source": "go-llm-provider"},
			})
			return
		}
	}

	// Use WaterfallClient for provider routing
	resp, err := s.waterfallClient.GenerateText(r.Context(), payload.TaskType, payload.Messages)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "LLM generation failed: " + err.Error(),
			"data":    nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"content":   resp.Content,
			"provider":  resp.Provider,
			"model":     resp.Model,
			"usage":     resp.Usage,
			"latencyMs": time.Since(startedAt).Milliseconds(),
		},
		"bridge": map[string]any{"source": "go-llm-waterfall"},
	})
}

func (s *Server) callModelDirectly(ctx context.Context, model string, messages []ai.Message) (*ai.LLMResponse, error) {
	providerName := detectProviderFromModel(model)

	var apiKey string
	switch providerName {
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey != "" {
			p := &ai.OpenAIProvider{APIKey: apiKey}
			return p.GenerateText(ctx, model, messages)
		}
	case "anthropic":
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey != "" {
			p := &ai.AnthropicProvider{APIKey: apiKey}
			return p.GenerateText(ctx, model, messages)
		}
	case "gemini":
		apiKey = os.Getenv("GOOGLE_API_KEY")
		if apiKey != "" {
			p := &ai.GeminiProvider{APIKey: apiKey}
			return p.GenerateText(ctx, model, messages)
		}
	case "deepseek":
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
		if apiKey != "" {
			p := &ai.DeepSeekProvider{APIKey: apiKey}
			return p.GenerateText(ctx, model, messages)
		}
	case "openrouter":
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		if apiKey != "" {
			p := &ai.OpenRouterProvider{APIKey: apiKey}
			return p.GenerateText(ctx, model, messages)
		}
	}
	return nil, nil // Fall back to WaterfallClient
}

func detectProviderFromModel(model string) string {
	model = strings.ToLower(model)
	switch {
	case strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3"):
		return "openai"
	case strings.HasPrefix(model, "claude"):
		return "anthropic"
	case strings.HasPrefix(model, "gemini"):
		return "gemini"
	case strings.HasPrefix(model, "deepseek"):
		return "deepseek"
	case strings.HasPrefix(model, "mistral"):
		return "mistral"
	case model == "auto" || strings.Contains(model, "openrouter"):
		return "openrouter"
	default:
		return ""
	}
}
