export const runtime = 'nodejs';

const DEFAULT_UPSTREAM_TRPC_URL = 'http://127.0.0.1:4100/trpc';
const DEFAULT_GO_API_BASE = 'http://127.0.0.1:4300';

function resolveUpstreamBase(): string {
  return process.env.BORG_TRPC_UPSTREAM?.trim() || DEFAULT_UPSTREAM_TRPC_URL;
}

function resolveGoApiBase(): string {
  return process.env.BORG_GO_API_BASE?.trim() || DEFAULT_GO_API_BASE;
}

function getProcedurePath(req: Request): string {
  const incomingUrl = new URL(req.url);
  const pathMatch = incomingUrl.pathname.match(/\/api\/trpc\/?(.*)$/);
  return pathMatch?.[1] ?? '';
}

function buildUpstreamUrl(req: Request): URL {
  const incomingUrl = new URL(req.url);
  const upstreamBase = resolveUpstreamBase().replace(/\/$/, '');
  const procedurePath = getProcedurePath(req);
  const upstreamUrl = new URL(`${upstreamBase}${procedurePath ? `/${procedurePath}` : ''}`);
  upstreamUrl.search = incomingUrl.search;
  return upstreamUrl;
}

function cloneHeaders(req: Request): Headers {
  const headers = new Headers(req.headers);
  headers.delete('host');
  headers.delete('content-length');
  return headers;
}

function getCompatRoute(procedurePath: string, input: unknown): string | null {
  if (procedurePath === 'billing.getProviderQuotas') {
    return '/api/billing/provider-quotas';
  }

  if (procedurePath === 'billing.getCostHistory') {
    const days = typeof input === 'object' && input !== null && 'days' in input
      ? Number((input as { days?: unknown }).days)
      : NaN;
    const normalizedDays = Number.isFinite(days) && days > 0 ? Math.min(Math.round(days), 90) : 30;
    return `/api/billing/cost-history?days=${normalizedDays}`;
  }

  if (procedurePath === 'billing.getModelPricing') {
    return '/api/billing/model-pricing';
  }

  if (procedurePath === 'billing.getFallbackChain') {
    const taskType = typeof input === 'object' && input !== null && 'taskType' in input
      ? (input as { taskType?: unknown }).taskType
      : undefined;
    const search = typeof taskType === 'string' && taskType.length > 0
      ? `?taskType=${encodeURIComponent(taskType)}`
      : '';
    return `/api/billing/fallback-chain${search}`;
  }

  if (procedurePath === 'billing.getTaskRoutingRules') {
    return '/api/billing/task-routing-rules';
  }

  if (procedurePath === 'memory.getRecentObservations') {
    const limit = typeof input === 'object' && input !== null && 'limit' in input
      ? Number((input as { limit?: unknown }).limit)
      : NaN;
    const namespace = typeof input === 'object' && input !== null && 'namespace' in input
      ? (input as { namespace?: unknown }).namespace
      : undefined;
    const type = typeof input === 'object' && input !== null && 'type' in input
      ? (input as { type?: unknown }).type
      : undefined;
    const params = new URLSearchParams();
    params.set('limit', String(Number.isFinite(limit) && limit > 0 ? Math.round(limit) : 6));
    if (typeof namespace === 'string' && namespace.length > 0) params.set('namespace', namespace);
    if (typeof type === 'string' && type.length > 0) params.set('type', type);
    return `/api/memory/observations/recent?${params.toString()}`;
  }

  if (procedurePath === 'memory.getRecentUserPrompts') {
    const limit = typeof input === 'object' && input !== null && 'limit' in input
      ? Number((input as { limit?: unknown }).limit)
      : NaN;
    const role = typeof input === 'object' && input !== null && 'role' in input
      ? (input as { role?: unknown }).role
      : undefined;
    const params = new URLSearchParams();
    params.set('limit', String(Number.isFinite(limit) && limit > 0 ? Math.round(limit) : 5));
    if (typeof role === 'string' && role.length > 0) params.set('role', role);
    return `/api/memory/user-prompts/recent?${params.toString()}`;
  }

  if (procedurePath === 'memory.getRecentSessionSummaries') {
    const limit = typeof input === 'object' && input !== null && 'limit' in input
      ? Number((input as { limit?: unknown }).limit)
      : NaN;
    const params = new URLSearchParams();
    params.set('limit', String(Number.isFinite(limit) && limit > 0 ? Math.round(limit) : 4));
    return `/api/memory/session-summaries/recent?${params.toString()}`;
  }

  if (procedurePath === 'mcpServers.syncTargets') {
    return '/api/mcp/servers/sync-targets';
  }

  if (procedurePath === 'mcpServers.exportClientConfig') {
    const client = typeof input === 'object' && input !== null && 'client' in input
      ? (input as { client?: unknown }).client
      : undefined;
    const path = typeof input === 'object' && input !== null && 'path' in input
      ? (input as { path?: unknown }).path
      : undefined;
    const params = new URLSearchParams();
    if (typeof client === 'string') params.set('client', client);
    if (typeof path === 'string') params.set('path', path);
    return `/api/mcp/servers/export-client-config?${params.toString()}`;
  }

  if (procedurePath === 'mcpServers.syncClientConfig') {
    return '/api/mcp/servers/sync-client-config';
  }

  return null;
}

