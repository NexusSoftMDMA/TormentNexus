/**
 * @file published-catalog-ingestor.ts
 * @module packages/core/src/services/published-catalog-ingestor
 *
 * WHAT:
 * The Archivist agent's ingestion subsystem. Fetches MCP server listings
 * from known external registries and upserts them into the published catalog.
 *
 * WHY:
 * Centralizing ingestion here allows the catalog to be populated from multiple
 * sources without each source becoming a separate maintenance burden. All
 * provenance is tracked in `published_mcp_server_sources`.
 *
 * HOW:
 * Each "adapter" implements the `CatalogSourceAdapter` interface and is registered
 * in `INGESTION_ADAPTERS`. The `ingestAll()` method runs all adapters sequentially.
 *
 * Currently implemented adapters:
 *   1. GlamaAiAdapter      — best-effort fetch against Glama's historical public listing endpoint
 *   2. SmitheryAiAdapter   — hits smithery.ai registry API
 *   3. McpRunAdapter       — best-effort fetch against mcp.run's historical listing endpoint (soft failure if unavailable)
 *   4. NpmRegistryAdapter  — searches npm for @modelcontextprotocol/* and mcp-server-* packages
 *   5. GitHubTopicAdapter  — searches GitHub repos tagged with topic:mcp-server (top 300 by stars)
 *
 * Security note: raw_payload is stored as JSON but NEVER executed. The Configurator
 * agent generates recipes separately based on the cataloged data. This adapter
 * only reads and stores metadata.
 */

import { publishedCatalogRepository } from "../db/repositories/published-catalog.repo.js";

function parseNpmPackageName(canonicalId: string): string | null {
    if (!canonicalId.startsWith("npm/")) return null;
    const pkg = canonicalId.slice(4).trim();
    return pkg.length > 0 ? pkg : null;
}

function readNpmVersionFromSources(sources: Array<{ source_name: string; raw_payload: unknown }>): string | null {
    for (const source of sources) {
        if (source.source_name !== "npm") continue;
        const payload = source.raw_payload;
        if (payload && typeof payload === "object" && !Array.isArray(payload)) {
            const version = (payload as Record<string, unknown>).version;
            if (typeof version === "string" && version.trim().length > 0) {
                return version.trim();
            }
        }
    }
    return null;
}

