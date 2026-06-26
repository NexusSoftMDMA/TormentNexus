package mcp

/**
 * @file conversational_predictor.go
 * @module go/internal/mcp
 *
 * WHAT: Go-native implementation of the ConversationalToolInjector mirror.
 *       Maintains a sliding conversation window and uses Ollama (Gemma 12b
 *       or configured model) to predict which tools should be preloaded into
 *       the session working set.
 *
 * WHY: The TypeScript ConversationalToolInjector (Phase 113) calls this
 *      sidecar endpoint first before falling back to cloud LLMs. By handling
 *      prediction here natively we keep latency low (<200ms for local Ollama)
 *      and avoid cloud API costs for routine tool routing.
 *
 * Endpoints served (registered by httpapi/server.go):
 *   POST /api/mcp/tools/predict-conversational  { prompt, systemPrompt } -> { tools: [] }
 *   POST /api/mcp/conversation/append            { role, text }          -> { ok: true }
 *   GET  /api/mcp/conversation/window                                     -> { turns, tokenCount }
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ConversationTurn holds a single message turn in the sliding window.
type ConversationTurn struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// ConversationalPredictor maintains a bounded sliding window of conversation
// turns and uses a local Ollama model to predict relevant tool names.
type ConversationalPredictor struct {
	mu              sync.RWMutex
	window          []ConversationTurn
	maxWindowTurns  int
	throttleMs      int64
	maxTools        int
	minWindowTokens int
	lastPredictedAt time.Time
	lastResult      []string
}

// NewConversationalPredictor creates a new predictor with sensible defaults.
// These match the TypeScript ConversationalToolInjector defaults.
func NewConversationalPredictor() *ConversationalPredictor {
	return &ConversationalPredictor{
		maxWindowTurns:  8,
		throttleMs:      3000,
		maxTools:        5,
		minWindowTokens: 30,
	}
}

// AppendTurn adds a turn to the sliding window, keeping it bounded.
func (p *ConversationalPredictor) AppendTurn(role, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if len(text) > 1000 {
		text = text[:1000]
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.window = append(p.window, ConversationTurn{
		Role:      role,
		Text:      text,
		Timestamp: time.Now(),
	})

	for len(p.window) > p.maxWindowTurns {
		p.window = p.window[1:]
	}
}

// WindowSnapshot returns a copy of the current sliding window.
func (p *ConversationalPredictor) WindowSnapshot() []ConversationTurn {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]ConversationTurn, len(p.window))
	copy(result, p.window)
	return result
}

// WindowTokenCount returns a rough word-based token count for the window.
func (p *ConversationalPredictor) WindowTokenCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var sb strings.Builder
	for _, t := range p.window {
		sb.WriteString(t.Text)
		sb.WriteString(" ")
	}
	return countRoughTokens(sb.String())
}

// PredictFromPrompt is the primary entry point called by the TypeScript
// ConversationalToolInjector via POST /api/mcp/tools/predict-conversational.
// It accepts a fully-formed systemPrompt and user prompt and returns a
// JSON array of tool name strings by calling Ollama locally.
func (p *ConversationalPredictor) PredictFromPrompt(
	ctx context.Context,
	systemPrompt, userPrompt string,
) ([]string, error) {
	p.mu.Lock()
	elapsed := time.Since(p.lastPredictedAt).Milliseconds()
	if elapsed < p.throttleMs && len(p.lastResult) > 0 {
		cached := make([]string, len(p.lastResult))
		copy(cached, p.lastResult)
		p.mu.Unlock()
		return cached, nil
	}
	p.mu.Unlock()

	result, err := p.callOllama(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.lastPredictedAt = time.Now()
	p.lastResult = result
	p.mu.Unlock()

	return result, nil
}

// PredictFromWindow builds its own prompt from the current conversation window
// and a provided compact catalog, then calls Ollama. Used by the Go-native
// preemptive advertiser and the autonomous prediction path.
func (p *ConversationalPredictor) PredictFromWindow(
	ctx context.Context,
	catalog []PredictorCatalogEntry,
) ([]string, error) {
	p.mu.RLock()
	tokenCount := countRoughTokensSlice(p.window)
	turns := make([]ConversationTurn, len(p.window))
	copy(turns, p.window)
	p.mu.RUnlock()

	if tokenCount < p.minWindowTokens {
		return nil, nil
	}

	// Build compact catalog text (max 200 entries)
	var catalogSB strings.Builder
	for i, entry := range catalog {
		if i >= 200 {
			break
		}
		desc := entry.Description
		if len(desc) > 100 {
			desc = desc[:100]
		}
		fmt.Fprintf(&catalogSB, "%s: %s\n", entry.Name, desc)
	}

	systemPrompt := fmt.Sprintf(
		`You are a predictive tool routing assistant embedded in a developer AI system (TormentNexus).
Given the recent conversation window below, select up to %d tools from the provided catalog that the user is MOST LIKELY to need next.

Rules:
- Only select tools genuinely relevant to the user's apparent current task/direction.
- Return a JSON array of tool name strings ONLY, no explanation text.
- If no tools are clearly relevant, return an empty array [].
- Example valid responses: ["github__create_issue","filesystem__write_file"] or []

Available tools catalog (name: description):
%s`,
		p.maxTools,
		catalogSB.String(),
	)

	var windowSB strings.Builder
	for _, t := range turns {
		fmt.Fprintf(&windowSB, "[%s] %s\n", t.Role, t.Text)
	}
	userPrompt := fmt.Sprintf("Recent conversation window:\n%s\nWhich tools should be pre-loaded?", windowSB.String())

	return p.PredictFromPrompt(ctx, systemPrompt, userPrompt)
}

// PredictorCatalogEntry is the minimal tool metadata needed for prediction prompts.
type PredictorCatalogEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AlwaysOn    bool   `json:"alwaysOn"`
	Loaded      bool   `json:"loaded"`
}

// callOllama sends a chat completion request to the local Ollama server.
// Falls back to a zero-dependency heuristic on connection failure.
func (p *ConversationalPredictor) callOllama(ctx context.Context, system, user string) ([]string, error) {
	ollamaURL := os.Getenv("TORMENTNEXUS_OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://127.0.0.1:11434"
	}
	model := os.Getenv("TORMENTNEXUS_LOCAL_PREDICT_MODEL")
	if model == "" {
		model = "gemma3:12b"
	}

	reqBody, _ := json.Marshal(map[string]any{
		"model":  model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"options": map[string]any{
			"temperature": 0.1,
			"num_predict": 128,
		},
	})

	ctxTimeout, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodPost, ollamaURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("ollama request build: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama response read: %w", err)
	}

	var payload struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("ollama response parse: %w", err)
	}

	return extractJSONArray(payload.Message.Content), nil
}

// extractJSONArray safely extracts the first JSON string array from text output.
func extractJSONArray(text string) []string {
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start == -1 || end == -1 || end <= start {
		return nil
	}

	var result []string
	if err := json.Unmarshal([]byte(text[start:end+1]), &result); err != nil {
		return nil
	}

	// Sanitize: keep only valid tool name strings (alphanumeric + _ and __)
	var clean []string
	for _, name := range result {
		if isValidToolName(name) {
			clean = append(clean, name)
		}
	}
	return clean
}

func isValidToolName(name string) bool {
	if name == "" || len(name) > 128 {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func countRoughTokens(text string) int {
	return len(strings.Fields(text))
}

func countRoughTokensSlice(turns []ConversationTurn) int {
	var sb strings.Builder
	for _, t := range turns {
		sb.WriteString(t.Text)
		sb.WriteString(" ")
	}
	return countRoughTokens(sb.String())
}
