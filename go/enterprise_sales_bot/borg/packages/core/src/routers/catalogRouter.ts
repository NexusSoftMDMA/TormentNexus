/**
 * @file catalogRouter.ts
 * @module packages/core/src/routers/catalogRouter
 *
 * WHAT:
 * tRPC router for the MCP Registry Intelligence Published Catalog feature.
 * Exposes read-only catalog queries and admin-only ingestion/validation triggers.
 *
 * WHY:
 * This router is the API boundary between the dashboard UI and the
 * catalog data layer. All catalog logic stays in services; this router just
 * wires inputs → services → typed outputs.
 *
 * HOW:
 * - publicProcedure: list, search, get — safe for any authenticated session
 * - adminProcedure: ingest trigger, validate trigger — restricted to admins
 * - Graceful degradation: if ingestion/validation fails, error is returned
 *   without crashing the server (consistent with TormentNexus's resilience patterns)
 */

import { z } from "zod";
import { t, publicProcedure, adminProcedure } from "../lib/trpc-core.js";
import { publishedCatalogRepository } from "../db/repositories/published-catalog.repo.js";
import { mcpServersRepository } from "../db/repositories/mcp-servers.repo.js";
import { ingestPublishedCatalog } from "../services/published-catalog-ingestor.js";
import { validatePublishedServer } from "../services/published-catalog-validator.js";
import { rethrowSqliteUnavailableAsTrpc } from "./sqliteTrpc.js";

function toSafeServerName(input: string): string {
    return input
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9_-]+/g, "-")
        .replace(/-{2,}/g, "-")
        .replace(/_{2,}/g, "_")
        .replace(/^[-_]+|[-_]+$/g, "")
        .slice(0, 64);
}

async function ensureUniqueServerName(baseName: string): Promise<string> {
    let candidate = toSafeServerName(baseName) || "catalog-server";
    let suffix = 1;

    while (await mcpServersRepository.findByName(candidate)) {
        suffix += 1;
        candidate = `${toSafeServerName(baseName) || "catalog-server"}-${suffix}`.slice(0, 64);
    }

    return candidate;
}

function readString(value: unknown): string | undefined {
    return typeof value === "string" && value.trim().length > 0 ? value.trim() : undefined;
}

function readStringArray(value: unknown): string[] {
    if (!Array.isArray(value)) return [];
    return value.filter((item): item is string => typeof item === "string" && item.length > 0);
}

function readStringMap(value: unknown): Record<string, string> {
    if (!value || typeof value !== "object" || Array.isArray(value)) {
        return {};
    }

    const out: Record<string, string> = {};
    for (const [key, raw] of Object.entries(value)) {
        if (typeof raw === "string") {
            out[key] = raw;
        }
    }
    return out;
}

function normalizeType(rawType: unknown, hasCommand: boolean, hasUrl: boolean): "STDIO" | "SSE" | "STREAMABLE_HTTP" {
    const parsed = typeof rawType === "string" ? rawType.toLowerCase() : "";

    if (parsed === "stdio") return "STDIO";
    if (parsed === "sse") return "SSE";
    if (parsed === "streamable_http" || parsed === "streamablehttp") return "STREAMABLE_HTTP";

    if (hasCommand) return "STDIO";
    if (hasUrl) return "STREAMABLE_HTTP";

    return "STDIO";
}

// ---- Shared output schemas ----

const publishedServerSchema = z.object({
    uuid: z.string(),
    canonical_id: z.string(),
    display_name: z.string(),
    description: z.string().nullable(),
    author: z.string().nullable(),
    repository_url: z.string().nullable(),
    homepage_url: z.string().nullable(),
    icon_url: z.string().nullable(),
    transport: z.string(),
    install_method: z.string(),
    auth_model: z.string(),
    status: z.string(),
    confidence: z.number(),
    tags: z.array(z.string()),
    categories: z.array(z.string()),
    stars: z.number().nullable(),
    last_seen_at: z.date().nullable(),
    last_verified_at: z.date().nullable(),
    created_at: z.date(),
    updated_at: z.date(),
});

const validationRunSchema = z.object({
    uuid: z.string(),
    server_uuid: z.string(),
    run_mode: z.string(),
    started_at: z.date(),
    finished_at: z.date().nullable(),
    outcome: z.string(),
    failure_class: z.string().nullable(),
    tool_count: z.number().nullable(),
    findings_summary: z.record(z.unknown()).nullable(),
    performed_by: z.string(),
    created_at: z.date(),
});

const configRecipeSchema = z.object({
    uuid: z.string(),
    server_uuid: z.string(),
    recipe_version: z.number(),
    template: z.record(z.unknown()),
    required_secrets: z.array(z.string()),
    required_env: z.record(z.string()),
    confidence: z.number(),
    explanation: z.string().nullable(),
    is_active: z.boolean(),
    generated_by: z.string(),
    created_at: z.date(),
    updated_at: z.date(),
});

