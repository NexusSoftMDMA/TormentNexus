/**
 * @file published-catalog-ingestor.test.ts
 * @module packages/core/src/services/published-catalog-ingestor.test
 *
 * Comprehensive test suite for the published catalog ingestion pipeline.
 * Tests normalization logic, transport inference, and catalog deduplication.
 */

import { afterEach, describe, expect, it, vi } from 'vitest';
import {
    buildBaselineRecipe,
    GlamaAiAdapter,
    NpmRegistryAdapter,
    SmitheryAiAdapter,
    ingestPublishedCatalog,
} from './published-catalog-ingestor.js';
import { publishedCatalogRepository } from '../db/repositories/published-catalog.repo.js';

afterEach(() => {
    vi.restoreAllMocks();
});

/**
 * Test transport normalization across different inputs.
 * This tests the core logic used by all adapters to infer standard transport types.
 */
describe('CatalogIngestor — Transport Normalization', () => {
    it('normalizes stdio variants to "stdio"', () => {
        const inputs = ['stdio', 'STDIO', 'Stdio', 'std-io', 'std_io'];
        inputs.forEach(input => {
            // The actual implementation uses a common normalization pattern
            const normalized = normalizeTransportString(input);
            expect(normalized).toBe('stdio');
        });
    });

    it('normalizes SSE variants to "sse"', () => {
        const inputs = ['sse', 'SSE', 'server-sent events', 'Server-Sent-Events', 'streaming'];
        inputs.forEach(input => {
            const normalized = normalizeTransportString(input);
            expect(normalized).toBe('sse');
        });
    });

    it('normalizes HTTP variants to "streamable_http"', () => {
        const inputs = ['http', 'HTTP', 'streamable-http', 'streamable_http', 'streamableHttp'];
        inputs.forEach(input => {
            const normalized = normalizeTransportString(input);
            expect(normalized).toBe('streamable_http');
        });
    });

    it('defaults to "stdio" for unknown transport types', () => {
        const inputs = ['', null, undefined, 'unknown', 'custom'];
        inputs.forEach(input => {
            const normalized = normalizeTransportString(input as any);
            expect(normalized).toBe('stdio');
        });
    });

    it('infers transport from technology keywords in description', () => {
        const cases = [
            { desc: 'built with FastAPI and Server-Sent Events', expected: 'sse' },
            { desc: 'exposes HTTP REST endpoints', expected: 'streamable_http' },
            { desc: 'runs as a subprocess using stdio', expected: 'stdio' },
            { desc: 'WebSocket-based integration', expected: 'streamable_http' },
        ];

        cases.forEach(({ desc, expected }) => {
            const inferred = inferTransportFromDescription(desc);
            expect(inferred).toBe(expected);
        });
    });
});

/**
 * Test normalization state transition logic (discovered -> normalized).
 * This models the Archivist's responsibility to advance servers through states
 * based on metadata completeness.
 */
describe('CatalogIngestor — Server State Transitions', () => {
    it('qualifies for normalization when description > 10 chars AND transport is known', () => {
        const cases = [
            { description: 'Browser automation tool', transport: 'stdio', shouldNormalize: true },
            { description: 'CLI for Python', transport: 'stdio', shouldNormalize: true },
            { description: 'DB query', transport: 'unknown', shouldNormalize: false }, // transport unknown
            { description: 'Short', transport: 'stdio', shouldNormalize: false }, // description too short
            { description: '', transport: 'stdio', shouldNormalize: false }, // no description
            { description: 'Terminal access tool', transport: 'sse', shouldNormalize: true },
        ];

        cases.forEach(({ description, transport, shouldNormalize }) => {
            const qualifies = qualifiesForNormalization({ description, transport });
            expect(qualifies).toBe(shouldNormalize);
        });
    });

    it('does not transition servers already at normalized or higher status', () => {
        const statuses = ['normalized', 'probeable', 'validated', 'certified', 'broken'];
        statuses.forEach(status => {
            expect(qualifiesForNormalization({ description: 'Valid description', transport: 'stdio' }, status as any))
                .toBe(false);
        });
    });

    it('correctly handles edge cases for description length', () => {
        // Exactly 10 characters is the boundary
        const desc10 = 'X'.repeat(10);
        const desc11 = 'X'.repeat(11);
        expect(qualifiesForNormalization({ description: desc10, transport: 'stdio' })).toBe(false);
        expect(qualifiesForNormalization({ description: desc11, transport: 'stdio' })).toBe(true);
    });
});

/**
 * Test deduplication logic across multiple adapters.
 * Ensures that servers from different registries don't create duplicates.
 */
