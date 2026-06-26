
import { LLMService } from '@tormentnexus/ai';
import { SearchService } from '@tormentnexus/search';
import { MemoryManager } from './MemoryManager.js';
import type { MCPServer } from '../MCPServer.js';

/**
 * Minimal server interface — only the methods DeepResearchService actually needs.
 * Uses structural typing so we don't need a full MCPServer reference (avoids circular deps).
 */
type DeepResearchServer = {
    executeTool: MCPServer['executeTool'];
    wssInstance?: MCPServer['wssInstance'];
};

/** WebSocket client shape for broadcasting research progress events */
type WebSocketLikeClient = {
    readyState: number;
    send(data: string): void;
};

/** Individual search result entry from WebSearchTool */
interface SearchResultEntry {
    title?: string;
    url?: string;
}

/** Structured output of a research operation */
export interface ResearchResult {
    topic: string;
    summary: string;
    sources: { title: string, url: string, keyPoints: string[] }[];
    relatedTopics: string[];
    // Phase 96: Execution Telemetry
    modelMetadata?: {
        provider: string;
        modelId: string;
    };
}

/**
 * DeepResearchService
 *
 * The core research engine powering ResearcherAgent and the /research tRPC route.
 *
 * Pipeline:
 * 1. **Query Generation** — LLM generates 3 targeted search queries from the topic
 * 2. **Web Search** — WebSearchTool (DuckDuckGo via duck-duck-scrape) executes queries
 * 3. **Source Collection** — Top results are gathered with titles and URLs
 * 4. **LLM Synthesis** — A high-complexity model synthesizes findings into a structured report
 * 5. **Memory Storage** — Report is saved to MemoryManager for future context retrieval
 * 6. **Recursive Expansion** — Related topics are explored recursively (depth/breadth controlled)
 *
 * Dependencies:
 * - `LLMService` — for query generation and synthesis (uses ModelSelector for model choice)
 * - `MemoryManager` — for persisting research reports as retrievable context
 * - `WebSearchTool` — dynamically imported from `../tools/WebSearchTool.js`
 * - `MCPServer` (via DeepResearchServer) — for `ingest()` tool execution and WebSocket broadcasting
 *
 * Note: `SearchService` (local ripgrep) is accepted in constructor but intentionally unused.
 * Web search is handled by `WebSearchTool` which provides external internet access.
 */
export class DeepResearchService {
    /** LLM service for query generation and report synthesis */
    private llm: LLMService;
    /** Memory system for persisting research reports */
    private memory: MemoryManager;
    /** Server reference for tool execution (ingest) and WebSocket broadcasting */
    private server: DeepResearchServer;

    constructor(server: DeepResearchServer, llm: LLMService, _search: SearchService, memory: MemoryManager) {
        this.server = server;
        this.llm = llm;
        // Note: SearchService is local ripgrep — not used for web research.
        // Web search is handled by dynamically importing WebSearchTool.
        this.memory = memory;
    }

    public async recursiveResearch(topic: string, depth: number = 2, maxBreadth: number = 3): Promise<ResearchResult & { subTopics?: ResearchResult[] }> {
        console.log(`[DeepResearch] 🔄 Recursive Research: ${topic} (Depth Remaining: ${depth})`);

        // 1. Research current topic
        const result = await this.researchTopic(topic, 2); // Standard depth for individual node

        if (depth <= 0) {
            return result;
        }

        // 2. Recurse into related topics
        const subTopics: ResearchResult[] = [];
        const nextDepth = depth - 1;

        // Take top N related topics
        const branches = result.relatedTopics.slice(0, maxBreadth);

        for (const subTopic of branches) {
            // Check if we've already researched this recently? (Memory check would go here)
            // For now, just recurse
            try {
                const subResult = await this.recursiveResearch(subTopic, nextDepth, maxBreadth);
                subTopics.push(subResult);
            } catch (e) {
                console.error(`[DeepResearch] Failed to research sub-topic: ${subTopic}`, e);
            }
        }

        return { ...result, subTopics };
    }