async function getCompatPayload(procedurePath: string, input: unknown): Promise<unknown | null> {
  const compatRoute = getCompatRoute(procedurePath, input);
  if (!compatRoute) {
    return null;
  }

  const goApiBase = resolveGoApiBase().replace(/\/$/, '');

  try {
    const compatResponse = await fetch(`${goApiBase}${compatRoute}`);
    if (!compatResponse.ok) {
      return null;
    }

    const compatJson = await compatResponse.json();
    return Array.isArray(compatJson?.data) || typeof compatJson?.data === 'object'
      ? compatJson.data
      : compatJson;
  } catch {
    return null;
  }
}

function parseBatchInput(req: Request): Record<string, unknown> {
  const inputParam = new URL(req.url).searchParams.get('input');
  if (!inputParam) {
    return {};
  }

  try {
    const parsed = JSON.parse(inputParam);
    return parsed && typeof parsed === 'object' ? parsed as Record<string, unknown> : {};
  } catch {
    return {};
  }
}

async function fetchSingleProcedureEntry(procedurePath: string, input: unknown): Promise<any | null> {
  const upstreamBase = resolveUpstreamBase().replace(/\/$/, '');
  const upstreamUrl = new URL(`${upstreamBase}/${procedurePath}`);
  upstreamUrl.searchParams.set('batch', '1');
  upstreamUrl.searchParams.set('input', JSON.stringify(input ?? {}));

  try {
    const response = await fetch(upstreamUrl, { method: 'GET' });
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

async function tryCompatFallback(req: Request, procedurePath: string): Promise<Response | null> {
  if (!procedurePath.includes(',')) {
    const compatPayload = await getCompatPayload(procedurePath, {});
    if (compatPayload === null) {
      return null;
    }

    return new Response(JSON.stringify([{ result: { data: compatPayload } }]), {
      status: 200,
      headers: { 'content-type': 'application/json' },
    });
  }

  const procedures = procedurePath.split(',').map((entry) => entry.trim()).filter(Boolean);
  if (procedures.length === 0) {
    return null;
  }

  const batchInput = parseBatchInput(req);
  const entries = [];
  for (const [index, procedure] of procedures.entries()) {
    const entry = await fetchSingleProcedureEntry(procedure, batchInput[String(index)] ?? {});
    if (!entry) {
      return null;
    }
    entries.push(entry);
  }

  const hasErrors = entries.some((entry) => entry?.error);
  return new Response(JSON.stringify(entries), {
    status: hasErrors ? 207 : 200,
    headers: { 'content-type': 'application/json' },
  });
}

async function handler(req: Request): Promise<Response> {
  const procedurePath = getProcedurePath(req);
  const upstreamUrl = buildUpstreamUrl(req);
  const headers = cloneHeaders(req);
  const hasBody = req.method !== 'GET' && req.method !== 'HEAD';
  const body = hasBody ? await req.text() : undefined;

  let upstreamResponse: Response;
  try {
    console.log(`[TRPC-Proxy] Fetching from upstream: ${upstreamUrl.toString()} (${req.method})`);
    upstreamResponse = await fetch(upstreamUrl, {
      method: req.method,
      headers,
      body,
    });
    console.log(`[TRPC-Proxy] Upstream responded: ${upstreamResponse.status} ${upstreamResponse.statusText}`);
  } catch (error) {
    const compatFallback = await tryCompatFallback(req, procedurePath);
    if (compatFallback) {
      console.warn(`[TRPC-Proxy] Using compat fallback for ${procedurePath} after upstream fetch failure`);
      return compatFallback;
    }

    const message = error instanceof Error ? error.message : String(error);
    console.error(`[TRPC-Proxy] Upstream fetch failed: ${message}`);
    return new Response(
      JSON.stringify({
        error: 'TRPC_UPSTREAM_UNAVAILABLE',
        message,
        upstream: upstreamUrl.toString(),
      }),
      {
        status: 502,
        headers: { 'content-type': 'application/json' },
      },
    );
  }

  if (!upstreamResponse.ok) {
    const compatFallback = await tryCompatFallback(req, procedurePath);
    if (compatFallback) {
      console.warn(`[TRPC-Proxy] Using compat fallback for ${procedurePath} after upstream status ${upstreamResponse.status}`);
      return compatFallback;
    }
  }

  const responseHeaders = new Headers(upstreamResponse.headers);
  const isSse = responseHeaders.get('content-type') === 'text/event-stream';
  if (isSse) {
    responseHeaders.set('Connection', 'keep-alive');
    responseHeaders.set('Cache-Control', 'no-cache');
  }

  return new Response(upstreamResponse.body, {
    status: upstreamResponse.status,
    statusText: upstreamResponse.statusText,
    headers: responseHeaders,
  });
}

export { handler as GET, handler as POST };
