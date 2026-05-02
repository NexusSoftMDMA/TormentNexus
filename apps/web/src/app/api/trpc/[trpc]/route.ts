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

async function tryCompatFallback(procedurePath: string): Promise<Response | null> {
  if (procedurePath.includes(',')) {
    return null;
  }

  const goApiBase = resolveGoApiBase().replace(/\/$/, '');
  const compatRoute =
    procedurePath === 'billing.getProviderQuotas'
      ? '/api/billing/provider-quotas'
      : procedurePath === 'billing.getFallbackChain'
        ? '/api/billing/fallback-chain'
        : null;

  if (!compatRoute) {
    return null;
  }

  try {
    const compatResponse = await fetch(`${goApiBase}${compatRoute}`);
    if (!compatResponse.ok) {
      return null;
    }

    const compatJson = await compatResponse.json();
    const payload = Array.isArray(compatJson?.data) || typeof compatJson?.data === 'object'
      ? compatJson.data
      : compatJson;

    return new Response(JSON.stringify([{ result: { data: payload } }]), {
      status: 200,
      headers: { 'content-type': 'application/json' },
    });
  } catch {
    return null;
  }
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
    const compatFallback = await tryCompatFallback(procedurePath);
    if (compatFallback) {
      console.warn(`[TRPC-Proxy] Using Go compat fallback for ${procedurePath} after upstream fetch failure`);
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
    const compatFallback = await tryCompatFallback(procedurePath);
    if (compatFallback) {
      console.warn(`[TRPC-Proxy] Using Go compat fallback for ${procedurePath} after upstream status ${upstreamResponse.status}`);
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
