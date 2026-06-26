import { NextRequest, NextResponse } from "next/server";
import { resolveDaemonUrl } from "@/lib/daemon-discovery";

/**
 * Generic proxy: /api/daemon/<path> → <daemon URL>/api/v1/<path>
 *
 * Why: the browser cannot reach the daemon directly (127.0.0.1 in the user's
 * browser is the user's machine, not the server hosting the Next app).
 * Server-side fetches still hit the daemon URL directly via lib/api/client.ts.
 *
 * Daemon URL resolution: see lib/daemon-discovery.ts. Honours the v0.8
 * `~/.chorus/daemon.json` first, falls back to `CHORUS_DAEMON_URL`,
 * then the legacy http://127.0.0.1:7707. `chorus start` sets the env
 * var when spawning the cockpit process so the proxy bypasses
 * disk reads in the steady state.
 */

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

const API_PREFIX = "api/v1";

/**
 * Module-level cache for the resolved daemon URL.
 *
 * Without this, every proxied request runs a synchronous disk read
 * (daemon.json), a /proc PID-alive check, AND an HTTP health probe
 * before forwarding — that's 1+ ms of overhead per UI call (see
 * gemini review feedback on PR #v0.8). Caching once per Next.js
 * process keeps the proxy on the fast path. The daemon.json port
 * only changes between `chorus stop` and `chorus start`, by which
 * point the cockpit process has already restarted, so the cache
 * lifetime aligns with the value's actual stability.
 */
let cachedDaemonUrl: string | null = null;

async function getDaemonUrl(): Promise<string> {
  if (cachedDaemonUrl) return cachedDaemonUrl;
  cachedDaemonUrl = await resolveDaemonUrl();
  return cachedDaemonUrl;
}

interface ProxyContext {
  params: Promise<{ path: string[] }>;
}

async function proxy(req: NextRequest, ctx: ProxyContext): Promise<Response> {
  const { path } = await ctx.params;
  // Re-encode each segment so a slashed value (e.g. an opencode voice
  // id `opencode-cli:opencode-go/kimi-k2.5`) survives the round trip.
  // Next.js decodes URL path components when populating the [...path]
  // catch-all; without re-encoding, `path.join("/")` would emit the
  // literal `/` and Fastify's `:id` parameter route on the daemon
  // would only match one segment of the id, return a default 404, and
  // the caller would surface the unhelpful "Unknown error" because
  // Fastify's 404 envelope doesn't match chorus's `{ ok, error }`
  // shape.
  const segments = path.map(encodeURIComponent).join("/");
  // Auto-prepend /api/v1 so cockpit code can call /api/daemon/<route>
  // while the daemon itself only exposes the versioned shape. Exact
  // segment check — `startsWith("api/v1")` would naively match
  // `api/v10/...` or `api/v1foo/...` and skip prepending.
  const isPrefixed = segments === API_PREFIX || segments.startsWith(`${API_PREFIX}/`);
  const versionedSegments = isPrefixed
    ? segments
    : `${API_PREFIX}/${segments}`;
  const search = req.nextUrl.search;
  const daemonUrl = await getDaemonUrl();
  const target = `${daemonUrl}/${versionedSegments}${search}`;

  const headers = new Headers();
  const accept = req.headers.get("accept");
  if (accept) headers.set("accept", accept);

  const init: RequestInit = {
    method: req.method,
    headers,
  };

  // Body forwarding rule: read the request body when content-length says
  // one is present OR the request is using chunked transfer-encoding
  // (which sends no content-length). Without the chunked check, large
  // uploads sent via Transfer-Encoding: chunked would silently drop their
  // body. The Content-Length-only path is the common one — fetchFromDaemon
  // never streams. The whole guard exists because Next 16's body iterator
  // hangs (UND_ERR_BODY_TIMEOUT) when Content-Type is set with no body —
  // see DELETE through fetchFromDaemon. We also drop the upstream
  // Content-Type for empty requests because Fastify rejects
  // application/json + empty body with FST_ERR_CTP_EMPTY_JSON_BODY.
  const hasContentLength =
    Number(req.headers.get("content-length") ?? "0") > 0;
  const isChunked = (req.headers.get("transfer-encoding") ?? "")
    .toLowerCase()
    .includes("chunked");
  const hasBody =
    req.method !== "GET" &&
    req.method !== "HEAD" &&
    (hasContentLength || isChunked);
  if (hasBody) {
    const ct = req.headers.get("content-type");
    if (ct) headers.set("content-type", ct);
    // arrayBuffer over text() — text() decodes the body as UTF-8 and
    // replaces invalid bytes with U+FFFD, which would silently corrupt
    // any future binary upload (file import, screenshot, etc.). All
    // current callers send JSON, so the practical impact today is nil,
    // but the proxy is a generic forward and shouldn't impose an
    // encoding on bytes it's just relaying.
    init.body = await req.arrayBuffer();
  }

  try {
    const upstream = await fetch(target, init);

    // Pass-through stream for SSE
    if (upstream.headers.get("content-type")?.includes("text/event-stream")) {
      return new Response(upstream.body, {
        status: upstream.status,
        headers: {
          "content-type": "text/event-stream",
          "cache-control": "no-cache",
          connection: "keep-alive",
        },
      });
    }

    const body = await upstream.text();
    return new NextResponse(body, {
      status: upstream.status,
      headers: {
        "content-type":
          upstream.headers.get("content-type") || "application/json",
      },
    });
  } catch {
    // Daemon may have shifted ports since we cached; invalidate so the
    // next request re-resolves. If it's genuinely down, the next call
    // will land here again and the user will see the same 502.
    cachedDaemonUrl = null;
    return NextResponse.json(
      {
        ok: false,
        error: {
          code: "daemon_unreachable",
          message:
            "Chorus daemon is not running on this host. Start it with `chorus start`.",
        },
      },
      { status: 502 },
    );
  }
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
