package contextpruner

import (
	"strings"
	"sync"
)

type PruneStrategy string

const (
	StrategyLength    PruneStrategy = "length"
	StrategyAge       PruneStrategy = "age"
	StrategyRelevance PruneStrategy = "relevance"
)

type Service struct {
	mu        sync.Mutex
	maxTokens int
}

func NewService(maxTokens int) *Service {
	if maxTokens <= 0 {
		maxTokens = 100000
	}
	return &Service{maxTokens: maxTokens}
}

type PruneResult struct {
	OriginalSize int    `json:"originalSize"`
	PrunedSize   int    `json:"prunedSize"`
	RemovedCount int    `json:"removedCount"`
	Strategy     string `json:"strategy"`
}

func (s *Service) Prune(content string, strategy PruneStrategy) *PruneResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalSize := len(content)
	lines := strings.Split(content, "\n")
	removed := 0

	switch strategy {
	case StrategyLength:
		if len(content) > s.maxTokens {
			content = content[:s.maxTokens]
			removed = originalSize - len(content)
		}
	case StrategyAge:
		// Keep last N lines
		if len(lines) > 1000 {
			content = strings.Join(lines[len(lines)-1000:], "\n")
			removed = len(lines) - 1000
		}
	default:
		// Remove empty lines
		var kept []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				kept = append(kept, line)
			} else {
				removed++
			}
		}
		content = strings.Join(kept, "\n")
	}

	return &PruneResult{
		OriginalSize: originalSize,
		PrunedSize:   len(content),
		RemovedCount: removed,
		Strategy:     string(strategy),
	}
}

func (s *Service) SetMaxTokens(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxTokens = n
}