    public async researchTopic(topic: string, depth: number = 2): Promise<ResearchResult> {
        console.log(`[DeepResearch] Starting research on: ${topic} (Depth: ${depth})`);

        // 1. Generate Search Queries
        const queries = await this.generateQueries(topic);
        console.log(`[DeepResearch] Generated queries: ${queries.join(', ')}`);

        // 2. Execute Search & Gather URLs
        const allUrls = new Set<string>();
        const sources: { title: string, url: string, content?: string }[] = [];

        try {
            // Import WebSearchTool from local core tools
            const { WebSearchTool } = await import('../tools/WebSearchTool.js');

            for (const q of queries.slice(0, 3)) {
                try {
                    const result = await WebSearchTool.handler({ query: q });
                    const text = result.content[0].text;

                    let parsed: unknown[] = [];
                    try {
                        parsed = JSON.parse(text);
                    } catch {
                        parsed = [];
                    }

                    if (Array.isArray(parsed)) {
                        parsed.slice(0, 3).forEach((r: unknown) => {
                            const entry = r as SearchResultEntry;
                            if (entry.url) {
                                allUrls.add(entry.url);
                                sources.push({ title: entry.title || '', url: entry.url });
                            }
                        });
                    }
                } catch (e) {
                    console.error("Search failed for", q, e);
                }
            }
        } catch (e) {
            console.error("Failed to load WebSearchTool", e);
        }

        console.log(`[DeepResearch] found ${allUrls.size} unique sources.`);

        // 3. Visit & Scrape
        const scrapedData: string[] = [];
        let count = 0;
        for (const source of sources) {
            if (count >= 3) break;
            // Capture snippet if available (we assume title/url is good enough for now/summary)
            scrapedData.push(`Source: ${source.title} (${source.url})\n(Content: [Pending scrape, utilizing snippet context])`);
            count++;
        }

        if (scrapedData.length === 0) {
            scrapedData.push("No sources found. Synthesizing from internal knowledge.");
        }

        // 4. Synthesize
        const synthesis = await this.synthesize(topic, scrapedData.join('\n\n'));

        // 5. Memorize
        await this.memory.saveContext(`Research Report: ${topic}\n\n${synthesis.summary}`, {
            source: 'DeepResearchService',
            title: `Research: ${topic}`,
            type: 'research_report',
            tags: ['research', topic, ...synthesis.relatedTopics]
        });

        return synthesis;
    }

    private async generateQueries(topic: string): Promise<string[]> {
        const prompt = `Generate 3 specific web search queries to deeply research the topic: "${topic}". Return only the queries, one per line.`;

        try {
            const model = await this.llm.modelSelector.selectModel({ taskComplexity: 'low', routingTaskType: 'research' });
            const response = await this.llm.generateText(
                model.provider,
                model.modelId,
                "You are a research assistant.",
                prompt,
                {
                    taskComplexity: 'low',
                    routingTaskType: 'research',
                }
            );
            return response.content.split('\n').filter(l => l.trim().length > 0);
        } catch (e) {
            return [topic, `${topic} analysis`, `${topic} latest news`];
        }
    }

    private async synthesize(topic: string, rawData: string): Promise<ResearchResult> {
        const prompt = `
            You are a Research Scholar.
            Topic: ${topic}
            
            Raw Data:
            ${rawData.substring(0, 10000)}
            
            Task:
            1. Summarize the key findings.
            2. List the sources used.
            3. Suggest 3 related topics for further research.
            
            Format as JSON:
            {
                "topic": "${topic}",
                "summary": "...",
                "sources": [ { "title": "...", "url": "...", "keyPoints": [] } ],
                "relatedTopics": []
            }
        `;

        try {
            const model = await this.llm.modelSelector.selectModel({ taskComplexity: 'high', routingTaskType: 'research' });
            const response = await this.llm.generateText(
                model.provider,
                model.modelId,
                "You are a helpful research assistant. Output valid JSON only.",
                prompt,
                {
                    taskComplexity: 'high',
                    routingTaskType: 'research',
                }
            );
            const cleanJson = response.content.replace(/```json\n?|\n?```/g, '').trim();
            const result = JSON.parse(cleanJson) as ResearchResult;

            // Ensure relatedTopics is present
            if (!result.relatedTopics) result.relatedTopics = [];

            // Phase 96: Execution Telemetry
            result.modelMetadata = {
                provider: model.provider,
                modelId: model.modelId
            };

            return result;
        } catch (e) {
            return {
                topic,
                summary: "Failed to synthesize JSON.",
                sources: [],
                relatedTopics: []
            };
        }
    }

    // --- Ingestion & Broadcasting (Migrated from ResearchService) ---

    /**
     * Directly ingest a specific URL into memory
     */
    public async ingest(url: string): Promise<string> {
        console.log(`[DeepResearch] Ingesting: ${url}`);
        this.broadcast('RESEARCH_UPDATE', {
            status: 'reading',
            target: url,
            url: url
        });

        try {
            await this.server.executeTool("navigate", { url });
            const result = await this.server.executeTool("read_page", { url });
            const contentText = result.content?.[0]?.text || "";

            if (contentText.startsWith("Error")) {
                this.broadcast('RESEARCH_UPDATE', { status: 'error', target: url, error: contentText });
                return `FAILED: ${contentText}`;
            }

            const ctxId = await this.memory.saveContext(
                `INGESTED SOURCE: ${url}\n\n${contentText}`,
                {
                    title: url,
                    source: url,
                    type: 'research'
                }
            );

            this.broadcast('RESEARCH_UPDATE', { status: 'memorized', target: url, id: ctxId });
            return `MEMORIZED: ${url} (ID: ${ctxId})`;

        } catch (e: unknown) {
            const message = e instanceof Error ? e.message : String(e);
            this.broadcast('RESEARCH_UPDATE', { status: 'error', target: url, error: message });
            return `ERROR: ${message}`;
        }
    }

    private broadcast(type: string, payload: Record<string, unknown>) {
        if (this.server.wssInstance && this.server.wssInstance.clients) {
            this.server.wssInstance.clients.forEach((client: WebSocketLikeClient) => {
                if (client.readyState === 1) {
                    client.send(JSON.stringify({ type, payload }));
                }
            });
        }
    }
}