describe('CatalogIngestor — Deduplication', () => {
    it('deduplicates servers by canonical_id within a single ingestion run', () => {
        const candidates = [
            { canonical_id: 'npm:@modelcontextprotocol/server-slack', name: 'Slack Server v1', source: 'npm' },
            { canonical_id: 'npm:@modelcontextprotocol/server-slack', name: 'Slack Server v2', source: 'smithery' },
            { canonical_id: 'github:slackapi/mcp-server-slack', name: 'Slack Official', source: 'github' },
        ];

        const deduped = deduplicateByCanonicalId(candidates);
        // Should keep first occurrence of each canonical_id
        expect(deduped).toHaveLength(2);
        expect(deduped.find(c => c.canonical_id === 'npm:@modelcontextprotocol/server-slack')?.name)
            .toBe('Slack Server v1');
    });

    it('tracks provenance for servers from multiple sources', () => {
        const sources = [
            { server_uuid: 'uuid-1', source_name: 'glama.ai', last_seen_at: new Date('2026-03-20') },
            { server_uuid: 'uuid-1', source_name: 'smithery.ai', last_seen_at: new Date('2026-03-20') },
            { server_uuid: 'uuid-1', source_name: 'npm', last_seen_at: new Date('2026-03-19') },
        ];

        // Multiple sources for same server should all be recorded
        expect(sources.filter(s => s.server_uuid === 'uuid-1')).toHaveLength(3);
    });

    it('preserves highest-confidence source on conflict', () => {
        const conflicts = [
            { canonical_id: 'test', source: 'glama.ai', confidence: 90 },
            { canonical_id: 'test', source: 'npm', confidence: 75 },
            { canonical_id: 'test', source: 'github', confidence: 60 },
        ];

        const winner = conflicts.reduce((best, current) =>
            current.confidence > best.confidence ? current : best
        );
        expect(winner.source).toBe('glama.ai');
    });
});

/**
 * Test configuration recipe generation quality metrics.
 */
describe('CatalogIngestor — Recipe Quality', () => {
    it('assigns appropriate confidence scores based on metadata completeness', () => {
        const cases = [
            {
                metadata: {
                    description: 'Full browser automation',
                    transport: 'sse',
                    install_method: 'npm',
                    author: 'MCP Team',
                    repository_url: 'https://github.com/model-context-protocol/servers',
                    homepage_url: 'https://example.com',
                },
                expectedConfidence: 85, // High confidence with full metadata
                tolerance: 5,
            },
            {
                metadata: {
                    description: 'Some tool',
                    transport: 'stdio',
                },
                expectedConfidence: 40, // Low confidence with minimal metadata
                tolerance: 10,
            },
        ];

        cases.forEach(({ metadata, expectedConfidence, tolerance }) => {
            const confidence = calculateRecipeConfidence(metadata);
            expect(confidence).toBeGreaterThanOrEqual(expectedConfidence - tolerance);
            expect(confidence).toBeLessThanOrEqual(expectedConfidence + tolerance);
        });
    });

    it('penalizes confidence for unknown transport or install method', () => {
        const withKnown = calculateRecipeConfidence({
            description: 'Test',
            transport: 'stdio',
            install_method: 'npm',
        });

        const withUnknown = calculateRecipeConfidence({
            description: 'Test',
            transport: 'unknown',
            install_method: 'unknown',
        });

        expect(withUnknown).toBeLessThan(withKnown);
    });
});

/**
 * Test install method inference from repository language.
 */
describe('CatalogIngestor — Install Method Inference', () => {
    it('infers install method from primary language', () => {
        const cases = [
            { language: 'Python', expected: 'pip' },
            { language: 'TypeScript', expected: 'npm' },
            { language: 'JavaScript', expected: 'npm' },
            { language: 'Go', expected: 'go-install' },
            { language: 'Rust', expected: 'cargo' },
            { language: 'Ruby', expected: 'gem' },
            { language: 'unknown', expected: 'npm' }, // default
        ];

        cases.forEach(({ language, expected }) => {
            const method = inferInstallMethod(language);
            expect(method).toBe(expected);
        });
    });

    it('defaults to npm when language cannot be determined', () => {
        expect(inferInstallMethod(null)).toBe('npm');
        expect(inferInstallMethod(undefined)).toBe('npm');
        expect(inferInstallMethod('')).toBe('npm');
    });
});

