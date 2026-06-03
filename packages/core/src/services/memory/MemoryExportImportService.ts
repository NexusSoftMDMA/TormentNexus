/**
 * MemoryExportImportService – canonical and provider-native memory interchange
 *
 * Supports:
 * - Canonical exports: JSON, CSV, JSONL
 * - Provider-native snapshots: TormentNexus JSON provider, sectioned memory store
 * - Conversion between all supported formats via a canonical intermediate form
 */

import crypto from 'node:crypto';
import fs from 'fs/promises';
import path from 'path';

import { Memory } from '../../interfaces/IMemoryProvider.js';
import { SECTIONED_MEMORY_DEFAULT_SECTIONS } from './SectionedMemoryAdapter.js';

export type CanonicalMemoryFormat = 'json' | 'csv' | 'jsonl';
export type ProviderMemoryFormat = 'json-provider' | 'sectioned-memory-store';
export type MemoryInterchangeFormat = CanonicalMemoryFormat | ProviderMemoryFormat;

export const MEMORY_INTERCHANGE_FORMATS: Array<{
    id: MemoryInterchangeFormat;
    label: string;
    kind: 'canonical' | 'provider';
    extension: 'json' | 'csv' | 'jsonl';
    description: string;
}> = [
    {
        id: 'json',
        label: 'Canonical JSON',
        kind: 'canonical',
        extension: 'json',
        description: 'Portable TormentNexus memory export with metadata preserved.',
    },
    {
        id: 'csv',
        label: 'Canonical CSV',
        kind: 'canonical',
        extension: 'csv',
        description: 'Spreadsheet-friendly memory export.',
    },
    {
        id: 'jsonl',
        label: 'Canonical JSONL',
        kind: 'canonical',
        extension: 'jsonl',
        description: 'Streaming-friendly newline-delimited memory export.',
    },
    {
        id: 'json-provider',
        label: 'TormentNexus JSON Provider',
        kind: 'provider',
        extension: 'json',
        description: 'Native snapshot of TormentNexus\'s flat-file memory provider.',
    },
    {
        id: 'sectioned-memory-store',
        label: 'Sectioned Memory Store',
        kind: 'provider',
        extension: 'json',
        description: 'Native TormentNexus sectioned memory snapshot.',
    },
];

interface MemoryBackend {
    listMemories?: (userId: string, limit?: number, offset?: number) => Promise<Memory[]>;
    listContexts?: () => Promise<unknown[]>;
    saveMemory?: (content: string, metadata: Record<string, unknown>, userId: string, agentId?: string) => Promise<Memory>;
    saveContext?: (content: string, metadata?: Record<string, unknown>) => Promise<unknown>;
}

export class MemoryExportImportService {
    private backend: MemoryBackend;
    private workspaceRoot?: string;

    constructor(backend: MemoryBackend, options?: { workspaceRoot?: string }) {
        this.backend = backend;
        this.workspaceRoot = options?.workspaceRoot;
    }

    getSupportedFormats() {
        return MEMORY_INTERCHANGE_FORMATS;
    }

    async exportAll(userId: string, format: MemoryInterchangeFormat): Promise<string> {
        if (format === 'json-provider') {
            return this.exportJsonProvider(userId);
        }

        const memories = await this.readCanonicalMemories(userId);

        if (format === 'sectioned-memory-store') {
            return JSON.stringify(this.toSectionedMemoryStore(memories), null, 2);
        }

        return this.serializeCanonical(memories, format);
    }

    async importBulk(data: string, format: MemoryInterchangeFormat, userId: string): Promise<{ imported: number; errors: number }> {
        const records = this.parseToCanonical(data, format, userId);
        let imported = 0;
        let errors = 0;

        for (const record of records) {
            try {
                if (this.backend.saveMemory) {
                    await this.backend.saveMemory(record.content, record.metadata, userId, record.agentId);
                } else if (this.backend.saveContext) {
                    await this.backend.saveContext(record.content, record.metadata);
                }
                imported++;
            } catch {
                errors++;
            }
        }

        return { imported, errors };
    }