function normalizeRepoPath(repositoryUrl?: string | null): string | null {
    if (!repositoryUrl) return null;
    const m = repositoryUrl.match(/github\.com\/([^/]+)\/([^/#?]+)/i);
    if (!m) return null;
    return `${m[1]}/${m[2].replace(/\.git$/i, "")}`;
}

/**
 * Heuristically extract likely secret / auth env-var names for a catalog server.
 *
 * HOW:
 * 1. Check the auth_model field for known patterns ("oauth", "api_key", "bearer", "none").
 * 2. Scan description and tags for common credential placeholder patterns:
 *    - Anything matching UPPER_SNAKE_CASE ending in _API_KEY, _TOKEN, _SECRET, etc.
 *    - Common provider-specific names (GITHUB_TOKEN, OPENAI_API_KEY, etc.)
 * 3. Construct a best-guess env var name from the server name + suffix when pattern is
 *    weak (auth_model = api_key but no env var found in descriptions).
 *
 * Returns a deduplicated array of secret names to use in `required_secrets`.
 * Returns empty array when auth_model = "none" or no patterns detected.
 */
export function inferRequiredSecrets(server: {
    canonical_id: string;
    display_name: string;
    description?: string | null;
    auth_model?: string;
    tags?: string[];
    categories?: string[];
}): string[] {
    const authModel = (server.auth_model ?? "unknown").toLowerCase();

    // Explicitly no auth — don't add any secrets
    if (authModel === "none" || authModel === "public") {
        return [];
    }

    const candidates = new Set<string>();

    // Scan description + tags for UPPER_SNAKE_CASE secrets
    const scanTargets = [
        server.description ?? "",
        ...(server.tags ?? []),
        ...(server.categories ?? []),
    ].join(" ");

    // Pattern: UPPER_SNAKE_CASE env var name ending in a well-known secret suffix
    const pattern = new RegExp(
        `\\b([A-Z][A-Z0-9_]{1,}(?:_API_KEY|_TOKEN|_SECRET|_PASSWORD|_AUTH_TOKEN|_BEARER_TOKEN|_ACCESS_TOKEN|_ACCESS_KEY|_CLIENT_SECRET|_API_TOKEN|_AUTH_KEY|_PRIVATE_KEY|_SECRET_KEY))\\b`,
        "g"
    );
    const matches = scanTargets.matchAll(pattern);
    for (const m of matches) {
        if (m[1] && m[1].length <= 48) {
            candidates.add(m[1]);
        }
    }

    // Well-known provider env vars that appear in descriptions but without UPPER_SNAKE suffix
    const knownProviderPatterns: [RegExp, string][] = [
        [/github/i, "GITHUB_TOKEN"],
        [/openai/i, "OPENAI_API_KEY"],
        [/anthropic/i, "ANTHROPIC_API_KEY"],
        [/google|gemini/i, "GOOGLE_API_KEY"],
        [/slack/i, "SLACK_BOT_TOKEN"],
        [/stripe/i, "STRIPE_SECRET_KEY"],
        [/notion/i, "NOTION_API_KEY"],
        [/jira|atlassian/i, "JIRA_API_TOKEN"],
        [/linear/i, "LINEAR_API_KEY"],
        [/discord/i, "DISCORD_BOT_TOKEN"],
        [/twitter|x\.com/i, "TWITTER_API_KEY"],
        [/cloudflare/i, "CLOUDFLARE_API_TOKEN"],
        [/aws|amazon/i, "AWS_ACCESS_KEY_ID"],
        [/sendgrid/i, "SENDGRID_API_KEY"],
        [/twilio/i, "TWILIO_AUTH_TOKEN"],
        [/browserbase/i, "BROWSERBASE_API_KEY"],
        [/figma/i, "FIGMA_ACCESS_TOKEN"],
        [/gitlab/i, "GITLAB_TOKEN"],
        [/bitbucket/i, "BITBUCKET_TOKEN"],
        [/hubspot/i, "HUBSPOT_ACCESS_TOKEN"],
        [/salesforce/i, "SALESFORCE_ACCESS_TOKEN"],
    ];

    // Only add provider-specific hints if we don't already have candidates
    if (candidates.size === 0 && authModel !== "unknown") {
        const fullText = [server.canonical_id, server.display_name, scanTargets].join(" ");
        for (const [providerPat, envVar] of knownProviderPatterns) {
            if (providerPat.test(fullText)) {
                candidates.add(envVar);
                break; // One provider hint per server
            }
        }
    }

    // If auth_model signals auth is needed but we found nothing, synthesize a generic name
    if (candidates.size === 0 && (authModel === "api_key" || authModel === "bearer" || authModel === "token")) {
        const safeName = server.display_name
            .toUpperCase()
            .replace(/[^A-Z0-9]+/g, "_")
            .replace(/^_+|_+$/g, "")
            .slice(0, 24);
        if (safeName.length >= 2) {
            candidates.add(`${safeName}_API_KEY`);
        }
    }

    // Cap at 4 secrets to avoid false-positive overload
    return Array.from(candidates).slice(0, 4);
}

export function buildBaselineRecipe(
    server: {
        canonical_id: string;
        display_name: string;
        transport: string;
        install_method: string;
        repository_url?: string | null;
    },
    npmVersion: string | null,
): {
    template: Record<string, unknown>;
    confidence: number;
    explanation: string;
} {
    const packageName = parseNpmPackageName(server.canonical_id);
    const repoPath = normalizeRepoPath(server.repository_url ?? null);

    if (server.install_method === "npm" && packageName) {
        const pinnedSpec = npmVersion ? `${packageName}@${npmVersion}` : packageName;
        return {
            template: {
                type: "stdio",
                command: "npx",
                args: ["-y", pinnedSpec],
                env: {},
            },
            confidence: npmVersion ? 45 : 38,
            explanation: npmVersion
                ? `Baseline Configurator recipe (version-pinned): npx -y ${pinnedSpec}`
                : `Baseline Configurator recipe: npx -y ${pinnedSpec}`,
        };
    }

    if (server.install_method === "go-install" && repoPath) {
        return {
            template: {
                type: "stdio",
                command: "go",
                args: ["install", `${repoPath}@latest`],
                env: {},
            },
            confidence: 35,
            explanation: `Baseline Configurator recipe: go install ${repoPath}@latest`,
        };
    }

    if (server.install_method === "cargo" && repoPath) {
        const crate = repoPath.split("/")[1] ?? repoPath;
        return {
            template: {
                type: "stdio",
                command: "cargo",
                args: ["install", crate],
                env: {},
            },
            confidence: 34,
            explanation: `Baseline Configurator recipe: cargo install ${crate}`,
        };
    }

    if (server.install_method === "pip" && packageName) {
        return {
            template: {
                type: "stdio",
                command: "uvx",
                args: [packageName],
                env: {},
            },
            confidence: 32,
            explanation: `Baseline Configurator recipe: uvx ${packageName}`,
        };
    }

    // Docker-based servers: generate a `docker run` STDIO recipe.
    // We derive the image name from the canonical_id (e.g. github/<owner>/<repo>)
    // or fall back to the display name, lowercased.  The operator will need to
    // review the exact image tag and port/volume flags before install.
    if (server.install_method === "docker") {
        const imageGuess = repoPath
            ? repoPath.toLowerCase()
            : server.display_name.toLowerCase().replace(/\s+/g, "-");
        return {
            template: {
                type: "stdio",
                command: "docker",
                args: ["run", "--rm", "-i", imageGuess],
                env: {},
            },
            confidence: 28,
            explanation: `Baseline Configurator recipe (Docker): docker run --rm -i ${imageGuess} — operator review required to confirm image tag and port/volume flags.`,
        };
    }

    if (server.transport === "sse" || server.transport === "streamable_http") {
        return {
            template: {
                type: server.transport,
                url: "https://example.com/mcp",
                headers: {},
            },
            confidence: 22,
            explanation:
                "Baseline Configurator recipe created with placeholder URL. Operator review required before validation/install.",
        };
    }

    return {
        template: {
            type: "stdio",
            command: "npx",
            args: ["-y", server.display_name.toLowerCase().replace(/\s+/g, "-")],
            env: {},
        },
        confidence: 20,
        explanation: "Baseline Configurator fallback recipe generated from catalog metadata.",
    };
}

export type IngestResult = {
    source: string;
    /** How many records were fetched from the source */
    fetched: number;
    /** How many were upserted (new or updated) */
    upserted: number;
    errors: string[];
};

/**
 * A catalog source adapter must implement this interface.
 * `name` must be stable — it is stored as `source_name` in the DB.
 */
export interface CatalogSourceAdapter {
    readonly name: string;
    ingest(): Promise<IngestResult>;
}

// ---------------------------------------------------------------------------
// Utility: safe fetch with timeout + JSON parse
// ---------------------------------------------------------------------------

const FETCH_TIMEOUT_MS = 15_000;
const SEARCH_FETCH_TIMEOUT_MS = 30_000;

function toErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

function readHeader(
    headers: Response["headers"] | Record<string, string> | undefined,
    name: string,
): string | null {
    if (!headers) return null;
    if (typeof (headers as Response["headers"]).get === "function") {
        return (headers as Response["headers"]).get(name);
    }
    const record = headers as Record<string, string>;
    return record[name] ?? record[name.toLowerCase()] ?? record[name.toUpperCase()] ?? null;
}

function summarizeResponseText(body: string): string {
    const normalized = body.replace(/\s+/g, " ").trim();
    if (normalized.length === 0) return "<empty>";
    return normalized.slice(0, 160);
}

async function safeFetch(url: string, options?: { timeoutMs?: number }): Promise<unknown> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), options?.timeoutMs ?? FETCH_TIMEOUT_MS);
    try {
        const response = await fetch(url, {
            signal: controller.signal,
            headers: { "Accept": "application/json", "User-Agent": "TormentNexus/MCP-Catalog-Ingestor" },
        });
        const contentType = (readHeader(response.headers, "content-type") ?? "").toLowerCase();
        if (!response.ok) {
            if (contentType.includes("text/html")) {
                throw new Error(`HTTP ${response.status} (received HTML error page)`);
            }
            throw new Error(`HTTP ${response.status}`);
        }
        if (!contentType.includes("json")) {
            const body = typeof response.text === "function" ? await response.text() : "";
            throw new Error(
                `Expected JSON but received ${contentType || "unknown content-type"}: ${summarizeResponseText(body)}`
            );
        }
        try {
            return await response.json();
        } catch (e) {
            throw new Error(`Failed to parse JSON: ${toErrorMessage(e)}`);
        }
    } catch (err) {
        throw new Error(`Request to ${url} failed: ${toErrorMessage(err)}`);
    } finally {
        clearTimeout(timeout);
    }
}

