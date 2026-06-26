package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/providers"
)

type WaterfallClient struct {
	tiers    []Provider
	selector *providers.ModelSelector
}

func NewWaterfallClient(selector *providers.ModelSelector, tiers ...Provider) *WaterfallClient {
	return &WaterfallClient{
		tiers:    tiers,
		selector: selector,
	}
}

func (w *WaterfallClient) GenerateText(ctx context.Context, taskType string, messages []Message) (*LLMResponse, error) {
	// If a selector is configured, use it to pick a provider
	if w.selector != nil {
		providerName, err := w.selector.SelectProvider(ctx, taskType)
		if err != nil {
			return nil, err
		}
		var provider Provider
		selection, ok := getProviderSelection(providerName)
		if ok {
			provider = selection.Factory(selection.APIKey)
		}
		if provider == nil {
			return nil, fmt.Errorf("selected provider %s not available", providerName)
		}
		resp, err := provider.GenerateText(ctx, selection.DefaultModel, messages)
		if err == nil {
			return resp, nil
		}
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
			fmt.Printf("[Waterfall] Selected provider %s failed due to quota. Falling back...\n", providerName)
		}
		return nil, err
	}

	// Fallback: iterate through tiers (first successful wins)
	for _, tier := range w.tiers {
		resp, err := tier.GenerateText(ctx, taskType, messages)
		if err == nil {
			return resp, nil
		}
		// Only cascade on retryable errors (429, 500, 503)
		if strings.Contains(err.Error(), "429") ||
			strings.Contains(err.Error(), "500") ||
			strings.Contains(err.Error(), "503") {
			continue
		}
		// Non-retryable errors (4xx) fail immediately
		return nil, err
	}
	return nil, fmt.Errorf("all %d provider tiers failed", len(w.tiers))
}