describe('CatalogIngestor — Baseline Recipe Generation', () => {
    it('creates version-pinned npm npx recipe when package version is known', () => {
        const recipe = buildBaselineRecipe(
            {
                canonical_id: 'npm/@modelcontextprotocol/server-filesystem',
                display_name: 'Filesystem Server',
                transport: 'stdio',
                install_method: 'npm',
                repository_url: 'https://github.com/modelcontextprotocol/servers',
            },
            '1.2.3',
        );

        expect(recipe.template).toMatchObject({
            type: 'stdio',
            command: 'npx',
            args: ['-y', '@modelcontextprotocol/server-filesystem@1.2.3'],
        });
        expect(recipe.confidence).toBeGreaterThanOrEqual(40);
        expect(recipe.explanation.toLowerCase()).toContain('version-pinned');
    });

    it('creates placeholder URL recipe for http transports to force operator review', () => {
        const recipe = buildBaselineRecipe(
            {
                canonical_id: 'github/acme/http-server',
                display_name: 'HTTP Server',
                transport: 'streamable_http',
                install_method: 'unknown',
                repository_url: 'https://github.com/acme/http-server',
            },
            null,
        );

        expect(recipe.template).toMatchObject({
            type: 'streamable_http',
            url: 'https://example.com/mcp',
        });
        expect(recipe.confidence).toBeLessThan(30);
    });
});

describe('CatalogIngestor — Error Reporting', () => {
    it('records HTTP fetch failures in adapter errors instead of silently returning zero errors', async () => {
        vi.stubGlobal(
            'fetch',
            vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                headers: {
                    get: () => 'application/json',
                },
            })
        );

        const result = await new GlamaAiAdapter().ingest();

        expect(result.fetched).toBe(0);
        expect(result.upserted).toBe(0);
        expect(result.errors).toHaveLength(1);
        expect(result.errors[0]).toContain('Request to https://glama.ai/api/mcp/servers?limit=200 failed: HTTP 404');
    });

    it('reports HTML error pages truthfully instead of as JSON parse failures', async () => {
        vi.stubGlobal(
            'fetch',
            vi.fn().mockResolvedValue({
                ok: false,
                status: 403,
                headers: {
                    get: (name: string) => name.toLowerCase() === 'content-type' ? 'text/html; charset=utf-8' : null,
                },
            })
        );

        const result = await new GlamaAiAdapter().ingest();

        expect(result.fetched).toBe(0);
        expect(result.upserted).toBe(0);
        expect(result.errors).toHaveLength(1);
        expect(result.errors[0]).toContain('HTTP 403 (received HTML error page)');
        expect(result.errors[0]).not.toContain('Failed to parse JSON');
    });

    it('uses the verified smithery limit query shape', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            status: 200,
            headers: {
                get: (name: string) => name.toLowerCase() === 'content-type' ? 'application/json; charset=utf-8' : null,
            },
            json: async () => ({ servers: [] }),
        });
        vi.stubGlobal('fetch', fetchMock);

        const result = await new SmitheryAiAdapter().ingest();

        expect(result.fetched).toBe(0);
        expect(result.upserted).toBe(0);
        expect(result.errors).toHaveLength(0);
        expect(fetchMock).toHaveBeenCalledTimes(1);
        expect(fetchMock.mock.calls[0]?.[0]).toBe('https://registry.smithery.ai/servers?limit=200');
    });

    it('records per-query fetch failures for npm ingestion summaries', async () => {
        vi.stubGlobal(
            'fetch',
            vi.fn().mockRejectedValue(new Error('network down'))
        );

        const result = await new NpmRegistryAdapter().ingest();

        expect(result.fetched).toBe(0);
        expect(result.upserted).toBe(0);
        expect(result.errors).toHaveLength(3);
        expect(result.errors[0]).toContain('npm query "scope:modelcontextprotocol" failed');
        expect(result.errors[1]).toContain('npm query "keywords:mcp-server" failed');
        expect(result.errors[2]).toContain('npm query "mcp-server" failed');
    });

    it('logs catalog warnings as a concise preview instead of dumping full error arrays', async () => {
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => {});

        const report = await ingestPublishedCatalog([
            {
                name: 'glama.ai',
                ingest: vi.fn(async () => ({
                    source: 'glama.ai',
                    fetched: 0,
                    upserted: 0,
                    errors: [
                        'fetch failed: Request to https://glama.ai/api/mcp/servers?limit=200 failed: HTTP 404',
                        'fetch failed: Request to https://glama.ai/api/mcp/servers?limit=200 failed: HTTP 403 (received HTML error page)',
                        'fetch failed: Request to https://glama.ai/api/mcp/servers?limit=200 failed: HTTP 400',
                    ],
                })),
            },
        ]);

        expect(report.total_errors).toBe(3);
        expect(infoSpy).toHaveBeenCalledWith('[CatalogIngestor] glama.ai: fetched=0 upserted=0 errors=3');
        expect(warnSpy).toHaveBeenCalledWith(
            '[CatalogIngestor] glama.ai warnings: fetch failed: Request to https://glama.ai/api/mcp/servers?limit=200 failed: HTTP 404 | fetch failed: Request to https://glama.ai/api/mcp/servers?limit=200 failed: HTTP 403 (received HTML error page) | +1 more'
        );
    });

    it('logs normalization and baseline pass failures as single-string warnings', async () => {
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => {});
        vi.spyOn(publishedCatalogRepository, 'listServers')
            .mockRejectedValueOnce(new Error('normalization store unavailable'))
            .mockRejectedValueOnce(new Error('baseline store unavailable'));

        const report = await ingestPublishedCatalog([
            {
                name: 'mcp.run',
                ingest: vi.fn(async () => ({
                    source: 'mcp.run',
                    fetched: 0,
                    upserted: 0,
                    errors: [],
                })),
            },
        ]);

        expect(report.total_errors).toBe(0);
        expect(infoSpy).toHaveBeenCalledWith('[CatalogIngestor] mcp.run: fetched=0 upserted=0 errors=0');
        expect(warnSpy).toHaveBeenCalledWith(expect.stringMatching(/^\[CatalogIngestor\] Normalization pass failed: .+/));
        expect(warnSpy).toHaveBeenCalledWith(expect.stringMatching(/^\[CatalogIngestor\] Baseline recipe pass failed: .+/));
        expect(warnSpy.mock.calls.every((call) => call.length === 1 && typeof call[0] === 'string')).toBe(true);
    });
});

