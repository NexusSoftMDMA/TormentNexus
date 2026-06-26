/**
 * Context Harvesting Service (Phase I)
 *
 * Automatic context harvesting, pruning, compacting, reranking, and semantic chunking.
 * This service manages the lifecycle of context windows for LLM interactions.
 *
 * Key capabilities:
 * - Harvest context from active files, terminal output, git diffs, and conversation history
 * - Prune low-relevance context to stay within token budgets
 * - Compact repetitive or stale information
 * - Rerank context by relevance to the current task
 * - Semantic chunking for optimal retrieval
 */

export interface ContextChunk {
    id: string;
    source: ContextSource;
    content: string;
    relevanceScore: number;
    tokenCount: number;
    createdAt: number;
    lastAccessedAt: number;
    accessCount: number;
    metadata: Record<string, unknown>;
}

export type ContextSource =
    | 'active-file'
    | 'terminal-output'
    | 'git-diff'
    | 'conversation'
    | 'memory'
    | 'web-search'
    | 'documentation'
    | 'error-log'
    | 'test-output';

export interface HarvestConfig {
    maxTokenBudget: number;
    pruneThreshold: number;       // Below this relevance score, prune
    compactAfterMs: number;       // Compact chunks older than this
    rerankOnAccess: boolean;      // Re-score on every access
    chunkSize: number;            // Target words per chunk
    chunkOverlap: number;         // Overlap between adjacent chunks
    maxChunksPerSource: number;
    decayRatePerHour: number;     // Relevance decay over time
}

export interface HarvestReport {
    totalChunks: number;
    totalTokens: number;
    pruned: number;
    compacted: number;
    reranked: number;
    sourceBreakdown: Record<string, number>;
    budgetUtilization: number;    // 0-1
}

const DEFAULT_CONFIG: HarvestConfig = {
    maxTokenBudget: 128_000,
    pruneThreshold: 0.1,
    compactAfterMs: 30 * 60 * 1000, // 30 minutes
    rerankOnAccess: true,
    chunkSize: 300,
    chunkOverlap: 30,
    maxChunksPerSource: 50,
    decayRatePerHour: 0.05,
};

/**
 * Estimate token count from text (rough: ~4 chars per token).
 */
function estimateTokens(text: string): number {
    return Math.ceil(text.length / 4);
}

/**
 * Split text into semantic chunks with overlap.
 */
function semanticChunk(text: string, chunkSize: number, overlap: number): string[] {
    const sentences = text.split(/(?<=[.!?])\s+/);
    const chunks: string[] = [];
    let currentChunk: string[] = [];
    let wordCount = 0;

    for (const sentence of sentences) {
        const sentenceWords = sentence.split(/\s+/).length;

        if (wordCount + sentenceWords > chunkSize && currentChunk.length > 0) {
            chunks.push(currentChunk.join(' '));

            // Keep overlap sentences
            const overlapSentences: string[] = [];
            let overlapWords = 0;
            for (let i = currentChunk.length - 1; i >= 0 && overlapWords < overlap; i--) {
                overlapSentences.unshift(currentChunk[i]);
                overlapWords += currentChunk[i].split(/\s+/).length;
            }

            currentChunk = [...overlapSentences];
            wordCount = overlapWords;
        }

        currentChunk.push(sentence);
        wordCount += sentenceWords;
    }

    if (currentChunk.length > 0) {
        chunks.push(currentChunk.join(' '));
    }

    return chunks;
}

export class ContextHarvester {
    private config: HarvestConfig;
    private chunks: Map<string, ContextChunk> = new Map();
    private compactTimer: ReturnType<typeof setInterval> | null = null;

    constructor(config?: Partial<HarvestConfig>) {
        this.config = { ...DEFAULT_CONFIG, ...config };
    }

    getConfig(): HarvestConfig {
        return { ...this.config };
    }