// ---------------------------------------------------------------------------
// Adapter 1: Glama.ai
// Historical public endpoint: https://glama.ai/api/mcp/servers
// This source is currently best-effort because live responses may drift to HTML.
// ---------------------------------------------------------------------------

type GlamaServer = {
    id?: string;
    slug?: string;
    name?: string;
    description?: string;
    repository?: { url?: string; owner?: string; name?: string };
    vendor?: { name?: string };
    categories?: string[];
    tags?: string[];
    attributes?: { stars?: number };
    transport?: string;
};

export class GlamaAiAdapter implements CatalogSourceAdapter {
    readonly name = "glama.ai";
    // The original /api/mcp/servers endpoint now returns HTML.
    // Try multiple candidate endpoints in order; first success wins.
    private readonly candidateUrls = [
        "https://glama.ai/api/mcp/servers?limit=200",
        "https://glama.ai/api/mcp/servers",
    ];

    async ingest(): Promise<IngestResult> {
        const result: IngestResult = {
            source: this.name,
            fetched: 0,
            upserted: 0,
            errors: [],
        };

        try {
            // Glama returns a paginated or flat list — try each candidate URL
            let servers: GlamaServer[] = [];
            let lastError: unknown;
            for (const url of this.candidateUrls) {
                try {
                    const payload = await safeFetch(url) as any;
                    const candidates = Array.isArray(payload?.servers)
                        ? payload.servers
                        : Array.isArray(payload)
                        ? payload
                        : [];
                    if (candidates.length > 0) {
                        servers = candidates;
                        break;
                    }
                } catch (error) {
                    if (!lastError) {
                        lastError = error;
                    }
                }
            }

            if (servers.length === 0 && lastError) {
                throw lastError;
            }

            result.fetched = servers.length;

            for (const s of servers) {
                try {
                    const canonicalId = this.buildCanonicalId(s);
                    const server = await publishedCatalogRepository.upsertServer({
                        canonical_id: canonicalId,
                        display_name: s.name ?? canonicalId,
                        description: s.description ?? null,
                        author: s.vendor?.name ?? s.repository?.owner ?? null,
                        repository_url: s.repository?.url ?? null,
                        tags: [...(s.tags ?? []), ...(s.categories ?? [])],
                        categories: s.categories ?? [],
                        stars: s.attributes?.stars ?? null,
                        transport: this.normalizeTransport(s.transport),
                        install_method: "unknown",
                        auth_model: "unknown",
                    });

                    await publishedCatalogRepository.upsertSource({
                        server_uuid: server.uuid,
                        source_name: this.name,
                        source_url: `https://glama.ai/api/mcp/servers/${s.slug ?? s.id ?? ""}`,
                        raw_payload: s as Record<string, unknown>,
                    });

                    result.upserted++;
                } catch (err) {
                    result.errors.push(`glama server ${s.id ?? s.name}: ${String(err)}`);
                }
            }
        } catch (err) {
            result.errors.push(`fetch failed: ${String(err)}`);
        }

        return result;
    }

