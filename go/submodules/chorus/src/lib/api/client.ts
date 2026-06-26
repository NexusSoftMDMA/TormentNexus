// Fetch wrapper with error handling and base URL management
import { ApiResponse } from "@/lib/types";

// Both server-side and browser-side route through the Next.js proxy
// at /api/daemon. The proxy handler runs in nodejs runtime and has
// reliable access to ~/.chorus/daemon.json — no need to duplicate
// the discovery logic across runtimes (which broke historically).
//
// Browser side: relative URL works because the cockpit is the origin.
// Server side: needs absolute URL constructed from the request's host
// (via next/headers) — Next.js requires absolute URLs in fetch() when
// running outside the browser.
const CLIENT_BASE = "/api/daemon";
const API_PREFIX = "/api/v1";

/**
 * Server-side base URL — routes through the same `/api/daemon` proxy
 * the browser uses, by constructing the cockpit's own origin from
 * the PORT env var that Next.js itself reads to bind.
 *
 * Why this shape (after multiple iterations):
 *
 *   v0.8.0 used a module-load snapshot of process.env.CHORUS_DAEMON_URL.
 *   v0.8.1 added eval('require')('node:fs') to read daemon.json from
 *           inside the Next.js bundle. Worked locally but failed in
 *           Node 22 + production server (ESM context).
 *   v0.8.5 routed through /api/daemon via next/headers(). Worked
 *           locally but the dynamic chunk import (c.e(248).then(...))
 *           failed in some user environments — fell through to a 7707
 *           default that hit a VSCode tunnel squatter and hung 15s.
 *   v0.8.7 (this): use process.env.PORT directly. PORT is set by
 *           `chorus start` when spawning the cockpit child AND is the
 *           same env Next.js reads to bind, so it's unambiguously
 *           available in every Next.js context. Synchronous, no
 *           chunk loading, no request-scope dependency.
 *
 * The proxy at /api/daemon runs in the catch-all route's nodejs
 * runtime where ~/.chorus/daemon.json reads work reliably — so this
 * function only needs to know the cockpit's own port, not the
 * daemon's.
 */
async function getServerBase(): Promise<string> {
  const port = process.env.PORT;
  if (port) return `http://127.0.0.1:${port}/api/daemon`;
  // Fallback for the rare case where PORT isn't set (some custom
  // Next.js host configurations). The legacy 5050 default matches
  // the cockpit's pre-port-shift defaults.
  return "http://127.0.0.1:5050/api/daemon";
}

async function getBaseUrl(): Promise<string> {
  if (typeof window === "undefined") return await getServerBase();
  return new URL(CLIENT_BASE, window.location.origin).toString();
}

export class DaemonError extends Error {
  constructor(
    public code: string,
    public statusCode: number,
    message: string,
    /**
     * Optional structured payload mirrored from the server envelope's
     * `error.details`. Used for zod-issue lists from /templates POST so
     * the cockpit can pin each error to the field it references.
     */
    public details?: Record<string, unknown>,
  ) {
    super(message);
    this.name = "DaemonError";
  }
}

export async function fetchFromDaemon<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const base = await getBaseUrl();
  // Normalise to a leading-slash form first so the prefix check can
  // anchor on `/api/v1` exactly. Without this, `api/v1/foo` (no leading
  // slash) gets double-prefixed → `/api/v1/api/v1/foo`.
  const normalised = path.startsWith("/") ? path : `/${path}`;
  // Exact-segment match — naive `startsWith` would match `/api/v10/...`
  // or `/api/v1foo/...` and skip prepending.
  const isPrefixed =
    normalised === API_PREFIX || normalised.startsWith(`${API_PREFIX}/`);
  const versionedPath = isPrefixed ? normalised : `${API_PREFIX}${normalised}`;
  const url = `${base.replace(/\/$/, "")}${versionedPath}`;

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    const raw: unknown = await response.json().catch(() => ({
      ok: false,
      error: {
        code: "parse_error",
        message: "Failed to parse response",
      },
    }));
    const data = raw as ApiResponse<T>;

    if (!data.ok) {
      // Fastify's default 404/500 returns `{ statusCode, error: "<string>",
      // message }` — `error` is a STRING, not the chorus envelope's object.
      // Without this branch, `data.error?.message` is undefined and we'd
      // surface "Unknown error" instead of the actual reason.
      const fastifyShape = raw as {
        statusCode?: number;
        error?: unknown;
        message?: string;
      };
      const isFastifyError =
        typeof fastifyShape.error === "string" &&
        typeof fastifyShape.message === "string";
      if (isFastifyError) {
        throw new DaemonError(
          fastifyShape.error as string,
          response.status,
          fastifyShape.message as string,
        );
      }
      throw new DaemonError(
        data.error?.code || "unknown",
        response.status,
        data.error?.message || `Daemon returned ${response.status}`,
        data.error?.details,
      );
    }

    return data.data as T;
  } catch (error) {
    if (error instanceof DaemonError) {
      throw error;
    }

    if (error instanceof TypeError && error.message.includes("fetch")) {
      throw new DaemonError(
        "connection_failed",
        0,
        "Failed to connect to Chorus daemon. Is it running?",
      );
    }

    throw new DaemonError(
      "unknown",
      0,
      error instanceof Error ? error.message : "Unknown error",
    );
  }
}