    /**
     * Harvest content from a source — splits into chunks and stores.
     */
    harvest(source: ContextSource, content: string, metadata: Record<string, unknown> = {}): ContextChunk[] {
        const textChunks = semanticChunk(content, this.config.chunkSize, this.config.chunkOverlap);
        const harvested: ContextChunk[] = [];

        // Enforce per-source limit
        const existingFromSource = Array.from(this.chunks.values())
            .filter(c => c.source === source);

        if (existingFromSource.length >= this.config.maxChunksPerSource) {
            // Prune lowest-relevance chunks from this source
            existingFromSource
                .sort((a, b) => a.relevanceScore - b.relevanceScore)
                .slice(0, existingFromSource.length - this.config.maxChunksPerSource + textChunks.length)
                .forEach(c => this.chunks.delete(c.id));
        }

        for (const text of textChunks) {
            const chunk: ContextChunk = {
                id: `ctx_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
                source,
                content: text,
                relevanceScore: 1.0, // Fresh content starts at max relevance
                tokenCount: estimateTokens(text),
                createdAt: Date.now(),
                lastAccessedAt: Date.now(),
                accessCount: 0,
                metadata,
            };

            this.chunks.set(chunk.id, chunk);
            harvested.push(chunk);
        }

        return harvested;
    }

    /**
     * Retrieve context for a query — reranks and returns top chunks within budget.
     */
    retrieve(query: string, maxTokens?: number): ContextChunk[] {
        const budget = maxTokens ?? this.config.maxTokenBudget;
        const queryWords = new Set(query.toLowerCase().split(/\s+/));

        // Score each chunk against query
        const scored = Array.from(this.chunks.values()).map(chunk => {
            let score = chunk.relevanceScore;

            // Keyword relevance boost
            const chunkWords = chunk.content.toLowerCase().split(/\s+/);
            const overlap = chunkWords.filter(w => queryWords.has(w)).length;
            score += (overlap / queryWords.size) * 0.5;

            // Time decay
            const hoursOld = (Date.now() - chunk.createdAt) / (1000 * 60 * 60);
            score *= Math.max(0.1, 1 - hoursOld * this.config.decayRatePerHour);

            // Recency boost
            const hoursSinceAccess = (Date.now() - chunk.lastAccessedAt) / (1000 * 60 * 60);
            if (hoursSinceAccess < 1) score *= 1.2;

            // Access frequency boost (diminishing returns)
            score *= 1 + Math.log(1 + chunk.accessCount) * 0.1;

            return { ...chunk, relevanceScore: score };
        });

        // Sort by relevance, then fill budget
        scored.sort((a, b) => b.relevanceScore - a.relevanceScore);

        const result: ContextChunk[] = [];
        let usedTokens = 0;

        for (const chunk of scored) {
            if (usedTokens + chunk.tokenCount > budget) continue;
            if (chunk.relevanceScore < this.config.pruneThreshold) break;

            chunk.lastAccessedAt = Date.now();
            chunk.accessCount++;
            this.chunks.set(chunk.id, chunk);

            result.push(chunk);
            usedTokens += chunk.tokenCount;
        }

        return result;
    }

    /**
     * Prune chunks below the relevance threshold.
     */
    prune(): number {
        let pruned = 0;
        for (const [id, chunk] of this.chunks) {
            const hoursOld = (Date.now() - chunk.createdAt) / (1000 * 60 * 60);
            const decayedScore = chunk.relevanceScore * Math.max(0.1, 1 - hoursOld * this.config.decayRatePerHour);

            if (decayedScore < this.config.pruneThreshold) {
                this.chunks.delete(id);
                pruned++;
            }
        }
        return pruned;
    }

    /**
     * Compact old chunks — merge adjacent chunks from same source.
     */
    getChunks(): ContextChunk[] {
        return Array.from(this.chunks.values()).sort((a, b) => b.createdAt - a.createdAt);
    }
    compact(): number {
        const now = Date.now();
        let compacted = 0;
        const bySource = new Map<string, ContextChunk[]>();

        for (const chunk of this.chunks.values()) {
            if (now - chunk.createdAt < this.config.compactAfterMs) continue;

            const key = chunk.source;
            if (!bySource.has(key)) bySource.set(key, []);
            bySource.get(key)!.push(chunk);
        }

        for (const [, chunks] of bySource) {
            if (chunks.length < 2) continue;

            // Merge all old chunks from same source into one
            chunks.sort((a, b) => a.createdAt - b.createdAt);
            const merged = chunks.map(c => c.content).join('\n\n');
            const mergedChunk: ContextChunk = {
                id: `ctx_merged_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
                source: chunks[0].source,
                content: merged.substring(0, 10_000), // Cap at 10K chars
                relevanceScore: Math.max(...chunks.map(c => c.relevanceScore)),
                tokenCount: estimateTokens(merged.substring(0, 10_000)),
                createdAt: chunks[0].createdAt,
                lastAccessedAt: Math.max(...chunks.map(c => c.lastAccessedAt)),
                accessCount: chunks.reduce((sum, c) => sum + c.accessCount, 0),
                metadata: { mergedFrom: chunks.length },
            };

            // Remove old, add merged
            for (const chunk of chunks) this.chunks.delete(chunk.id);
            this.chunks.set(mergedChunk.id, mergedChunk);
            compacted += chunks.length - 1;
        }

        return compacted;
    }

    /**
     * Get a report on the current context window state.
     */
    getReport(): HarvestReport {
        const sourceBreakdown: Record<string, number> = {};
        let totalTokens = 0;

        for (const chunk of this.chunks.values()) {
            sourceBreakdown[chunk.source] = (sourceBreakdown[chunk.source] || 0) + 1;
            totalTokens += chunk.tokenCount;
        }

        return {
            totalChunks: this.chunks.size,
            totalTokens,
            pruned: 0,
            compacted: 0,
            reranked: 0,
            sourceBreakdown,
            budgetUtilization: totalTokens / this.config.maxTokenBudget,
        };
    }

    /**
     * Start automatic compaction on a timer.
     */
    startAutoCompaction(): void {
        if (this.compactTimer) return;
        this.compactTimer = setInterval(() => {
            this.prune();
            this.compact();
        }, this.config.compactAfterMs);
    }

    stopAutoCompaction(): void {
        if (this.compactTimer) {
            clearInterval(this.compactTimer);
            this.compactTimer = null;
        }
    }

    clear(): void {
        this.chunks.clear();
    }

    cleanup(): void {
        this.stopAutoCompaction();
        this.clear();
    }
}

export const contextHarvester = new ContextHarvester();
