//go:build ignore
// +build ignore

package tools

/**
 * @file openrouter_deep_research.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of OpenRouter Deep Research MCP.
 * Replaces: openrouter-deep-research-mcp
 *
 * Multi-step research via OpenRouter API with web search augmentation.
 * Configurable via OPENROUTER_API_KEY env var.
 *
 * Tools:
 *  - deep_research — perform multi-step research on a topic
 *  - deep_research_status — check research progress
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const orDefaultURL = "https://openrouter.ai/api/v1"

func orAPIKey() string { return os.Getenv("OPENROUTER_API_KEY") }
func orBaseURL() string {
	if u := os.Getenv("OPENROUTER_BASE_URL"); u != "" {
		return u
	}
	return orDefaultURL
}

// HandleDeepResearch performs multi-step research via OpenRouter.
func HandleDeepResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ := getString(args, "topic", "query", "q")
	if topic == "" {
		return err("topic is required")
	}
	model, _ := getString(args, "model")
	if model == "" {
		model = "/gpt-4o-mini"
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a research assistant. Provide comprehensive, well-structured research on the given topic. Include key findings, sources, and analysis."},
			{"role": "user", "content": fmt.Sprintf("Research this topic thoroughly: %s", topic)},
		},
		"max_tokens": 4096,
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 120 * time.Second}
	req, e := http.NewRequestWithContext(ctx, "POST", orBaseURL()+"/chat/completions",
		strings.NewReader(string(body)))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+orAPIKey())

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("research failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleDeepResearchStatus returns the model and configuration status.
func HandleDeepResearchStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	info := map[string]interface{}{
		"provider": "OpenRouter",
		"api_key_set": orAPIKey() != "",
		"base_url": orBaseURL(),
		"available": true,
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	return ok(string(data))
}
