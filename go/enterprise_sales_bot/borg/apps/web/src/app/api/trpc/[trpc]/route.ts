import { getCacheTTL, getCached, setCached } from "../cache";

export const runtime = "nodejs";

const DEFAULT_UPSTREAM_TRPC_URL = "http://127.0.0.1:7779/trpc";
const DEFAULT_GO_API_BASE = "http://127.0.0.1:7778";

function resolveUpstreamBase(): string {
	return (
		process.env.TORMENTNEXUS_TRPC_UPSTREAM?.trim() || DEFAULT_UPSTREAM_TRPC_URL
	);
}

function resolveGoApiBase(): string {
	return process.env.TORMENTNEXUS_GO_API_BASE?.trim() || DEFAULT_GO_API_BASE;
}

function getProcedurePath(req: Request): string {
	const incomingUrl = new URL(req.url);
	const pathMatch = incomingUrl.pathname.match(/\/api\/trpc\/?(.*)$/);
	return pathMatch?.[1] ?? "";
}

function buildUpstreamUrl(req: Request): URL {
	const incomingUrl = new URL(req.url);
	const upstreamBase = resolveUpstreamBase().replace(/\/$/, "");
	const procedurePath = getProcedurePath(req);
	const upstreamUrl = new URL(
		`${upstreamBase}${procedurePath ? `/${procedurePath}` : ""}`,
	);
	upstreamUrl.search = incomingUrl.search;
	return upstreamUrl;
}

function cloneHeaders(req: Request): Headers {
	const headers = new Headers(req.headers);
	headers.delete("host");
	headers.delete("content-length");
	return headers;
}

function getCompatRoute(procedurePath: string, input: unknown): string | null {
	// ── Go-native fast paths (cached, <5ms) ──
	if (procedurePath === "startupStatus") {
		return "/api/startup/status";
	}
	if (procedurePath === "mcp.getStatus") {
		return "/api/mcp/status";
	}
	if (procedurePath === "mcp.listServers") {
		return "/api/mcp/servers";
	}
	if (procedurePath === "mcp.getToolSelectionTelemetry") {
		return "/api/mcp/tool-selection-telemetry";
	}
	if (procedurePath === "mcp.clearToolSelectionTelemetry") {
		return "/api/mcp/tool-selection-telemetry/clear";
	}
	if (procedurePath === "mcp.runServerTest") {
		return "/api/mcp/server-test";
	}
	if (procedurePath === "session.list") {
		return "/api/native/session/list";
	}
	if (procedurePath === "session.importedMaintenanceStats") {
		return "/api/sessions/imported/maintenance-stats";
	}
	// ── Billing (Go-native with local fallback) ──
	if (procedurePath === "billing.getStatus") {
		return "/api/billing/status";
	}
	if (procedurePath === "billing.getProviderQuotas") {
		return "/api/billing/provider-quotas";
	}
	if (procedurePath === "billing.getCostHistory") {
		const days =
			typeof input === "object" && input !== null && "days" in input
				? Number((input as { days?: unknown }).days)
				: NaN;
		const normalizedDays =
			Number.isFinite(days) && days > 0 ? Math.min(Math.round(days), 90) : 30;
		return `/api/billing/cost-history?days=${normalizedDays}`;
	}
	if (procedurePath === "billing.getModelPricing") {
		return "/api/billing/model-pricing";
	}
	if (procedurePath === "billing.getFallbackChain") {
		const taskType =
			typeof input === "object" && input !== null && "taskType" in input
				? (input as { taskType?: unknown }).taskType
				: undefined;
		const search =
			typeof taskType === "string" && taskType.length > 0
				? `?taskType=${encodeURIComponent(taskType)}`
				: "";
		return `/api/billing/fallback-chain${search}`;
	}
	if (procedurePath === "billing.getTaskRoutingRules") {
		return "/api/billing/task-routing-rules";
	}
	if (procedurePath === "billing.getDepletedModels") {
		return "/api/billing/depleted-models";
	}
	if (procedurePath === "billing.getFallbackHistory") {
		const limit =
			typeof input === "object" && input !== null && "limit" in input
				? Number((input as { limit?: unknown }).limit)
				: NaN;
		const normalizedLimit =
			Number.isFinite(limit) && limit > 0
				? Math.min(Math.round(limit), 50)
				: 20;
		return `/api/billing/fallback-history?limit=${normalizedLimit}`;
	}
	// ── Director (Go-native with local fallback) ──
	if (procedurePath === "director.status") {
		return "/api/director/status";
	}
	if (procedurePath === "directorConfig.get") {
		return "/api/director-config";
	}
	// ── LLM (Go-native with WaterfallClient) ──
	if (procedurePath === "llm.generate") {
		return "/api/llm/generate";
	}
	// ── Memory (Go-native with local fallback) ──
	if (procedurePath === "memory.getRecentObservations") {
		const limit =
			typeof input === "object" && input !== null && "limit" in input
				? Number((input as { limit?: unknown }).limit)
				: NaN;
		const namespace =
			typeof input === "object" && input !== null && "namespace" in input
				? (input as { namespace?: unknown }).namespace
				: undefined;
		const type =
			typeof input === "object" && input !== null && "type" in input
				? (input as { type?: unknown }).type
				: undefined;
		const params = new URLSearchParams();
		params.set(
			"limit",
			String(Number.isFinite(limit) && limit > 0 ? Math.round(limit) : 6),
		);
		if (typeof namespace === "string" && namespace.length > 0)
			params.set("namespace", namespace);
		if (typeof type === "string" && type.length > 0) params.set("type", type);
		return `/api/memory/observations/recent?${params.toString()}`;
	}
	if (procedurePath === "memory.getRecentUserPrompts") {
		const limit =
			typeof input === "object" && input !== null && "limit" in input
				? Number((input as { limit?: unknown }).limit)
				: NaN;
		const role =
			typeof input === "object" && input !== null && "role" in input
				? (input as { role?: unknown }).role
				: undefined;
		const params = new URLSearchParams();
		params.set(
			"limit",
			String(Number.isFinite(limit) && limit > 0 ? Math.round(limit) : 5),
		);
		if (typeof role === "string" && role.length > 0) params.set("role", role);
		return `/api/memory/user-prompts/recent?${params.toString()}`;
	}
	if (procedurePath === "memory.getRecentSessionSummaries") {
		const limit =
			typeof input === "object" && input !== null && "limit" in input
				? Number((input as { limit?: unknown }).limit)
				: NaN;
		const params = new URLSearchParams();
		params.set(
			"limit",
			String(Number.isFinite(limit) && limit > 0 ? Math.round(limit) : 4),
		);
		return `/api/memory/session-summaries/recent?${params.toString()}`;
	}

	return null;
}

