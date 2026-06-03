/**
 * TormentNexusAdapter – IMemoryProvider backed by a CLAUDE.md-style flat-file
 *
 * tormentnexus stores per-project context as structured markdown sections inside
 * a single `.claude/CLAUDE.md` file.  This adapter translates that format into
 * TormentNexus's Memory/IMemoryProvider interface so it can participate in the
 * redundant memory pipeline alongside JsonMemoryProvider (and future vector
 * stores).
 *
 * Design notes:
 * - Reads/writes a `claude_mem.json` file that mirrors the tormentnexus schema
 *   (project context, user facts, style preferences, commands).
 * - Sections are stored as individual Memory entries with metadata.section
 *   tracking which tormentnexus section they belong to.
 * - Search is keyword-based (same as JsonMemoryProvider). Future: could be
 *   upgraded to cosine similarity if embeddings are available.
 *
 * Integration with the submodule `packages/tormentnexus`:
 * When the submodule is fully checked out, this adapter can optionally read
 * the actual CLAUDE.md file. Until then, it acts as a standalone provider
 * using its own JSON store inspired by the tormentnexus schema.
 */

import fs from 'fs/promises';
import path from 'path';
import { IMemoryProvider, Memory } from '../../interfaces/IMemoryProvider.js';
import { v4 as uuidv4 } from 'uuid';

export const TORMENTNEXUS_DEFAULT_SECTIONS = [
    'project_context',
    'user_facts',
    'style_preferences',
    'commands',
    'general',
] as const;

// The tormentnexus schema sections
interface TormentNexusSection {
    section: string;          // e.g. "project_context", "user_facts", "style"
    entries: TormentNexusEntry[];
}

interface TormentNexusEntry {
    uuid: string;
    content: string;
    tags: string[];
    createdAt: string;
    source: string;           // "user" | "agent" | "auto"
}

interface TormentNexusStore {
    version: string;
    sections: TormentNexusSection[];
}

export class TormentNexusAdapter implements IMemoryProvider {
    private storePath: string;
    private store: TormentNexusStore = { version: '1.0.0', sections: [] };
    private initialized = false;

    constructor(workspaceRoot: string) {
        // Store in .tormentnexus/claude_mem.json alongside other TormentNexus data
        this.storePath = path.join(workspaceRoot, '.tormentnexus', 'claude_mem.json');
    }

    async init(): Promise<void> {
        if (this.initialized) return;

        try {
            // Ensure directory exists
            await fs.mkdir(path.dirname(this.storePath), { recursive: true });
            const data = await fs.readFile(this.storePath, 'utf-8');
            this.store = JSON.parse(data);
        } catch (error: any) {
            if (error.code === 'ENOENT') {
                // Initialize with default sections inspired by tormentnexus
                this.store = {
                    version: '1.0.0',
                    sections: [...TORMENTNEXUS_DEFAULT_SECTIONS].map((section) => ({ section, entries: [] })),
                };
                await this.persist();
            } else {
                console.error(`[TormentNexusAdapter] Failed to load store from ${this.storePath}:`, error);
                throw error;
            }
        }
        this.initialized = true;
    }

    async saveMemory(
        content: string,
        metadata: Record<string, any>,
        userId: string,
        agentId?: string
    ): Promise<Memory> {
        await this.init();

        const section = metadata.section || 'general';
        const entry: TormentNexusEntry = {
            uuid: uuidv4(),
            content,
            tags: metadata.tags || [],
            createdAt: new Date().toISOString(),
            source: agentId ? 'agent' : 'user',
        };

        // Find or create section
        let sec = this.store.sections.find(s => s.section === section);
        if (!sec) {
            sec = { section, entries: [] };
            this.store.sections.push(sec);
        }
        sec.entries.push(entry);
        await this.persist();

        return this.toMemory(entry, section, userId, agentId);
    }

    async searchMemories(
        query: string,
        userId: string,
        limit: number = 5,
        _threshold: number = 0.7
    ): Promise<Memory[]> {
        await this.init();

        const queryLower = query.toLowerCase();
        const results: Memory[] = [];

        for (const sec of this.store.sections) {
            for (const entry of sec.entries) {
                if (
                    entry.content.toLowerCase().includes(queryLower) ||
                    entry.tags.some(t => t.toLowerCase().includes(queryLower))
                ) {
                    results.push(this.toMemory(entry, sec.section, userId));
                }
            }
        }

        // Sort by recency
        results.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
        return results.slice(0, limit);
    }

    async listMemories(userId: string, limit: number = 20, offset: number = 0): Promise<Memory[]> {
        await this.init();

        const all: Memory[] = [];
        for (const sec of this.store.sections) {
            for (const entry of sec.entries) {
                all.push(this.toMemory(entry, sec.section, userId));
            }
        }

        all.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
        return all.slice(offset, offset + limit);
    }

    async deleteMemory(uuid: string, _userId: string): Promise<void> {
        await this.init();

        for (const sec of this.store.sections) {
            const idx = sec.entries.findIndex(e => e.uuid === uuid);
            if (idx !== -1) {
                sec.entries.splice(idx, 1);
                await this.persist();
                return;
            }
        }
    }

    // ---- Helpers ----

    private toMemory(
        entry: TormentNexusEntry,
        section: string,
        userId: string,
        agentId?: string
    ): Memory {
        return {
            uuid: entry.uuid,
            content: entry.content,
            metadata: {
                section,
                tags: entry.tags,
                source: entry.source,
                provider: 'tormentnexus',
            },
            userId,
            agentId,
            createdAt: new Date(entry.createdAt),
        };
    }

    private async persist(): Promise<void> {
        await fs.writeFile(this.storePath, JSON.stringify(this.store, null, 2), 'utf-8');
    }
}