    async convert(data: string, from: MemoryInterchangeFormat, to: MemoryInterchangeFormat, userId: string): Promise<string> {
        const records = this.parseToCanonical(data, from, userId);

        if (to === 'sectioned-memory-store') {
            return JSON.stringify(this.toSectionedMemoryStore(records), null, 2);
        }

        if (to === 'json-provider') {
            return JSON.stringify(this.toJsonProvider(records), null, 2);
        }

        return this.serializeCanonical(records, to);
    }

    private async readCanonicalMemories(userId: string): Promise<Memory[]> {
        const providerSnapshot = await this.tryReadJsonProviderSnapshot();
        if (providerSnapshot) {
            return this.parseJsonProvider(providerSnapshot, userId);
        }

        if (this.backend.listMemories) {
            return this.backend.listMemories(userId, 100000, 0);
        }

        if (this.backend.listContexts) {
            const contexts = await this.backend.listContexts();
            return contexts.map((context: any, index) => this.normalizeMemory({
                uuid: context?.uuid ?? context?.id ?? `context-${index + 1}`,
                content: context?.content ?? '',
                metadata: context?.metadata ?? {},
                createdAt: context?.createdAt,
                agentId: context?.agentId,
            }, userId));
        }

        return [];
    }

    private async exportJsonProvider(userId: string): Promise<string> {
        const providerSnapshot = await this.tryReadJsonProviderSnapshot();
        if (providerSnapshot) {
            return providerSnapshot;
        }

        const memories = await this.readCanonicalMemories(userId);
        return JSON.stringify(this.toJsonProvider(memories), null, 2);
    }

    private async tryReadJsonProviderSnapshot(): Promise<string | null> {
        if (!this.workspaceRoot) {
            return null;
        }

        try {
            return await fs.readFile(path.join(this.workspaceRoot, 'memory.json'), 'utf-8');
        } catch (error: unknown) {
            const code = typeof error === 'object' && error !== null && 'code' in error
                ? String((error as { code?: unknown }).code)
                : '';
            if (code === 'ENOENT') {
                return null;
            }
            throw error;
        }
    }

    private parseToCanonical(data: string, format: MemoryInterchangeFormat, userId: string): Memory[] {
        switch (format) {
            case 'json':
                return this.parseCanonicalJson(data, userId);
            case 'csv':
                return this.fromCSV(data, userId);
            case 'jsonl':
                return data
                    .split('\n')
                    .filter(line => line.trim())
                    .map(line => this.normalizeMemory(JSON.parse(line), userId));
            case 'json-provider':
                return this.parseJsonProvider(data, userId);
            case 'sectioned-memory-store':
                return this.parseSectionedMemoryStore(data, userId);
            default:
                throw new Error(`Unsupported memory format: ${format}`);
        }
    }

    private serializeCanonical(memories: Memory[], format: CanonicalMemoryFormat): string {
        switch (format) {
            case 'json':
                return JSON.stringify(this.toJsonProvider(memories), null, 2);
            case 'csv':
                return this.toCSV(memories);
            case 'jsonl':
                return memories.map(memory => JSON.stringify(this.toSerializableMemory(memory))).join('\n');
            default:
                throw new Error(`Unsupported export format: ${format}`);
        }
    }

    private parseCanonicalJson(data: string, userId: string): Memory[] {
        const parsed = JSON.parse(data);
        if (!Array.isArray(parsed)) {
            return [];
        }

        return parsed.map(record => this.normalizeMemory(record, userId));
    }

    private parseJsonProvider(data: string, userId: string): Memory[] {
        return this.parseCanonicalJson(data, userId);
    }

    private parseSectionedMemoryStore(data: string, userId: string): Memory[] {
        const parsed = JSON.parse(data) as {
            sections?: Array<{
                section?: string;
                entries?: Array<{
                    uuid?: string;
                    content?: string;
                    tags?: string[];
                    createdAt?: string;
                    source?: string;
                }>;
            }>;
        };

        const sections = Array.isArray(parsed.sections) ? parsed.sections : [];
        const memories: Memory[] = [];

        for (const section of sections) {
            const sectionName = typeof section?.section === 'string' && section.section.trim().length > 0
                ? section.section
                : 'general';
            const entries = Array.isArray(section?.entries) ? section.entries : [];

            for (const entry of entries) {
                memories.push(this.normalizeMemory({
                    uuid: entry?.uuid,
                    content: entry?.content,
                    createdAt: entry?.createdAt,
                    metadata: {
                        section: sectionName,
                        tags: Array.isArray(entry?.tags) ? entry.tags : [],
                        source: entry?.source ?? 'user',
                        provider: 'sectioned-store',
                    },
                }, userId));
            }
        }

        return memories;
    }

