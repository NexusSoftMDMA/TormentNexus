package providers

import (
	"context"
	"fmt"
)

type ModelSelector struct {
	quotaManager *QuotaManager
}

func NewModelSelector(qm *QuotaManager) *ModelSelector {
	return &ModelSelector{quotaManager: qm}
}

func (s *ModelSelector) SelectProvider(ctx context.Context, taskType string) (string, error) {
	order := taskProviderOrder[taskType]
	if len(order) == 0 {
		order = taskProviderOrder["general"]
	}

	for _, provider := range order {
		if s.quotaManager.IsHealthy(provider) {
			return provider, nil
		}
	}

	return "", fmt.Errorf("no healthy providers found for task: %s", taskType)
}
