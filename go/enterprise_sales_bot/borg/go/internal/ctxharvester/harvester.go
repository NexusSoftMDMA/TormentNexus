// Package ctxharvester provides automatic context harvesting, pruning,
// compacting, reranking, and semantic chunking — ported from
// packages/core/src/services/ContextHarvester.ts.
//
// Key capabilities:
//   - Harvest context from active files, terminal output, git diffs, conversation, etc.
//   - Prune low-relevance context to stay within token budgets
//   - Compact repetitive or stale information
//   - Rerank context by relevance to the current query
//   - Semantic chunking for optimal retrieval
package ctxharvester

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// nextChunkSeq ensures unique chunk IDs even within the same millisecond.
var nextChunkSeq atomic.Int64

// ContextSource identifies the origin of a context chunk.
type ContextSource string

const (
	SourceActiveFile    ContextSource = "active-file"
	SourceTerminal      ContextSource = "terminal-output"
	SourceGitDiff       ContextSource = "git-diff"
	SourceConversation  ContextSource = "conversation"
	SourceMemory        ContextSource = "memory"
	SourceWebSearch     ContextSource = "web-search"
	SourceDocumentation ContextSource = "documentation"
	SourceErrorLog      ContextSource = "error-log"
	SourceTestOutput    ContextSource = "test-output"
)

