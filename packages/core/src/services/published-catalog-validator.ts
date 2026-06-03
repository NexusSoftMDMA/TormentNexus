/**
 * @file published-catalog-validator.ts
 * @module packages/core/src/services/published-catalog-validator
 *
 * WHAT:
 * The Verifier agent's validation subsystem. Attempts to connect to a
 * published MCP server using its active config recipe, runs `tools/list`,
 * and records the result in `published_mcp_validation_runs`.
 *
 * WHY:
 * "We only certify what we have safely observed." — Verifier mantra.
 * Validation evidence is required before a server can advance to
 * `validated` or `certified` status. Without it, we can only claim
 * "this was seen in a registry" — not "this actually works."
 *
 * HOW:
 * 1. `validateServer(serverUuid)` — main entry point:
 *    a. Load the server + active recipe from the DB.
 *    b. Start a `pending` validation run row.
 *    c. Attempt transport probe (can we establish a connection?).
 *    d. If successful, attempt `tools/list` and count tools.
 *    e. Finish the run with outcome + findings.
 *    f. Update server status / confidence accordingly.
 *
 * Security note: Validation runs use the MCP SDK client directly.
 * They NEVER execute arbitrary commands beyond the MCP protocol handshake.
 * STDIO servers are NOT probed in the validator by design — only SSE/HTTP
 * and known-safe npx patterns may be attempted (transport_probe only).
 *
 * Isolation: Future phases will run these in Docker containers.
 * For now, only URL-based (SSE/Streamable HTTP) servers are probed.
 * STDIO servers remain catalog-visible and get `skipped` outcome with
 * failure_class = 'stdio_transport' so operators can distinguish transport
 * shape from actual validation failure.
 */

import { publishedCatalogRepository } from "../db/repositories/published-catalog.repo.js";
import type { PublishedMcpServer, PublishedMcpConfigRecipe } from "../db/repositories/published-catalog.repo.js";

export type ValidationOutput = {
    server_uuid: string;
    run_uuid: string;
    outcome: "passed" | "failed" | "error" | "timeout" | "skipped";
    failure_class?: string;
    tool_count?: number;
    findings_summary?: Record<string, unknown>;
};

const VALIDATION_TIMEOUT_MS = 10_000;

/**
 * Attempt to validate a single published server.
 * Uses the active recipe's template to determine how to connect.
 *
 * Returns true if the validation passed (tools/list succeeded).
 */
export async function validatePublishedServer(serverUuid: string): Promise<ValidationOutput> {
    const server = await publishedCatalogRepository.findServerByUuid(serverUuid);
    if (!server) {
        throw new Error(`Published server not found: ${serverUuid}`);
    }

    const recipe = await publishedCatalogRepository.getActiveRecipe(serverUuid);

    // Begin tracking the run
    const run = await publishedCatalogRepository.startValidationRun({
        server_uuid: serverUuid,
        run_mode: "tools_list",
        performed_by: "Verifier",
    });

    const output: ValidationOutput = {
        server_uuid: serverUuid,
        run_uuid: run.uuid,
        outcome: "skipped",
    };

    try {
        // If no recipe exists, we cannot probe — skip
        if (!recipe) {
            output.outcome = "skipped";
            output.failure_class = "no_recipe";
            await _finishRun(run.uuid, output);
            return output;
        }

        const template = recipe.template as Record<string, unknown>;
        const transport = String(server.transport ?? template.type ?? "unknown");

        // Validation currently only probes URL-based transports.
        // STDIO entries are preserved and labeled as stdio transport, but not
        // actively probed until sandboxed validation is available.
        if (transport === "stdio" || transport === "STDIO") {
            output.outcome = "skipped";
            output.failure_class = "stdio_transport";
            output.findings_summary = {
                transport: "stdio",
                validation_mode: "not_probed",
            };
            await _finishRun(run.uuid, output);
            return output;
        }

        const url = template.url as string | undefined;
        if (!url) {
            output.outcome = "skipped";
            output.failure_class = "no_url_in_recipe";
            await _finishRun(run.uuid, output);
            return output;
        }

        // Attempt a lightweight HTTP probe first (not a full MCP handshake)
        const probeResult = await _httpProbe(url);
        if (!probeResult.reachable) {
            output.outcome = "failed";
            output.failure_class = probeResult.failureClass;
            output.findings_summary = { probe_error: probeResult.error };
            await _finishRun(run.uuid, output);
            await _updateServerStatusFromOutcome(server, "failed", 10);
            return output;
        }

        // Attempt MCP tools/list via the SDK
        const toolsResult = await _mcpToolsList(url, transport, VALIDATION_TIMEOUT_MS);
        if (toolsResult.success) {
            output.outcome = "passed";
            output.tool_count = toolsResult.toolCount;
            output.findings_summary = {
                tool_count: toolsResult.toolCount,
                tool_names: toolsResult.toolNames,
            };
            await _finishRun(run.uuid, output);
            await _updateServerStatusFromOutcome(server, "passed", 80);
        } else {
            output.outcome = "failed";
            output.failure_class = toolsResult.failureClass;
            output.findings_summary = { error: toolsResult.error };
            await _finishRun(run.uuid, output);
            await _updateServerStatusFromOutcome(server, "failed", 30);
        }
    } catch (err) {
        output.outcome = "error";
        output.failure_class = "unexpected_error";
        output.findings_summary = { error: String(err) };
        try {
            await _finishRun(run.uuid, output);
        } catch {
            // best-effort
        }
    }

    return output;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

async function _finishRun(runUuid: string, output: ValidationOutput): Promise<void> {
    await publishedCatalogRepository.finishValidationRun({
        uuid: runUuid,
        outcome: output.outcome,
        failure_class: output.failure_class ?? null,
        tool_count: output.tool_count ?? null,
        findings_summary: output.findings_summary ?? null,
    });
}

async function _updateServerStatusFromOutcome(
    server: PublishedMcpServer,
    result: "passed" | "failed",
    confidence: number
): Promise<void> {
    if (result === "passed") {
        await publishedCatalogRepository.updateServerStatus(server.uuid, "validated", confidence);
    } else {
        // Only mark as broken if confidence is already low (multiple failures)
        const newStatus = server.confidence < 20 ? "broken" : server.status;
        const newConfidence = Math.max(0, (server.confidence ?? 0) - 10);
        await publishedCatalogRepository.updateServerStatus(server.uuid, newStatus as any, newConfidence);
    }
}

type ProbeResult = {
    reachable: boolean;
    failureClass?: string;
    error?: string;
};

/**
 * Lightweight HTTP reachability probe.
 * Does NOT attempt MCP protocol — just verifies the URL is responding.
 */
async function _httpProbe(url: string): Promise<ProbeResult> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5_000);

    try {
        const response = await fetch(url, {
            method: "GET",
            signal: controller.signal,
            headers: { "Accept": "application/json, text/event-stream, */*" },
        });

        clearTimeout(timeout);

        // 401/403 means the server is reachable but requires auth — still "reachable" for probe
        if (response.status === 401 || response.status === 403) {
            return { reachable: true };
        }
        if (response.status >= 200 && response.status < 500) {
            return { reachable: true };
        }
        return {
            reachable: false,
            failureClass: `http_${response.status}`,
            error: `HTTP ${response.status} ${response.statusText}`,
        };
    } catch (err: any) {
        clearTimeout(timeout);
        if (err?.name === "AbortError") {
            return { reachable: false, failureClass: "timeout", error: "Connection timed out" };
        }
        const msg = String(err?.message ?? err);
        const failureClass = msg.includes("ENOTFOUND") || msg.includes("ECONNREFUSED")
            ? "network_unreachable"
            : "connection_error";
        return { reachable: false, failureClass, error: msg };
    }
}