async function getCompatPayload(
	procedurePath: string,
	input: unknown,
): Promise<unknown | null> {
	const compatRoute = getCompatRoute(procedurePath, input);
	if (!compatRoute) {
		return null;
	}

	const goApiBase = resolveGoApiBase().replace(/\/$/, "");

	try {
		const compatResponse = await fetch(`${goApiBase}${compatRoute}`);
		if (!compatResponse.ok) {
			return null;
		}

		const compatJson = await compatResponse.json();
		return Array.isArray(compatJson?.data) ||
			typeof compatJson?.data === "object"
			? compatJson.data
			: compatJson;
	} catch {
		return null;
	}
}

function parseBatchInput(req: Request): Record<string, unknown> {
	const inputParam = new URL(req.url).searchParams.get("input");
	if (!inputParam) {
		return {};
	}

	try {
		const parsed = JSON.parse(inputParam);
		return parsed && typeof parsed === "object"
			? (parsed as Record<string, unknown>)
			: {};
	} catch {
		return {};
	}
}

async function fetchSingleProcedureEntry(
	procedurePath: string,
	input: unknown,
): Promise<any | null> {
	const upstreamBase = resolveUpstreamBase().replace(/\/$/, "");
	const upstreamUrl = new URL(`${upstreamBase}/${procedurePath}`);
	upstreamUrl.searchParams.set("batch", "1");
	upstreamUrl.searchParams.set("input", JSON.stringify(input ?? {}));

	try {
		const response = await fetch(upstreamUrl, { method: "GET" });
		if (response.ok) {
			const json = await response.json();
			if (Array.isArray(json) && json.length > 0) {
				return json[0];
			}
			return json;
		}
	} catch {
		// Fall through to compat path below.
	}

	const compatPayload = await getCompatPayload(procedurePath, input);
	if (compatPayload !== null) {
		return { result: { data: compatPayload } };
	}

	return null;
}

async function tryCompatFallback(
	req: Request,
	procedurePath: string,
): Promise<Response | null> {
	if (!procedurePath.includes(",")) {
		const compatPayload = await getCompatPayload(procedurePath, {});
		if (compatPayload === null) {
			return null;
		}

		return new Response(JSON.stringify([{ result: { data: compatPayload } }]), {
			status: 200,
			headers: { "content-type": "application/json" },
		});
	}

	const procedures = procedurePath
		.split(",")
		.map((entry) => entry.trim())
		.filter(Boolean);
	if (procedures.length === 0) {
		return null;
	}

	const batchInput = parseBatchInput(req);
	const entries = [];
	for (const [index, procedure] of procedures.entries()) {
		const entry = await fetchSingleProcedureEntry(
			procedure,
			batchInput[String(index)] ?? {},
		);
		if (!entry) {
			return null;
		}
		entries.push(entry);
	}

	const hasErrors = entries.some((entry) => entry?.error);
	return new Response(JSON.stringify(entries), {
		status: hasErrors ? 207 : 200,
		headers: { "content-type": "application/json" },
	});
}

