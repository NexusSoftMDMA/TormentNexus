/**
 * NotebookLM-Style Citation Service
 *
 * Provides source-grounded, citation-backed answers by:
 * 1. Accepting a user query + a set of source documents
 * 2. Retrieving relevant chunks via vector similarity (LanceDB)
 * 3. Generating an answer with inline citations [1], [2], etc.
 * 4. Returning the answer + citation metadata for UI rendering
 *
 * This is the TormentNexus equivalent of NotebookLM's "Grounded Answers" feature.
 */

export interface CitationSource {
    id: string;
    title: string;
    content: string;
    url?: string;
    sourceType: 'document' | 'email' | 'web-memory' | 'bookmark' | 'file';
    metadata?: Record<string, unknown>;
}

export interface CitationChunk {
    sourceId: string;
    sourceTitle: string;
    chunkText: string;
    relevanceScore: number;
    startOffset?: number;
    endOffset?: number;
}

export interface InlineCitation {
    index: number;        // [1], [2], etc.
    sourceId: string;
    sourceTitle: string;
    excerpt: string;      // The exact passage cited
    url?: string;
}

export interface GroundedAnswer {
    answer: string;       // Text with inline references like [1], [2]
    citations: InlineCitation[];
    sourcesUsed: number;
    totalSourcesAvailable: number;
    confidence: number;   // 0-1 based on how well grounded the answer is
    generatedAt: Date;
}

export interface CitationServiceConfig {
    maxChunksPerQuery: number;
    chunkSize: number;
    chunkOverlap: number;
    minRelevanceScore: number;
    maxCitationsPerAnswer: number;
}

const DEFAULT_CONFIG: CitationServiceConfig = {
    maxChunksPerQuery: 20,
    chunkSize: 500,
    chunkOverlap: 50,
    minRelevanceScore: 0.3,
    maxCitationsPerAnswer: 10,
};

/**
 * Splits a document into overlapping chunks for embedding/retrieval.
 */
export function chunkDocument(content: string, chunkSize: number, overlap: number): string[] {
    const chunks: string[] = [];
    const words = content.split(/\s+/);

    if (words.length <= chunkSize) {
        return [content];
    }

    for (let i = 0; i < words.length; i += chunkSize - overlap) {
        const chunk = words.slice(i, i + chunkSize).join(' ');
        if (chunk.trim().length > 0) {
            chunks.push(chunk);
        }
        if (i + chunkSize >= words.length) break;
    }

    return chunks;
}

/**
 * Extracts citation references from an LLM-generated answer.
 * Looks for patterns like [1], [2], [Source 1], etc.
 */
export function extractCitationReferences(answer: string): number[] {
    const matches = answer.matchAll(/\[(\d+)\]/g);
    const refs = new Set<number>();
    for (const match of matches) {
        refs.add(parseInt(match[1], 10));
    }
    return Array.from(refs).sort((a, b) => a - b);
}

/**
 * Builds a grounded prompt that instructs the LLM to cite sources.
 */
export function buildGroundedPrompt(query: string, chunks: CitationChunk[]): string {
    const sourceContext = chunks
        .map((chunk, i) => `[Source ${i + 1}] (${chunk.sourceTitle}):\n${chunk.chunkText}`)
        .join('\n\n');

    return `You are a research assistant that provides accurate, source-grounded answers.

SOURCES:
${sourceContext}

INSTRUCTIONS:
- Answer the user's question using ONLY the information from the sources above.
- Cite your sources using numbered references like [1], [2], etc.
- If the sources don't contain enough information, say so explicitly.
- Keep your answer concise and well-structured.
- Every claim must be backed by at least one source citation.

USER QUESTION: ${query}

GROUNDED ANSWER:`;
}

export class CitationService {
    private config: CitationServiceConfig;
    private lancedbPath: string;

    constructor(config?: Partial<CitationServiceConfig>) {
        this.config = { ...DEFAULT_CONFIG, ...config };
        
        // Store vector DB locally in the global user .tormentnexus directory
        const os = require('os');
        this.lancedbPath = `${os.homedir()}/.tormentnexus/citations_db`;
    }

    getConfig(): CitationServiceConfig {
        return { ...this.config };
    }