// ---- Router ----

export const catalogRouter = t.router({
    /**
     * List published servers with optional filtering and pagination.
     * Used by the Registry dashboard table.
     */
    list: publicProcedure
        .input(
            z.object({
                limit: z.number().min(1).max(200).default(50),
                offset: z.number().min(0).default(0),
                search: z.string().optional(),
                status: z
                    .enum([
                        "discovered",
                        "normalized",
                        "probeable",
                        "validated",
                        "certified",
                        "broken",
                        "archived",
                    ])
                    .optional(),
                transport: z.string().optional(),
                install_method: z.string().optional(),
            }).optional()
        )
        .query(async ({ input }) => {
            try {
                const [servers, total] = await Promise.all([
                    publishedCatalogRepository.listServers(input),
                    publishedCatalogRepository.countServers(input),
                ]);
                return { servers, total };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),

    /**
     * Get a single published server by UUID, including its latest validation run
     * and active recipe.
     */
    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                const [server, latestRun, activeRecipe, sources] = await Promise.all([
                    publishedCatalogRepository.findServerByUuid(input.uuid),
                    publishedCatalogRepository
                        .listRunsForServer(input.uuid, 1)
                        .then((runs) => runs[0] ?? null),
                    publishedCatalogRepository.getActiveRecipe(input.uuid),
                    publishedCatalogRepository.findSourcesByServerUuid(input.uuid),
                ]);

                return {
                    server: server ?? null,
                    latestRun: latestRun ?? null,
                    activeRecipe: activeRecipe ?? null,
                    sources: sources.map((s) => ({
                        uuid: s.uuid,
                        source_name: s.source_name,
                        source_url: s.source_url ?? null,
                        first_seen_at: s.first_seen_at,
                        last_seen_at: s.last_seen_at,
                    })),
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),

    /**
     * Get validation run history for a server.
     */
    listRuns: publicProcedure
        .input(z.object({ server_uuid: z.string(), limit: z.number().min(1).max(50).default(10) }))
        .query(async ({ input }) => {
            try {
                return await publishedCatalogRepository.listRunsForServer(input.server_uuid, input.limit);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),

    /**
     * Trigger ingestion from all registered external registry sources.
     * Admin-only. Returns a report of what was fetched and upserted.
     */
    triggerIngestion: adminProcedure
        .mutation(async () => {
            const report = await ingestPublishedCatalog();
            return {
                started_at: report.started_at,
                finished_at: report.finished_at,
                total_upserted: report.total_upserted,
                total_errors: report.total_errors,
                results: report.results.map((r) => ({
                    source: r.source,
                    fetched: r.fetched,
                    upserted: r.upserted,
                    error_count: r.errors.length,
                    // First 5 errors for diagnostics (avoid leaking full stack traces)
                    sample_errors: r.errors.slice(0, 5),
                })),
            };
        }),

    /**
     * Trigger validation for a specific server.
     * Admin-only. Returns the validation outcome.
     */
    triggerValidation: adminProcedure
        .input(z.object({ server_uuid: z.string() }))
        .mutation(async ({ input }) => {
            const output = await validatePublishedServer(input.server_uuid);
            return {
                run_uuid: output.run_uuid,
                outcome: output.outcome,
                failure_class: output.failure_class ?? null,
                tool_count: output.tool_count ?? null,
                findings_summary: output.findings_summary ?? null,
            };
        }),

    /**
     * Install a validated/certified published server recipe into managed MCP servers.
     * Admin-only. Uses active recipe template with optional env overrides.
     */
    installFromRecipe: adminProcedure
        .input(
            z.object({
                server_uuid: z.string(),
                env: z.record(z.string()).optional(),
                name: z.string().min(1).max(64).optional(),
            })
        )
        .mutation(async ({ input }) => {
            try {
                const [server, recipe] = await Promise.all([
                    publishedCatalogRepository.findServerByUuid(input.server_uuid),
                    publishedCatalogRepository.getActiveRecipe(input.server_uuid),
                ]);

                if (!server) {
                    throw new Error("Published server not found.");
                }
                if (!recipe) {
                    throw new Error("No active install recipe exists for this server.");
                }
                if (server.status !== "validated" && server.status !== "certified") {
                    throw new Error("Only validated or certified catalog entries can be installed.");
                }

                const template = recipe.template;
                if (!template || typeof template !== "object" || Array.isArray(template)) {
                    throw new Error("Recipe template is invalid.");
                }

                const baseName = input.name ?? server.canonical_id ?? server.display_name;
                const name = await ensureUniqueServerName(baseName);

                const command = readString((template as Record<string, unknown>).command);
                const args = readStringArray((template as Record<string, unknown>).args);
                const url = readString((template as Record<string, unknown>).url);
                const bearerToken = readString((template as Record<string, unknown>).bearerToken);
                const headers = readStringMap((template as Record<string, unknown>).headers);

                const templateEnv = readStringMap((template as Record<string, unknown>).env);
                const mergedEnv = {
                    ...recipe.required_env,
                    ...templateEnv,
                    ...(input.env ?? {}),
                };

                const missingSecrets = recipe.required_secrets.filter((key) => {
                    const value = mergedEnv[key];
                    return !value || value.trim().length === 0;
                });

                if (missingSecrets.length > 0) {
                    throw new Error(`Missing required secret values: ${missingSecrets.join(", ")}`);
                }

                const type = normalizeType((template as Record<string, unknown>).type, Boolean(command), Boolean(url));

                const created = await mcpServersRepository.create({
                    name,
                    description:
                        server.description ??
                        `Installed from published catalog recipe (${server.display_name})`,
                    type,
                    command: type === "STDIO" ? command ?? null : null,
                    args: type === "STDIO" ? args : [],
                    env: mergedEnv,
                    url: type === "STDIO" ? null : url ?? null,
                    bearerToken: type === "STDIO" ? null : bearerToken ?? null,
                    headers: type === "STDIO" ? {} : headers,
                    always_on: false,
                    user_id: "system",
                    source_published_server_uuid: input.server_uuid,
                });

                return {
                    installed: true,
                    server_uuid: created.uuid,
                    name: created.name,
                    source_published_server_uuid: input.server_uuid,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),

    /**
     * Trigger batch validation for servers matching a status filter.
     * Admin-only. Validates up to `max_servers` servers sequentially.
     * Defaults to validating all servers with status "normalized" or "probeable".
     *
     * WHY: Manual per-row validation is fine for spot-checks, but operators
     * need a way to bulk-validate newly ingested servers after a sync.
     */
    triggerBatchValidation: adminProcedure
        .input(
            z.object({
                statuses: z.array(z.string()).default(["normalized", "probeable"]),
                max_servers: z.number().min(1).max(50).default(10),
            })
        )
        .mutation(async ({ input }) => {
            try {
                const statuses = input.statuses;
                const maxServers = input.max_servers;

                const uuids = await publishedCatalogRepository.listUuidsByStatus(statuses, maxServers);

                if (uuids.length === 0) {
                    return { queued: 0, results: [] };
                }

                // Validate sequentially to avoid overwhelming remote servers.
                // Each call is bounded by validatePublishedServer's internal timeout.
                const results: Array<{
                    server_uuid: string;
                    run_uuid: string;
                    outcome: string;
                    failure_class: string | null;
                    tool_count: number | null;
                }> = [];

                for (const uuid of uuids) {
                    try {
                        const out = await validatePublishedServer(uuid);
                        results.push({
                            server_uuid: uuid,
                            run_uuid: out.run_uuid,
                            outcome: out.outcome,
                            failure_class: out.failure_class ?? null,
                            tool_count: out.tool_count ?? null,
                        });
                    } catch {
                        results.push({
                            server_uuid: uuid,
                            run_uuid: "error",
                            outcome: "error",
                            failure_class: "internal_error",
                            tool_count: null,
                        });
                    }
                }

                const passed = results.filter((r) => r.outcome === "passed").length;
                const failed = results.filter((r) => r.outcome === "failed" || r.outcome === "error").length;
                const skipped = results.filter((r) => r.outcome === "skipped").length;

                return { queued: uuids.length, passed, failed, skipped, results };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),

    /**
     * Get catalog statistics for the dashboard summary cards.
     */
    stats: publicProcedure
        .query(async () => {
            try {
                const [total, validated, broken, recentlyUpdated] = await Promise.all([
                    publishedCatalogRepository.countServers({}),
                    publishedCatalogRepository.countServers({ status: "validated" }),
                    publishedCatalogRepository.countServers({ status: "broken" }),
                    publishedCatalogRepository.countRecentlyUpdated(24),
                ]);

                const statuses = [
                    "discovered",
                    "normalized",
                    "probeable",
                    "validated",
                    "certified",
                    "broken",
                    "archived",
                ] as const;

                const statusCounts = await Promise.all(
                    statuses.map(async (s) => ({
                        status: s,
                        count: await publishedCatalogRepository.countServers({ status: s }),
                    }))
                );

                const byStatus = Object.fromEntries(statusCounts.map((s) => [s.status, s.count]));

                return {
                    total,
                    byStatus,
                    validated,
                    broken,
                    recentlyUpdated,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),

    /**
     * List managed MCP servers that were installed from a specific published catalog entry.
     * Public-accessible. Returns server records linked via `source_published_server_uuid`.
     */
    listLinkedServers: publicProcedure
        .input(z.object({ published_server_uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                const servers = await mcpServersRepository.findAll();
                return (servers ?? []).filter((s) => s.source_published_server_uuid === input.published_server_uuid);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc("Published catalog is unavailable", error);
            }
        }),
});
