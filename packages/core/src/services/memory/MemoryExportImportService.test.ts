import { describe, expect, it } from 'vitest';

import { MemoryExportImportService } from './MemoryExportImportService.js';

describe('MemoryExportImportService interchange', () => {
    const service = new MemoryExportImportService({});

    it('converts sectioned store snapshots into TormentNexus JSON provider snapshots', async () => {
        const source = JSON.stringify({
            version: '1.0.0',
            sections: [
                {
                    section: 'project_context',
                    entries: [
                        {
                            uuid: 'mem-1',
                            content: 'Remember the active billing refactor.',
                            tags: ['billing', 'routing'],
                            createdAt: '2026-03-11T10:00:00.000Z',
                            source: 'agent',
                        },
                    ],
                },
            ],
        });

        const converted = await service.convert(source, 'sectioned-memory-store', 'json-provider', 'default');
        const parsed = JSON.parse(converted);

        expect(parsed).toEqual([
            {
                uuid: 'mem-1',
                content: 'Remember the active billing refactor.',
                metadata: {
                    section: 'project_context',
                    tags: ['billing', 'routing'],
                    source: 'agent',
                    provider: 'sectioned-store',
                },
                userId: 'default',
                createdAt: '2026-03-11T10:00:00.000Z',
            },
        ]);
    });

    it('converts canonical JSON exports into sectioned store snapshots', async () => {
        const source = JSON.stringify([
            {
                uuid: 'mem-2',
                content: 'Use the MCP dashboard logs shortcut for server diagnostics.',
                metadata: {
                    section: 'commands',
                    tags: ['mcp', 'logs'],
                    source: 'user',
                },
                userId: 'default',
                createdAt: '2026-03-11T12:00:00.000Z',
            },
        ]);

        const converted = await service.convert(source, 'json', 'sectioned-memory-store', 'default');
        const parsed = JSON.parse(converted);

        expect(parsed.sections).toEqual(expect.arrayContaining([
            {
                section: 'commands',
                entries: [
                    {
                        uuid: 'mem-2',
                        content: 'Use the MCP dashboard logs shortcut for server diagnostics.',
                        tags: ['mcp', 'logs'],
                        createdAt: '2026-03-11T12:00:00.000Z',
                        source: 'user',
                    },
                ],
            },
        ]));
    });

    it('imports provider-native snapshots through the canonical pipeline', async () => {
        const saved: Array<{ content: string; metadata: Record<string, unknown>; userId: string; agentId?: string }> = [];
        const importingService = new MemoryExportImportService({
            saveMemory: async (content, metadata, userId, agentId) => {
                saved.push({ content, metadata, userId, agentId });
                return {
                    uuid: 'saved-1',
                    content,
                    metadata,
                    userId,
                    agentId,
                    createdAt: new Date('2026-03-11T13:00:00.000Z'),
                };
            },
        });

        const result = await importingService.importBulk(JSON.stringify({
            version: '1.0.0',
            sections: [
                {
                    section: 'general',
                    entries: [
                        {
                            uuid: 'mem-3',
                            content: 'Native provider import should preserve tags.',
                            tags: ['import'],
                            createdAt: '2026-03-11T13:00:00.000Z',
                            source: 'user',
                        },
                    ],
                },
            ],
        }), 'sectioned-memory-store', 'default');

        expect(result).toEqual({ imported: 1, errors: 0 });
        expect(saved).toEqual([
            {
                content: 'Native provider import should preserve tags.',
                metadata: {
                    section: 'general',
                    tags: ['import'],
                    source: 'user',
                    provider: 'sectioned-store',
                },
                userId: 'default',
                agentId: undefined,
            },
        ]);
    });
});