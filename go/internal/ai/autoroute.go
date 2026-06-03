package ai

import (
	"context"
	"fmt"
	"os"
    "github.com/tormentnexushq/tormentnexus-go/internal/providers"
)

type Router struct {
    selector *providers.ModelSelector
}

func NewRouter(selector *providers.ModelSelector) *Router {
    return &Router{selector: selector}
}

func (r *Router) Route(ctx context.Context, taskType string, messages []Message) (*LLMResponse, error) {
    providerName, err := r.selector.SelectProvider(ctx, taskType)
    if err != nil {
        return nil, err
    }

    selection, ok := getProviderSelection(providerName)
    if !ok {
        return nil, fmt.Errorf("provider %s not found", providerName)
    }

    return selection.Factory(selection.APIKey).GenerateText(ctx, selection.DefaultModel, messages)
}

func getProviderSelection(name string) (providerSelection, bool) {
    for _, entry := range ProviderPriority {
        if entry.ProviderName == name {
            if key := os.Getenv(entry.EnvVar); key != "" {
                return providerSelection{
                    EnvVar:       entry.EnvVar,
                    ProviderName: entry.ProviderName,
                    DefaultModel: entry.DefaultModel,
                    Factory:      entry.Factory,
                    APIKey:       key,
                }, true
            }
        }
    }
    return providerSelection{}, false
}
