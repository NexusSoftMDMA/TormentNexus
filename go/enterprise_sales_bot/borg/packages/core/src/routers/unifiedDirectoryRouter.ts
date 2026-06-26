import { z } from "zod";
import { t, publicProcedure } from "../lib/trpc-core.js";
import { publishedCatalogRepository } from "../db/repositories/published-catalog.repo.js";
import { linksBacklogRepository } from "../db/repositories/links-backlog.repo.js";
import { rethrowSqliteUnavailableAsTrpc } from "./sqliteTrpc.js";

type UnifiedDirectorySource = "catalog" | "backlog";

type UnifiedCatalogItem = {
    source: "catalog";
    id: string;
    updated_at: Date | null;
    created_at: Date | null;
    title: string;
    subtitle: string | null;
    description: string | null;
    status: string | null;
    transport: string | null;
    install_method: string | null;
    url: string | null;
    tags: string[];
    confidence: number | null;
    is_duplicate: null;
    duplicate_of: null;
};

type UnifiedBacklogItem = {
    source: "backlog";
    id: string;
    updated_at: Date | null;
    created_at: Date | null;
    title: string;
    subtitle: string | null;
    description: string | null;
    status: string | null;
    transport: null;
    install_method: null;
    url: string | null;
    tags: string[];
    confidence: null;
    is_duplicate: boolean;
    duplicate_of: string | null;
};

export type UnifiedDirectoryItem = UnifiedCatalogItem | UnifiedBacklogItem;

function toDateValue(value: unknown): number {
    if (value instanceof Date) return value.getTime();
    if (typeof value === "string") {
        const parsed = Date.parse(value);
        return Number.isNaN(parsed) ? 0 : parsed;
    }
    return 0;
}

function normalizeCatalogItem(server: Awaited<ReturnType<typeof publishedCatalogRepository.listServers>>[number]): UnifiedCatalogItem {
    return {
        source: "catalog",
        id: server.uuid,
        updated_at: server.updated_at ?? null,
        created_at: server.created_at ?? null,
        title: server.display_name,
        subtitle: server.canonical_id,
        description: server.description ?? null,
        status: server.status,
        transport: server.transport,
        install_method: server.install_method,
        url: server.repository_url ?? server.homepage_url ?? null,
        tags: [...(server.categories ?? []), ...(server.tags ?? [])],
        confidence: typeof server.confidence === "number" ? server.confidence : null,
        is_duplicate: null,
        duplicate_of: null,
    };
}

function normalizeBacklogItem(item: Awaited<ReturnType<typeof linksBacklogRepository.listLinks>>[number]): UnifiedBacklogItem {
    return {
        source: "backlog",
        id: item.uuid,
        updated_at: item.updated_at ?? null,
        created_at: item.created_at ?? null,
        title: item.title ?? item.page_title ?? item.normalized_url,
        subtitle: item.normalized_url,
        description: item.description ?? item.page_description ?? null,
        status: item.research_status,
        transport: null,
        install_method: null,
        url: item.url,
        tags: item.tags ?? [],
        confidence: null,
        is_duplicate: Boolean(item.is_duplicate),
        duplicate_of: item.duplicate_of ?? null,
    };
}

export const unifiedDirectoryRouter = t.router({
    list: publicProcedure
        .input(
            z
                .object({
                    limit: z.number().min(1).max(200).default(50),
                    offset: z.number().min(0).default(0),
                    search: z.string().optional(),
                    source: z.enum(["all", "catalog", "backlog"]).default("all"),
                    show_duplicates: z.boolean().default(false),
                    duplicates_only: z.boolean().default(false),
                    research_status: z.enum(["pending", "running", "done", "failed", "skipped"]).optional(),
                })
                .optional(),
        )
        .query(async ({ input }) => {
            try {
                const limit = input?.limit ?? 50;
                const offset = input?.offset ?? 0;
                const source = input?.source ?? "all";
                const search = input?.search;
                const showDuplicates = input?.show_duplicates ?? false;
                const duplicatesOnly = input?.duplicates_only ?? false;
                const researchStatus = input?.research_status;
                const effectiveShowDuplicates = showDuplicates || duplicatesOnly;

                const wantCatalog = source === "all" || source === "catalog";
                const wantBacklog = source === "all" || source === "backlog";

                const fetchWindow = Math.min(1000, Math.max(200, offset + limit * 3));

                const [catalogItems, backlogItems, catalogTotal, backlogTotal] = await Promise.all([
                    wantCatalog
                        ? publishedCatalogRepository
                              .listServers({ limit: fetchWindow, offset: 0, search })
                              .then((rows) => rows.map(normalizeCatalogItem))
                        : Promise.resolve([] as UnifiedCatalogItem[]),
                    wantBacklog
                        ? linksBacklogRepository
                              .listLinks({
                                  limit: fetchWindow,
                                  offset: 0,
                                  search,
                                  show_duplicates: effectiveShowDuplicates,
                                  duplicates_only: duplicatesOnly,
                                  research_status: researchStatus,
                              })
                              .then((rows) => rows.map(normalizeBacklogItem))
                        : Promise.resolve([] as UnifiedBacklogItem[]),
                    wantCatalog
                        ? publishedCatalogRepository.countServers({ search })
                        : Promise.resolve(0),
                    wantBacklog
                        ? linksBacklogRepository.countLinks({
                              search,
                              show_duplicates: effectiveShowDuplicates,
                              duplicates_only: duplicatesOnly,
                              research_status: researchStatus,
                          })
                        : Promise.resolve(0),
                ]);

                const merged: UnifiedDirectoryItem[] = [...catalogItems, ...backlogItems].sort((a, b) => {
                    const bTime = toDateValue(b.updated_at) || toDateValue(b.created_at);
                    const aTime = toDateValue(a.updated_at) || toDateValue(a.created_at);
                    return bTime - aTime;
                });

                return {
                    items: merged.slice(offset, offset + limit),
                    total: catalogTotal + backlogTotal,
                    totals: {
                        catalog: catalogTotal,
                        backlog: backlogTotal,
                    },
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Unified directory is unavailable", error);
            }
        }),

    stats: publicProcedure.query(async () => {
        try {
            const [catalogTotal, catalogValidated, catalogBroken, catalogRecent, backlogStats] = await Promise.all([
                publishedCatalogRepository.countServers(),
                publishedCatalogRepository.countServers({ status: "validated" }),
                publishedCatalogRepository.countServers({ status: "broken" }),
                publishedCatalogRepository.countRecentlyUpdated(24),
                linksBacklogRepository.getStats(),
            ]);

            return {
                catalog: {
                    total: catalogTotal,
                    validated: catalogValidated,
                    broken: catalogBroken,
                    updated_24h: catalogRecent,
                },
                backlog: backlogStats,
                combined_total: catalogTotal + backlogStats.total,
            };
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc("Unified directory is unavailable", error);
        }
    }),
});