    private buildCanonicalId(s: GlamaServer): string {
        if (s.repository?.owner && s.repository?.name) {
            return `github/${s.repository.owner}/${s.repository.name}`;
        }
        if (s.slug) return `glama/${s.slug}`;
        if (s.id) return `glama/${s.id}`;
        return `glama/${(s.name ?? "unknown").toLowerCase().replace(/\s+/g, "-")}`;
    }

    private normalizeTransport(raw?: string): string {
        const t = (raw ?? "").toLowerCase();
        if (t.includes("stdio")) return "stdio";
        if (t.includes("sse")) return "sse";
        if (t.includes("http")) return "streamable_http";
        return "unknown";
    }
}

// ---------------------------------------------------------------------------
// Adapter 2: Smithery.ai
// Public endpoint: https://registry.smithery.ai/servers
// ---------------------------------------------------------------------------

type SmitheryServer = {
    qualifiedName?: string;
    displayName?: string;
    description?: string;
    homepage?: string;
    iconUrl?: string;
    useCount?: number;
    isDeployed?: boolean;
    transport?: string[];
    tags?: string[];
};

export class SmitheryAiAdapter implements CatalogSourceAdapter {
    readonly name = "smithery.ai";
    private readonly baseUrl = "https://registry.smithery.ai/servers";