    /**
     * Index sources — splits each source into chunks, generates embeddings,
     * and stores them in a LanceDB table for retrieval.
     */
    async indexSources(sessionId: string, sources: CitationSource[]): Promise<void> {
        // Dynamic imports for heavy ML / DB dependencies
        const { pipeline } = await import('@xenova/transformers');
        const lancedb = await import('@lancedb/lancedb');

        const embedder = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2');
        const db = await lancedb.connect(this.lancedbPath);

        const data: any[] = [];
        
        for (const source of sources) {
            const textChunks = chunkDocument(
                source.content,
                this.config.chunkSize,
                this.config.chunkOverlap,
            );

            for (let i = 0; i < textChunks.length; i++) {
                const chunkText = textChunks[i];
                const output = await embedder(chunkText, { pooling: 'mean', normalize: true });
                const vector = Array.from(output.data);
                
                data.push({
                    vector,
                    sourceId: source.id,
                    sourceTitle: source.title,
                    chunkText,
                    chunkIndex: i
                });
            }
        }
        
        if (data.length === 0) return;

        // Use a unique table per session/query batch to scope citations
        const tableName = `citations_${sessionId}`;
        try {
            await db.dropTable(tableName);
        } catch (e) { /* Ignore if it doesn't exist */ }
        
        await db.createTable(tableName, data);
    }
    
    /**
     * Retrieves the most relevant chunks from the indexed sources for a given query.
     */
    async retrieveRelevantChunks(sessionId: string, query: string): Promise<CitationChunk[]> {
        const { pipeline } = await import('@xenova/transformers');
        const lancedb = await import('@lancedb/lancedb');
        
        const embedder = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2');
        const db = await lancedb.connect(this.lancedbPath);
        
        const output = await embedder(query, { pooling: 'mean', normalize: true });
        const queryVector = Array.from(output.data) as number[];
        
        const tableName = `citations_${sessionId}`;
        try {
            const table = await db.openTable(tableName);
            const results = await table.search(queryVector)
                .limit(this.config.maxChunksPerQuery)
                .toArray();
                
            return results
                .filter(row => 1 - (row._distance || 0) >= this.config.minRelevanceScore)
                .map(row => ({
                    sourceId: row.sourceId as string,
                    sourceTitle: row.sourceTitle as string,
                    chunkText: row.chunkText as string,
                    relevanceScore: 1 - (row._distance || 0)
                }));
        } catch (e) {
            console.error("Failed to retrieve chunks from LanceDB:", e);
            return [];
        }
    }

    /**
     * Build citation metadata from an LLM answer and the chunks that were used.
     */
    buildCitations(answer: string, usedChunks: CitationChunk[], sources: CitationSource[]): InlineCitation[] {
        const refNumbers = extractCitationReferences(answer);
        const citations: InlineCitation[] = [];

        for (const refNum of refNumbers) {
            const chunkIndex = refNum - 1; // [1] → index 0
            if (chunkIndex >= 0 && chunkIndex < usedChunks.length) {
                const chunk = usedChunks[chunkIndex];
                const source = sources.find(s => s.id === chunk.sourceId);

                citations.push({
                    index: refNum,
                    sourceId: chunk.sourceId,
                    sourceTitle: chunk.sourceTitle,
                    excerpt: chunk.chunkText.substring(0, 200),
                    url: source?.url,
                });
            }
        }

        return citations;
    }

    /**
     * Generate a grounded answer with citations.
     */
    async generateGroundedAnswer(
        sessionId: string,
        query: string,
        sources: CitationSource[],
        llmCall: (prompt: string) => Promise<string>,
    ): Promise<GroundedAnswer> {
        
        // 1. Index the sources into the vector DB
        await this.indexSources(sessionId, sources);
        
        // 2. Retrieve relevant chunks
        const scoredChunks = await this.retrieveRelevantChunks(sessionId, query);
        
        // 3. Build prompt and call LLM
        const prompt = buildGroundedPrompt(query, scoredChunks);
        const rawAnswer = await llmCall(prompt);
        
        // 4. Build citations
        const citations = this.buildCitations(rawAnswer, scoredChunks, sources);
        
        // Ensure confidence score based on citation counts
        const confidence = scoredChunks.length > 0 ? (citations.length / scoredChunks.length) : 0;
        
        return {
            answer: rawAnswer,
            citations,
            sourcesUsed: citations.length,
            totalSourcesAvailable: sources.length,
            confidence: Math.min(Math.max(confidence, 0.1), 1.0), // Bound 0.1 - 1.0 based on citation density
            generatedAt: new Date()
        };
    }
}

export const citationService = new CitationService();
