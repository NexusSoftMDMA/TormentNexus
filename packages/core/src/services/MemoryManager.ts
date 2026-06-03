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
  public graph: any = null;

  constructor(rootDir: string) {
    this.rootDir = rootDir;
    this.registryPath = path.join(rootDir, 'data', 'memory_registry.json');
    this.pruner = new ContextPruner();
  }

  async initialize() {
    if (this.initialized) return;
    try {
      const { LanceDBStore, MemoryVectorStore } = await import('@tormentnexus/memory');
      try {
        this.provider = new LanceDBStore(this.rootDir);
        await this.provider.initialize();
      } catch (e) {
        this.provider = new MemoryVectorStore();
        await this.provider.initialize();
      }
      // Initialize graph memory
      try {
        const { GraphMemory } = await import('@tormentnexus/memory');
        const gm = new GraphMemory(this.rootDir);
        if (gm.initialize) await gm.initialize();
        this.graph = gm;
      } catch {
        // Graph memory is optional
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
      id: outcomeId,
      type: 'tool_outcome',
      tool: toolName,
      success,
      heat_score: success ? 80 : 20,
      last_accessed_at: Date.now()
    });
  }

  public async search(query: string, limit: number = 5): Promise<SearchResult[]> {
    if (!this.initialized) await this.initialize();
    const results = await this.provider.search(query, limit);
    return results.map((r: any) => ({
      id: r.id || '',
      content: r.text || '',
      score: r._distance !== undefined ? 1 - r._distance : 0,
      metadata: { ...r, heat_score: r.heat_score }
    })).sort((a: any, b: any) => {
      const heatA = (a.metadata.heat_score || 50) / 500;
      const heatB = (b.metadata.heat_score || 50) / 500;
      return (b.score + heatB) - (a.score + heatA);
    });
  }

  public async indexCodebase(rootDir: string): Promise<number> {
    if (!this.initialized) await this.initialize();
    try {
      const { Indexer } = await import('@tormentnexus/memory');
      const indexer = new Indexer(this.provider);
      return await indexer.indexDirectory(rootDir);
    } catch {
      // Fallback: simple file-walk indexing
      const glob = await import('fast-glob');
      const files = await glob.default('**/*.{ts,tsx,js,jsx,go,py,rs}', {
        cwd: rootDir,
        absolute: true,
        ignore: ['**/node_modules/**', '**/dist/**', '**/.git/**']
      });
      let count = 0;
      for (const file of files) {
        try {
          const content = await fs.readFile(file, 'utf-8');
          const chunks = content.length > 2000
            ? content.match(/[\s\S]{1,2000}/g) || [content]
            : [content];
          for (const chunk of chunks) {
            await this.provider.addMemory(chunk, {
              id: `codebase/${path.relative(rootDir, file)}/${count}`,
              type: 'codebase',
              source: path.relative(rootDir, file),
              heat_score: 50,
              last_accessed_at: Date.now()
            });
            count++;
          }
        } catch { /* skip unreadable files */ }
      }
      return count;
    }
  }

  public async indexSymbols(rootDir: string): Promise<number> {
    if (!this.initialized) await this.initialize();
    try {
      const { Indexer } = await import('@tormentnexus/memory');
      const indexer = new Indexer(this.provider);
      return await indexer.indexSymbols(rootDir);
    } catch {
      // Fallback: simple symbol extraction via regex
      const glob = await import('fast-glob');
      const files = await glob.default('**/*.{ts,tsx,js,jsx,go}', {
        cwd: rootDir,
        absolute: true,
        ignore: ['**/node_modules/**', '**/dist/**', '**/.git/**']
      });
      let count = 0;
      const symbolPattern = /(?:export\s+)?(?:function|class|interface|type|const|var|let)\s+(\w+)/g;
      for (const file of files) {
        try {
          const content = await fs.readFile(file, 'utf-8');
          let match: RegExpExecArray | null;
          while ((match = symbolPattern.exec(content)) !== null) {
            await this.provider.addMemory(`${match[1]} in ${path.relative(rootDir, file)}`, {
              id: `symbol/${match[1]}/${path.relative(rootDir, file)}`,
              type: 'symbol',
              name: match[1],
              source: path.relative(rootDir, file),
              heat_score: 60,
              last_accessed_at: Date.now()
            });
            count++;
          }
        } catch { /* skip unreadable files */ }
      }
      return count;
    }
  }

  public async searchSymbols(query: string, limit: number = 10): Promise<SearchResult[]> {
    return this.search(`symbol ${query}`, limit);
  }

  public async listContexts(): Promise<any[]> {
    if (!this.initialized) await this.initialize();
    try {
      if (this.provider.listMemories) {
        return await this.provider.listMemories();
      }
      // Fallback: search with wildcard
      const results = await this.search('*', 100);
      return results.map(r => ({
        id: r.id,
        content: r.content?.substring(0, 200),
        type: r.metadata?.type || 'unknown',
        source: r.metadata?.source || '',
        heat_score: r.metadata?.heat_score || 50
      }));
    } catch {
      return [];
    }
  }

  public async getContext(id: string): Promise<any | null> {
    if (!this.initialized) await this.initialize();
    try {
      if (this.provider.getMemory) {
        return await this.provider.getMemory(id);
      }
      // Fallback: search by ID prefix
      const results = await this.search(id, 1);
      return results.length > 0 ? results[0] : null;
    } catch {
      return null;
    }
  }

  public async deleteContext(id: string): Promise<boolean> {
    if (!this.initialized) await this.initialize();
    try {
      if (this.provider.deleteMemory) {
        await this.provider.deleteMemory(id);
        return true;
      }
      return false;
    } catch {
      return false;
    }
  }

  public getContextSize(messages: Message[]): number {
    return this.pruner.estimateTokens(messages);
  }

  public pruneContext(messages: Message[], options?: Partial<PruningOptions>): Message[] {
    if (options) {
      return new ContextPruner(options).prune(messages);
    }
    return this.pruner.prune(messages);
  }

  public getPipelineSummary(): { provider: string; initialized: boolean } | null {
    if (!this.provider) return null;
    return {
      provider: this.provider.constructor?.name ?? 'unknown',
      initialized: this.initialized
    };
  }
}