type ToolsListResult = {
    success: boolean;
    toolCount?: number;
    toolNames?: string[];
    failureClass?: string;
    error?: string;
};

/**
 * Attempt MCP tools/list via the @modelcontextprotocol/sdk client.
 * Only URL-based transports (SSE, StreamableHTTP) are attempted here.
 * STDIO probing requires sandbox and is handled by a future phase.
 */
async function _mcpToolsList(
    url: string,
    transport: string,
    timeoutMs: number
): Promise<ToolsListResult> {
    // Dynamic import to avoid circular dependencies and allow graceful fallback
    try {
        const { Client } = await import("@modelcontextprotocol/sdk/client/index.js");
        const { SSEClientTransport } = await import("@modelcontextprotocol/sdk/client/sse.js");
        const { StreamableHTTPClientTransport } = await import(
            "@modelcontextprotocol/sdk/client/streamableHttp.js"
        );

        const client = new Client({ name: "tormentnexus-catalog-verifier", version: "0.1.0" });

        let transportImpl: any;
        const normalizedTransport = transport.toLowerCase();

        if (normalizedTransport === "sse" || normalizedTransport.includes("sse")) {
            transportImpl = new SSEClientTransport(new URL(url));
        } else {
            // Default to StreamableHTTP for anything else URL-based
            transportImpl = new StreamableHTTPClientTransport(new URL(url));
        }

        const connectPromise = client.connect(transportImpl);
        const timeoutPromise = new Promise<never>((_, reject) =>
            setTimeout(() => reject(new Error("Connection timeout")), timeoutMs)
        );

        await Promise.race([connectPromise, timeoutPromise]);

        const listResult = await Promise.race([
            client.listTools(),
            new Promise<never>((_, reject) =>
                setTimeout(() => reject(new Error("tools/list timeout")), timeoutMs)
            ),
        ]);

        await client.close();

        const tools = (listResult as any)?.tools ?? [];
        return {
            success: true,
            toolCount: tools.length,
            toolNames: tools.slice(0, 20).map((t: any) => String(t.name ?? "")),
        };
    } catch (err: any) {
        const msg = String(err?.message ?? err);
        let failureClass = "protocol_error";
        if (msg.includes("timeout")) failureClass = "timeout";
        else if (msg.includes("ENOTFOUND") || msg.includes("ECONNREFUSED")) failureClass = "network_unreachable";
        else if (msg.includes("auth") || msg.includes("401") || msg.includes("403")) failureClass = "auth_required";
        return { success: false, failureClass, error: msg };
    }
}