    async ingest(): Promise<IngestResult> {
        const result: IngestResult = {
            source: this.name,
            fetched: 0,
            upserted: 0,
            errors: [],
        };

        try {
            const payload = await safeFetch(`${this.baseUrl}?limit=200`) as any;
            const servers: SmitheryServer[] = Array.isArray(payload?.servers)
                ? payload.servers
                : [];

            result.fetched = servers.length;

            for (const s of servers) {
                try {
                    const canonicalId = s.qualifiedName
                        ? `smithery/${s.qualifiedName.replace("@", "").replace(/\//g, "__")}`
                        : `smithery/${(s.displayName ?? "unknown").toLowerCase().replace(/\s+/g, "-")}`;

                    const transport = Array.isArray(s.transport) && s.transport.length > 0
                        ? this.normalizeTransport(s.transport[0])
                        : "unknown";

                    const installMethod = s.isDeployed ? "url" : "unknown";

                    const server = await publishedCatalogRepository.upsertServer({
                        canonical_id: canonicalId,
                        display_name: s.displayName ?? canonicalId,
                        description: s.description ?? null,
                        author: null,
                        homepage_url: s.homepage ?? null,
                        icon_url: s.iconUrl ?? null,
                        tags: s.tags ?? [],
                        categories: [],
                        transport,
                        install_method: installMethod,
                        auth_model: "unknown",
                    });

                    await publishedCatalogRepository.upsertSource({
                        server_uuid: server.uuid,
                        source_name: this.name,
                        source_url: s.homepage ?? `${this.baseUrl}/${s.qualifiedName ?? ""}`,
                        raw_payload: s as Record<string, unknown>,
                    });

                    result.upserted++;
                } catch (err) {
                    result.errors.push(`smithery server ${s.qualifiedName ?? s.displayName}: ${String(err)}`);
                }
            }
        } catch (err) {
            result.errors.push(`fetch failed: ${String(err)}`);
        }

        return result;
    }

    private normalizeTransport(raw: string): string {
        const t = raw.toLowerCase();
        if (t.includes("stdio")) return "stdio";
        if (t.includes("sse")) return "sse";
        if (t.includes("http")) return "streamable_http";
        return "unknown";
    }
}

// ---------------------------------------------------------------------------
// Adapter 3: MCP.run public registry
// Historical public endpoint: https://mcp.run/api/servers
// Keep this source non-fatal until a live listing endpoint is re-verified.
// ---------------------------------------------------------------------------

export class McpRunAdapter implements CatalogSourceAdapter {
    readonly name = "mcp.run";
    private readonly baseUrl = "https://mcp.run/api/servers";

    async ingest(): Promise<IngestResult> {
        const result: IngestResult = {
            source: this.name,
            fetched: 0,
            upserted: 0,
            errors: [],
        };

        try {
            const payload = await safeFetch(this.baseUrl) as any;
            const servers: Array<Record<string, unknown>> = Array.isArray(payload)
                ? payload
                : Array.isArray(payload?.data)
                ? payload.data
                : [];

            result.fetched = servers.length;

            for (const s of servers) {
                try {
                    const name = String(s.name ?? s.id ?? "unknown");
                    const canonicalId = `mcp.run/${name.replace(/\s+/g, "-").toLowerCase()}`;

                    const server = await publishedCatalogRepository.upsertServer({
                        canonical_id: canonicalId,
                        display_name: String(s.displayName ?? s.name ?? canonicalId),
                        description: s.description ? String(s.description) : null,
                        author: s.author ? String(s.author) : null,
                        homepage_url: s.url ? String(s.url) : null,
                        tags: Array.isArray(s.tags) ? s.tags.map(String) : [],
                        transport: "unknown",
                        install_method: "url",
                        auth_model: "unknown",
                    });

                    await publishedCatalogRepository.upsertSource({
                        server_uuid: server.uuid,
                        source_name: this.name,
                        source_url: `${this.baseUrl}/${name}`,
                        raw_payload: s,
                    });

                    result.upserted++;
                } catch (err) {
                    result.errors.push(`mcp.run server ${String(s.name ?? s.id)}: ${String(err)}`);
                }
            }
        } catch (err) {
            // Not all environments can reach mcp.run — treat as soft failure
            result.errors.push(`fetch failed (non-fatal): ${String(err)}`);
        }

        return result;
    }
}

// ---------------------------------------------------------------------------
// Adapter 4: npm registry search
// Searches npm for packages matching common MCP server naming patterns:
//   - Packages in the @modelcontextprotocol scope
//   - Packages with "mcp" as a keyword
//   - Packages named "mcp-server-*" or "@mcp-*"
//
// npm search API: https://registry.npmjs.org/-/v1/search
//   Supports `text`, `size` (max 250), and result fields include package.links.repository
//
// WHY: npm is the largest registry for STDIO-based MCP servers because most
// are Node.js packages run via `npx` or `node`. This gives far more coverage
// than the Glama/Smithery lists alone.
// ---------------------------------------------------------------------------

