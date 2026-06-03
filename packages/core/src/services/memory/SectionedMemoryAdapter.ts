/**
 * SectionedMemoryAdapter – IMemoryProvider backed by a TormentNexus-managed sectioned store.
 *
 * This provider persists structured project context in a single TormentNexus-owned JSON
 * snapshot so it can participate in the redundant memory pipeline alongside the
 * JSON provider and future vector/database stores.
 */

import fs from 'fs/promises';
import path from 'path';

import { v4 as uuidv4 } from 'uuid';

import { IMemoryProvider, Memory } from '../../interfaces/IMemoryProvider.js';

export const SECTIONED_MEMORY_DEFAULT_SECTIONS = [
    'project_context',
    'user_facts',
    'style_preferences',
    'commands',
    'general',
] as const;

interface SectionedMemorySection {
    section: string;
    entries: SectionedMemoryEntry[];
}

interface SectionedMemoryEntry {
    uuid: string;
    content: string;
    tags: string[];
    createdAt: string;
    source: string;
}

interface SectionedMemoryStore {
    version: string;
    sections: SectionedMemorySection[];
}

const LEGACY_STORE_FILE = ['claude', '_mem.json'].join('');

export class SectionedMemoryAdapter implements IMemoryProvider {
    private readonly storePath: string;
    private readonly legacyStorePath: string;
    private store: SectionedMemoryStore = { version: '1.0.0', sections: [] };
    private initialized = false;

    constructor(workspaceRoot: string) {
        this.storePath = path.join(workspaceRoot, '.tormentnexus', 'sectioned_memory.json');
        this.legacyStorePath = path.join(workspaceRoot, '.tormentnexus', LEGACY_STORE_FILE);
    }

    async init(): Promise<void> {
        if (this.initialized) return;

        await fs.mkdir(path.dirname(this.storePath), { recursive: true });

        const loaded = await this.tryLoadStore(this.storePath) ?? await this.tryLoadStore(this.legacyStorePath);
        if (loaded) {
            this.store = loaded;
            if (await this.fileExists(this.legacyStorePath) && !(await this.fileExists(this.storePath))) {
                await this.persist();
            }
            this.initialized = true;
            return;
        }

        this.store = {
            version: '1.0.0',
            sections: [...SECTIONED_MEMORY_DEFAULT_SECTIONS].map((section) => ({ section, entries: [] })),
        };
        await this.persist();
        this.initialized = true;
    }

    async saveMemory(
        content: string,
        metadata: Record<string, any>,
        userId: string,
        agentId?: string,
    ): Promise<Memory> {
        await this.init();

        const section = metadata.section || 'general';
        const entry: SectionedMemoryEntry = {
            uuid: uuidv4(),
            content,
            tags: metadata.tags || [],
            createdAt: new Date().toISOString(),
            source: agentId ? 'agent' : 'user',
        };

        let targetSection = this.store.sections.find((candidate) => candidate.section === section);
        if (!targetSection) {
            targetSection = { section, entries: [] };
            this.store.sections.push(targetSection);
        }

        targetSection.entries.push(entry);
        await this.persist();

        return this.toMemory(entry, section, userId, agentId);
    }

    async searchMemories(
        query: string,
        userId: string,
        limit: number = 5,
        _threshold: number = 0.7,
    ): Promise<Memory[]> {
        await this.init();

        const queryLower = query.toLowerCase();
        const results: Memory[] = [];

        for (const section of this.store.sections) {
            for (const entry of section.entries) {
                if (
                    entry.content.toLowerCase().includes(queryLower)
                    || entry.tags.some((tag) => tag.toLowerCase().includes(queryLower))
                ) {
                    results.push(this.toMemory(entry, section.section, userId));
                }
            }
        }

        results.sort((left, right) => right.createdAt.getTime() - left.createdAt.getTime());
        return results.slice(0, limit);
    }

    async listMemories(userId: string, limit: number = 20, offset: number = 0): Promise<Memory[]> {
        await this.init();

        const all: Memory[] = [];
        for (const section of this.store.sections) {
            for (const entry of section.entries) {
                all.push(this.toMemory(entry, section.section, userId));
            }
        }

        all.sort((left, right) => right.createdAt.getTime() - left.createdAt.getTime());
        return all.slice(offset, offset + limit);
    }

    async deleteMemory(uuid: string, _userId: string): Promise<void> {
        await this.init();

        for (const section of this.store.sections) {
            const index = section.entries.findIndex((entry) => entry.uuid === uuid);
            if (index !== -1) {
                section.entries.splice(index, 1);
                await this.persist();
                return;
            }
        }
    }

    private async tryLoadStore(filePath: string): Promise<SectionedMemoryStore | null> {
        try {
            const raw = await fs.readFile(filePath, 'utf-8');
            return JSON.parse(raw) as SectionedMemoryStore;
        } catch (error: any) {
            if (error?.code === 'ENOENT') {
                return null;
            }

            console.error(`[SectionedMemoryAdapter] Failed to load store from ${filePath}:`, error);
            throw error;
        }
    }

    private async fileExists(filePath: string): Promise<boolean> {
        try {
            await fs.access(filePath);
            return true;
        } catch {
            return false;
        }
    }

    private toMemory(
        entry: SectionedMemoryEntry,
        section: string,
        userId: string,
        agentId?: string,
    ): Memory {
        return {
            uuid: entry.uuid,
            content: entry.content,
            metadata: {
                section,
                tags: entry.tags,
                source: entry.source,
                provider: 'sectioned-store',
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