    private normalizeMemory(record: any, userId: string): Memory {
        const metadata = typeof record?.metadata === 'object' && record?.metadata !== null
            ? record.metadata
            : {};

        return {
            uuid: String(record?.uuid ?? record?.id ?? crypto.randomUUID()),
            content: String(record?.content ?? ''),
            metadata,
            userId: typeof record?.userId === 'string' && record.userId.length > 0 ? record.userId : userId,
            agentId: typeof record?.agentId === 'string' && record.agentId.length > 0 ? record.agentId : undefined,
            createdAt: record?.createdAt ? new Date(record.createdAt) : new Date(),
        };
    }

    private toSerializableMemory(memory: Memory) {
        return {
            ...memory,
            createdAt: memory.createdAt.toISOString(),
        };
    }

    private toJsonProvider(memories: Memory[]) {
        return memories.map(memory => this.toSerializableMemory(memory));
    }

    private toSectionedMemoryStore(memories: Memory[]) {
        const sections = new Map<string, Array<{
            uuid: string;
            content: string;
            tags: string[];
            createdAt: string;
            source: string;
        }>>();

        for (const defaultSection of SECTIONED_MEMORY_DEFAULT_SECTIONS) {
            sections.set(defaultSection, []);
        }

        for (const memory of memories) {
            const sectionName = typeof memory.metadata?.section === 'string' && memory.metadata.section.trim().length > 0
                ? memory.metadata.section
                : 'general';
            const entries = sections.get(sectionName) ?? [];
            entries.push({
                uuid: memory.uuid,
                content: memory.content,
                tags: Array.isArray(memory.metadata?.tags)
                    ? memory.metadata.tags.map((tag: unknown) => String(tag))
                    : [],
                createdAt: memory.createdAt.toISOString(),
                source: typeof memory.metadata?.source === 'string'
                    ? memory.metadata.source
                    : (memory.agentId ? 'agent' : 'user'),
            });
            sections.set(sectionName, entries);
        }

        return {
            version: '1.0.0',
            sections: Array.from(sections.entries()).map(([section, entries]) => ({ section, entries })),
        };
    }

    private toCSV(memories: Memory[]): string {
        const header = 'uuid,content,userId,agentId,createdAt,metadata';
        const rows = memories.map(memory => {
            const escapedContent = `"${(memory.content || '').replace(/"/g, '""')}"`;
            const escapedMeta = `"${JSON.stringify(memory.metadata || {}).replace(/"/g, '""')}"`;
            return [
                memory.uuid,
                escapedContent,
                memory.userId || '',
                memory.agentId || '',
                memory.createdAt.toISOString(),
                escapedMeta,
            ].join(',');
        });

        return [header, ...rows].join('\n');
    }

    private fromCSV(csv: string, userId: string): Memory[] {
        const lines = csv.split('\n').filter(line => line.trim());
        if (lines.length < 2) {
            return [];
        }

        return lines.slice(1).map(line => {
            const fields = this.parseCSVLine(line);
            return this.normalizeMemory({
                uuid: fields[0],
                content: fields[1],
                userId: fields[2] || userId,
                agentId: fields[3] || undefined,
                createdAt: fields[4] ? new Date(fields[4]) : new Date(),
                metadata: fields[5]
                    ? (() => {
                        try {
                            return JSON.parse(fields[5]);
                        } catch {
                            return {};
                        }
                    })()
                    : {},
            }, userId);
        });
    }

    private parseCSVLine(line: string): string[] {
        const result: string[] = [];
        let current = '';
        let inQuotes = false;

        for (let i = 0; i < line.length; i++) {
            const char = line[i];
            if (char === '"') {
                if (inQuotes && line[i + 1] === '"') {
                    current += '"';
                    i++;
                } else {
                    inQuotes = !inQuotes;
                }
            } else if (char === ',' && !inQuotes) {
                result.push(current);
                current = '';
            } else {
                current += char;
            }
        }

        result.push(current);
        return result;
    }
}
