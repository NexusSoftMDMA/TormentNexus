import { beforeEach, describe, expect, it, vi } from 'vitest';

import { linksBacklogRepository } from '../db/repositories/links-backlog.repo.js';
import { publishedCatalogRepository } from '../db/repositories/published-catalog.repo.js';
import { unifiedDirectoryRouter } from './unifiedDirectoryRouter.js';

function createCaller() {
    return unifiedDirectoryRouter.createCaller({} as never);
}

describe('unifiedDirectoryRouter', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    it('merges catalog + backlog results and sorts by recency', async () => {
        vi.spyOn(publishedCatalogRepository, 'listServers').mockResolvedValue([
            {
                uuid: 'catalog-old',
                display_name: 'Catalog Old',
                canonical_id: 'catalog/old',
                description: 'older catalog entry',
                status: 'normalized',
                transport: 'stdio',
                install_method: 'npm',
                repository_url: 'https://example.com/old',
                homepage_url: null,
                categories: ['infra'],
                tags: ['old'],
                confidence: 40,
                updated_at: new Date('2026-03-20T10:00:00.000Z'),
                created_at: new Date('2026-03-20T09:00:00.000Z'),
            },
            {
                uuid: 'catalog-new',
                display_name: 'Catalog New',
                canonical_id: 'catalog/new',
                description: 'newer catalog entry',
                status: 'validated',
                transport: 'sse',
                install_method: 'url',
                repository_url: 'https://example.com/new',
                homepage_url: null,
                categories: ['productivity'],
                tags: ['new'],
                confidence: 88,
                updated_at: new Date('2026-03-21T10:00:00.000Z'),
                created_at: new Date('2026-03-21T09:00:00.000Z'),
            },
        ] as any);

        vi.spyOn(linksBacklogRepository, 'listLinks').mockResolvedValue([
            {
                uuid: 'backlog-mid',
                normalized_url: 'https://example.com/mid',
                url: 'https://example.com/mid',
                title: 'Backlog Mid',
                description: 'mid recency backlog link',
                page_title: null,
                page_description: null,
                research_status: 'pending',
                is_duplicate: true,
                duplicate_of: 'catalog/new',
                tags: ['bookmark'],
                updated_at: new Date('2026-03-21T09:30:00.000Z'),
                created_at: new Date('2026-03-21T09:00:00.000Z'),
            },
        ] as any);

        vi.spyOn(publishedCatalogRepository, 'countServers').mockResolvedValue(2);
        vi.spyOn(linksBacklogRepository, 'countLinks').mockResolvedValue(1);

        const caller = createCaller();
        const result = await caller.list({ source: 'all', limit: 10, offset: 0 });

        expect(result.total).toBe(3);
        expect(result.totals).toEqual({ catalog: 2, backlog: 1 });
        expect(result.items.map((item) => item.id)).toEqual([
            'catalog-new',
            'backlog-mid',
            'catalog-old',
        ]);

        const backlogItem = result.items.find((item) => item.id === 'backlog-mid');
        expect(backlogItem?.source).toBe('backlog');
        if (backlogItem?.source === 'backlog') {
            expect(backlogItem.is_duplicate).toBe(true);
            expect(backlogItem.duplicate_of).toBe('catalog/new');
        }
    });

    it('catalog-only source does not query backlog repository', async () => {
        const listServersSpy = vi.spyOn(publishedCatalogRepository, 'listServers').mockResolvedValue([] as any[]);
        const countServersSpy = vi.spyOn(publishedCatalogRepository, 'countServers').mockResolvedValue(0);
        const listLinksSpy = vi.spyOn(linksBacklogRepository, 'listLinks').mockResolvedValue([] as any[]);
        const countLinksSpy = vi.spyOn(linksBacklogRepository, 'countLinks').mockResolvedValue(0);

        const caller = createCaller();
        await caller.list({ source: 'catalog', search: 'mcp', limit: 5, offset: 0 });

        expect(listServersSpy).toHaveBeenCalledTimes(1);
        expect(countServersSpy).toHaveBeenCalledTimes(1);
        expect(listLinksSpy).not.toHaveBeenCalled();
        expect(countLinksSpy).not.toHaveBeenCalled();
    });

    it('backlog-only source forwards duplicates_only, effective show_duplicates, and research_status to backlog queries', async () => {
        vi.spyOn(publishedCatalogRepository, 'listServers').mockResolvedValue([] as any[]);
        vi.spyOn(publishedCatalogRepository, 'countServers').mockResolvedValue(0);

        const listLinksSpy = vi.spyOn(linksBacklogRepository, 'listLinks').mockResolvedValue([] as any[]);
        const countLinksSpy = vi.spyOn(linksBacklogRepository, 'countLinks').mockResolvedValue(0);

        const caller = createCaller();
        await caller.list({
            source: 'backlog',
            show_duplicates: false,
            duplicates_only: true,
            research_status: 'pending',
            limit: 20,
            offset: 0,
            search: 'bookmark',
        });

        expect(listLinksSpy).toHaveBeenCalledWith(
            expect.objectContaining({
                show_duplicates: true,
                duplicates_only: true,
                search: 'bookmark',
                research_status: 'pending',
            }),
        );
        expect(countLinksSpy).toHaveBeenCalledWith(
            expect.objectContaining({
                show_duplicates: true,
                duplicates_only: true,
                search: 'bookmark',
                research_status: 'pending',
            }),
        );
    });

    it('supports pagination on the merged set', async () => {
        vi.spyOn(publishedCatalogRepository, 'listServers').mockResolvedValue([
            {
                uuid: 'c-1',
                display_name: 'C1',
                canonical_id: 'c/1',
                status: 'normalized',
                transport: 'stdio',
                install_method: 'npm',
                confidence: 10,
                updated_at: new Date('2026-03-21T10:00:00.000Z'),
                created_at: new Date('2026-03-21T09:00:00.000Z'),
                categories: [],
                tags: [],
                description: null,
                repository_url: null,
                homepage_url: null,
            },
        ] as any);
        vi.spyOn(linksBacklogRepository, 'listLinks').mockResolvedValue([
            {
                uuid: 'b-1',
                normalized_url: 'https://b1.test',
                url: 'https://b1.test',
                research_status: 'pending',
                tags: [],
                title: 'B1',
                description: null,
                page_title: null,
                page_description: null,
                updated_at: new Date('2026-03-21T09:30:00.000Z'),
                created_at: new Date('2026-03-21T08:00:00.000Z'),
            },
            {
                uuid: 'b-2',
                normalized_url: 'https://b2.test',
                url: 'https://b2.test',
                research_status: 'done',
                tags: [],
                title: 'B2',
                description: null,
                page_title: null,
                page_description: null,
                updated_at: new Date('2026-03-21T09:10:00.000Z'),
                created_at: new Date('2026-03-21T07:00:00.000Z'),
            },
        ] as any);

        vi.spyOn(publishedCatalogRepository, 'countServers').mockResolvedValue(1);
        vi.spyOn(linksBacklogRepository, 'countLinks').mockResolvedValue(2);

        const caller = createCaller();
        const page1 = await caller.list({ source: 'all', limit: 2, offset: 0 });
        const page2 = await caller.list({ source: 'all', limit: 2, offset: 2 });

        expect(page1.items.map((item) => item.id)).toEqual(['c-1', 'b-1']);
        expect(page2.items.map((item) => item.id)).toEqual(['b-2']);
    });

    it('stats aggregates catalog and backlog counters', async () => {
        const countServersSpy = vi.spyOn(publishedCatalogRepository, 'countServers')
            .mockResolvedValueOnce(120)
            .mockResolvedValueOnce(45)
            .mockResolvedValueOnce(7);

        vi.spyOn(publishedCatalogRepository, 'countRecentlyUpdated').mockResolvedValue(18);
        vi.spyOn(linksBacklogRepository, 'getStats').mockResolvedValue({
            total: 80,
            unique: 70,
            duplicates: 10,
            pending: 30,
            researched: 45,
            failed: 5,
            sources: 3,
        });

        const caller = createCaller();
        const stats = await caller.stats();

        expect(countServersSpy).toHaveBeenNthCalledWith(1);
        expect(countServersSpy).toHaveBeenNthCalledWith(2, { status: 'validated' });
        expect(countServersSpy).toHaveBeenNthCalledWith(3, { status: 'broken' });

        expect(stats.catalog).toEqual({
            total: 120,
            validated: 45,
            broken: 7,
            updated_24h: 18,
        });
        expect(stats.backlog.total).toBe(80);
        expect(stats.combined_total).toBe(200);
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        vi.spyOn(publishedCatalogRepository, 'listServers').mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list({ source: 'catalog', limit: 10, offset: 0 })).rejects.toMatchObject({
            message: 'Unified directory is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for stats when SQLite is unavailable', async () => {
        vi.spyOn(publishedCatalogRepository, 'countServers').mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.stats()).rejects.toMatchObject({
            message: 'Unified directory is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