type NpmSearchObject = {
    package: {
        name: string;
        scope?: string;
        version?: string;
        description?: string;
        keywords?: string[];
        date?: string;
        links?: {
            npm?: string;
            homepage?: string;
            repository?: string;
        };
        author?: { name?: string; username?: string };
        publisher?: { username?: string };
    };
    score?: { final?: number };
};

type NpmSearchResponse = {
    objects?: NpmSearchObject[];
    total?: number;
};

export class NpmRegistryAdapter implements CatalogSourceAdapter {
    readonly name = "npm.registry";
    private readonly searchBase = "https://registry.npmjs.org/-/v1/search";

    async ingest(): Promise<IngestResult> {
        const result: IngestResult = {
            source: this.name,
            fetched: 0,
            upserted: 0,
            errors: [],
        };

        // Run multiple focused queries to maximize coverage.
        // Each query can return up to 250 results; we cap at a few queries to
        // stay well within NPM's rate limits for this non-authenticated path.
        const queries = [
            // Official MCP scope
            "scope:modelcontextprotocol",
            // Keyword-tagged packages
            "keywords:mcp-server",
            // Common naming patterns
            "mcp-server",
        ];

        const seen = new Set<string>();

        for (const text of queries) {
            try {
                const url = `${this.searchBase}?text=${encodeURIComponent(text)}&size=250`;
                const payload = await safeFetch(url, { timeoutMs: SEARCH_FETCH_TIMEOUT_MS }) as NpmSearchResponse;
                const objects = payload?.objects ?? [];

                for (const obj of objects) {
                    const pkg = obj.package;
                    if (!pkg?.name) continue;

                    // Deduplicate within this ingestion run
                    if (seen.has(pkg.name)) continue;
                    seen.add(pkg.name);

                    result.fetched++;

                    // Screen out packages that are very unlikely to be MCP servers.
                    // We keep all @modelcontextprotocol/* and keyword-tagged ones,
                    // but for the "mcp-server" text query we check the name pattern.
                    if (
                        pkg.scope !== "modelcontextprotocol" &&
                        !(pkg.keywords ?? []).some((k) =>
                            k === "mcp" || k === "mcp-server" || k === "model-context-protocol"
                        ) &&
                        !pkg.name.includes("mcp-server") &&
                        !pkg.name.includes("mcp_server")
                    ) {
                        // Skip — probably a false positive from generic text match
                        continue;
                    }

                    try {
                        // Determine install method (npm - standard for Node.js packages)
                        // and infer transport from description/keywords
                        const transport = this.inferTransport(pkg.description ?? "", pkg.keywords ?? []);

                        // Canonical ID uses npm package name as stable identifier
                        const canonicalId = `npm/${pkg.name}`;

                        const server = await publishedCatalogRepository.upsertServer({
                            canonical_id: canonicalId,
                            display_name: pkg.name,
                            description: pkg.description ?? null,
                            author:
                                pkg.author?.name ??
                                pkg.author?.username ??
                                pkg.publisher?.username ??
                                null,
                            repository_url: pkg.links?.repository ?? null,
                            homepage_url: pkg.links?.homepage ?? pkg.links?.npm ?? null,
                            tags: pkg.keywords ?? [],
                            categories: [],
                            transport,
                            // npm packages are installed via npm/npx
                            install_method: "npm",
                            auth_model: "unknown",
                            stars: null,
                        });

                        await publishedCatalogRepository.upsertSource({
                            server_uuid: server.uuid,
                            source_name: this.name,
                            source_url: pkg.links?.npm ?? `https://www.npmjs.com/package/${pkg.name}`,
                            raw_payload: pkg as unknown as Record<string, unknown>,
                        });

                        result.upserted++;
                    } catch (err) {
                        result.errors.push(`npm package ${pkg.name}: ${String(err)}`);
                    }
                }
            } catch (err) {
                result.errors.push(`npm query "${text}" failed: ${String(err)}`);
            }
        }

        return result;
    }

    /**
     * Infer transport type from package description and keywords.
     * Most npm MCP servers use STDIO — hosted/SSE variants mention it explicitly.
     */
    private inferTransport(description: string, keywords: string[]): string {
        const combined = (description + " " + keywords.join(" ")).toLowerCase();
        if (combined.includes("sse") || combined.includes("server-sent")) return "sse";
        if (combined.includes("streamable") || combined.includes("http server")) return "streamable_http";
        // Default for npm packages: stdio (run via npx/node)
        return "stdio";
    }
}

