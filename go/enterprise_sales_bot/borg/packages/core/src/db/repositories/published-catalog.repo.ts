/**
 * @file published-catalog.repo.ts
 * @module packages/core/src/db/repositories/published-catalog.repo
 *
 * WHAT:
 * Repository for the MCP Registry Intelligence published catalog tables:
 *   - published_mcp_servers          — canonical catalog of known public MCP servers
 *   - published_mcp_server_sources   — provenance tracking per (server × registry)
 *   - published_mcp_config_recipes   — install/config templates (Configurator agent)
 *   - published_mcp_validation_runs  — validation results (Verifier agent)
 *
 * WHY:
 * Separating catalog data from the operator-managed `mcp_servers` table keeps
 * "things discovered in the wild" cleanly distinct from "things I actually use."
 * The Archivist, Configurator, Verifier, and Operator agents all operate through
 * this layer.
 *
 * HOW:
 * - Plain Drizzle ORM queries on SQLite via `better-sqlite3`.
 * - UpsertServer uses `canonical_id` as the natural dedup key.
 * - UpsertSource uses the (server_uuid, source_name) unique index for dedup.
 */

import { and, desc, eq, gte, inArray, like, or, sql } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { db } from "../index.js";
import {
    publishedMcpConfigRecipesTable,
    publishedMcpServerSourcesTable,
    publishedMcpServersTable,
    publishedMcpValidationRunsTable,
    type PublishedServerStatusEnum,
    type ValidationRunModeEnum,
    type ValidationRunOutcomeEnum,
} from "../tormentnexus-schema.js";

// ---- Type aliases ----

export type PublishedMcpServer =
    typeof publishedMcpServersTable.$inferSelect;

export type PublishedMcpServerInsert =
    Omit<typeof publishedMcpServersTable.$inferInsert, "uuid" | "created_at" | "updated_at">;

export type PublishedMcpServerSource =
    typeof publishedMcpServerSourcesTable.$inferSelect;

export type PublishedMcpConfigRecipe =
    typeof publishedMcpConfigRecipesTable.$inferSelect;

export type PublishedMcpValidationRun =
    typeof publishedMcpValidationRunsTable.$inferSelect;

// ---- Query helpers ----

export type CatalogListInput = {
    limit?: number;
    offset?: number;
    search?: string;
    status?: (typeof PublishedServerStatusEnum)[number];
    transport?: string;
    install_method?: string;
};

export type UpsertServerInput = {
    canonical_id: string;
    display_name: string;
    description?: string | null;
    author?: string | null;
    repository_url?: string | null;
    homepage_url?: string | null;
    icon_url?: string | null;
    transport?: string;
    install_method?: string;
    auth_model?: string;
    tags?: string[];
    categories?: string[];
    stars?: number | null;
};

export type UpsertSourceInput = {
    server_uuid: string;
    source_name: string;
    source_url?: string | null;
    raw_payload?: Record<string, unknown> | null;
};

export type CreateRecipeInput = {
    server_uuid: string;
    template: Record<string, unknown>;
    required_secrets?: string[];
    required_env?: Record<string, string>;
    confidence?: number;
    explanation?: string;
    generated_by?: string;
};

export type CreateValidationRunInput = {
    server_uuid: string;
    run_mode: (typeof ValidationRunModeEnum)[number];
    performed_by?: string;
};

export type FinishValidationRunInput = {
    uuid: string;
    outcome: (typeof ValidationRunOutcomeEnum)[number];
    failure_class?: string | null;
    tool_count?: number | null;
    findings_summary?: Record<string, unknown> | null;
};

// ===========================================================================
// PublishedCatalogRepository
// ===========================================================================

export class PublishedCatalogRepository {
    // ---- published_mcp_servers ----