/**
 * Procedures that have fast Go-native implementations.
 * These are served directly from the Go sidecar (<5ms) instead of
 * proxying through the TS Core tRPC server (~100-300ms).
 */
const GO_NATIVE_PROCEDURES = new Set([
	"mcp.getStatus",
	"mcp.listServers",
	"mcp.getToolSelectionTelemetry",
	"session.importedMaintenanceStats",
	"billing.getProviderQuotas",
	"billing.getModelPricing",
	"billing.getStatus",
	"billing.getDepletedModels",
	"billing.getFallbackHistory",
	"billing.getFallbackChain",
	"billing.getTaskRoutingRules",
	"director.status",
	"directorConfig.get",
]);

async function handler(req: Request): Promise<Response> {
	const procedurePath = getProcedurePath(req);

	// Go-native fast path: serve from Go sidecar first (<5ms),
	// fall back to tRPC upstream if Go sidecar is unavailable.
	// Only use Go-native fast path for single-procedure requests
	// (batch requests need all procedures to go through the same path)
	const isBatch = procedurePath.includes(",");
	const firstProc = isBatch ? "" : (procedurePath.trim() ?? "");
	if (firstProc && GO_NATIVE_PROCEDURES.has(firstProc)) {
		const batchInput = parseBatchInput(req);
		const compatPayload = await getCompatPayload(
			firstProc,
			batchInput["0"] ?? {},
		);
		if (compatPayload !== null) {
			return new Response(
				JSON.stringify([{ result: { data: compatPayload } }]),
				{ status: 200, headers: { "content-type": "application/json" } },
			);
		}
		// Go sidecar unavailable; fall through to tRPC upstream
	}

	const upstreamUrl = buildUpstreamUrl(req);
	const headers = cloneHeaders(req);
	const hasBody = req.method !== "GET" && req.method !== "HEAD";
	const body = hasBody ? await req.text() : undefined;

	// Check tRPC response cache for frequently-polled procedures
	const cacheTTL = getCacheTTL(procedurePath);
	const cacheInput = new URL(req.url).searchParams.get("input") ?? "{}";
	if (cacheTTL !== null && req.method === "GET") {
		const cached = getCached(procedurePath, cacheInput);
		if (cached) {
			return new Response(JSON.stringify(cached.data), {
				status: cached.status,
				headers: cached.headers,
			});
		}
	}

	let upstreamResponse: Response;
	try {
		upstreamResponse = await fetch(upstreamUrl, {
			method: req.method,
			headers,
			body,
		});
	} catch (error) {
		const compatFallback = await tryCompatFallback(req, procedurePath);
		if (compatFallback) {
			console.warn(
				`[TRPC-Proxy] Using compat fallback for ${procedurePath} after upstream fetch failure`,
			);
			return compatFallback;
		}
		const message = error instanceof Error ? error.message : String(error);
		console.error(`[TRPC-Proxy] Upstream fetch failed: ${message}`);
		return new Response(
			JSON.stringify({
				error: "TRPC_UPSTREAM_UNAVAILABLE",
				message,
				upstream: upstreamUrl.toString(),
			}),
			{ status: 502, headers: { "content-type": "application/json" } },
		);
	}

	if (!upstreamResponse.ok) {
		const compatFallback = await tryCompatFallback(req, procedurePath);
		if (compatFallback) {
			console.warn(
				`[TRPC-Proxy] Using compat fallback for ${procedurePath} after upstream status ${upstreamResponse.status}`,
			);
			return compatFallback;
		}
	}

	const responseHeaders = new Headers(upstreamResponse.headers);
	const isSse = responseHeaders.get("content-type") === "text/event-stream";
	if (isSse) {
		responseHeaders.set("Connection", "keep-alive");
		responseHeaders.set("Cache-Control", "no-cache");
	}
	// Cache the successful response for frequently-polled procedures
	if (cacheTTL !== null && upstreamResponse.ok && req.method === "GET") {
		try {
			const bodyClone = upstreamResponse.clone();
			const bodyText = await bodyClone.text();
			const responseData = JSON.parse(bodyText);
			const headerObj: Record<string, string> = {};
			responseHeaders.forEach((v, k) => {
				headerObj[k] = v;
			});
			setCached(
				procedurePath,
				cacheInput,
				responseData,
				upstreamResponse.status,
				headerObj,
				cacheTTL,
			);
		} catch {
			// Cache write failure is non-critical
		}
	}

	return new Response(upstreamResponse.body, {
		status: upstreamResponse.status,
		statusText: upstreamResponse.statusText,
		headers: responseHeaders,
	});
}

export { handler as GET, handler as POST };