// ---------------------------------------------------------------------------
// Adapter 5: GitHub topic: mcp-server
// Searches GitHub's public repository search API for repos tagged with the
// `mcp-server` topic. Requires no auth for up to 10 requests/min (unauthenticated).
// We fetch up to 3 pages × 100 items = 300 repos, sorted by stars descending.
// ---------------------------------------------------------------------------

type GitHubRepo = {
    id: number;
    full_name: string;
    name: string;
    description: string | null;
    html_url: string;
    clone_url?: string;
    stargazers_count: number;
    topics?: string[];
    language?: string | null;
    owner?: { login?: string; type?: string };
    homepage?: string | null;
    default_branch?: string;
};

type GitHubSearchResponse = {
    total_count?: number;
    items?: GitHubRepo[];
};

export class GitHubTopicAdapter implements CatalogSourceAdapter {
    readonly name = "github-topics";
    private readonly searchBase = "https://api.github.com/search/repositories";
    /** Max pages to fetch. 100 items per page, unauthenticated rate limit = 10 req/min. */
    private readonly maxPages = 3;

    async ingest(): Promise<IngestResult> {
        const result: IngestResult = {
            source: this.name,
            fetched: 0,
            upserted: 0,
            errors: [],
        };

        const seen = new Set<string>();

        for (let page = 1; page <= this.maxPages; page++) {
            const url =
                `${this.searchBase}?q=topic:mcp-server` +
                `&sort=stars&order=desc&per_page=100&page=${page}`;

            try {
                const payload = await safeFetch(url, { timeoutMs: SEARCH_FETCH_TIMEOUT_MS }) as GitHubSearchResponse;
                const items: GitHubRepo[] = payload?.items ?? [];

                if (items.length === 0) break; // no more pages

                for (const repo of items) {
                    const key = `github/${repo.full_name}`;
                    if (seen.has(key)) continue;
                    seen.add(key);

                    result.fetched++;

                    try {
                        const transport = this.inferTransport(
                            repo.description ?? "",
                            repo.topics ?? [],
                            repo.language ?? "",
                        );

                        const server = await publishedCatalogRepository.upsertServer({
                            canonical_id: key,
                            display_name: repo.name,
                            description: repo.description ?? null,
                            author: repo.owner?.login ?? null,
                            repository_url: repo.html_url,
                            homepage_url: repo.homepage ?? null,
                            tags: repo.topics ?? [],
                            categories: [],
                            stars: repo.stargazers_count,
                            transport,
                            install_method: this.inferInstallMethod(repo.topics ?? [], repo.language ?? ""),
                            auth_model: "unknown",
                        });

                        await publishedCatalogRepository.upsertSource({
                            server_uuid: server.uuid,
                            source_name: this.name,
                            source_url: repo.html_url,
                            raw_payload: repo as unknown as Record<string, unknown>,
                        });

                        result.upserted++;
                    } catch (err) {
                        result.errors.push(`github repo ${repo.full_name}: ${String(err)}`);
                    }
                }

                // Respect GitHub's unauthenticated rate limit: add a small delay between pages
                if (page < this.maxPages && items.length === 100) {
                    await new Promise<void>(res => setTimeout(res, 1500));
                }
            } catch (err) {
                result.errors.push(`github page ${page}: ${String(err)}`);
                break; // abort pagination on network error
            }
        }

        return result;
    }

    private inferTransport(description: string, topics: string[], language: string): string {
        const combined = (description + " " + topics.join(" ") + " " + language).toLowerCase();
        if (combined.includes("sse") || combined.includes("server-sent")) return "sse";
        if (combined.includes("streamable") || combined.includes("http server")) return "streamable_http";
        // Most GitHub-hosted MCP servers expose stdio; default to that
        return "stdio";
    }

    private inferInstallMethod(topics: string[], language: string): string {
        const lang = language.toLowerCase();
        // Docker-first: repos that tag themselves as containerized
        if (topics.includes("docker") || topics.includes("dockerfile") || topics.includes("container")) return "docker";
        if (lang === "python" || topics.includes("python")) return "pip";
        if (lang === "go") return "go-install";
        if (lang === "rust" || topics.includes("rust")) return "cargo";
        // JavaScript/TypeScript and unknown → npm/npx
        return "npm";
    }
}

// ---------------------------------------------------------------------------
// Main ingestion service
// ---------------------------------------------------------------------------