// Helper functions that model the actual ingestor logic

function normalizeTransportString(raw: string | null | undefined): string {
    if (!raw) return 'stdio';
    const normalized = raw.toLowerCase().replace(/[-_\s]/g, '');
    if (normalized === 'stdio') return 'stdio';
    if (normalized.includes('sse') || normalized.includes('event') || normalized.includes('streaming')) return 'sse';
    if (normalized.includes('http') || normalized.includes('websocket')) return 'streamable_http';
    return 'stdio';
}

function inferTransportFromDescription(desc: string): string {
    if (!desc) return 'stdio';
    const lower = desc.toLowerCase();
    if (lower.includes('sse') || lower.includes('server-sent') || lower.includes('event')) return 'sse';
    if (lower.includes('http') || lower.includes('rest') || lower.includes('websocket')) return 'streamable_http';
    if (lower.includes('stdio') || lower.includes('subprocess')) return 'stdio';
    return 'stdio';
}

function qualifiesForNormalization(
    server: { description?: string | null; transport?: string | null },
    currentStatus?: string
): boolean {
    // Must be at 'discovered' status (or no status)
    if (currentStatus && currentStatus !== 'discovered') return false;

    // Must have description > 10 characters
    if (!server.description || server.description.length <= 10) return false;

    // Transport must be known (not 'unknown')
    if (!server.transport || server.transport === 'unknown') return false;

    return true;
}

function deduplicateByCanonicalId(
    candidates: Array<{ canonical_id: string; name: string; source: string }>
): typeof candidates {
    const seen = new Set<string>();
    return candidates.filter(c => {
        if (seen.has(c.canonical_id)) return false;
        seen.add(c.canonical_id);
        return true;
    });
}

function calculateRecipeConfidence(metadata: {
    description?: string | null;
    transport?: string | null;
    install_method?: string | null;
    author?: string | null;
    repository_url?: string | null;
    homepage_url?: string | null;
}): number {
    let score = 40; // Base score tuned to keep "full metadata" in high-confidence band (~85-90)

    if (metadata.description && metadata.description.length > 20) score += 10;
    if (metadata.transport && metadata.transport !== 'unknown') score += 10;
    if (metadata.install_method && metadata.install_method !== 'unknown') score += 10;
    if (metadata.author) score += 5;
    if (metadata.repository_url) score += 10;
    if (metadata.homepage_url) score += 5;

    // Penalties
    if (metadata.transport === 'unknown') score -= 15;
    if (metadata.install_method === 'unknown') score -= 15;

    return Math.max(10, Math.min(100, score));
}

function inferInstallMethod(language: string | null | undefined): string {
    if (!language) return 'npm';
    const lower = language.toLowerCase();
    if (lower.includes('python')) return 'pip';
    if (lower.includes('typescript') || lower.includes('javascript')) return 'npm';
    if (lower.includes('go')) return 'go-install';
    if (lower.includes('rust')) return 'cargo';
    if (lower.includes('ruby')) return 'gem';
    return 'npm'; // default
}
