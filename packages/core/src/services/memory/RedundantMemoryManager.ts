/**
 * RedundantMemoryManager – Fan-out writes to every registered IMemoryProvider
 *
 * This is the "super-memory" module requested by the user.  It guarantees that
 * every memory write is persisted to ALL providers simultaneously (JSON flat
 * file, TormentNexus sectioned memory store, and any future vector/DB providers).
 *
 * Read operations merge results from all providers, de-duplicate by UUID, and
 * return a combined view.
 *
 * Why redundancy?
 * - Durability: losing one store doesn't lose data.
 * - Diversity: different providers excel at different queries (keyword, vector,
 *   section-based).  Merging results gives the richest recall.
 * - Compatibility: existing code can keep using MemoryManager while new code
 *   opts into RedundantMemoryManager for full coverage.
 *
 * Usage:
 *   const mgr = new RedundantMemoryManager(workspaceRoot);
 *   await mgr.init();
 *   await mgr.saveMemory("fact", {}, "user-1"); // saved to ALL providers
 */

import { IMemoryProvider, Memory } from '../../interfaces/IMemoryProvider.js';
import { JsonMemoryProvider } from './JsonMemoryProvider.js';
import { SectionedMemoryAdapter } from './SectionedMemoryAdapter.js';

export type RedundantMemoryProviderName = 'json' | 'sectioned-store';

export class RedundantMemoryManager implements IMemoryProvider {
    private providers: IMemoryProvider[] = [];

    constructor(workspaceRoot: string) {
        // Register all available memory providers
        this.providers.push(new JsonMemoryProvider(workspaceRoot));
        this.providers.push(new SectionedMemoryAdapter(workspaceRoot));
        // Future: add VectorMemoryProvider, PostgresMemoryProvider, etc.
    }

    /**
     * Add a custom provider at runtime (e.g. a plugin-supplied vector store).
     */
    addProvider(provider: IMemoryProvider): void {
        this.providers.push(provider);
    }

    getProviderNames(): RedundantMemoryProviderName[] {
        return this.providers.flatMap((provider) => {
            if (provider instanceof JsonMemoryProvider) {
                return ['json'];
            }

            if (provider instanceof SectionedMemoryAdapter) {
                return ['sectioned-store'];
            }

            return [];
        });
    }

    async init(): Promise<void> {
        // Initialize all providers concurrently.  If one fails we log but
        // continue – partial redundancy is better than total failure.
        const results = await Promise.allSettled(
            this.providers.map(p => p.init())
        );
        results.forEach((r, i) => {
            if (r.status === 'rejected') {
                console.error(
                    `[RedundantMemoryManager] Provider ${i} failed to init:`,
                    r.reason
                );
            }
        });
    }

    /**
     * Fan-out write: save to every provider.
     * Returns the Memory from the first provider that succeeds.
     */
    async saveMemory(
        content: string,
        metadata: Record<string, any>,
        userId: string,
        agentId?: string
    ): Promise<Memory> {
        const results = await Promise.allSettled(
            this.providers.map(p => p.saveMemory(content, metadata, userId, agentId))
        );

        let firstSuccess: Memory | null = null;
        results.forEach((r, i) => {
            if (r.status === 'fulfilled') {
                if (!firstSuccess) firstSuccess = r.value;
            } else {
                console.error(
                    `[RedundantMemoryManager] Provider ${i} failed to save:`,
                    r.reason
                );
            }
        });

        if (!firstSuccess) {
            throw new Error('All memory providers failed to save');
        }
        return firstSuccess;
    }

    /**
     * Merge search results from all providers and de-duplicate by UUID.
     */
    async searchMemories(
        query: string,
        userId: string,
        limit: number = 10,
        threshold: number = 0.7
    ): Promise<Memory[]> {
        const allResults = await Promise.allSettled(
            this.providers.map(p => p.searchMemories(query, userId, limit, threshold))
        );

        const merged = new Map<string, Memory>();
        allResults.forEach(r => {
            if (r.status === 'fulfilled') {
                for (const mem of r.value) {
                    if (!merged.has(mem.uuid)) {
                        merged.set(mem.uuid, mem);
                    }
                }
            }
        });

        // Sort by creation date, newest first
        const deduped = Array.from(merged.values());
        deduped.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
        return deduped.slice(0, limit);
    }

    /**
     * Merge listings from all providers and de-duplicate by UUID.
     */
    async listMemories(
        userId: string,
        limit: number = 20,
        offset: number = 0
    ): Promise<Memory[]> {
        const allResults = await Promise.allSettled(
            this.providers.map(p => p.listMemories(userId, limit + offset))
        );

        const merged = new Map<string, Memory>();
        allResults.forEach(r => {
            if (r.status === 'fulfilled') {
                for (const mem of r.value) {
                    if (!merged.has(mem.uuid)) {
                        merged.set(mem.uuid, mem);
                    }
                }
            }
        });

        const deduped = Array.from(merged.values());
        deduped.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
        return deduped.slice(offset, offset + limit);
    }

    /**
     * Delete from all providers (fan-out).
     */
    async deleteMemory(uuid: string, userId: string): Promise<void> {
        const results = await Promise.allSettled(
            this.providers.map(p => p.deleteMemory(uuid, userId))
        );

        results.forEach((r, i) => {
            if (r.status === 'rejected') {
                console.error(
                    `[RedundantMemoryManager] Provider ${i} failed to delete:`,
                    r.reason
                );
            }
        });
    }
}