/** All configured adapters. Add new ones here. */
const INGESTION_ADAPTERS: CatalogSourceAdapter[] = [
    new GlamaAiAdapter(),
    new SmitheryAiAdapter(),
    new McpRunAdapter(),
    new NpmRegistryAdapter(),
    new GitHubTopicAdapter(),
];

export type IngestionReport = {
    started_at: string;
    finished_at: string;
    results: IngestResult[];
    total_upserted: number;
    total_errors: number;
};

function summarizeCatalogErrors(errors: string[], limit: number = 2): string {
    if (errors.length === 0) {
        return 'none';
    }

    const preview = errors.slice(0, limit).join(' | ');
    const remaining = errors.length - limit;
    return remaining > 0
        ? `${preview} | +${remaining} more`
        : preview;
}

function summarizeCatalogFailure(error: unknown): string {
    const message = error instanceof Error ? error.message : String(error);
    return message.replace(/\s+/g, ' ').trim();
}

/**
 * Run all registered ingestion adapters and return a consolidated report.
 * Each adapter's failures are captured without aborting the others.
 *
 * After ingestion, newly discovered servers that have enough normalized data
 * (display_name + description + a known transport) are automatically advanced
 * from `discovered` → `normalized` by the Archivist.
 *
 * Usage (from a cron / admin endpoint):
 *   const report = await ingestPublishedCatalog();
 */
export async function ingestPublishedCatalog(
    adapters: CatalogSourceAdapter[] = INGESTION_ADAPTERS
): Promise<IngestionReport> {
    const startedAt = new Date().toISOString();
    const results: IngestResult[] = [];

    for (const adapter of adapters) {
        const result = await adapter.ingest();
        results.push(result);
        console.info(
            `[CatalogIngestor] ${adapter.name}: fetched=${result.fetched} upserted=${result.upserted} errors=${result.errors.length}`
        );
        if (result.errors.length > 0) {
            console.warn(
                `[CatalogIngestor] ${adapter.name} warnings: ${summarizeCatalogErrors(result.errors)}`
            );
        }
    }

    // --- Archivist normalization pass ---
    // Advance `discovered` servers with sufficient metadata to `normalized`.
    // This keeps status advancement automatic and avoids the catalog being
    // stuck at `discovered` for well-understood npm/glama entries.
    try {
        const discovered = await publishedCatalogRepository.listServers({ status: "discovered", limit: 200 });
        let normalized = 0;
        for (const server of discovered) {
            // Qualify as normalized if: has description, and transport is known
            if (server.description && server.description.trim().length > 10 && server.transport !== "unknown") {
                await publishedCatalogRepository.updateServerStatus(server.uuid, "normalized", 30);
                normalized++;
            }
        }
        if (normalized > 0) {
            console.info(`[CatalogIngestor] Normalization pass: advanced ${normalized} discovered → normalized`);
        }
    } catch (err) {
        console.warn(`[CatalogIngestor] Normalization pass failed: ${summarizeCatalogFailure(err)}`);
    }

    // --- Configurator baseline recipe pass ---
    // Ensure each ingested server has at least one active baseline recipe so operators
    // can install from metadata immediately and improve the recipe over time.
    try {
        const servers = await publishedCatalogRepository.listServers({ limit: 400 });
        let created = 0;

        for (const server of servers) {
            const active = await publishedCatalogRepository.getActiveRecipe(server.uuid);
            if (active) continue;

            const sources = await publishedCatalogRepository.findSourcesByServerUuid(server.uuid);
            const npmVersion = readNpmVersionFromSources(sources as Array<{ source_name: string; raw_payload: unknown }>);
            const baseline = buildBaselineRecipe(server, npmVersion);

            await publishedCatalogRepository.createRecipe({
                server_uuid: server.uuid,
                template: baseline.template,
                required_secrets: inferRequiredSecrets(server),
                required_env: {},
                confidence: baseline.confidence,
                explanation: baseline.explanation,
                generated_by: "Configurator",
            });

            created++;
        }

        if (created > 0) {
            console.info(`[CatalogIngestor] Baseline recipe pass: created ${created} active recipes`);
        }
    } catch (err) {
        console.warn(`[CatalogIngestor] Baseline recipe pass failed: ${summarizeCatalogFailure(err)}`);
    }

    const finishedAt = new Date().toISOString();
    return {
        started_at: startedAt,
        finished_at: finishedAt,
        results,
        total_upserted: results.reduce((acc, r) => acc + r.upserted, 0),
        total_errors: results.reduce((acc, r) => acc + r.errors.length, 0),
    };
}
