package mcp

/**
 * @file preemptive_advertiser.go
 * @module go/internal/mcp
 *
 * WHAT: Preemptive tool advertising — watches conversation topics and injects
 *       relevant tool advertisements before the model has to search for them.
 *
 * WHY: Models often don't know what tools are available. By monitoring the
 *      conversation and detecting topics, we can proactively suggest the best
 *      tools for the current task context.
 *
 * DESIGN:
 *   - Topic detection from conversation messages
 *   - Score-based matching against the full tool catalog
 *   - Progressive disclosure: only show top 3-5 most relevant tools
 *   - Configurable injection thresholds
 *   - Debouncing to avoid spam
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"context"
	"strings"
	"sync"
	"time"
)

// ToolAdvertisement is a suggested tool injection for the model.
type ToolAdvertisement struct {
	ToolName    string  `json:"toolName"`
	Reason      string  `json:"reason"`
	Score       float64 `json:"score"`
	Category    string  `json:"category,omitempty"`
	Description string  `json:"description,omitempty"`
}

// ConversationMessage represents a message in the conversation being watched.
type ConversationMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// AdvertiseConfig controls the preemptive advertising behavior.
type AdvertiseConfig struct {
	// MinScore: minimum match score to advertise a tool
	MinScore float64 `json:"minScore"`
	// MaxAds: maximum advertisements per injection
	MaxAds int `json:"maxAds"`
	// Cooldown: minimum time between advertisements on the same topic
	Cooldown time.Duration `json:"cooldown"`
	// HistoryWindow: how far back to look in conversation for topics
	HistoryWindow time.Duration `json:"historyWindow"`
}

func DefaultAdvertiseConfig() AdvertiseConfig {
	return AdvertiseConfig{
		MinScore:      15.0,
		MaxAds:        5,
		Cooldown:      30 * time.Second,
		HistoryWindow: 5 * time.Minute,
	}
}

// PreemptiveAdvertiser watches conversations and suggests relevant tools.
type PreemptiveAdvertiser struct {
	cfg      AdvertiseConfig
	ds       *DecisionSystem
	mu       sync.Mutex
	history  []ConversationMessage
	injected map[string]time.Time // tool name → last injection time
}

// NewPreemptiveAdvertiser creates a new advertiser bound to a decision system.
func NewPreemptiveAdvertiser(ds *DecisionSystem, cfg AdvertiseConfig) *PreemptiveAdvertiser {
	return &PreemptiveAdvertiser{
		cfg:      cfg,
		ds:       ds,
		history:  []ConversationMessage{},
		injected: make(map[string]time.Time),
	}
}

// OnMessage is called when a new conversation message arrives.
func (pa *PreemptiveAdvertiser) OnMessage(msg ConversationMessage) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.history = append(pa.history, msg)

	// Trim old messages
	cutoff := time.Now().Add(-pa.cfg.HistoryWindow)
	var trimmed []ConversationMessage
	for _, m := range pa.history {
		if m.Timestamp.After(cutoff) {
			trimmed = append(trimmed, m)
		}
	}
	pa.history = trimmed
}

// GetAdvertisements analyzes recent conversation and returns tool suggestions.
func (pa *PreemptiveAdvertiser) GetAdvertisements(ctx context.Context) []ToolAdvertisement {
	pa.mu.Lock()
	topics := pa.extractTopics()
	history := pa.history
	pa.mu.Unlock()

	if len(topics) == 0 || len(history) == 0 {
		return nil
	}

	// Search for each topic and collect candidates
	allTools := pa.ds.getAllKnownTools()
	candidateMap := make(map[string]*ToolAdvertisement)

	for _, topic := range topics {
		ranked := RankTools(topic, allTools, pa.cfg.MaxAds)
		for _, r := range ranked {
			if r.Score < pa.cfg.MinScore {
				continue
			}

			if existing, ok := candidateMap[r.AdvertisedName]; ok {
				// Keep the higher score
				if r.Score > existing.Score {
					existing.Score = r.Score
					existing.Reason = r.MatchReason
				}
			} else {
				candidateMap[r.AdvertisedName] = &ToolAdvertisement{
					ToolName:    r.AdvertisedName,
					Reason:      r.MatchReason,
					Score:       r.Score,
					Description: truncateStr(r.Description, 100),
				}
			}
		}
	}

	// Filter out recently injected tools
	now := time.Now()
	var ads []ToolAdvertisement
	for _, ad := range candidateMap {
		if lastInjected, ok := pa.injected[ad.ToolName]; ok {
			if now.Sub(lastInjected) < pa.cfg.Cooldown {
				continue
			}
		}
		ads = append(ads, *ad)
	}

	// Sort by score descending and limit
	sortAdvertisements(ads)

	if len(ads) > pa.cfg.MaxAds {
		ads = ads[:pa.cfg.MaxAds]
	}

	// Mark as injected
	pa.mu.Lock()
	for _, ad := range ads {
		pa.injected[ad.ToolName] = now
	}
	pa.mu.Unlock()

	return ads
}

// extractTopics extracts key topics from the recent conversation history.
func (pa *PreemptiveAdvertiser) extractTopics() []string {
	var allContent []string
	for _, msg := range pa.history {
		if msg.Role == "user" || msg.Role == "assistant" {
			allContent = append(allContent, msg.Content)
		}
	}

	if len(allContent) == 0 {
		return nil
	}

	combined := strings.Join(allContent, " ")
	tokens := Tokenize(combined)

	// Count token frequency
	freq := make(map[string]int)
	for _, t := range tokens {
		freq[t]++
	}

	// Get top topics by frequency
	type topicScore struct {
		topic string
		count int
	}
	var scored []topicScore
	for t, c := range freq {
		if c >= 1 && len(t) > 3 { // Skip very short tokens
			scored = append(scored, topicScore{t, c})
		}
	}

	// Sort by frequency
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].count > scored[i].count {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Also extract bigrams (two-word phrases)
	words := strings.Fields(combined)
	bigrams := make(map[string]int)
	for i := 0; i < len(words)-1; i++ {
		w1 := strings.ToLower(strings.Trim(words[i], ".,!?;:()[]{}\"'"))
		w2 := strings.ToLower(strings.Trim(words[i+1], ".,!?;:()[]{}\"'"))
		if len(w1) > 2 && len(w2) > 2 {
			bigrams[w1+" "+w2]++
		}
	}

	for bg, count := range bigrams {
		if count >= 2 {
			scored = append(scored, topicScore{bg, count})
		}
	}

	// Take top 5 topics
	var topics []string
	for i, s := range scored {
		if i >= 5 {
			break
		}
		topics = append(topics, s.topic)
	}

	return topics
}

func sortAdvertisements(ads []ToolAdvertisement) {
	for i := 0; i < len(ads); i++ {
		for j := i + 1; j < len(ads); j++ {
			if ads[j].Score > ads[i].Score {
				ads[i], ads[j] = ads[j], ads[i]
			}
		}
	}
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
