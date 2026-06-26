package providers

import (
	"sync"
	"time"
)

type ProviderQuota struct {
	Provider     string    `json:"provider"`
	TokensUsed   int64     `json:"tokensUsed"`
	TokenLimit   int64     `json:"tokenLimit"`
	CreditsLeft  float64   `json:"creditsLeft"`
	LastUpdateAt time.Time `json:"lastUpdateAt"`
}

type QuotaManager struct {
	mu     sync.RWMutex
	quotas map[string]*ProviderQuota
}

func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]*ProviderQuota),
	}
}

func (qm *QuotaManager) UpdateUsage(provider string, tokens int64, cost float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	q, ok := qm.quotas[provider]
	if !ok {
		q = &ProviderQuota{Provider: provider, TokenLimit: 1000000, CreditsLeft: 10.0}
		qm.quotas[provider] = q
	}

	q.TokensUsed += tokens
	q.CreditsLeft -= cost
	q.LastUpdateAt = time.Now()
}

func (qm *QuotaManager) IsHealthy(provider string) bool {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	q, ok := qm.quotas[provider]
	if !ok {
		return true // Assume healthy if unknown
	}

	return q.CreditsLeft > 0.05 && q.TokensUsed < q.TokenLimit
}

func (qm *QuotaManager) GetQuotas() []*ProviderQuota {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	var list []*ProviderQuota
	for _, q := range qm.quotas {
		copy := *q
		list = append(list, &copy)
	}
	return list
}