    /**
     * List catalog entries with optional search/filter/pagination.
     * Sorted by confidence DESC, updated_at DESC for freshness.
     */
    async listServers(input: CatalogListInput = {}): Promise<PublishedMcpServer[]> {
        const { limit = 50, offset = 0, search, status, transport, install_method } = input;

        let query = db
            .select()
            .from(publishedMcpServersTable)
            .orderBy(
                desc(publishedMcpServersTable.confidence),
                desc(publishedMcpServersTable.updated_at)
            )
            .limit(limit)
            .offset(offset);

        // Build filter conditions
        const conditions = [];

        if (status) {
            conditions.push(eq(publishedMcpServersTable.status, status));
        }
        if (transport) {
            conditions.push(eq(publishedMcpServersTable.transport, transport as any));
        }
        if (install_method) {
            conditions.push(eq(publishedMcpServersTable.install_method, install_method as any));
        }
        if (search) {
            const term = `%${search}%`;
            conditions.push(
                or(
                    like(publishedMcpServersTable.display_name, term),
                    like(publishedMcpServersTable.description, term),
                    like(publishedMcpServersTable.author, term),
                    like(publishedMcpServersTable.canonical_id, term),
                )
            );
        }

        if (conditions.length > 0) {
            return await db
                .select()
                .from(publishedMcpServersTable)
                .where(conditions.length === 1 ? conditions[0] : and(...conditions))
                .orderBy(
                    desc(publishedMcpServersTable.confidence),
                    desc(publishedMcpServersTable.updated_at)
                )
                .limit(limit)
                .offset(offset);
        }

        return await query;
    }

    /**
     * Count total servers matching filters (for pagination).
     */
    async countServers(input: Omit<CatalogListInput, "limit" | "offset"> = {}): Promise<number> {
        const { search, status, transport, install_method } = input;

        const conditions = [];
        if (status) conditions.push(eq(publishedMcpServersTable.status, status));
        if (transport) conditions.push(eq(publishedMcpServersTable.transport, transport as any));
        if (install_method) conditions.push(eq(publishedMcpServersTable.install_method, install_method as any));
        if (search) {
            const term = `%${search}%`;
            conditions.push(
                or(
                    like(publishedMcpServersTable.display_name, term),
                    like(publishedMcpServersTable.description, term),
                    like(publishedMcpServersTable.canonical_id, term),
                )
            );
        }

        const [row] = await db
            .select({ count: sql<number>`count(*)` })
            .from(publishedMcpServersTable)
            .where(conditions.length > 0 ? (conditions.length === 1 ? conditions[0] : and(...conditions)) : undefined);

        return row?.count ?? 0;
    }

    /**
     * Count servers updated within the last N hours.
     * Used by dashboard freshness metrics.
     */
    async countRecentlyUpdated(hours = 24): Promise<number> {
        const threshold = new Date(Date.now() - Math.max(1, hours) * 60 * 60 * 1000);
        const [row] = await db
            .select({ count: sql<number>`count(*)` })
            .from(publishedMcpServersTable)
            .where(gte(publishedMcpServersTable.updated_at, threshold));

        return row?.count ?? 0;
    }

    async findServerByUuid(uuid: string): Promise<PublishedMcpServer | undefined> {
        const [row] = await db
            .select()
            .from(publishedMcpServersTable)
            .where(eq(publishedMcpServersTable.uuid, uuid));
        return row;
    }

    async findServerByCanonicalId(canonicalId: string): Promise<PublishedMcpServer | undefined> {
        const [row] = await db
            .select()
            .from(publishedMcpServersTable)
            .where(eq(publishedMcpServersTable.canonical_id, canonicalId));
        return row;
    }

