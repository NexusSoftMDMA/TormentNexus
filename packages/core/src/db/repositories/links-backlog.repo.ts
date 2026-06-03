import { and, desc, eq, like, not, sql } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { db } from "../index.js";
import { linksBacklogTable } from "../tormentnexus-schema.js";

export type LinkBacklogItem = typeof linksBacklogTable.$inferSelect;

export type LinkBacklogListInput = {
    limit?: number;
    offset?: number;
    search?: string;
    source?: string;
    research_status?: string;
    cluster_id?: string;
    show_duplicates?: boolean;
    duplicates_only?: boolean;
};

export type UpsertLinkBacklogInput = {
    url: string;
    normalized_url: string;
    title?: string | null;
    description?: string | null;
    tags?: string[];
    source: string;
    is_duplicate?: boolean;
    duplicate_of?: string | null;
    research_status?: string;
    http_status?: number | null;
    page_title?: string | null;
    page_description?: string | null;
    favicon_url?: string | null;
    researched_at?: Date | null;
    cluster_id?: string | null;
    bobbybookmarks_bookmark_id?: number | null;
    import_session_id?: number | null;
    raw_payload?: Record<string, unknown> | null;
    synced_at?: Date | null;
};

function buildConditions(input: LinkBacklogListInput) {
    const conditions = [];

    if (input.duplicates_only) {
        conditions.push(eq(linksBacklogTable.is_duplicate, true));
    } else if (!input.show_duplicates) {
        conditions.push(eq(linksBacklogTable.is_duplicate, false));
    }
    if (input.source) {
        conditions.push(eq(linksBacklogTable.source, input.source));
    }
    if (input.research_status) {
        conditions.push(eq(linksBacklogTable.research_status, input.research_status as any));
    }
    if (input.cluster_id) {
        conditions.push(eq(linksBacklogTable.cluster_id, input.cluster_id));
    }
    if (input.search && input.search.trim().length > 0) {
        const term = `%${input.search.trim()}%`;
        conditions.push(
            sql`(
                ${linksBacklogTable.url} LIKE ${term}
                OR ${linksBacklogTable.normalized_url} LIKE ${term}
                OR ${linksBacklogTable.title} LIKE ${term}
                OR ${linksBacklogTable.description} LIKE ${term}
            )`
        );
    }

    return conditions;
}

export class LinksBacklogRepository {
    async listLinks(input: LinkBacklogListInput = {}): Promise<LinkBacklogItem[]> {
        const { limit = 50, offset = 0 } = input;
        const conditions = buildConditions(input);

        return db
            .select()
            .from(linksBacklogTable)
            .where(conditions.length === 0 ? undefined : conditions.length === 1 ? conditions[0] : and(...conditions))
            .orderBy(desc(linksBacklogTable.updated_at), desc(linksBacklogTable.created_at))
            .limit(limit)
            .offset(offset);
    }

    async countLinks(input: Omit<LinkBacklogListInput, "limit" | "offset"> = {}): Promise<number> {
        const conditions = buildConditions(input);
        const [row] = await db
            .select({ count: sql<number>`count(*)` })
            .from(linksBacklogTable)
            .where(conditions.length === 0 ? undefined : conditions.length === 1 ? conditions[0] : and(...conditions));

        return row?.count ?? 0;
    }

    async getStats(): Promise<{
        total: number;
        unique: number;
        duplicates: number;
        pending: number;
        researched: number;
        failed: number;
        sources: number;
    }> {
        const [aggregate] = await db
            .select({
                total: sql<number>`count(*)`,
                duplicates: sql<number>`sum(case when ${linksBacklogTable.is_duplicate} = 1 then 1 else 0 end)`,
                pending: sql<number>`sum(case when ${linksBacklogTable.research_status} = 'pending' then 1 else 0 end)`,
                researched: sql<number>`sum(case when ${linksBacklogTable.research_status} = 'done' then 1 else 0 end)`,
                failed: sql<number>`sum(case when ${linksBacklogTable.research_status} = 'failed' then 1 else 0 end)`,
                sources: sql<number>`count(distinct ${linksBacklogTable.source})`,
            })
            .from(linksBacklogTable);

        const total = aggregate?.total ?? 0;
        const duplicates = aggregate?.duplicates ?? 0;

        return {
            total,
            unique: Math.max(0, total - duplicates),
            duplicates,
            pending: aggregate?.pending ?? 0,
            researched: aggregate?.researched ?? 0,
            failed: aggregate?.failed ?? 0,
            sources: aggregate?.sources ?? 0,
        };
    }

    async findByUuid(uuid: string): Promise<LinkBacklogItem | undefined> {
        const [row] = await db
            .select()
            .from(linksBacklogTable)
            .where(eq(linksBacklogTable.uuid, uuid));
        return row;
    }

    async findByNormalizedUrl(normalizedUrl: string): Promise<LinkBacklogItem | undefined> {
        const [row] = await db
            .select()
            .from(linksBacklogTable)
            .where(eq(linksBacklogTable.normalized_url, normalizedUrl));
        return row;
    }

    async updateLinkStatus(uuid: string, status: string, score: number = 0): Promise<void> {
        await db
            .update(linksBacklogTable)
            .set({
                research_status: status as any,
                updated_at: new Date(),
            })
            .where(eq(linksBacklogTable.uuid, uuid));
    }

    async upsertLink(input: UpsertLinkBacklogInput): Promise<LinkBacklogItem> {
        const now = new Date();
        const payload = {
            uuid: randomUUID(),
            url: input.url,
            normalized_url: input.normalized_url,
            title: input.title ?? null,
            description: input.description ?? null,
            tags: input.tags ?? [],
            source: input.source,
            is_duplicate: input.is_duplicate ?? false,
            duplicate_of: input.duplicate_of ?? null,
            research_status: (input.research_status ?? "pending") as any,
            http_status: input.http_status ?? null,
            page_title: input.page_title ?? null,
            page_description: input.page_description ?? null,
            favicon_url: input.favicon_url ?? null,
            researched_at: input.researched_at ?? null,
            cluster_id: input.cluster_id ?? null,
            bobbybookmarks_bookmark_id: input.bobbybookmarks_bookmark_id ?? null,
            import_session_id: input.import_session_id ?? null,
            raw_payload: input.raw_payload ?? null,
            synced_at: input.synced_at ?? now,
            updated_at: now,
        };

        const [row] = await db
            .insert(linksBacklogTable)
            .values(payload)
            .onConflictDoUpdate({
                target: linksBacklogTable.normalized_url,
                set: {
                    url: sql`excluded.url`,
                    title: sql`excluded.title`,
                    description: sql`excluded.description`,
                    tags: sql`excluded.tags`,
                    source: sql`excluded.source`,
                    is_duplicate: sql`excluded.is_duplicate`,
                    duplicate_of: sql`excluded.duplicate_of`,
                    research_status: sql`excluded.research_status`,
                    http_status: sql`excluded.http_status`,
                    page_title: sql`excluded.page_title`,
                    page_description: sql`excluded.page_description`,
                    favicon_url: sql`excluded.favicon_url`,
                    researched_at: sql`excluded.researched_at`,
                    cluster_id: sql`excluded.cluster_id`,
                    bobbybookmarks_bookmark_id: sql`excluded.bobbybookmarks_bookmark_id`,
                    import_session_id: sql`excluded.import_session_id`,
                    raw_payload: sql`excluded.raw_payload`,
                    synced_at: sql`excluded.synced_at`,
                    updated_at: sql`excluded.updated_at`,
                },
            })
            .returning();

        return row;
    }
}

export const linksBacklogRepository = new LinksBacklogRepository();
