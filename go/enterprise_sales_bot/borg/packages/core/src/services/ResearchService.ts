
import { MemoryManager } from '../services/MemoryManager.js';
import type { MCPServer } from '../MCPServer.js';
import { search, SafeSearchType } from 'duck-duck-scrape';

type ToolTextEnvelope = {
    content?: Array<{ text?: string }>;
};

type BroadcastPayload = Record<string, unknown>;

type WebSocketLikeClient = {
    readyState: number;
    send(data: string): void;
};

function getFirstTextContent(value: unknown): string {
    if (!value || typeof value !== 'object') {
        return '';
    }

    const envelope = value as ToolTextEnvelope;
    const first = envelope.content?.[0];
    return typeof first?.text === 'string' ? first.text : '';
}

function getErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message;
    }

    return typeof error === 'string' ? error : String(error);
}

/**
 * Deep Research Service
 * Autonomously research a topic by searching, reading, and memorizing.
 */
export class ResearchService {
    private memory: MemoryManager;
    private server: MCPServer;
    private visitedUrls: Set<string> = new Set();

    constructor(server: MCPServer, memory: MemoryManager) {
        this.server = server;
        this.memory = memory;
    }

    /**
     * Conduct deep research on a topic
     * @param topic - The research query
     * @param depth - How many pages to read (default 3)
     */
    public async research(topic: string, depth: number = 3): Promise<string> {
        console.log(`[Research] Starting deep dive on: "${topic}"`);
        const report: string[] = [];
        this.visitedUrls.clear();

        // 1. Initial Search
        const searchResults = await search(topic, {
            safeSearch: SafeSearchType.MODERATE
        });
        const targets = searchResults.results.slice(0, depth);

        report.push(`# Research Report: ${topic}`);
        report.push(`Found ${targets.length} primary sources.`);

        // 2. Process Sources
        for (const target of targets) {
            if (this.visitedUrls.has(target.url)) continue;
            this.visitedUrls.add(target.url);

            this.broadcast('RESEARCH_UPDATE', {
                status: 'reading',
                target: target.title,
                url: target.url,
                progress: (this.visitedUrls.size / targets.length) * 100
            });

            console.log(`[Research] Reading: ${target.title}`);
            try {
                // Navigate first
                await this.server.executeTool("navigate", { url: target.url });

                // Then scrape
                const result = await this.server.executeTool("read_page", { url: target.url });
                const contentText = getFirstTextContent(result);

                if (contentText.startsWith("Error")) {
                    report.push(`- [FAILED] ${target.title}: ${contentText}`);
                    this.broadcast('RESEARCH_UPDATE', { status: 'error', target: target.title, error: contentText });
                    continue;
                }

                // 3. Memorize
                const ctxId = await this.memory.saveContext(
                    `RESEARCH TOPIC: ${topic}\nSOURCE: ${target.url}\n\n${contentText}`,
                    {
                        title: target.title,
                        source: target.url,
                        type: 'research'
                    }
                );

                report.push(`- [MEMORIZED] ${target.title} (ID: ${ctxId})`);
                this.broadcast('RESEARCH_UPDATE', { status: 'memorized', target: target.title, id: ctxId });

            } catch (err: unknown) {
                const errorMessage = getErrorMessage(err);
                console.error(`[Research] Failed to process ${target.url}:`, err);
                report.push(`- [ERROR] ${target.title}: ${errorMessage}`);
                this.broadcast('RESEARCH_UPDATE', { status: 'error', target: target.title, error: errorMessage });
            }
        }

        const finalReport = report.join('\n');
        this.broadcast('RESEARCH_COMPLETE', { report: finalReport });
        return finalReport;
    }

    /**
     * Directly ingest a specific URL
     */
    public async ingest(url: string): Promise<string> {
        console.log(`[Research] Ingesting: ${url}`);
        this.broadcast('RESEARCH_UPDATE', {
            status: 'reading',
            target: url,
            url: url
        });

        try {
            await this.server.executeTool("navigate", { url });
            const result = await this.server.executeTool("read_page", { url });
            const contentText = getFirstTextContent(result);

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
            const errorMessage = getErrorMessage(e);
            this.broadcast('RESEARCH_UPDATE', { status: 'error', target: url, error: errorMessage });
            return `ERROR: ${errorMessage}`;
        }
    }

    private broadcast(type: string, payload: BroadcastPayload) {
        if (this.server.wssInstance && this.server.wssInstance.clients) {
            this.server.wssInstance.clients.forEach((client: WebSocketLikeClient) => {
                if (client.readyState === 1) {
                    client.send(JSON.stringify({ type, payload }));
                }
            });
        }
    }
}
