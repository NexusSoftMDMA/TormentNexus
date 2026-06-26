package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type HTTPProvider struct {
	config TierConfig
	client *http.Client
}

func NewHTTPProvider(cfg TierConfig) *HTTPProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &HTTPProvider{
		config: cfg,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *HTTPProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	url := strings.TrimRight(p.config.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}
	for k, v := range p.config.Headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, &ProviderError{ProviderID: p.config.ID, StatusCode: 0, Body: err.Error(), Retryable: true}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		var chatResp ChatResponse
		if err := json.Unmarshal(body, &chatResp); err != nil {
			return nil, fmt.Errorf("decode response from %s: %w", p.config.ID, err)
		}
		chatResp.ProviderID = p.config.ID
		return &chatResp, nil
	}
	retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
	return nil, &ProviderError{ProviderID: p.config.ID, StatusCode: resp.StatusCode, Body: string(body), Retryable: retryable}
}

type tierProvider struct {
	priority int
	provider *HTTPProvider
}

type WaterfallRouter struct {
	providers      []tierProvider
	mu             sync.RWMutex
	totalRequests  atomic.Int64
	totalFallbacks atomic.Int64
	totalFailures  atomic.Int64
}

func NewWaterfallRouter(configs []TierConfig) *WaterfallRouter {
	providers := make([]tierProvider, 0, len(configs))
	for _, cfg := range configs {
		providers = append(providers, tierProvider{priority: cfg.Priority, provider: NewHTTPProvider(cfg)})
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].priority < providers[j].priority })
	return &WaterfallRouter{providers: providers}
}

func (r *WaterfallRouter) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	r.totalRequests.Add(1)
	exclude := make(map[string]bool)
	for _, id := range req.Exclude {
		exclude[id] = true
	}
	var lastErr error
	attempts := 0
	for _, tp := range r.providers {
		if exclude[tp.provider.config.ID] {
			continue
		}
		attempts++
		resp, err := tp.provider.Chat(ctx, req)
		if err == nil {
			resp.Attempts = attempts
			return resp, nil
		}
		lastErr = err
		if pe, ok := err.(*ProviderError); ok && pe.Retryable {
			r.totalFallbacks.Add(1)
			continue
		}
		break
	}
	r.totalFailures.Add(1)
	if lastErr == nil {
		return nil, fmt.Errorf("llm: no providers configured")
	}
	return nil, lastErr
}

func (r *WaterfallRouter) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	req.Stream = false
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		resp, err := r.Chat(ctx, req)
		if err != nil {
			ch <- StreamChunk{Err: err, Done: true}
			return
		}
		if len(resp.Choices) > 0 {
			ch <- StreamChunk{Delta: &resp.Choices[0].Message, Done: false}
		}
		ch <- StreamChunk{Done: true, Usage: &resp.Usage}
	}()
	return ch, nil
}

func (r *WaterfallRouter) SetConfig(configs []TierConfig) {
	providers := make([]tierProvider, 0, len(configs))
	for _, cfg := range configs {
		providers = append(providers, tierProvider{priority: cfg.Priority, provider: NewHTTPProvider(cfg)})
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].priority < providers[j].priority })
	r.mu.Lock()
	r.providers = providers
	r.mu.Unlock()
}

func (r *WaterfallRouter) Stats() (requests, fallbacks, failures int64) {
	return r.totalRequests.Load(), r.totalFallbacks.Load(), r.totalFailures.Load()
}

var _ LLMClient = (*WaterfallRouter)(nil)
