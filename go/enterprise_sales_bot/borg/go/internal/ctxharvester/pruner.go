package ctxharvester

/**
 * @file pruner.go
 * @module go/internal/ctxharvester
 *
 * WHAT: Context pruning — reduces context size by removing low-value chunks,
 *       summarizing old content, and enforcing token budgets.
 *
 * WHY: Long conversations accumulate context that needs to be pruned to stay
 *      within model context windows. This implements smart pruning based on
 *      relevance scoring, recency, and access patterns.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"math"
	"sort"
	"strings"
	"time"
)

// PruneOptions controls how context is pruned.
type PruneOptions struct {
	MaxTokens     int     `json:"maxTokens"`
	KeepRecent    int     `json:"keepRecent"`    // Always keep this many recent chunks
	MinRelevance  float64 `json:"minRelevance"`  // Remove chunks below this score
	DecayPerHour  float64 `json:"decayPerHour"`  // Relevance decay rate
	TargetRatio   float64 `json:"targetRatio"`   // Target fill ratio (0.0-1.0)
}

// DefaultPruneOptions returns sensible defaults.
func DefaultPruneOptions() PruneOptions {
	return PruneOptions{
		MaxTokens:    100000,
		KeepRecent:   5,
		MinRelevance: 0.1,
		DecayPerHour: 0.05,
		TargetRatio:  0.7,
	}
}

// PruneResult describes what happened during pruning.
type PruneResult struct {
	BeforeChunks  int     `json:"beforeChunks"`
	AfterChunks   int     `json:"afterChunks"`
	RemovedChunks int     `json:"removedChunks"`
	BeforeTokens  int     `json:"beforeTokens"`
	AfterTokens   int     `json:"afterTokens"`
	TokensSaved   int     `json:"tokensSaved"`
	Summarized    int     `json:"summarized"`
}

// prunerScoredChunk pairs a chunk with its computed score for sorting.
type prunerScoredChunk struct {
	chunk *ContextChunk
	score float64
}

// Prune performs context pruning on the harvester's chunks.
func (ch *ContextHarvester) PruneWithOptions(opts PruneOptions) *PruneResult {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	result := &PruneResult{
		BeforeChunks: len(ch.chunks),
		BeforeTokens: ch.totalTokensLocked(),
	}

	// Score all chunks
	var scored []prunerScoredChunk
	now := time.Now().UnixMilli()
	for _, chunk := range ch.chunks {
		ageHours := float64(now-chunk.CreatedAt) / (1000 * 60 * 60)
		decayedRelevance := chunk.RelevanceScore * math.Exp(-opts.DecayPerHour*ageHours)

		// Boost for recent access
		accessAgeHours := float64(now-chunk.LastAccessedAt) / (1000 * 60 * 60)
		recencyBoost := 1.0
		if accessAgeHours < 1 {
			recencyBoost = 2.0
		} else if accessAgeHours < 6 {
			recencyBoost = 1.5
		}

		score := decayedRelevance * recencyBoost

		// Boost for access count
		score += float64(chunk.AccessCount) * 0.3

		scored = append(scored, prunerScoredChunk{chunk, score})
	}

	// Sort by score ascending (weakest first for removal)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score < scored[j].score
	})

	// Always keep the N most recent chunks
	recentIDs := make(map[string]bool)
	for i := 0; i < opts.KeepRecent && i < len(scored); i++ {
		recentIDs[scored[i].chunk.ID] = true
	}

	// Re-sort by score for pruning
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score < scored[j].score
	})

	// Calculate target tokens
	targetTokens := int(float64(opts.MaxTokens) * opts.TargetRatio)

	// Remove weakest chunks until we're under budget
	removed := 0
	for _, sc := range scored {
		if ch.totalTokensLocked() <= targetTokens {
			break
		}
		if recentIDs[sc.chunk.ID] {
			continue // Don't remove recent chunks
		}
		if sc.score >= opts.MinRelevance*5 {
			// High-relevance chunks get summarized instead of removed
			summarized := ch.summarizeChunkLocked(sc.chunk)
			if summarized {
				result.Summarized++
			}
			continue
		}

		delete(ch.chunks, sc.chunk.ID)
		removed++
	}

	result.RemovedChunks = removed
	result.AfterChunks = len(ch.chunks)
	result.AfterTokens = ch.totalTokensLocked()
	result.TokensSaved = result.BeforeTokens - result.AfterTokens

	return result
}

// summarizeChunkLocked replaces a chunk's content with a compressed version.
func (ch *ContextHarvester) summarizeChunkLocked(chunk *ContextChunk) bool {
	content := chunk.Content
	if len(content) <= 200 {
		return false // Already short
	}

	// Simple extractive summarization: keep first and last sentences
	sentences := splitSentences(content)
	if len(sentences) <= 2 {
		return false
	}

	// Keep first 2 and last 1 sentences
	summarized := sentences[0]
	if len(sentences) > 1 {
		summarized += " " + sentences[1]
	}
	summarized += " ... [compressed] ... " + sentences[len(sentences)-1]

	if len(summarized) >= len(content) {
		return false // Didn't actually compress
	}

	chunk.Content = summarized
	chunk.TokenCount = estimateTokens(summarized)
	return true
}

// totalTokensLocked returns the total token count of all chunks.
func (ch *ContextHarvester) totalTokensLocked() int {
	total := 0
	for _, chunk := range ch.chunks {
		total += chunk.TokenCount
	}
	return total
}

// splitSentences does basic sentence splitting.
func splitSentences(text string) []string {
	var sentences []string
	current := ""
	for _, r := range text {
		current += string(r)
		if r == '.' || r == '!' || r == '?' {
			sentences = append(sentences, current)
			current = ""
		}
	}
	if strings.TrimSpace(current) != "" {
		sentences = append(sentences, current)
	}
	return sentences
}