    /**
     * Upsert a server by canonical_id.
     * On conflict, updates mutable fields (display_name, description, stars, last_seen_at, etc.)
     * but preserves status/confidence/created_at.
     * Returns the upserted row.
     */
    async upsertServer(input: UpsertServerInput): Promise<PublishedMcpServer> {
        const now = new Date();
        const payload = {
            uuid: randomUUID(),
            canonical_id: input.canonical_id,
            display_name: input.display_name,
            description: input.description ?? null,
            author: input.author ?? null,
            repository_url: input.repository_url ?? null,
            homepage_url: input.homepage_url ?? null,
            icon_url: input.icon_url ?? null,
            transport: (input.transport ?? "unknown") as any,
            install_method: (input.install_method ?? "unknown") as any,
            auth_model: (input.auth_model ?? "unknown") as any,
            tags: input.tags ?? [],
            categories: input.categories ?? [],
            stars: input.stars ?? null,
            last_seen_at: now,
            updated_at: now,
        };

        const [row] = await db
            .insert(publishedMcpServersTable)
            .values(payload)
            .onConflictDoUpdate({
                target: publishedMcpServersTable.canonical_id,
                set: {
                    display_name: sql`excluded.display_name`,
                    description: sql`excluded.description`,
                    author: sql`excluded.author`,
                    repository_url: sql`excluded.repository_url`,
                    homepage_url: sql`excluded.homepage_url`,
                    icon_url: sql`excluded.icon_url`,
                    transport: sql`excluded.transport`,
                    install_method: sql`excluded.install_method`,
                    auth_model: sql`excluded.auth_model`,
                    tags: sql`excluded.tags`,
                    categories: sql`excluded.categories`,
                    stars: sql`excluded.stars`,
                    last_seen_at: sql`excluded.last_seen_at`,
                    updated_at: sql`excluded.updated_at`,
                },
            })
            .returning();

        return row;
    }

    /**
     * Advance the status of a server (controlled by Archivist/Verifier).
     * Also updates confidence and last_verified_at when appropriate.
     */
    async updateServerStatus(
        uuid: string,
        status: (typeof PublishedServerStatusEnum)[number],
        confidence?: number
    ): Promise<void> {
        const now = new Date();
        const set: Record<string, unknown> = {
            status,
            updated_at: now,
        };
        if (confidence !== undefined) {
            set.confidence = confidence;
        }
        if (status === "validated" || status === "certified") {
            set.last_verified_at = now;
        }

        await db
            .update(publishedMcpServersTable)
            .set(set as any)
            .where(eq(publishedMcpServersTable.uuid, uuid));
    }

    // ---- published_mcp_server_sources ----

    /**
     * Upsert a source record.
     * On conflict (server_uuid, source_name) updates last_seen_at and raw_payload.
     */
    async upsertSource(input: UpsertSourceInput): Promise<PublishedMcpServerSource> {
        const now = new Date();
        const payload = {
            uuid: randomUUID(),
            server_uuid: input.server_uuid,
            source_name: input.source_name,
            source_url: input.source_url ?? null,
            raw_payload: input.raw_payload ?? null,
            first_seen_at: now,
            last_seen_at: now,
        };

        const [row] = await db
            .insert(publishedMcpServerSourcesTable)
            .values(payload)
            .onConflictDoUpdate({
                target: [
                    publishedMcpServerSourcesTable.server_uuid,
                    publishedMcpServerSourcesTable.source_name,
                ],
                set: {
                    source_url: sql`excluded.source_url`,
                    raw_payload: sql`excluded.raw_payload`,
                    last_seen_at: sql`excluded.last_seen_at`,
                },
            })
            .returning();

        return row;
    }

    async findSourcesByServerUuid(serverUuid: string): Promise<PublishedMcpServerSource[]> {
        return db
            .select()
            .from(publishedMcpServerSourcesTable)
            .where(eq(publishedMcpServerSourcesTable.server_uuid, serverUuid));
    }

    // ---- published_mcp_config_recipes ----

    /**
     * Create a new config recipe.
     * Deactivates all prior active recipes for this server first.
     */
    async createRecipe(input: CreateRecipeInput): Promise<PublishedMcpConfigRecipe> {
        // Deactivate prior active recipes for this server
        await db
            .update(publishedMcpConfigRecipesTable)
            .set({ is_active: false, updated_at: new Date() } as any)
            .where(
                and(
                    eq(publishedMcpConfigRecipesTable.server_uuid, input.server_uuid),
                    eq(publishedMcpConfigRecipesTable.is_active, true)
                )
            );

        // Get next version number
        const [versionRow] = await db
            .select({ max: sql<number>`coalesce(max(recipe_version), 0)` })
            .from(publishedMcpConfigRecipesTable)
            .where(eq(publishedMcpConfigRecipesTable.server_uuid, input.server_uuid));

        const nextVersion = (versionRow?.max ?? 0) + 1;

        const [row] = await db
            .insert(publishedMcpConfigRecipesTable)
            .values({
                uuid: randomUUID(),
                server_uuid: input.server_uuid,
                recipe_version: nextVersion,
                template: input.template,
                required_secrets: input.required_secrets ?? [],
                required_env: input.required_env ?? {},
                confidence: input.confidence ?? 0,
                explanation: input.explanation ?? null,
                is_active: true,
                generated_by: input.generated_by ?? "Configurator",
            })
            .returning();

        return row;
    }

