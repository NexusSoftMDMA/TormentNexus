import { Document, SearchResult, Message } from '../interfaces/VectorProvider.js';
import path from 'path';
import fs from 'fs/promises';
import { existsSync } from 'fs';
import { ContextPruner, PruningOptions } from './ContextPruner.js';

export class MemoryManager {
    private provider: any = null;
    private initialized: boolean = false;
    private rootDir: string;
    private registryPath: string;
    private pruner: ContextPruner;

    constructor(rootDir: string) {
        this.rootDir = rootDir;
        this.registryPath = path.join(rootDir, 'data', 'memory_registry.json');
        this.pruner = new ContextPruner();
    }

    async initialize() {
        if (this.initialized) return;
        try {
            const { LanceDBStore, MemoryVectorStore } = await import('@borg/memory');
            try {
                this.provider = new LanceDBStore(this.rootDir);
                await this.provider.initialize();
            } catch (e) {
                this.provider = new MemoryVectorStore();
                await this.provider.initialize();
            }
            this.initialized = true;
        } catch (e) {
            console.error("Init failed:", e);
        }
    }

    public async saveContext(content: string, metadata: any = {}) {
        if (!this.initialized) await this.initialize();
        const docId = `ctx/${Date.now()}/${Math.random().toString(36).substring(7)}`;
        const meta = { ...metadata, heat_score: metadata.heat_score ?? 50, last_accessed_at: Date.now() };
        await this.provider.addMemory(content, { id: docId, ...meta });
        return docId;
    }

    public async recordToolOutcome(toolName: string, success: boolean, context: string) {
        if (!this.initialized) await this.initialize();
        const outcomeId = `outcome/${toolName}/${Date.now()}`;
        await this.provider.addMemory(`Tool ${toolName}: ${success ? 'SUCCESS' : 'FAILURE'}`, {
            id: outcomeId, type: 'tool_outcome', tool: toolName, success,
            heat_score: success ? 80 : 20, last_accessed_at: Date.now()
        });
    }

    public async search(query: string, limit: number = 5): Promise<SearchResult[]> {
        if (!this.initialized) await this.initialize();
        const results = await this.provider.search(query, limit);
        return results.map((r: any) => ({
            id: r.id || '', content: r.text || '', score: r._distance !== undefined ? 1 - r._distance : 0,
            metadata: { ...r, heat_score: r.heat_score }
        })).sort((a: any, b: any) => {
            const heatA = (a.metadata.heat_score || 50) / 500;
            const heatB = (b.metadata.heat_score || 50) / 500;
            return (b.score + heatB) - (a.score + heatA);
        });
    }
}