// ContextChunk is a single unit of harvested context.
type ContextChunk struct {
	ID             string         `json:"id"`
	Source         ContextSource  `json:"source"`
	Content        string         `json:"content"`
	RelevanceScore float64        `json:"relevanceScore"`
	TokenCount     int            `json:"tokenCount"`
	CreatedAt      int64          `json:"createdAt"`
	LastAccessedAt int64          `json:"lastAccessedAt"`
	AccessCount    int            `json:"accessCount"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// HarvestConfig controls harvesting behavior.
type HarvestConfig struct {
	MaxTokenBudget    int     `json:"maxTokenBudget"`
	PruneThreshold    float64 `json:"pruneThreshold"`
	CompactAfterMs    int64   `json:"compactAfterMs"`
	RerankOnAccess    bool    `json:"rerankOnAccess"`
	ChunkSize         int     `json:"chunkSize"`       // target words per chunk
	ChunkOverlap      int     `json:"chunkOverlap"`
	MaxChunksPerSource int    `json:"maxChunksPerSource"`
	DecayRatePerHour  float64 `json:"decayRatePerHour"`
}

// HarvestReport summarizes a harvesting pass.
type HarvestReport struct {
	TotalChunks       int           `json:"totalChunks"`
	TotalTokens       int           `json:"totalTokens"`
	Pruned            int           `json:"pruned"`
	Compacted         int           `json:"compacted"`
	Reranked          int           `json:"reranked"`
	SourceBreakdown   map[string]int `json:"sourceBreakdown"`
	BudgetUtilization float64       `json:"budgetUtilization"`
}

// DefaultHarvestConfig returns the default configuration.
func DefaultHarvestConfig() HarvestConfig {
	return HarvestConfig{
		MaxTokenBudget:    128000,
		PruneThreshold:    0.1,
		CompactAfterMs:    30 * 60 * 1000, // 30 min
		RerankOnAccess:    true,
		ChunkSize:         300,
		ChunkOverlap:      30,
		MaxChunksPerSource: 50,
		DecayRatePerHour:  0.05,
	}
}

// ContextHarvester manages the lifecycle of context windows for LLM interactions.
type ContextHarvester struct {
	mu     sync.RWMutex
	config HarvestConfig
	chunks map[string]*ContextChunk
	stopCh chan struct{}
}

// NewContextHarvester creates a new harvester with optional config overrides.
func NewContextHarvester(cfg *HarvestConfig) *ContextHarvester {
	c := DefaultHarvestConfig()
	if cfg != nil {
		// Merge overrides
		if cfg.MaxTokenBudget > 0 {
			c.MaxTokenBudget = cfg.MaxTokenBudget
		}
		if cfg.PruneThreshold > 0 {
			c.PruneThreshold = cfg.PruneThreshold
		}
		if cfg.CompactAfterMs > 0 {
			c.CompactAfterMs = cfg.CompactAfterMs
		}
		if cfg.ChunkSize > 0 {
			c.ChunkSize = cfg.ChunkSize
		}
		if cfg.ChunkOverlap > 0 {
			c.ChunkOverlap = cfg.ChunkOverlap
		}
		if cfg.MaxChunksPerSource > 0 {
			c.MaxChunksPerSource = cfg.MaxChunksPerSource
		}
		if cfg.DecayRatePerHour > 0 {
			c.DecayRatePerHour = cfg.DecayRatePerHour
		}
		c.RerankOnAccess = cfg.RerankOnAccess
	}

	return &ContextHarvester{
		config: c,
		chunks: make(map[string]*ContextChunk),
		stopCh: make(chan struct{}),
	}
}

// GetConfig returns the current harvest config.
func (ch *ContextHarvester) GetConfig() HarvestConfig {
	return ch.config
}

var sentenceRe = regexp.MustCompile(`(?m)(.+?[.!?])\s+`)

// Harvest splits content into semantic chunks and stores them.
func (ch *ContextHarvester) Harvest(source ContextSource, content string, metadata map[string]interface{}) []*ContextChunk {
	textChunks := semanticChunk(content, ch.config.ChunkSize, ch.config.ChunkOverlap)

	ch.mu.Lock()
	defer ch.mu.Unlock()

	// Enforce per-source limit
	var existingFromSource []*ContextChunk
	for _, c := range ch.chunks {
		if c.Source == source {
			existingFromSource = append(existingFromSource, c)
		}
	}

	// Prune lowest-relevance if over limit
	overhead := (len(existingFromSource) + len(textChunks)) - ch.config.MaxChunksPerSource
	if overhead > 0 {
		// Sort existing by relevance ascending, remove the weakest
		sorted := make([]*ContextChunk, len(existingFromSource))
		copy(sorted, existingFromSource)
		sortChunksByRelevance(sorted)
		for i := 0; i < overhead && i < len(sorted); i++ {
			delete(ch.chunks, sorted[i].ID)
		}
	}

	harvested := make([]*ContextChunk, 0, len(textChunks))
	for _, text := range textChunks {
		chunk := &ContextChunk{
			ID:             fmt.Sprintf("ctx_%d_%d_%s", time.Now().UnixMilli(), nextChunkSeq.Add(1), randomSuffix(6)),
			Source:         source,
			Content:        text,
			RelevanceScore: 1.0,
			TokenCount:     estimateTokens(text),
			CreatedAt:      time.Now().UnixMilli(),
			LastAccessedAt: time.Now().UnixMilli(),
			AccessCount:    0,
			Metadata:       metadata,
		}
		ch.chunks[chunk.ID] = chunk
		harvested = append(harvested, chunk)
	}

	return harvested
}

// Retrieve returns the best chunks for a query within the token budget.
func (ch *ContextHarvester) Retrieve(query string, maxTokens int) []*ContextChunk {
	if maxTokens <= 0 {
		maxTokens = ch.config.MaxTokenBudget
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	queryWords := wordSet(strings.ToLower(query))

	// Score each chunk and collect
	type scoredEntry struct {
		chunk *ContextChunk
		score float64
	}
	var scoredChunks []scoredEntry
	for _, c := range ch.chunks {
		score := c.RelevanceScore

		// Keyword relevance boost
		chunkWords := wordSet(strings.ToLower(c.Content))
		overlap := 0
		for w := range queryWords {
			if chunkWords[w] {
				overlap++
			}
		}
		if len(queryWords) > 0 {
			score += (float64(overlap) / float64(len(queryWords))) * 0.5
		}

		// Time decay
		hoursOld := float64(time.Now().UnixMilli()-c.CreatedAt) / (1000 * 60 * 60)
		score *= math.Max(0.1, 1-hoursOld*ch.config.DecayRatePerHour)

		// Recency boost
		hoursSinceAccess := float64(time.Now().UnixMilli()-c.LastAccessedAt) / (1000 * 60 * 60)
		if hoursSinceAccess < 1 {
			score *= 1.2
		}

		// Access frequency boost (diminishing returns)
		score *= 1 + math.Log(1+float64(c.AccessCount))*0.1

		scoredChunks = append(scoredChunks, scoredEntry{chunk: c, score: score})
	}

	// Sort by score descending (inline insertion sort)
	for i := 1; i < len(scoredChunks); i++ {
		for j := i; j > 0 && scoredChunks[j].score > scoredChunks[j-1].score; j-- {
			scoredChunks[j], scoredChunks[j-1] = scoredChunks[j-1], scoredChunks[j]
		}
	}

	var result []*ContextChunk
	usedTokens := 0
	now := time.Now().UnixMilli()

	for _, sc := range scoredChunks {
		if usedTokens+sc.chunk.TokenCount > maxTokens {
			continue
		}
		if sc.score < ch.config.PruneThreshold {
			break
		}
		sc.chunk.LastAccessedAt = now
		sc.chunk.AccessCount++
		sc.chunk.RelevanceScore = sc.score
		result = append(result, sc.chunk)
		usedTokens += sc.chunk.TokenCount
	}

	return result
}

// Prune removes chunks below the relevance threshold. Returns count pruned.
func (ch *ContextHarvester) Prune() int {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	pruned := 0
	now := time.Now().UnixMilli()
	for id, c := range ch.chunks {
		hoursOld := float64(now-c.CreatedAt) / (1000 * 60 * 60)
		decayed := c.RelevanceScore * math.Max(0.1, 1-hoursOld*ch.config.DecayRatePerHour)
		if decayed < ch.config.PruneThreshold {
			delete(ch.chunks, id)
			pruned++
		}
	}
	return pruned
}

// Compact merges adjacent old chunks from the same source. Returns count compacted.
func (ch *ContextHarvester) Compact() int {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	now := time.Now().UnixMilli()
	bySource := make(map[ContextSource][]*ContextChunk)

	for _, c := range ch.chunks {
		if now-c.CreatedAt < ch.config.CompactAfterMs {
			continue
		}
		bySource[c.Source] = append(bySource[c.Source], c)
	}

	compacted := 0
	for _, chunks := range bySource {
		if len(chunks) < 2 {
			continue
		}

		// Sort by createdAt
		sortChunksByAge(chunks)

		// Merge content
		var merged strings.Builder
		for i, c := range chunks {
			if i > 0 {
				merged.WriteString("\n\n")
			}
			merged.WriteString(c.Content)
		}
		mergedContent := merged.String()
		if len(mergedContent) > 10000 {
			mergedContent = mergedContent[:10000]
		}

		// Find max relevance
		maxRel := 0.0
		for _, c := range chunks {
			if c.RelevanceScore > maxRel {
				maxRel = c.RelevanceScore
			}
		}

		// Find max lastAccessedAt
		maxAccess := int64(0)
		totalAccess := 0
		for _, c := range chunks {
			if c.LastAccessedAt > maxAccess {
				maxAccess = c.LastAccessedAt
			}
			totalAccess += c.AccessCount
		}

		// Remove old, add merged
		for _, c := range chunks {
			delete(ch.chunks, c.ID)
		}

		newChunk := &ContextChunk{
			ID:             fmt.Sprintf("ctx_merged_%d_%s", now, randomSuffix(6)),
			Source:         chunks[0].Source,
			Content:        mergedContent,
			RelevanceScore: maxRel,
			TokenCount:     estimateTokens(mergedContent),
			CreatedAt:      chunks[0].CreatedAt,
			LastAccessedAt: maxAccess,
			AccessCount:    totalAccess,
			Metadata:       map[string]interface{}{"mergedFrom": len(chunks)},
		}
		ch.chunks[newChunk.ID] = newChunk
		compacted += len(chunks) - 1
	}

	return compacted
}

// GetChunks returns all chunks sorted by creation time descending.
func (ch *ContextHarvester) GetChunks() []*ContextChunk {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	result := make([]*ContextChunk, 0, len(ch.chunks))
	for _, c := range ch.chunks {
		result = append(result, c)
	}
	sortChunksByAgeDesc(result)
	return result
}

// GetReport returns a summary of the current context state.
func (ch *ContextHarvester) GetReport() *HarvestReport {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	breakdown := make(map[string]int)
	totalTokens := 0
	for _, c := range ch.chunks {
		breakdown[string(c.Source)]++
		totalTokens += c.TokenCount
	}

	return &HarvestReport{
		TotalChunks:       len(ch.chunks),
		TotalTokens:       totalTokens,
		SourceBreakdown:   breakdown,
		BudgetUtilization: float64(totalTokens) / float64(ch.config.MaxTokenBudget),
	}
}

// StartAutoCompaction begins periodic pruning and compaction.
func (ch *ContextHarvester) StartAutoCompaction() {
	interval := time.Duration(ch.config.CompactAfterMs) * time.Millisecond
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ch.Prune()
				ch.Compact()
			case <-ch.stopCh:
				return
			}
		}
	}()
}

// StopAutoCompaction stops the background compaction goroutine.
func (ch *ContextHarvester) StopAutoCompaction() {
	close(ch.stopCh)
}

// Clear removes all chunks.
func (ch *ContextHarvester) Clear() {
	ch.mu.Lock()
	ch.chunks = make(map[string]*ContextChunk)
	ch.mu.Unlock()
}

// Cleanup stops compaction and clears all chunks.
func (ch *ContextHarvester) Cleanup() {
	ch.StopAutoCompaction()
	ch.Clear()
}

// --- helpers ---

func estimateTokens(text string) int {
	return (len(text) + 3) / 4
}

func semanticChunk(text string, chunkSize, overlap int) []string {
	sentences := sentenceRe.Split(text, -1)
	// Clean up empty strings
	var clean []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			clean = append(clean, s)
		}
	}
	if len(clean) == 0 {
		return nil
	}

	var chunks []string
	var currentChunk []string
	wordCount := 0

	for _, sentence := range clean {
		words := len(strings.Fields(sentence))

		if wordCount+words > chunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, strings.Join(currentChunk, " "))

			// Keep overlap sentences
			var overlapSentences []string
			overlapWords := 0
			for i := len(currentChunk) - 1; i >= 0 && overlapWords < overlap; i-- {
				overlapSentences = append([]string{currentChunk[i]}, overlapSentences...)
				overlapWords += len(strings.Fields(currentChunk[i]))
			}
			currentChunk = overlapSentences
			wordCount = overlapWords
		}

		currentChunk = append(currentChunk, sentence)
		wordCount += words
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

func wordSet(s string) map[string]bool {
	words := strings.Fields(s)
	set := make(map[string]bool, len(words))
	for _, w := range words {
		set[w] = true
	}
	return set
}

func sortChunksByRelevance(chunks []*ContextChunk) {
	for i := 1; i < len(chunks); i++ {
		for j := i; j > 0 && chunks[j].RelevanceScore < chunks[j-1].RelevanceScore; j-- {
			chunks[j], chunks[j-1] = chunks[j-1], chunks[j]
		}
	}
}

func sortChunksByAge(chunks []*ContextChunk) {
	for i := 1; i < len(chunks); i++ {
		for j := i; j > 0 && chunks[j].CreatedAt < chunks[j-1].CreatedAt; j-- {
			chunks[j], chunks[j-1] = chunks[j-1], chunks[j]
		}
	}
}

func sortChunksByAgeDesc(chunks []*ContextChunk) {
	for i := 1; i < len(chunks); i++ {
		for j := i; j > 0 && chunks[j].CreatedAt > chunks[j-1].CreatedAt; j-- {
			chunks[j], chunks[j-1] = chunks[j-1], chunks[j]
		}
	}
}

type scoredChunk struct {
	chunk *ContextChunk
	score float64
}

func randomSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[(time.Now().UnixNano()+int64(i)*7)%int64(len(chars))]
	}
	return string(b)
}

// Ensure unused import check passes
var _ = 0