    async getActiveRecipe(serverUuid: string): Promise<PublishedMcpConfigRecipe | undefined> {
        const [row] = await db
            .select()
            .from(publishedMcpConfigRecipesTable)
            .where(
                and(
                    eq(publishedMcpConfigRecipesTable.server_uuid, serverUuid),
                    eq(publishedMcpConfigRecipesTable.is_active, true)
                )
            )
            .orderBy(desc(publishedMcpConfigRecipesTable.recipe_version))
            .limit(1);
        return row;
    }

    async listRecipesForServer(serverUuid: string): Promise<PublishedMcpConfigRecipe[]> {
        return db
            .select()
            .from(publishedMcpConfigRecipesTable)
            .where(eq(publishedMcpConfigRecipesTable.server_uuid, serverUuid))
            .orderBy(desc(publishedMcpConfigRecipesTable.recipe_version));
    }

    // ---- published_mcp_validation_runs ----

    /**
     * Start a validation run (returns pending row).
     */
    async startValidationRun(input: CreateValidationRunInput): Promise<PublishedMcpValidationRun> {
        const [row] = await db
            .insert(publishedMcpValidationRunsTable)
            .values({
                uuid: randomUUID(),
                server_uuid: input.server_uuid,
                run_mode: input.run_mode,
                started_at: new Date(),
                outcome: "pending",
                performed_by: input.performed_by ?? "Verifier",
            })
            .returning();
        return row;
    }

    /**
     * Record the final outcome of a validation run.
     */
    async finishValidationRun(input: FinishValidationRunInput): Promise<void> {
        await db
            .update(publishedMcpValidationRunsTable)
            .set({
                outcome: input.outcome,
                finished_at: new Date(),
                failure_class: input.failure_class ?? null,
                tool_count: input.tool_count ?? null,
                findings_summary: input.findings_summary ?? null,
            } as any)
            .where(eq(publishedMcpValidationRunsTable.uuid, input.uuid));
    }

    async listRunsForServer(
        serverUuid: string,
        limit = 10
    ): Promise<PublishedMcpValidationRun[]> {
        return db
            .select()
            .from(publishedMcpValidationRunsTable)
            .where(eq(publishedMcpValidationRunsTable.server_uuid, serverUuid))
            .orderBy(desc(publishedMcpValidationRunsTable.created_at))
            .limit(limit);
    }

    async getLatestRun(serverUuid: string): Promise<PublishedMcpValidationRun | undefined> {
        const [row] = await db
            .select()
            .from(publishedMcpValidationRunsTable)
            .where(eq(publishedMcpValidationRunsTable.server_uuid, serverUuid))
            .orderBy(desc(publishedMcpValidationRunsTable.created_at))
            .limit(1);
        return row;
    }

    /**
     * List UUIDs of servers with the given status, up to a specified limit.
     * Used by batch validation to avoid loading full server rows.
     */
    async listUuidsByStatus(
        statuses: string[],
        limit = 20
    ): Promise<string[]> {
        if (statuses.length === 0) return [];

        const rows = await db
            .select({ uuid: publishedMcpServersTable.uuid })
            .from(publishedMcpServersTable)
            .where(inArray(publishedMcpServersTable.status, statuses as any[]))
            .orderBy(desc(publishedMcpServersTable.updated_at))
            .limit(limit);

        return rows.map((r) => r.uuid);
    }
}

// Singleton export — mirrors pattern used by mcpServersRepository, toolsRepository etc.
export const publishedCatalogRepository = new PublishedCatalogRepository();
