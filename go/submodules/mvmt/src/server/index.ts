import { createHash, randomUUID } from 'crypto';
import { execFile } from 'node:child_process';
import { createReadStream, createWriteStream } from 'fs';
import fsp from 'fs/promises';
import { Server as HttpServer } from 'node:http';
import path from 'path';
import { Transform } from 'stream';
import { pipeline } from 'stream/promises';
import { promisify } from 'node:util';
import express, { Request, Response } from 'express';
import type { AccessToken } from './oauth.js';
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { StreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/streamableHttp.js';
import { CallToolRequestSchema, ListToolsRequestSchema } from '@modelcontextprotocol/sdk/types.js';
import { ToolRouter } from '../apps/mcp/router.js';
import { log } from '../utils/logger.js';
import { defaultClientsPath, defaultRefreshTokensPath, defaultSigningKeyPath, ensureSessionToken, ensureSigningKey, readSigningKey, TOKEN_PATH, validateSessionToken } from '../utils/token.js';
import { verifyApiToken } from '../utils/api-token-hash.js';
import { isExpired, parseTokenTtl } from '../utils/token-ttl.js';
import {
  CodeChallengeMethod,
  OAuthClientAlreadyRegisteredError,
  OAuthClientPersistenceError,
  OAuthClientRegistryLimitError,
  OAuthError,
  OAuthStore,
  getBaseUrl,
  renderAuthorizePage,
} from './oauth.js';
import { rateLimit } from './rate-limit.js';
import { ClientConfig, LocalFolderMountConfig, PermissionConfig } from '../config/schema.js';
import { readConfig, saveConfig, withConfigLock } from '../config/loader.js';
import { MountRegistry, normalizeVirtualPath, RegisteredMount } from '../context/mount-registry.js';
import { LocalFolderStorageProvider, type StorageProviderReadStream, type StorageProviderStat } from '../context/storage-provider.js';
import { isTextPath, MAX_TEXT_BYTES } from '../context/text-index.js';
import { addMountToConfig, editMountInConfig, MountInput, removeMountFromConfig } from '../cli/mounts.js';
import { addApiTokenToConfig, removeApiTokenFromConfig, setApiTokenPublishedInConfig } from '../cli/api-tokens.js';
import { resolveSetupPath } from '../connectors/setup-paths.js';
import { normalizeDashboardPath, resolveDashboardFileTarget } from '../apps/dashboard/files.js';
import { getApp, INSTALLED_APPS } from '../apps/registry.js';
import {
  defaultPrivilegedUsersPath,
  findPrivilegedUser,
  PrivilegedUser,
  recordPrivilegedUserLogin,
  verifyPrivilegedUserPassword,
} from '../apps/dashboard/users.js';
import { listLeaseDirectory, resolveLeaseFileTarget, resolveLeaseUploadTarget } from '../core/leases/files.js';
import {
  createLease,
  defaultLeasesPath,
  findLease,
  findLeaseByToken,
  leaseAllows,
  LeaseRecord,
  LeaseResource,
  leaseUnavailableReason,
  listLeases,
  recordLeaseUse,
  reserveLeaseDownload,
  revokeLease,
  rotateLeaseToken,
  setLeasePublished,
  validateLeaseToken,
} from '../core/leases/store.js';
import { findLeaseSecret, leaseSecretsPathForLeaseStore, removeLeaseSecret, saveLeaseSecret } from '../core/leases/secrets.js';
import { clientConfigToGrant, isGrantPublished, leaseGrantScope } from '../grant/model.js';
import {
  attachClientIdentity,
  ClientIdentity,
  clientBindingMatches,
  clientCredentialVersion,
  isQuarantined,
  readClientIdentity,
  resolveClientIdentity,
} from './client-identity.js';
import { pathAllowed, pathMayExposeEntry } from '../core/auth/permissions.js';
import { AuditLogger, summarizeArgs } from '../utils/audit.js';

// Rate limits are defense-in-depth against brute-force and DoS,
// primarily meaningful when mvmt is exposed via a tunnel. Auth-gated
// data-plane routes (/mcp) get a generous cap that comfortably covers
// MCP polling. Auth surface routes (/authorize, /token, /register) get
// a tight cap to slow session-token guessing. /health is polled
// frequently and only checks the bearer.
const DEFAULT_AUTH_RATE_LIMIT = { windowMs: 60_000, max: 30 };
const DEFAULT_MCP_RATE_LIMIT = { windowMs: 60_000, max: 600 };
const DEFAULT_HEALTH_RATE_LIMIT = { windowMs: 60_000, max: 120 };
const DEFAULT_MAX_LEASE_UPLOAD_BYTES = 10 * 1024 * 1024 * 1024;
const DEFAULT_MAX_FS_UPLOAD_BYTES = DEFAULT_MAX_LEASE_UPLOAD_BYTES;
const DEFAULT_DASHBOARD_LEASE_TTL = '24h';
const DASHBOARD_SESSION_TTL_MS = 12 * 60 * 60 * 1000;
const DASHBOARD_SESSION_COOKIE = 'mvmt_dashboard';
const execFileAsync = promisify(execFile);

type McpSession = {
  transport: StreamableHTTPServerTransport;
  server: Server;
  clientIdentityRef: ClientIdentityRef;
  clientIdentity?: ClientIdentity;
  lastActivity: number;
};

type ClientIdentityRef = {
  current?: ClientIdentity;
};

type DashboardSession = {
  id: string;
  userId: string;
  username: string;
  expiresAt: number;
  admin: boolean;
};

export interface RateLimitOverrides {
  auth?: { windowMs: number; max: number };
  mcp?: { windowMs: number; max: number };
  health?: { windowMs: number; max: number };
}

export interface HttpServerOptions {
  port: number;
  allowedOrigins?: string[];
  tokenPath?: string;
  signingKeyPath?: string;
  // Called at request time to resolve the public-facing base URL for
  // OAuth metadata and redirect issuer values. When set, its result
  // overrides the request Host header. This is how the tunnel URL gets
  // propagated into OAuth responses without trusting X-Forwarded-* headers.
  resolvePublicBaseUrl?: () => string | undefined;
  // Override rate-limit buckets; used by tests to exercise the 429 path
  // without blasting thousands of requests.
  rateLimits?: RateLimitOverrides;
  requestLog?: (entry: HttpRequestLogEntry) => void;
  audit?: AuditLogger;
  maxFsUploadBytes?: number;
  // Per-client policy entries (from `config.clients`). When undefined or
  // empty, requests authenticated via the session token resolve to a
  // synthesized default identity that preserves pre-PR single-token
  // behavior. Pass an array to enable per-client identity resolution.
  clients?: readonly ClientConfig[] | (() => readonly ClientConfig[] | undefined);
  leaseMounts?: readonly LocalFolderMountConfig[] | (() => readonly LocalFolderMountConfig[] | undefined);
  leaseStorePath?: string;
  leaseSecretsPath?: string;
  privilegedUsersPath?: string;
  localPathPicker?: (kind: 'file' | 'folder') => Promise<string | undefined>;
  // Path to the on-disk config file. When set, the dashboard can persist
  // mount changes (add/edit/remove) through this file using the same
  // helpers and lock that the CLI uses. When undefined, dashboard mount
  // mutation endpoints respond 403.
  configPath?: string;
  // Defaults to true for local backward compatibility. Tunnel mode passes
  // false unless MVMT_ALLOW_LEGACY_TUNNEL is set, so a public endpoint can
  // start with no API tokens while exposing no data.
  allowLegacyDefaultClient?: boolean | (() => boolean);
}

export interface StartedHttpServer {
  port: number;
  close(): Promise<void>;
}

export interface HttpRequestLogEntry {
  ts: string;
  kind: string;
  method: string;
  path: string;
  status: number;
  detail?: string;
  clientId?: string;
  ip?: string;
}

export const MVMT_SERVER_INSTRUCTIONS = [
  'mvmt exposes selected local files through permissioned tools.',
  '',
  'Use mvmt when the user asks about their own notes, files, projects, workspace, local docs, or previously mounted content. For content questions, call search first. If a result looks relevant, call read with the returned path before answering.',
  '',
  'Use list only to discover available mounts or browse directories. Prefer search over list+read for topic questions.',
  '',
  'Do not use mvmt for general web or current-events questions unless the user asks about local files. Never write or remove unless the user explicitly asks to create, overwrite, or delete a specific file. If search returns no useful results, say that instead of inventing local file contents.',
].join('\n');

export function createMcpServer(router: ToolRouter, clientIdentity?: ClientIdentity | (() => ClientIdentity | undefined)): Server {
  const resolveClientIdentityForRequest = typeof clientIdentity === 'function'
    ? clientIdentity
    : () => clientIdentity;
  const server = new Server(
    { name: 'mvmt', version: '0.1.0' },
    {
      capabilities: { tools: {} },
      instructions: MVMT_SERVER_INSTRUCTIONS,
    },
  );

  server.setRequestHandler(ListToolsRequestSchema, async () => ({
    tools: router.getAllTools(resolveClientIdentityForRequest()).map((tool) => ({
      name: tool.namespacedName,
      description: tool.description,
      inputSchema: tool.inputSchema as { type: 'object'; properties?: Record<string, object>; required?: string[] },
    })),
  }));

  server.setRequestHandler(CallToolRequestSchema, async (request) => {
    try {
      const result = await router.callTool(
        request.params.name,
        request.params.arguments ?? {},
        resolveClientIdentityForRequest(),
      );
      return result as any;
    } catch (err) {
      return {
        content: [
          {
            type: 'text' as const,
            text: `Error: ${err instanceof Error ? err.message : 'Unknown error'}`,
          },
        ],
        isError: true,
      };
    }
  });

  return server;
}

// Returns a close handle so the caller can tear down the transport on
// shutdown. Without this, stdio child processes could outlive the parent.
export async function startStdioServer(router: ToolRouter): Promise<{ close(): Promise<void> }> {
  const server = createMcpServer(router);
  const transport = new StdioServerTransport();
  await server.connect(transport);
  return {
    close: () => transport.close(),
  };
}

export async function startHttpServer(router: ToolRouter, options: HttpServerOptions): Promise<StartedHttpServer> {
  const { port } = options;
  const tokenPath = options.tokenPath;
  const dashboardLogs: HttpRequestLogEntry[] = [];
  const requestLog = (entry: HttpRequestLogEntry) => {
    dashboardLogs.push(entry);
    if (dashboardLogs.length > 100) dashboardLogs.shift();
    options.requestLog?.(entry);
  };
  const resolvePublicBaseUrl = options.resolvePublicBaseUrl;
  // Proxied/tunneled deployments should set resolvePublicBaseUrl so OAuth
  // resource audience checks do not depend on the request Host header.
  const baseUrlFor = (req: Request): string => getBaseUrl(req, resolvePublicBaseUrl?.());
  // baseUrlFor() returns just protocol+host — fine for origin comparison,
  // but it strips any relay workspace prefix (e.g. /t/demo). Every URL
  // we hand to a client or browser must keep that prefix so the relay
  // can route it back to this agent — share links, and the OAuth issuer/
  // metadata/endpoints/resource. Without the prefix those URLs hit the
  // relay's catch-all "no explicit mapping for /error" 404 and the OAuth
  // discovery chain dead-ends. When no relay URL is configured this
  // falls back to baseUrlFor(), so local-only behaviour is unchanged.
  const userFacingBaseUrlFor = (req: Request): string => resolvePublicBaseUrl?.() ?? baseUrlFor(req);
  const app = express();
  app.use(express.json({
    limit: '10mb',
    type: (req) => {
      if (req.method === 'PUT' && typeof req.url === 'string' && req.url.startsWith('/api/fs/file')) return false;
      const header = Array.isArray(req.headers['content-type']) ? req.headers['content-type'][0] : req.headers['content-type'];
      return typeof header === 'string' && /^\s*application\/(?:[\w.+-]+\+)?json\b/i.test(header);
    },
  }));
  app.use(express.urlencoded({ extended: false, limit: '64kb' }));

  ensureSessionToken(tokenPath);
  const signingKeyPath = options.signingKeyPath ?? defaultSigningKeyPath(tokenPath ?? TOKEN_PATH);
  const leaseStorePath = options.leaseStorePath ?? defaultLeasesPath(tokenPath ?? TOKEN_PATH);
  const privilegedUsersPath = options.privilegedUsersPath ?? defaultPrivilegedUsersPath(tokenPath ?? TOKEN_PATH);
  // Create the signing key file on first boot, then re-read it on every
  // HMAC op. This way internal session-token rotation (which rewrites the file)
  // invalidates outstanding OAuth access tokens immediately, without
  // requiring a server restart.
  ensureSigningKey(signingKeyPath);
  const clientsPath = defaultClientsPath(tokenPath ?? TOKEN_PATH);
  const refreshTokensPath = defaultRefreshTokensPath(tokenPath ?? TOKEN_PATH);
  const leaseSecretsPath = options.leaseSecretsPath ?? leaseSecretsPathForLeaseStore(leaseStorePath);
  const maxFsUploadBytes = options.maxFsUploadBytes ?? DEFAULT_MAX_FS_UPLOAD_BYTES;
  const oauth = new OAuthStore({
    signingKey: () => readSigningKey(signingKeyPath) ?? ensureSigningKey(signingKeyPath),
    clientsPath,
    refreshTokensPath,
  });
  const oauthCleanup = setInterval(() => oauth.cleanup(), 60 * 1000);
  oauthCleanup.unref();

  const sessions = new Map<string, McpSession>();
  const dashboardSessions = new Map<string, DashboardSession>();
  const cleanupInterval = setInterval(() => {
    const now = Date.now();
    const staleTimeoutMs = 30 * 60 * 1000;

    for (const [id, session] of sessions) {
      if (now - session.lastActivity > staleTimeoutMs) {
        session.transport.close().catch(() => undefined);
        sessions.delete(id);
      }
    }

    for (const [id, session] of dashboardSessions) {
      if (session.expiresAt <= now) dashboardSessions.delete(id);
    }
  }, 5 * 60 * 1000);
  cleanupInterval.unref();

  const authLimiter = rateLimit(options.rateLimits?.auth ?? DEFAULT_AUTH_RATE_LIMIT);
  const mcpLimiter = rateLimit(options.rateLimits?.mcp ?? DEFAULT_MCP_RATE_LIMIT);
  const healthLimiter = rateLimit(options.rateLimits?.health ?? DEFAULT_HEALTH_RATE_LIMIT);

  const originCheck = buildOriginCheck(options.allowedOrigins ?? []);

  const oauthOriginMiddleware: express.RequestHandler = (req, res, next) => {
    if (originCheck(req) || originMatchesBaseUrl(req, baseUrlFor(req))) {
      next();
      return;
    }
    logHttpRequest(requestLog, req, 403, 'oauth.origin', 'origin_not_allowed');
    res.status(403).json({ error: 'Origin not allowed' });
  };

  const authMiddleware: express.RequestHandler = (req, res, next) => {
    if (!originCheck(req)) {
      logHttpRequest(requestLog, req, 403, authLogKind(req), 'origin_not_allowed');
      res.status(403).json({ error: 'Origin not allowed' });
      return;
    }
    const authHeader = firstHeaderValue(req.headers.authorization);
    if (!authHeader) {
      const baseUrl = userFacingBaseUrlFor(req);
      res.setHeader(
        'WWW-Authenticate',
        `Bearer realm="mvmt", resource_metadata="${baseUrl}/.well-known/oauth-protected-resource"`,
      );
      logHttpRequest(requestLog, req, 401, authLogKind(req), 'missing_bearer');
      res.status(401).json({ error: 'Invalid or missing bearer token' });
      return;
    }
    if (!authHeader.startsWith('Bearer ')) {
      const baseUrl = userFacingBaseUrlFor(req);
      res.setHeader(
        'WWW-Authenticate',
        `Bearer realm="mvmt", resource_metadata="${baseUrl}/.well-known/oauth-protected-resource"`,
      );
      logHttpRequest(requestLog, req, 401, authLogKind(req), 'non_bearer_auth');
      res.status(401).json({ error: 'Invalid or missing bearer token' });
      return;
    }
    // Expected audience is the canonical resource identifier we
    // advertise in protected-resource-metadata. The authorize flow
    // copies the client-supplied `resource` into the token's `aud`;
    // we validate it here so a token minted for a different resource
    // cannot be replayed at this server.
    const expectedAudience = `${userFacingBaseUrlFor(req)}/mcp`;
    const oauthAccessToken = oauth.validateAccessToken(authHeader, {
      expectedAudience,
      allowLegacyNoAudience: true,
    });
    let identity = resolveClientIdentity({
      authHeader,
      clients: resolveClients(options.clients),
      oauthAccessToken,
      validateSession: (header) => validateSessionToken(header, tokenPath),
      allowLegacyDefault: resolveAllowLegacyDefaultClient(options.allowLegacyDefaultClient),
      clientHint: requestClientHint(req, oauthAccessToken?.clientId),
    });
    if (!identity) {
      const lease = findLeaseByToken(leaseStorePath, bearerToken(req));
      if (lease && !leaseUnavailableReason(lease)) identity = identityFromLease(lease);
    }
    if (identity && isQuarantined(identity)) {
      // Quarantined identities are authenticated (the OAuth access token
      // is valid) but the OAuth client_id has no mapping to a configured
      // client. Reject at auth time until a separate enforcement layer
      // exists; otherwise quarantine would be in-name-only and the
      // unknown client would still reach the global tool surface.
      logHttpRequest(
        requestLog,
        req,
        403,
        authLogKind(req),
        `quarantined oauth_client_id=${identity.oauthClientId ?? '(none)'}`,
        identity.id,
      );
      res.status(403).json({
        error: 'oauth_client_quarantined',
        error_description: 'OAuth client_id is not mapped to a configured mvmt client; admin must approve',
      });
      return;
    }
    if (identity && !isLocalRequest(req) && !isGrantPublished(identity.published)) {
      // Exposure boundary: a capability-only grant (explicitly published:
      // false) is reachable only by apps on this machine. Any request
      // arriving through a public tunnel - the mvmt relay, Cloudflare,
      // pinggy, localhost.run, a custom tunnel - is rejected. Grants
      // without an explicit published value are grandfathered as
      // published.
      logHttpRequest(requestLog, req, 403, authLogKind(req), 'grant_not_published', identity.id);
      res.status(403).json({
        error: 'grant_not_published',
        error_description: 'This grant is not published for remote access',
      });
      return;
    }
    if (identity) {
      attachClientIdentity(req, identity);
      next();
      return;
    }

    const baseUrl = userFacingBaseUrlFor(req);
    res.setHeader(
      'WWW-Authenticate',
      `Bearer realm="mvmt", resource_metadata="${baseUrl}/.well-known/oauth-protected-resource"`,
    );
    logHttpRequest(requestLog, req, 401, authLogKind(req), 'invalid_bearer');
    res.status(401).json({ error: 'Invalid or missing bearer token' });
  };

  const authorizationServerMetadata = (req: Request, res: Response) => {
    const baseUrl = userFacingBaseUrlFor(req);
    logHttpRequest(requestLog, req, 200, 'oauth.discovery');
    res.json({
      issuer: baseUrl,
      authorization_endpoint: `${baseUrl}/authorize`,
      token_endpoint: `${baseUrl}/token`,
      registration_endpoint: `${baseUrl}/register`,
      response_types_supported: ['code'],
      grant_types_supported: ['authorization_code', 'refresh_token'],
      code_challenge_methods_supported: ['S256'],
      token_endpoint_auth_methods_supported: ['none'],
      scopes_supported: ['mcp', 'offline_access'],
    });
  };
  app.get('/.well-known/oauth-authorization-server', authorizationServerMetadata);
  app.get('/.well-known/oauth-authorization-server/mcp', authorizationServerMetadata);

  const protectedResourceMetadata = (req: Request, res: Response) => {
    const baseUrl = userFacingBaseUrlFor(req);
    logHttpRequest(requestLog, req, 200, 'oauth.resource');
    res.json({
      resource: `${baseUrl}/mcp`,
      authorization_servers: [baseUrl],
      bearer_methods_supported: ['header'],
      scopes_supported: ['mcp', 'offline_access'],
    });
  };
  app.get('/.well-known/oauth-protected-resource', protectedResourceMetadata);
  app.get('/.well-known/oauth-protected-resource/mcp', protectedResourceMetadata);

  app.post('/register', authLimiter, oauthOriginMiddleware, (req, res) => {
    const requested = (req.body ?? {}) as Record<string, unknown>;
    const requestedClientId = stringField(requested.client_id);
    if (requestedClientId && !validateSessionToken(firstHeaderValue(req.headers.authorization), tokenPath)) {
      logHttpRequest(requestLog, req, 401, 'oauth.register', 'client_id_requires_session_token', requestedClientId);
      res.status(401).json({
        error: 'invalid_client',
        error_description: 'Supplying client_id requires the mvmt session token',
      });
      return;
    }
    const clientId = requestedClientId ?? `mvmt-${randomUUID()}`;
    const clientName = typeof requested.client_name === 'string' ? requested.client_name : undefined;
    const scope = typeof requested.scope === 'string' ? requested.scope : undefined;
    const redirectUris = Array.isArray(requested.redirect_uris)
      ? requested.redirect_uris.filter((uri): uri is string => typeof uri === 'string' && uri.length > 0)
      : [];
    if (redirectUris.length === 0) {
      logHttpRequest(
        requestLog, req, 400, 'oauth.register',
        'missing_redirect_uris', clientId,
      );
      res.status(400).json({ error: 'invalid_redirect_uri', error_description: 'redirect_uris is required' });
      return;
    }
    const redirectHosts: string[] = [];
    for (const uri of redirectUris) {
      try {
        const parsed = new URL(uri);
        redirectHosts.push(parsed.host);
      } catch {
        logHttpRequest(requestLog, req, 400, 'oauth.register', 'invalid_redirect_uri', clientId);
        res.status(400).json({ error: 'invalid_redirect_uri' });
        return;
      }
    }
    let registered;
    try {
      registered = oauth.registerClient({ clientId, redirectUris });
    } catch (err) {
      if (err instanceof OAuthClientAlreadyRegisteredError) {
        logHttpRequest(requestLog, req, 409, 'oauth.register', 'client_id_already_registered', clientId);
        res.status(409).json({
          error: 'invalid_client_metadata',
          error_description: 'client_id is already registered',
        });
        return;
      }
      if (err instanceof OAuthClientPersistenceError) {
        logHttpRequest(requestLog, req, 500, 'oauth.register', 'persist_failed', clientId);
        res.status(500).json({
          error: 'server_error',
          error_description: 'Failed to persist OAuth client registration',
        });
        return;
      }
      if (err instanceof OAuthClientRegistryLimitError) {
        logHttpRequest(requestLog, req, 429, 'oauth.register', 'registry_limit', clientId);
        res.status(429).json({
          error: 'invalid_client_metadata',
          error_description: err.message,
        });
        return;
      }
      throw err;
    }
    const detail = `redirect_uris=${registered.redirectUris.length} hosts=${[...new Set(redirectHosts)].join(',')}${clientName ? ` name="${clientName}"` : ''}`;
    logHttpRequest(requestLog, req, 201, 'oauth.register', detail, clientId);
    // Response shape follows RFC 7591 §3.2.1 (Client Information Response).
    // We are a public client (no client_secret), so token_endpoint_auth_method
    // is "none". client_id_issued_at helps clients audit when they were
    // provisioned. client_name and scope are echoed only when the client
    // supplied them.
    res.status(201).json({
      client_id: registered.clientId,
      client_id_issued_at: Math.floor(Date.now() / 1000),
      token_endpoint_auth_method: 'none',
      grant_types: ['authorization_code', 'refresh_token'],
      response_types: ['code'],
      redirect_uris: registered.redirectUris,
      ...(clientName ? { client_name: clientName } : {}),
      ...(scope ? { scope } : {}),
    });
  });

  app.get('/authorize', authLimiter, oauthOriginMiddleware, (req, res) => {
    const parsed = parseAuthorizeParams(req.query);
    if ('error' in parsed) {
      logHttpRequest(requestLog, req, 400, 'oauth.authorize', parsed.error);
      res.status(400).type('text/plain').send(parsed.error);
      return;
    }
    const canonicalResource = `${userFacingBaseUrlFor(req)}/mcp`;
    const { params, resourceDefaulted } = defaultAuthorizeResource(parsed, canonicalResource);
    if (!oauth.isRedirectUriAllowed(params.clientId, params.redirectUri)) {
      const rejectedHost = safeHost(params.redirectUri);
      logHttpRequest(
        requestLog, req, 400, 'oauth.authorize',
        `unregistered_redirect_uri host=${rejectedHost}`, params.clientId,
      );
      res.status(400).type('text/plain').send('redirect_uri is not registered for this client');
      return;
    }
    const resourceError = validateAuthorizeResource(params.resource, canonicalResource);
    if (resourceError) {
      const redirect = authorizeErrorRedirect(params, 'invalid_target', resourceError.description);
      logHttpRequest(requestLog, req, 302, 'oauth.authorize', resourceError.code, params.clientId);
      res.redirect(302, redirect);
      return;
    }
    const requestId = params.requestId ?? randomUUID();
    const promptDetail = formatAuthorizeLogDetail({
      phase: 'prompt',
      requestId,
      redirectUri: params.redirectUri,
      resource: params.resource,
      resourceDefaulted,
      state: params.state,
    });
    logHttpRequest(requestLog, req, 200, 'oauth.authorize', promptDetail, params.clientId);
    res.type('text/html').send(renderAuthorizePage({ ...params, requestId }));
  });

  app.post('/authorize', authLimiter, oauthOriginMiddleware, (req, res) => {
    const parsed = parseAuthorizeParams(req.body ?? {});
    if ('error' in parsed) {
      logHttpRequest(requestLog, req, 400, 'oauth.authorize', parsed.error);
      res.status(400).type('text/plain').send(parsed.error);
      return;
    }
    const canonicalResource = `${userFacingBaseUrlFor(req)}/mcp`;
    const { params, resourceDefaulted } = defaultAuthorizeResource(parsed, canonicalResource);
    if (!oauth.isRedirectUriAllowed(params.clientId, params.redirectUri)) {
      const rejectedHost = safeHost(params.redirectUri);
      logHttpRequest(
        requestLog, req, 400, 'oauth.authorize',
        `unregistered_redirect_uri host=${rejectedHost}`, params.clientId,
      );
      res.status(400).type('text/plain').send('redirect_uri is not registered for this client');
      return;
    }
    const resourceError = validateAuthorizeResource(params.resource, canonicalResource);
    if (resourceError) {
      const redirect = authorizeErrorRedirect(params, 'invalid_target', resourceError.description);
      logHttpRequest(requestLog, req, 302, 'oauth.authorize', resourceError.code, params.clientId);
      res.redirect(302, redirect);
      return;
    }

    const requestId = params.requestId ?? randomUUID();
    const approval = resolveAuthorizeApproval(
      req.body ?? {},
      resolveClients(options.clients),
      tokenPath,
      requestClientHint(req, params.clientId),
    );
    if (!approval.ok) {
      const denyDetail = formatAuthorizeLogDetail({
        phase: approval.phase,
        requestId,
        redirectUri: params.redirectUri,
        resource: params.resource,
        resourceDefaulted,
        state: params.state,
      });
      logHttpRequest(requestLog, req, 401, 'oauth.authorize', denyDetail, params.clientId);
      res
        .status(401)
        .type('text/html')
        .send(renderAuthorizePage({ ...params, requestId, error: approval.message }));
      return;
    }

    const authCode = oauth.issueCode({
      clientId: params.clientId,
      mvmtClientId: approval.mvmtClientId,
      mvmtClientCredentialVersion: approval.mvmtClientCredentialVersion,
      redirectUri: params.redirectUri,
      resource: params.resource,
      codeChallenge: params.codeChallenge,
      codeChallengeMethod: params.codeChallengeMethod,
      scope: params.scope,
    });

    const redirect = new URL(params.redirectUri);
    redirect.searchParams.set('code', authCode.code);
    if (params.state) redirect.searchParams.set('state', params.state);
    const approveDetail = formatAuthorizeLogDetail({
      phase: 'approved_redirect',
      requestId,
      redirectUri: params.redirectUri,
      resource: params.resource,
      resourceDefaulted,
      state: params.state,
      authorizedClientId: approval.mvmtClientId,
    });
    logHttpRequest(requestLog, req, 302, 'oauth.authorize', approveDetail, params.clientId);
    res.redirect(302, redirect.toString());
  });

  app.post('/token', authLimiter, oauthOriginMiddleware, (req, res) => {
    const body = (req.body ?? {}) as Record<string, unknown>;
    const grantType = typeof body.grant_type === 'string' ? body.grant_type : undefined;
    const requestClientId = typeof body.client_id === 'string' ? body.client_id : undefined;
    if (grantType !== 'authorization_code' && grantType !== 'refresh_token') {
      logHttpRequest(requestLog, req, 400, 'oauth.token', 'unsupported_grant_type');
      res.status(400).json({ error: 'unsupported_grant_type' });
      return;
    }

    try {
      if (grantType === 'authorization_code') {
        const code = typeof body.code === 'string' ? body.code : undefined;
        const clientId = requestClientId;
        const redirectUri = typeof body.redirect_uri === 'string' ? body.redirect_uri : undefined;
        const resource = typeof body.resource === 'string' ? body.resource : undefined;
        const codeVerifier = typeof body.code_verifier === 'string' ? body.code_verifier : undefined;

        if (!code || !clientId || !redirectUri || !codeVerifier) {
          logHttpRequest(requestLog, req, 400, 'oauth.token', 'invalid_request', clientId);
          res.status(400).json({ error: 'invalid_request' });
          return;
        }

        const tokens = oauth.exchangeCode({ code, clientId, redirectUri, resource, codeVerifier });
        if (!oauthGrantMatchesCurrentClient(
          tokens.accessToken,
          resolveClients(options.clients),
          requestClientHint(req, clientId),
        )) {
          throw new OAuthError('invalid_grant', 'Scoped API token was rotated or removed');
        }
        const tokenDetail = `issued grant=authorization_code aud=${tokens.accessToken.audience ? safeHost(tokens.accessToken.audience) : '(none)'}`;
        logHttpRequest(requestLog, req, 200, 'oauth.token', tokenDetail, clientId);
        res.json({
          access_token: tokens.accessToken.token,
          token_type: 'Bearer',
          expires_in: oauth.tokenTtlSeconds,
          refresh_token: tokens.refreshToken.token,
          scope: tokens.accessToken.scope,
        });
        return;
      }

      const refreshToken = typeof body.refresh_token === 'string' ? body.refresh_token : undefined;
      const clientId = requestClientId;
      const scope = typeof body.scope === 'string' ? body.scope : undefined;
      if (!refreshToken || !clientId) {
        logHttpRequest(requestLog, req, 400, 'oauth.token', 'invalid_request', clientId);
        res.status(400).json({ error: 'invalid_request' });
        return;
      }

      const tokens = oauth.exchangeRefreshToken({ refreshToken, clientId, scope });
      if (!oauthGrantMatchesCurrentClient(
        tokens.accessToken,
        resolveClients(options.clients),
        requestClientHint(req, clientId),
      )) {
        throw new OAuthError('invalid_grant', 'Scoped API token was rotated or removed');
      }
      const tokenDetail = `issued grant=refresh_token aud=${tokens.accessToken.audience ? safeHost(tokens.accessToken.audience) : '(none)'}`;
      logHttpRequest(requestLog, req, 200, 'oauth.token', tokenDetail, clientId);
      res.json({
        access_token: tokens.accessToken.token,
        token_type: 'Bearer',
        expires_in: oauth.tokenTtlSeconds,
        refresh_token: tokens.refreshToken.token,
        scope: tokens.accessToken.scope,
      });
    } catch (err) {
      if (err instanceof OAuthError) {
        const status = err.code === 'server_error' ? 500 : 400;
        logHttpRequest(requestLog, req, status, 'oauth.token', err.code, requestClientId);
        res.status(status).json({ error: err.code, error_description: err.message });
        return;
      }
      log.warn(`Token exchange failed: ${err instanceof Error ? err.message : 'unknown'}`);
      logHttpRequest(requestLog, req, 500, 'oauth.token', 'server_error', requestClientId);
      res.status(500).json({ error: 'server_error' });
    }
  });

  const dashboardOriginMiddleware: express.RequestHandler = (req, res, next) => {
    if (originCheck(req) || originMatchesBaseUrl(req, baseUrlFor(req)) || firstHeaderValue(req.headers['x-mvmt-transport']) === 'relay') {
      next();
      return;
    }
    logHttpRequest(requestLog, req, 403, 'dashboard.origin', 'origin_not_allowed');
    res.status(403).json({ error: 'Origin not allowed' });
  };

  const dashboardAuthMiddleware: express.RequestHandler = (req, res, next) => {
    const session = dashboardSessionForRequest(req, dashboardSessions);
    if (!session) {
      logHttpRequest(requestLog, req, 401, 'dashboard.auth', 'missing_session');
      res.status(401).json({ error: 'dashboard_login_required' });
      return;
    }
    const user = findPrivilegedUser(privilegedUsersPath, session.userId);
    if (!user || user.disabled) {
      dashboardSessions.delete(session.id);
      res.setHeader('Set-Cookie', clearDashboardSessionCookie(req));
      logHttpRequest(requestLog, req, 401, 'dashboard.auth', 'user_removed', session.username);
      res.status(401).json({ error: 'dashboard_login_required' });
      return;
    }
    const refreshedSession = { ...session, username: user.username, admin: Boolean(user.admin) };
    dashboardSessions.set(session.id, refreshedSession);
    res.locals.dashboardSession = refreshedSession;
    next();
  };

  // Admin-only endpoints (mount mutation). Non-admin
  // dashboard users can still log in, browse the configured mounts, and
  // create/revoke leases — they just can't change the mount config.
  const dashboardAdminMiddleware: express.RequestHandler = (req, res, next) => {
    const session = res.locals.dashboardSession as DashboardSession | undefined;
    if (!session || !session.admin) {
      logHttpRequest(requestLog, req, 403, 'dashboard.admin', 'not_admin', session?.username);
      res.status(403).json({ error: 'admin_required' });
      return;
    }
    next();
  };

  const dashboardLocalMiddleware: express.RequestHandler = (req, res, next) => {
    if (!isLocalRequest(req)) {
      logHttpRequest(requestLog, req, 403, 'dashboard.local', 'local_required');
      res.status(403).json({ error: 'local_dashboard_required' });
      return;
    }
    next();
  };

  app.get('/dashboard', (req, res) => {
    if (req.originalUrl.includes('?')) {
      res.setHeader('Location', 'dashboard');
      res.status(303).end();
      return;
    }
    res.status(200).type('html').send(DASHBOARD_PAGE_HTML);
  });

  app.post('/dashboard', authLimiter, dashboardOriginMiddleware, (req, res) => {
    const body = (req.body ?? {}) as Record<string, unknown>;
    const username = typeof body.username === 'string' ? body.username : undefined;
    const password = typeof body.password === 'string' ? body.password : undefined;
    const user = verifyPrivilegedUserPassword(privilegedUsersPath, username, password);
    if (!user) {
      logHttpRequest(requestLog, req, 401, 'dashboard.login', 'invalid_credentials', username);
      res.status(401).type('html').send('<!doctype html><meta charset="utf-8"><title>mvmt</title><p>Wrong username or password.</p><p><a href="dashboard">Back to sign in</a></p>');
      return;
    }
    recordPrivilegedUserLogin(privilegedUsersPath, user.id);
    const session = createDashboardSession(user.id, user.username, Boolean(user.admin));
    dashboardSessions.set(session.id, session);
    res.setHeader('Set-Cookie', dashboardSessionCookie(session, req));
    res.setHeader('Location', 'dashboard');
    logHttpRequest(requestLog, req, 303, 'dashboard.login', 'ok', user.username);
    res.status(303).end();
  });

  app.post('/dashboard/api/login', authLimiter, dashboardOriginMiddleware, (req, res) => {
    const body = (req.body ?? {}) as Record<string, unknown>;
    const username = typeof body.username === 'string' ? body.username : undefined;
    const password = typeof body.password === 'string' ? body.password : undefined;
    const user = verifyPrivilegedUserPassword(privilegedUsersPath, username, password);
    if (!user) {
      logHttpRequest(requestLog, req, 401, 'dashboard.login', 'invalid_credentials', username);
      res.status(401).json({ error: 'invalid_credentials' });
      return;
    }
    recordPrivilegedUserLogin(privilegedUsersPath, user.id);
    const session = createDashboardSession(user.id, user.username, Boolean(user.admin));
    dashboardSessions.set(session.id, session);
    res.setHeader('Set-Cookie', dashboardSessionCookie(session, req));
    logHttpRequest(requestLog, req, 200, 'dashboard.login', 'ok', user.username);
    res.json({ user: { ...publicDashboardUser(user), admin: Boolean(user.admin) }, localOwner: isLocalRequest(req) });
  });

  app.post('/dashboard/api/logout', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const session = res.locals.dashboardSession as DashboardSession;
    dashboardSessions.delete(session.id);
    res.setHeader('Set-Cookie', clearDashboardSessionCookie(req));
    logHttpRequest(requestLog, req, 200, 'dashboard.logout', 'ok', session.username);
    res.json({ ok: true });
  });

  app.get('/dashboard/api/me', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const session = res.locals.dashboardSession as DashboardSession;
    res.json({ user: { username: session.username, admin: session.admin }, localOwner: isLocalRequest(req) });
  });

  app.get('/dashboard/api/status', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, (req, res) => {
    const session = res.locals.dashboardSession as DashboardSession;
    const configuredPublicBase = resolvePublicBaseUrl?.();
    const publicDashboardUrl = configuredPublicBase ? `${configuredPublicBase.replace(/\/+$/, '')}/dashboard` : undefined;
    let tunnel: { configured: boolean; provider?: string; publicUrl?: string; canReconfigure: boolean } = {
      configured: Boolean(configuredPublicBase),
      ...(publicDashboardUrl ? { publicUrl: publicDashboardUrl } : {}),
      canReconfigure: Boolean(options.configPath) && session.admin && isLocalRequest(req),
    };
    if (options.configPath) {
      try {
        const config = readConfig(options.configPath);
        tunnel = {
          configured: config.server.access === 'tunnel' && Boolean(config.server.tunnel),
          ...(config.server.tunnel?.provider ? { provider: config.server.tunnel.provider } : {}),
          ...(publicDashboardUrl ? { publicUrl: publicDashboardUrl } : {}),
          canReconfigure: Boolean(options.configPath) && session.admin && isLocalRequest(req),
        };
      } catch {
        // Keep the runtime-only status if the config cannot be read.
      }
    }
    logHttpRequest(requestLog, req, 200, 'dashboard.status');
    res.json({
      server: {
        localUrl: `http://127.0.0.1:${port}/dashboard`,
        publicUrl: publicDashboardUrl,
      },
      tunnel,
      logs: dashboardLogs.slice(-50),
    });
  });

  app.get('/dashboard/api/mounts', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const session = res.locals.dashboardSession as DashboardSession;
    const canManage = Boolean(options.configPath) && session.admin && isLocalRequest(req);
    const mounts = resolveLeaseMounts(options.leaseMounts);
    logHttpRequest(requestLog, req, 200, 'dashboard.mounts');
    res.json({
      canManage,
      mounts: mounts.map((mount) => ({
        name: mount.name,
        path: mount.path,
        // Local filesystem paths are admin-only info; non-admins see the
        // virtual path and base permission, not the on-disk root.
        ...(canManage ? { root: mount.root } : {}),
        description: mount.description,
        writeAccess: Boolean(mount.writeAccess),
        enabled: mount.enabled !== false,
      })),
    });
  });

  app.post('/dashboard/api/local-path-picker', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const body = (req.body ?? {}) as Record<string, unknown>;
    const kind = body.kind === 'file' || body.kind === 'folder' ? body.kind : undefined;
    if (!kind) {
      res.status(400).json({ error: 'invalid_picker_kind' });
      return;
    }
    const picker = options.localPathPicker ?? pickLocalPathWithFinder;
    try {
      const picked = await picker(kind);
      if (!picked) {
        res.json({ cancelled: true });
        return;
      }
      res.json({ path: picked });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'picker_unavailable';
      logHttpRequest(requestLog, req, 400, 'dashboard.path_picker', detail);
      res.status(400).json({ error: 'picker_unavailable', detail });
    }
  });

  app.get('/api/fs/sources', dashboardOriginMiddleware, dashboardAuthMiddleware, async (req, res) => {
    const identity = readClientIdentity(req);
    const fsOptions = fsListOptionsForRequest(req, res);
    try {
      const sources = await fsSourcesForRequest(resolveLeaseMounts(options.leaseMounts), identity, fsOptions);
      logHttpRequest(requestLog, req, 200, 'fs.sources');
      res.json({ sources });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      logHttpRequest(requestLog, req, 500, 'fs.sources', detail);
      res.status(500).json({ error: 'fs_sources_unavailable' });
    }
  });

  app.get('/api/fs/list', dashboardOriginMiddleware, dashboardAuthMiddleware, async (req, res) => {
    const requestPath = firstStringQuery(req.query.path) ?? '/';
    const identity = readClientIdentity(req);
    const fsOptions = fsListOptionsForRequest(req, res);
    try {
      const listing = await fsListForRequest(resolveLeaseMounts(options.leaseMounts), requestPath, identity, fsOptions);
      logHttpRequest(requestLog, req, 200, 'fs.list', listing.path);
      res.json(listing);
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      const status = detail.includes('missing_permission') ? 403 : 404;
      logHttpRequest(requestLog, req, status, 'fs.list', detail);
      res.status(status).json({ error: status === 403 ? 'fs_permission_denied' : 'fs_path_unavailable' });
    }
  });

  app.get('/api/fs/stat', dashboardOriginMiddleware, dashboardAuthMiddleware, async (req, res) => {
    const requestPath = firstStringQuery(req.query.path);
    const identity = readClientIdentity(req);
    if (!requestPath) {
      res.status(400).json({ error: 'path_required' });
      return;
    }
    try {
      const stat = await fsStatForRequest(resolveLeaseMounts(options.leaseMounts), requestPath, identity);
      logHttpRequest(requestLog, req, 200, 'fs.stat', stat.path);
      res.json(stat);
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      const status = detail.includes('missing_permission') ? 403 : 404;
      logHttpRequest(requestLog, req, status, 'fs.stat', detail);
      res.status(status).json({ error: status === 403 ? 'fs_permission_denied' : 'fs_path_unavailable' });
    }
  });

  const fsFileHandler: express.RequestHandler = async (req, res) => {
    const requestPath = firstStringQuery(req.query.path);
    const identity = readClientIdentity(req);
    const dashboardSession = res.locals.dashboardSession as DashboardSession | undefined;
    const startedAt = Date.now();
    if (!requestPath) {
      res.status(400).json({ error: 'path_required' });
      return;
    }

    let file;
    try {
      file = await fsStatForRequest(resolveLeaseMounts(options.leaseMounts), requestPath, identity, { requireExactRead: true });
      if (file.type !== 'file') throw new Error('not_file');
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      const status = detail.includes('missing_permission') ? 403 : 404;
      logHttpRequest(requestLog, req, status, 'fs.file', detail);
      recordFsAudit(options.audit, 'read', requestPath, startedAt, true, identity, dashboardSession, detail);
      res.status(status).json({ error: status === 403 ? 'fs_permission_denied' : 'fs_path_unavailable' });
      return;
    }

    const range = parseRangeHeader(firstHeaderValue(req.headers.range), file.size);
    if (range === 'invalid') {
      res.setHeader('Content-Range', `bytes */${file.size}`);
      logHttpRequest(requestLog, req, 416, 'fs.file', 'invalid_range');
      recordFsAudit(options.audit, 'read', file.path, startedAt, true, identity, dashboardSession, 'invalid_range');
      res.status(416).end();
      return;
    }

    const start = range?.start ?? 0;
    const end = range?.end ?? Math.max(0, file.size - 1);
    const status = range ? 206 : 200;
    const etag = fsFileEtag(file);
    const lastModified = new Date(file.mtimeMs).toUTCString();
    if (fsFileNotModified(req, file, etag)) {
      res.status(304);
      setFsFileCacheHeaders(res, etag, lastModified);
      logHttpRequest(requestLog, req, 304, 'fs.file', 'cache_hit');
      recordFsAudit(options.audit, 'read', file.path, startedAt, false, identity, dashboardSession, 'cache_hit');
      res.end();
      return;
    }

    res.status(status);
    setFsFileCacheHeaders(res, etag, lastModified);
    res.setHeader('Content-Disposition', `inline; filename="${escapeHeaderValue(path.basename(file.path))}"`);
    res.setHeader('Content-Length', String(file.size === 0 ? 0 : end - start + 1));
    res.setHeader('Content-Type', contentTypeForPath(file.path));
    res.setHeader('X-Content-Type-Options', 'nosniff');
    if (range) res.setHeader('Content-Range', `bytes ${start}-${end}/${file.size}`);
    logHttpRequest(requestLog, req, status, 'fs.file', file.path);
    recordFsAudit(options.audit, 'read', file.path, startedAt, false, identity, dashboardSession);
    if (req.method === 'HEAD') {
      res.end();
      return;
    }
    let opened: StorageProviderReadStream;
    try {
      opened = await fsOpenForRequest(resolveLeaseMounts(options.leaseMounts), file.path, identity, range ? { start, end } : undefined);
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      logHttpRequest(requestLog, req, 404, 'fs.file', detail);
      recordFsAudit(options.audit, 'read', file.path, startedAt, true, identity, dashboardSession, detail);
      res.destroy();
      return;
    }
    opened.stream
      .on('error', () => {
        if (!res.headersSent) res.status(500).end();
        else res.destroy();
      })
      .pipe(res);
  };

  app.get('/api/fs/file', dashboardOriginMiddleware, dashboardAuthMiddleware, fsFileHandler);
  app.head('/api/fs/file', dashboardOriginMiddleware, dashboardAuthMiddleware, fsFileHandler);

  app.put('/api/fs/file', dashboardOriginMiddleware, dashboardAuthMiddleware, async (req, res) => {
    const requestPath = firstStringQuery(req.query.path);
    const identity = readClientIdentity(req);
    const dashboardSession = res.locals.dashboardSession as DashboardSession | undefined;
    const startedAt = Date.now();
    if (!requestPath) {
      res.status(400).json({ error: 'path_required' });
      return;
    }
    const contentLength = parseContentLength(firstHeaderValue(req.headers['content-length']));
    if (contentLength === 'invalid') {
      logHttpRequest(requestLog, req, 400, 'fs.write', 'invalid_content_length');
      recordFsAudit(options.audit, 'write', requestPath, startedAt, true, identity, dashboardSession, 'invalid_content_length');
      res.status(400).json({ error: 'invalid_content_length' });
      return;
    }
    if (contentLength !== undefined && contentLength > maxFsUploadBytes) {
      logHttpRequest(requestLog, req, 413, 'fs.write', 'upload_too_large');
      recordFsAudit(options.audit, 'write', requestPath, startedAt, true, identity, dashboardSession, 'upload_too_large');
      res.status(413).json({ error: 'fs_upload_too_large' });
      return;
    }
    try {
      const written = await fsWriteForRequest(resolveLeaseMounts(options.leaseMounts), requestPath, req, maxFsUploadBytes, identity);
      logHttpRequest(requestLog, req, 200, 'fs.write', written.path);
      recordFsAudit(options.audit, 'write', written.path, startedAt, false, identity, dashboardSession);
      res.json(written);
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'write_failed';
      const status = err instanceof FsUploadTooLargeError
        ? 413
        : detail.includes('missing_permission') || detail.includes('read-only') || detail.includes('protected') ? 403 : 400;
      logHttpRequest(requestLog, req, status, 'fs.write', detail);
      recordFsAudit(options.audit, 'write', requestPath, startedAt, true, identity, dashboardSession, detail);
      res.status(status).json({ error: status === 413 ? 'fs_upload_too_large' : status === 403 ? 'fs_permission_denied' : 'fs_write_failed' });
    }
  });

  app.delete('/api/fs/file', dashboardOriginMiddleware, dashboardAuthMiddleware, async (req, res) => {
    const requestPath = firstStringQuery(req.query.path);
    const identity = readClientIdentity(req);
    const dashboardSession = res.locals.dashboardSession as DashboardSession | undefined;
    const startedAt = Date.now();
    if (!requestPath) {
      res.status(400).json({ error: 'path_required' });
      return;
    }
    try {
      const removed = await fsRemoveForRequest(resolveLeaseMounts(options.leaseMounts), requestPath, identity);
      logHttpRequest(requestLog, req, 200, 'fs.remove', removed.path);
      recordFsAudit(options.audit, 'remove', removed.path, startedAt, false, identity, dashboardSession);
      res.json({ ...removed, removed: true });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'remove_failed';
      const status = detail.includes('missing_permission') || detail.includes('read-only') || detail.includes('protected') ? 403 : 404;
      logHttpRequest(requestLog, req, status, 'fs.remove', detail);
      recordFsAudit(options.audit, 'remove', requestPath, startedAt, true, identity, dashboardSession, detail);
      res.status(status).json({ error: status === 403 ? 'fs_permission_denied' : 'fs_remove_failed' });
    }
  });

  // First-party app registry. INSTALLED_APPS is a static list; third-party
  // app install / dynamic discovery is intentionally out of scope. Apps
  // inherit dashboard cookie auth — they are tabs in the dashboard, not a
  // separate trust tier.
  app.get('/dashboard/api/apps', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    logHttpRequest(requestLog, req, 200, 'dashboard.apps');
    res.json({
      apps: INSTALLED_APPS.map((app) => ({
        id: app.id,
        label: app.label,
        description: app.description,
      })),
    });
  });

  app.get('/apps/:appId', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const appId = firstStringQuery(req.params.appId) ?? '';
    const manifest = getApp(appId);
    if (!manifest) {
      logHttpRequest(requestLog, req, 404, 'apps.serve', `unknown_app=${appId}`);
      res.status(404).type('text/plain').send('App not found');
      return;
    }
    logHttpRequest(requestLog, req, 200, 'apps.serve', manifest.id);
    res.type('html').send(manifest.html);
  });

  const dashboardLeasePayload = (req: Request, lease: LeaseRecord): LeaseRecord & { url?: string } => {
    const secret = findLeaseSecret(leaseSecretsPath, lease.id);
    const canRecoverUrl = secret && !leaseUnavailableReason(lease) && validateLeaseToken(lease, secret.token);
    return {
      ...lease,
      ...(canRecoverUrl ? { url: leasePublicUrl(userFacingBaseUrlFor(req), lease.id, secret.token) } : {}),
    };
  };

  app.get('/dashboard/api/leases', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    logHttpRequest(requestLog, req, 200, 'dashboard.leases');
    res.json({ leases: listLeases(leaseStorePath).map((lease) => dashboardLeasePayload(req, lease)) });
  });

  app.post('/dashboard/api/leases', dashboardOriginMiddleware, dashboardAuthMiddleware, async (req, res) => {
    const body = (req.body ?? {}) as Record<string, unknown>;
    const label = typeof body.label === 'string' ? body.label.trim() : '';
    const mode = typeof body.mode === 'string' ? body.mode.trim().toLowerCase() : 'read';
    const paths = dashboardLeasePaths(body);
    const maxDownloads = dashboardMaxDownloads(body);
    if (!label) {
      res.status(400).json({ error: 'label_required' });
      return;
    }
    if (paths.length === 0) {
      res.status(400).json({ error: 'path_required' });
      return;
    }
    // Dashboard exposes three modes, mapped onto the underlying primitives:
    //   read     → ['read']           browse + download
    //   upload   → ['upload']         drop-box (no download, no overwrite)
    //   two-way  → ['read', 'upload'] browse + download + add new files
    // `write` (overwrite + delete) is intentionally not exposed by the
    // dashboard — see core/leases/files.ts for collision-suffix semantics.
    const permissions = leasePermissionsForDashboardMode(mode);
    if (!permissions) {
      res.status(400).json({ error: 'invalid_mode' });
      return;
    }
    if (maxDownloads === 'invalid') {
      res.status(400).json({ error: 'invalid_download_limit' });
      return;
    }
    if (maxDownloads !== undefined && !permissions.includes('read')) {
      res.status(400).json({ error: 'download_limit_requires_read' });
      return;
    }
    try {
      const resources = await dashboardLeaseResources(resolveLeaseMounts(options.leaseMounts), paths);
      if (mode !== 'read' && resources.some((resource) => resource.type === 'file')) {
        res.status(400).json({ error: 'mode_requires_folder' });
        return;
      }
      if (permissions.includes('upload') && !(await dashboardLeaseResourcesAreWritable(resolveLeaseMounts(options.leaseMounts), paths))) {
        res.status(400).json({ error: 'mount_read_only' });
        return;
      }
      const ttl = parseTokenTtl(typeof body.expires === 'string' ? body.expires : DEFAULT_DASHBOARD_LEASE_TTL);
      const created = createLease(leaseStorePath, {
        label,
        path: resources[0]!.sourcePath,
        resources,
        expiresAt: ttl.expiresAt,
        permissions: [...permissions],
        ...(maxDownloads === undefined ? {} : { maxDownloads }),
        // Creating a share link in the dashboard is the explicit publish
        // gesture — the lease is meant to be reachable over the relay.
        published: true,
      });
      saveLeaseSecret(leaseSecretsPath, created.record.id, created.token);
      const url = leasePublicUrl(userFacingBaseUrlFor(req), created.record.id, created.token);
      logHttpRequest(requestLog, req, 201, 'dashboard.leases', `created ${created.record.id}`);
      res.status(201).json({ lease: { ...created.record, url } });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'lease_failed';
      logHttpRequest(requestLog, req, 400, 'dashboard.leases', detail);
      res.status(400).json({ error: 'lease_failed', detail });
    }
  });

  app.post('/dashboard/api/leases/:id/revoke', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const id = firstStringQuery(req.params.id);
    if (!id || !revokeLease(leaseStorePath, id)) {
      logHttpRequest(requestLog, req, 404, 'dashboard.leases', 'unknown_lease');
      res.status(404).json({ error: 'lease_not_found' });
      return;
    }
    removeLeaseSecret(leaseSecretsPath, id);
    logHttpRequest(requestLog, req, 200, 'dashboard.leases', `revoked ${id}`);
    res.json({ ok: true });
  });

  // Toggles the exposure boundary for a lease. Unpublishing blocks public
  // tunnels without revoking the lease; local apps can still reach
  // it over 127.0.0.1. Distinct from revoke, which kills the lease.
  app.post('/dashboard/api/leases/:id/publish', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const id = firstStringQuery(req.params.id);
    const body = (req.body ?? {}) as Record<string, unknown>;
    if (!id) {
      res.status(400).json({ error: 'id_required' });
      return;
    }
    // Require an explicit boolean so a missing or malformed body cannot
    // silently unpublish; an unpublish is a real exposure change.
    if (typeof body.published !== 'boolean') {
      res.status(400).json({ error: 'published_required' });
      return;
    }
    const published = body.published;
    const lease = findLease(leaseStorePath, id);
    if (!lease) {
      logHttpRequest(requestLog, req, 404, 'dashboard.leases', 'unknown_lease');
      res.status(404).json({ error: 'lease_not_found' });
      return;
    }
    if (lease.revokedAt) {
      logHttpRequest(requestLog, req, 410, 'dashboard.leases', 'revoked_lease');
      res.status(410).json({ error: 'lease_revoked' });
      return;
    }
    const updated = setLeasePublished(leaseStorePath, id, published);
    if (!updated) {
      logHttpRequest(requestLog, req, 404, 'dashboard.leases', 'unknown_lease');
      res.status(404).json({ error: 'lease_not_found' });
      return;
    }
    logHttpRequest(requestLog, req, 200, 'dashboard.leases', `${published ? 'published' : 'unpublished'} ${id}`);
    res.json({ lease: updated });
  });

  // Rotates a lease's token, invalidating any previously-issued URL. Used
  // by the dashboard when an admin explicitly wants to replace a link whose
  // original token is not recoverable from the local lease-secret store.
  app.post('/dashboard/api/leases/:id/rotate', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const id = typeof req.params.id === 'string' ? req.params.id : '';
    if (!id) {
      res.status(400).json({ error: 'id_required' });
      return;
    }
    const lease = findLease(leaseStorePath, id);
    if (!lease) {
      logHttpRequest(requestLog, req, 404, 'dashboard.leases', 'unknown_lease');
      res.status(404).json({ error: 'lease_not_found' });
      return;
    }
    if (lease.revokedAt) {
      logHttpRequest(requestLog, req, 410, 'dashboard.leases', 'revoked_lease');
      res.status(410).json({ error: 'lease_revoked' });
      return;
    }
    const rotated = rotateLeaseToken(leaseStorePath, id);
    if (!rotated) {
      logHttpRequest(requestLog, req, 404, 'dashboard.leases', 'unknown_lease');
      res.status(404).json({ error: 'lease_not_found' });
      return;
    }
    saveLeaseSecret(leaseSecretsPath, rotated.record.id, rotated.token);
    const url = leasePublicUrl(userFacingBaseUrlFor(req), rotated.record.id, rotated.token);
    logHttpRequest(requestLog, req, 200, 'dashboard.leases', `rotated ${id}`);
    res.json({ lease: { ...rotated.record, url } });
  });

  // Mount management endpoints. These persist through the same config file
  // and lock that the CLI uses, so a privileged dashboard user can add,
  // edit, or remove mounts without local shell access. Mutation endpoints
  // require options.configPath; without it they respond 403.
  const requireConfigPath = (req: Request, res: Response): string | undefined => {
    if (!options.configPath) {
      logHttpRequest(requestLog, req, 403, 'dashboard.mounts', 'no_config_path');
      res.status(403).json({ error: 'mount_management_disabled' });
      return undefined;
    }
    return options.configPath;
  };

  app.post('/dashboard/api/mounts', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const configPath = requireConfigPath(req, res);
    if (!configPath) return;
    const body = (req.body ?? {}) as Record<string, unknown>;
    try {
      const saved = await withConfigLock(configPath, async () => {
        const current = readConfig(configPath);
        const mountInput = mountInputFromBody(body, current.mounts);
        if (typeof mountInput === 'string') throw new Error(mountInput);
        await assertMountRootUsable(mountInput.root);
        const next = addMountToConfig(current, mountInput);
        await saveConfig(configPath, next);
        return next.mounts.find((mount) => mount.name === mountInput.name);
      });
      if (!saved) throw new Error('mount_failed');
      logHttpRequest(requestLog, req, 201, 'dashboard.mounts', `added ${saved.name}`);
      res.status(201).json({ mount: saved });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'mount_failed';
      const knownValidation = new Set(['root_required', 'invalid_name', 'invalid_root', 'invalid_path']);
      const error = knownValidation.has(detail) ? detail : 'mount_failed';
      logHttpRequest(requestLog, req, 400, 'dashboard.mounts', detail);
      res.status(400).json({ error, detail });
    }
  });

  app.patch('/dashboard/api/mounts/:name', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const configPath = requireConfigPath(req, res);
    if (!configPath) return;
    const name = typeof req.params.name === 'string' ? req.params.name : '';
    if (!name) {
      res.status(400).json({ error: 'name_required' });
      return;
    }
    const body = (req.body ?? {}) as Record<string, unknown>;
    const patch = mountPatchFromBody(body);
    if (typeof patch === 'string') {
      logHttpRequest(requestLog, req, 400, 'dashboard.mounts', patch);
      res.status(400).json({ error: patch });
      return;
    }
    try {
      const saved = await withConfigLock(configPath, async () => {
        const current = readConfig(configPath);
        if (patch.root) await assertMountRootUsable(patch.root);
        const next = editMountInConfig(current, name, patch);
        await saveConfig(configPath, next);
        return next.mounts.find((mount) => mount.name === name);
      });
      logHttpRequest(requestLog, req, 200, 'dashboard.mounts', `edited ${name}`);
      res.json({ mount: saved });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'mount_failed';
      logHttpRequest(requestLog, req, 400, 'dashboard.mounts', detail);
      const knownValidation = new Set(['invalid_root', 'invalid_path']);
      res.status(400).json({ error: knownValidation.has(detail) ? detail : 'mount_failed', detail });
    }
  });

  app.delete('/dashboard/api/mounts/:name', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const configPath = requireConfigPath(req, res);
    if (!configPath) return;
    const name = typeof req.params.name === 'string' ? req.params.name : '';
    if (!name) {
      res.status(400).json({ error: 'name_required' });
      return;
    }
    try {
      await withConfigLock(configPath, async () => {
        const current = readConfig(configPath);
        const next = removeMountFromConfig(current, name);
        await saveConfig(configPath, next);
      });
      logHttpRequest(requestLog, req, 200, 'dashboard.mounts', `removed ${name}`);
      res.json({ ok: true });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'mount_failed';
      logHttpRequest(requestLog, req, 400, 'dashboard.mounts', detail);
      res.status(400).json({ error: 'mount_failed', detail });
    }
  });

  // MCP-access grants — clients[] API tokens the owner mints from the
  // dashboard's MCP tab. A grant is either "all mounts" (one /** scope
  // that follows the mount list) or a per-mount selection. Mutating
  // endpoints are admin + local-only and need options.configPath; the
  // list is readable by any signed-in dashboard user so a non-admin can
  // see what exists without a silent dead-end.
  const grantScopeSummary = (client: ClientConfig): string => {
    const grant = clientConfigToGrant(client);
    if (grant.scope.length === 1 && grant.scope[0]!.path === '/**') return 'All mounts';
    if (grant.scope.length === 0) return 'No access';
    return grant.scope.length === 1 ? '1 mount' : `${grant.scope.length} mounts`;
  };

  app.get('/dashboard/api/grants', dashboardOriginMiddleware, dashboardAuthMiddleware, (req, res) => {
    const session = res.locals.dashboardSession as DashboardSession;
    const tokenClients = resolveClients(options.clients).filter((client) => client.auth.type === 'token');
    logHttpRequest(requestLog, req, 200, 'dashboard.grants');
    res.json({
      // Minting grants writes to config and is local-only (see the
      // dashboardLocalMiddleware on POST /dashboard/api/grants), so the
      // local check must be part of canManage; otherwise a remote admin
      // sees a "New MCP token" button whose POST is then blocked.
      canManage: Boolean(options.configPath) && session.admin && isLocalRequest(req),
      grants: tokenClients.map((client) => ({
        id: client.id,
        label: client.name,
        scope: grantScopeSummary(client),
        published: isGrantPublished(client.published),
        expiresAt: client.expiresAt,
        lastUsedAt: client.lastUsedAt,
      })),
    });
  });

  // Mint an MCP-access grant. `allMounts: true` stores a single /** scope
  // that tracks the mount list; otherwise `scopes` is a per-mount list,
  // each entry resolved against the configured mounts before it becomes
  // a permission entry — a scope that does not land inside a mount is
  // rejected, so a grant can never reach outside the namespace.
  app.post('/dashboard/api/grants', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const configPath = requireConfigPath(req, res);
    if (!configPath) return;
    const body = (req.body ?? {}) as Record<string, unknown>;
    const label = typeof body.label === 'string' ? body.label.trim() : '';
    const published = body.published === true;
    const allMounts = body.allMounts === true;
    const rawScopes = Array.isArray(body.scopes) ? body.scopes : [];
    if (!label) {
      res.status(400).json({ error: 'label_required' });
      return;
    }
    if (!allMounts && rawScopes.length === 0) {
      res.status(400).json({ error: 'scope_required' });
      return;
    }
    try {
      let permissions: PermissionConfig[];
      if (allMounts) {
        // One /** scope. Mount-level writeAccess still gates real writes,
        // and the grant automatically follows mounts added later.
        permissions = [{ path: '/**', actions: ['search', 'read', 'write'] }];
      } else {
        const mounts = resolveLeaseMounts(options.leaseMounts);
        permissions = [];
        for (const entry of rawScopes) {
          const scope = (entry ?? {}) as Record<string, unknown>;
          const scopePath = typeof scope.path === 'string' ? scope.path.trim() : '';
          const mode = scope.mode === 'write' ? 'write' : 'read';
          if (!scopePath) {
            res.status(400).json({ error: 'scope_path_required' });
            return;
          }
          const target = await resolveDashboardFileTarget(mounts, scopePath);
          if (mode === 'write' && !target.writeAccess) {
            res.status(400).json({ error: 'mount_read_only', detail: target.virtualPath });
            return;
          }
          permissions.push({
            path: target.type === 'file'
              ? target.virtualPath
              : `${stripTrailingSlashes(target.virtualPath)}/**`,
            actions: mode === 'write' ? ['search', 'read', 'write'] : ['search', 'read'],
          });
        }
      }
      const expires = typeof body.expires === 'string' && body.expires ? body.expires : undefined;
      const grantId = `grant-${randomUUID().slice(0, 8)}`;
      const saved = await withConfigLock(configPath, async () => {
        const current = readConfig(configPath);
        const update = addApiTokenToConfig(current, {
          id: grantId,
          name: label,
          // permissions is ignored when resolvedPermissions is present.
          permissions: [],
          resolvedPermissions: permissions,
          ...(expires ? { expires } : {}),
          published,
        });
        await saveConfig(configPath, update.config);
        return update;
      });
      logHttpRequest(requestLog, req, 201, 'dashboard.grants', `created ${saved.client.id}`);
      res.status(201).json({
        grant: { id: saved.client.id, label: saved.client.name, published },
        token: saved.plaintextToken,
        endpoint: `${userFacingBaseUrlFor(req)}/mcp`,
      });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'grant_failed';
      logHttpRequest(requestLog, req, 400, 'dashboard.grants', detail);
      res.status(400).json({ error: 'grant_failed', detail });
    }
  });

  // Toggle a grant's public-tunnel exposure. Unpublishing keeps the grant
  // usable by local apps over 127.0.0.1; it just loses remote reachability.
  app.post('/dashboard/api/grants/:id/publish', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const configPath = requireConfigPath(req, res);
    if (!configPath) return;
    const id = typeof req.params.id === 'string' ? req.params.id : '';
    const body = (req.body ?? {}) as Record<string, unknown>;
    // Require an explicit boolean so a missing or malformed body cannot
    // silently unpublish; an unpublish is a real exposure change.
    if (typeof body.published !== 'boolean') {
      res.status(400).json({ error: 'published_required' });
      return;
    }
    const published = body.published;
    try {
      const saved = await withConfigLock(configPath, async () => {
        const current = readConfig(configPath);
        const update = setApiTokenPublishedInConfig(current, id, published);
        await saveConfig(configPath, update.config);
        return update;
      });
      logHttpRequest(requestLog, req, 200, 'dashboard.grants', `${published ? 'published' : 'unpublished'} ${id}`);
      res.json({ grant: { id: saved.client.id, label: saved.client.name, published } });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'grant_not_found';
      logHttpRequest(requestLog, req, 404, 'dashboard.grants', detail);
      res.status(404).json({ error: 'grant_not_found', detail });
    }
  });

  // Revoke a grant: removes the clients[] entry entirely. Distinct from
  // unpublishing, which only blocks public tunnels.
  app.delete('/dashboard/api/grants/:id', dashboardOriginMiddleware, dashboardAuthMiddleware, dashboardLocalMiddleware, dashboardAdminMiddleware, async (req, res) => {
    const configPath = requireConfigPath(req, res);
    if (!configPath) return;
    const id = typeof req.params.id === 'string' ? req.params.id : '';
    try {
      await withConfigLock(configPath, async () => {
        const current = readConfig(configPath);
        const next = removeApiTokenFromConfig(current, id);
        await saveConfig(configPath, next);
      });
      logHttpRequest(requestLog, req, 200, 'dashboard.grants', `revoked ${id}`);
      res.json({ ok: true });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'grant_not_found';
      logHttpRequest(requestLog, req, 404, 'dashboard.grants', detail);
      res.status(404).json({ error: 'grant_not_found', detail });
    }
  });

  const authorizeLeaseRequest = (req: Request, res: Response): ReturnType<typeof findLease> => {
    const id = regexParam(req, 0) ?? firstStringQuery(req.params.id);
    const lease = id ? findLease(leaseStorePath, id) : undefined;
    if (!lease) {
      logHttpRequest(requestLog, req, 404, 'lease.request', 'unknown_lease');
      res.status(404).json({ error: 'lease_not_found' });
      return undefined;
    }

    const token = bearerToken(req) ?? firstStringQuery(req.query.token) ?? firstStringQuery(req.query.t);
    if (!validateLeaseToken(lease, token)) {
      logHttpRequest(requestLog, req, 401, 'lease.request', 'invalid_token', lease.id);
      res.status(401).json({ error: 'invalid_lease_token' });
      return undefined;
    }

    const unavailable = leaseUnavailableReason(lease);
    if (unavailable) {
      logHttpRequest(requestLog, req, 410, 'lease.request', unavailable, lease.id);
      res.status(410).json({ error: `lease_${unavailable}` });
      return undefined;
    }

    // Exposure boundary: a capability-only lease (explicitly published:
    // false) is reachable only by apps on this machine. Any request
    // through a public tunnel - the mvmt relay, Cloudflare, pinggy,
    // localhost.run, a custom tunnel - is rejected. Leases without an
    // explicit published value are grandfathered as published.
    if (!isLocalRequest(req) && !isGrantPublished(lease.published)) {
      logHttpRequest(requestLog, req, 403, 'lease.request', 'not_published', lease.id);
      res.status(403).json({ error: 'lease_not_published' });
      return undefined;
    }
    return lease;
  };

  const leasePageHandler: express.RequestHandler = async (req, res) => {
    const lease = authorizeLeaseRequest(req, res);
    if (!lease) return;
    if (!leaseAllows(lease, 'read')) {
      if (leaseAllows(lease, 'upload')) {
        recordLeaseUse(leaseStorePath, lease.id);
        logHttpRequest(requestLog, req, 200, 'lease.request', 'upload_page', lease.id);
        res.status(200).type('html').send(LEASE_UPLOAD_PAGE_HTML);
        return;
      }
      logHttpRequest(requestLog, req, 403, 'lease.request', 'permission_denied', lease.id);
      res.status(403).json({ error: 'lease_permission_denied' });
      return;
    }
    recordLeaseUse(leaseStorePath, lease.id);
    logHttpRequest(requestLog, req, 200, 'lease.request', 'browser_page', lease.id);
    res.status(200).type('html').send(LEASE_BROWSER_PAGE_HTML);
  };

  const leaseFilesHandler: express.RequestHandler = async (req, res) => {
    const lease = authorizeLeaseRequest(req, res);
    if (!lease) return;
    if (!leaseAllows(lease, 'read')) {
      logHttpRequest(requestLog, req, 403, 'lease.request', 'permission_denied', lease.id);
      res.status(403).json({ error: 'lease_permission_denied' });
      return;
    }
    const requestPath = regexParam(req, 1) ?? firstStringQuery(req.query.path) ?? '';

    try {
      const listing = await listLeaseDirectory(resolveLeaseMounts(options.leaseMounts), lease, requestPath);
      recordLeaseUse(leaseStorePath, lease.id);
      logHttpRequest(requestLog, req, 200, 'lease.request', listing.path, lease.id);
      res.status(200).json({
        ...listing,
        canUpload: leaseAllows(lease, 'upload') || leaseAllows(lease, 'write'),
      });
      return;
    } catch {
      // If it is not a directory, try serving it as a file below.
    }

    let target;
    try {
      target = await resolveLeaseFileTarget(resolveLeaseMounts(options.leaseMounts), lease, requestPath);
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      logHttpRequest(requestLog, req, 404, 'lease.request', detail, lease.id);
      res.status(404).json({ error: 'lease_target_unavailable' });
      return;
    }

    const range = parseRangeHeader(firstHeaderValue(req.headers.range), target.size);
    if (range === 'invalid') {
      res.setHeader('Content-Range', `bytes */${target.size}`);
      logHttpRequest(requestLog, req, 416, 'lease.request', 'invalid_range', lease.id);
      res.status(416).end();
      return;
    }

    const start = range?.start ?? 0;
    const end = range?.end ?? Math.max(0, target.size - 1);
    const status = range ? 206 : 200;
    if (req.method !== 'HEAD') {
      const reserved = reserveLeaseDownload(leaseStorePath, lease.id);
      if (reserved === 'download_limit_reached') {
        logHttpRequest(requestLog, req, 410, 'lease.request', 'download_limit_reached', lease.id);
        res.status(410).json({ error: 'lease_download_limit_reached' });
        return;
      }
      if (!reserved) {
        logHttpRequest(requestLog, req, 404, 'lease.request', 'unknown_lease', lease.id);
        res.status(404).json({ error: 'lease_not_found' });
        return;
      }
    }
    res.status(status);
    res.setHeader('Accept-Ranges', 'bytes');
    res.setHeader('Cache-Control', 'no-store');
    res.setHeader('Content-Disposition', `attachment; filename="${escapeHeaderValue(target.filename)}"`);
    res.setHeader('Content-Length', String(target.size === 0 ? 0 : end - start + 1));
    res.setHeader('Content-Type', 'application/octet-stream');
    res.setHeader('Last-Modified', new Date(target.mtimeMs).toUTCString());
    res.setHeader('Referrer-Policy', 'no-referrer');
    res.setHeader('X-Content-Type-Options', 'nosniff');
    if (range) res.setHeader('Content-Range', `bytes ${start}-${end}/${target.size}`);

    logHttpRequest(requestLog, req, status, 'lease.request', target.leaseRelativePath || '/', lease.id);
    if (req.method === 'HEAD') {
      recordLeaseUse(leaseStorePath, lease.id);
      res.end();
      return;
    }
    createReadStream(target.realPath, target.size === 0 ? {} : { start, end })
      .on('error', () => {
        if (!res.headersSent) res.status(500).end();
        else res.destroy();
      })
      .pipe(res);
  };

  const leaseUploadHandler: express.RequestHandler = async (req, res) => {
    const lease = authorizeLeaseRequest(req, res);
    if (!lease) return;
    const allowOverwrite = leaseAllows(lease, 'write');
    const allowUpload = leaseAllows(lease, 'upload');
    if (!allowUpload && !allowOverwrite) {
      logHttpRequest(requestLog, req, 403, 'lease.request', 'permission_denied', lease.id);
      res.status(403).json({ error: 'lease_permission_denied' });
      return;
    }
    // Drop-box semantics: when the lease has upload but not write, the
    // recipient cannot overwrite an existing file. On a name collision we
    // suffix the upload (`note.txt` → `note (2).txt`) so two recipients
    // dropping the same filename never clobber each other.
    const suffixOnCollision = allowUpload && !allowOverwrite;

    const contentLength = parseContentLength(firstHeaderValue(req.headers['content-length']));
    if (contentLength === 'invalid') {
      logHttpRequest(requestLog, req, 400, 'lease.upload', 'invalid_content_length', lease.id);
      res.status(400).json({ error: 'invalid_content_length' });
      return;
    }
    if (contentLength !== undefined && contentLength > DEFAULT_MAX_LEASE_UPLOAD_BYTES) {
      logHttpRequest(requestLog, req, 413, 'lease.upload', 'upload_too_large', lease.id);
      res.status(413).json({ error: 'lease_upload_too_large' });
      return;
    }

    const requestPath = regexParam(req, 1) ?? firstStringQuery(req.query.path) ?? '';
    let target;
    try {
      target = await resolveLeaseUploadTarget(resolveLeaseMounts(options.leaseMounts), lease, requestPath, { allowOverwrite, suffixOnCollision });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      const exists = detail.includes('already exists');
      logHttpRequest(requestLog, req, exists ? 409 : 404, 'lease.upload', detail, lease.id);
      res.status(exists ? 409 : 404).json({ error: exists ? 'lease_upload_exists' : 'lease_target_unavailable' });
      return;
    }

    try {
      const bytes = await writeLeaseUpload(req, target.parentRealPath, target.realPath, DEFAULT_MAX_LEASE_UPLOAD_BYTES, { overwrite: allowOverwrite });
      recordLeaseUse(leaseStorePath, lease.id, { uploaded: true });
      const status = allowOverwrite ? 200 : 201;
      const savedPath = target.leaseRelativePath ? `/${target.leaseRelativePath}` : '/';
      logHttpRequest(requestLog, req, status, 'lease.upload', target.leaseRelativePath, lease.id);
      res.status(status).json({ path: savedPath, filename: target.filename, bytes });
    } catch (err) {
      const status = err instanceof LeaseUploadTooLargeError ? 413 : isNodeErrorCode(err, 'EEXIST') ? 409 : 500;
      const error = status === 413 ? 'lease_upload_too_large' : status === 409 ? 'lease_upload_exists' : 'lease_upload_failed';
      const detail = err instanceof Error ? err.message : error;
      logHttpRequest(requestLog, req, status, 'lease.upload', detail, lease.id);
      res.status(status).json({ error });
    }
  };

  const leaseDeleteHandler: express.RequestHandler = async (req, res) => {
    const lease = authorizeLeaseRequest(req, res);
    if (!lease) return;
    if (!leaseAllows(lease, 'write')) {
      logHttpRequest(requestLog, req, 403, 'lease.delete', 'permission_denied', lease.id);
      res.status(403).json({ error: 'lease_permission_denied' });
      return;
    }
    const requestPath = regexParam(req, 1) ?? firstStringQuery(req.query.path) ?? '';
    let target;
    try {
      target = await resolveLeaseFileTarget(resolveLeaseMounts(options.leaseMounts), lease, requestPath);
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'unavailable';
      logHttpRequest(requestLog, req, 404, 'lease.delete', detail, lease.id);
      res.status(404).json({ error: 'lease_target_unavailable' });
      return;
    }
    try {
      await fsp.unlink(target.realPath);
      logHttpRequest(requestLog, req, 200, 'lease.delete', target.leaseRelativePath || '/', lease.id);
      res.json({ ok: true });
    } catch (err) {
      const detail = err instanceof Error ? err.message : 'delete_failed';
      logHttpRequest(requestLog, req, 500, 'lease.delete', detail, lease.id);
      res.status(500).json({ error: 'lease_delete_failed' });
    }
  };

  const leaseFilesRoute = /^\/lease\/([^/]+)\/files(?:\/(.*))?$/;
  app.get('/lease/:id', mcpLimiter, leasePageHandler);
  app.get(leaseFilesRoute, mcpLimiter, leaseFilesHandler);
  app.head(leaseFilesRoute, mcpLimiter, leaseFilesHandler);
  app.put(leaseFilesRoute, mcpLimiter, leaseUploadHandler);
  app.delete(leaseFilesRoute, mcpLimiter, leaseDeleteHandler);

  app.post('/mcp', mcpLimiter, authMiddleware, async (req, res) => {
    await handleMcpRequest(req, res, router, sessions);
    logHttpRequest(requestLog, req, res.statusCode, 'mcp.request');
  });

  app.get('/mcp', mcpLimiter, authMiddleware, async (req, res) => {
    await handleMcpRequest(req, res, router, sessions);
    logHttpRequest(requestLog, req, res.statusCode, 'mcp.request');
  });

  app.delete('/mcp', mcpLimiter, authMiddleware, async (req, res) => {
    const sessionId = getSessionId(req);
    if (!sessionId || !sessions.has(sessionId)) {
      logHttpRequest(requestLog, req, 404, 'mcp.request', 'session_not_found');
      res.status(404).json({ error: 'Session not found' });
      return;
    }

    const session = sessions.get(sessionId)!;
    await session.transport.close();
    sessions.delete(sessionId);
    logHttpRequest(requestLog, req, 200, 'mcp.request');
    res.status(200).json({ ok: true });
  });

  app.get('/health', healthLimiter, authMiddleware, (_req, res) => {
    logHttpRequest(requestLog, _req, 200, 'health.request');
    res.json({
      status: 'ok',
      uptime: Math.floor(process.uptime()),
      tools: router.getAllTools(readClientIdentity(_req)).length,
      sessions: sessions.size,
    });
  });

  let httpServer: HttpServer;
  try {
    httpServer = await new Promise<HttpServer>((resolve, reject) => {
      const server = app.listen(port, '127.0.0.1', () => {
        resolve(server);
      });
      server.on('error', reject);
    });
  } catch (err) {
    clearInterval(oauthCleanup);
    clearInterval(cleanupInterval);
    throw err;
  }

  const address = httpServer.address();
  const actualPort = typeof address === 'object' && address !== null ? address.port : port;

  let closed = false;
  return {
    port: actualPort,
    close: async () => {
      if (closed) return;
      closed = true;

      clearInterval(oauthCleanup);
      clearInterval(cleanupInterval);

      const activeSessions = [...sessions.values()];
      sessions.clear();
      await Promise.allSettled(activeSessions.map((session) => session.transport.close()));
      await closeHttpServer(httpServer);
    },
  };
}

async function handleMcpRequest(
  req: Request,
  res: Response,
  router: ToolRouter,
  sessions: Map<string, McpSession>,
): Promise<void> {
  const sessionId = getSessionId(req);

  if (sessionId && sessions.has(sessionId)) {
    const session = sessions.get(sessionId)!;
    const requestIdentity = readClientIdentity(req);
    if (!sameClientIdentity(session.clientIdentity, requestIdentity)) {
      res.status(403).json({ error: 'mcp_session_client_mismatch' });
      return;
    }
    session.clientIdentity = requestIdentity;
    session.clientIdentityRef.current = requestIdentity;
    session.lastActivity = Date.now();
    if (isStandaloneSseRequest(req)) {
      session.transport.closeStandaloneSSEStream();
    }
    await session.transport.handleRequest(req, res, req.body);
    return;
  }

  if (sessionId) {
    await handleStatelessMcpRequest(req, res, router);
    return;
  }

  const transport = new StreamableHTTPServerTransport({
    sessionIdGenerator: () => randomUUID(),
  });
  const clientIdentityRef: ClientIdentityRef = { current: readClientIdentity(req) };
  const server = createMcpServer(router, () => clientIdentityRef.current);

  transport.onerror = (error) => {
    if (isBenignDuplicateSseConflict(error)) {
      log.debug(`MCP transport notice: ${error.message}`);
      return;
    }
    log.warn(`MCP transport error: ${error.message}`);
  };

  await server.connect(transport);

  transport.onclose = () => {
    if (transport.sessionId) sessions.delete(transport.sessionId);
  };

  await transport.handleRequest(req, res, req.body);

  if (transport.sessionId) {
    sessions.set(transport.sessionId, {
      transport,
      server,
      clientIdentityRef,
      clientIdentity: clientIdentityRef.current,
      lastActivity: Date.now(),
    });
  }
}

async function handleStatelessMcpRequest(req: Request, res: Response, router: ToolRouter): Promise<void> {
  const transport = new StreamableHTTPServerTransport({
    sessionIdGenerator: undefined,
  });
  const server = createMcpServer(router, readClientIdentity(req));

  transport.onerror = (error) => {
    if (isBenignDuplicateSseConflict(error)) {
      log.debug(`MCP transport notice: ${error.message}`);
      return;
    }
    log.warn(`MCP transport error: ${error.message}`);
  };

  await server.connect(transport);
  await transport.handleRequest(req, res, req.body);
}

function sameClientIdentity(left: ClientIdentity | undefined, right: ClientIdentity | undefined): boolean {
  return left?.id === right?.id;
}

function authLogKind(req: Request): string {
  if (req.path === '/mcp') return 'mcp.auth';
  if (req.path === '/health') return 'health.auth';
  return 'http.auth';
}

type FsEntryPayload = StorageProviderStat & {
  name: string;
  writeAccess: boolean;
  unavailable?: boolean;
};

type FsStatPayload = StorageProviderStat & {
  writeAccess: boolean;
};

type FsListingPayload = FsStatPayload & {
  entries: FsEntryPayload[];
};

// includeUnavailable surfaces mounts whose root path does not resolve (deleted,
// permission error, unmounted disk) as entries flagged with `unavailable: true`.
// Strictly an operator-facing diagnostic — callers should gate it on
// local-admin context. The unauthenticated/remote/non-admin path keeps the
// strict behavior of silently omitting broken mounts.
interface FsListOptions {
  includeUnavailable?: boolean;
}

async function fsSourcesForRequest(
  mounts: readonly LocalFolderMountConfig[],
  identity?: ClientIdentity,
  options: FsListOptions = {},
): Promise<FsEntryPayload[]> {
  const registry = new MountRegistry([...mounts]);
  const sources: FsEntryPayload[] = [];
  for (const mount of registry.mounts()) {
    if (!pathMayExposeEntry(mount.config.path, 'read', identity)) continue;
    try {
      const provider = fsProviderForMount(mount);
      const stat = await provider.stat('');
      sources.push(fsEntryPayload(stat, mount, identity, mount.config.name));
    } catch {
      if (options.includeUnavailable) {
        sources.push(unavailableMountEntry(mount.config));
      }
    }
  }
  return sources.sort((a, b) => a.path.localeCompare(b.path));
}

async function fsListForRequest(
  mounts: readonly LocalFolderMountConfig[],
  inputPath: string,
  identity?: ClientIdentity,
  options: FsListOptions = {},
): Promise<FsListingPayload> {
  const normalizedPath = normalizeVirtualPath(inputPath);
  if (normalizedPath === '/') {
    return {
      mount: '',
      path: '/',
      relativePath: '',
      type: 'directory',
      size: 0,
      mtimeMs: 0,
      writeAccess: false,
      entries: await fsSourcesForRequest(mounts, identity, options),
    };
  }

  if (!pathMayExposeEntry(normalizedPath, 'read', identity)) throw new Error('missing_permission');
  const target = fsTargetForPath(mounts, normalizedPath);
  const stat = await target.provider.stat(target.relativePath);
  if (stat.type === 'file' && !pathAllowed(stat.path, 'read', identity)) throw new Error('missing_permission');
  const entries = stat.type === 'directory'
    ? (await target.provider.list(target.relativePath))
        .filter((entry) => pathMayExposeEntry(entry.path, 'read', identity))
        .map((entry) => fsEntryPayload(entry, target.mount, identity))
    : [];
  return {
    ...fsStatPayload(stat, target.mount, identity),
    entries,
  };
}

async function fsStatForRequest(
  mounts: readonly LocalFolderMountConfig[],
  inputPath: string,
  identity?: ClientIdentity,
  options: { requireExactRead?: boolean } = {},
): Promise<FsStatPayload> {
  const normalizedPath = normalizeVirtualPath(inputPath);
  if (!pathMayExposeEntry(normalizedPath, 'read', identity)) throw new Error('missing_permission');
  const target = fsTargetForPath(mounts, normalizedPath);
  const stat = await target.provider.stat(target.relativePath);
  if ((options.requireExactRead || stat.type === 'file') && !pathAllowed(stat.path, 'read', identity)) {
    throw new Error('missing_permission');
  }
  return fsStatPayload(stat, target.mount, identity);
}

async function fsOpenForRequest(
  mounts: readonly LocalFolderMountConfig[],
  inputPath: string,
  identity?: ClientIdentity,
  range?: { start?: number; end?: number },
): Promise<StorageProviderReadStream> {
  const normalizedPath = normalizeVirtualPath(inputPath);
  if (!pathAllowed(normalizedPath, 'read', identity)) throw new Error('missing_permission');
  const target = fsTargetForPath(mounts, normalizedPath);
  return target.provider.openReadStream(target.relativePath, range);
}

async function fsWriteForRequest(
  mounts: readonly LocalFolderMountConfig[],
  inputPath: string,
  input: NodeJS.ReadableStream,
  maxBytes: number,
  identity?: ClientIdentity,
): Promise<FsStatPayload> {
  const normalizedPath = normalizeVirtualPath(inputPath);
  if (!pathAllowed(normalizedPath, 'write', identity)) throw new Error('missing_permission');
  const target = fsTargetForPath(mounts, normalizedPath);
  const written = await target.provider.writeStream(target.relativePath, limitFsUploadStream(input, maxBytes));
  return fsStatPayload(written, target.mount, identity);
}

async function fsRemoveForRequest(
  mounts: readonly LocalFolderMountConfig[],
  inputPath: string,
  identity?: ClientIdentity,
): Promise<{ mount: string; path: string }> {
  const normalizedPath = normalizeVirtualPath(inputPath);
  if (!pathAllowed(normalizedPath, 'write', identity)) throw new Error('missing_permission');
  const target = fsTargetForPath(mounts, normalizedPath);
  await target.provider.remove(target.relativePath);
  return { mount: target.mount.config.name, path: target.virtualPath };
}

function fsTargetForPath(mounts: readonly LocalFolderMountConfig[], inputPath: string): {
  mount: RegisteredMount;
  relativePath: string;
  virtualPath: string;
  provider: LocalFolderStorageProvider;
} {
  const resolved = new MountRegistry([...mounts]).resolve(inputPath);
  return {
    mount: resolved.mount,
    relativePath: resolved.relativePath,
    virtualPath: resolved.virtualPath,
    provider: fsProviderForMount(resolved.mount),
  };
}

function fsProviderForMount(mount: RegisteredMount): LocalFolderStorageProvider {
  return new LocalFolderStorageProvider(mount, { isTextPath, maxTextBytes: MAX_TEXT_BYTES });
}

function fsEntryPayload(
  stat: StorageProviderStat,
  mount: RegisteredMount,
  identity?: ClientIdentity,
  name = path.basename(stat.path),
): FsEntryPayload {
  return {
    ...stat,
    name,
    writeAccess: Boolean(mount.config.writeAccess) && pathAllowed(stat.path, 'write', identity),
  };
}

function fsStatPayload(stat: StorageProviderStat, mount: RegisteredMount, identity?: ClientIdentity): FsStatPayload {
  return {
    ...stat,
    writeAccess: Boolean(mount.config.writeAccess) && pathAllowed(stat.path, 'write', identity),
  };
}

function fsFileEtag(file: Pick<FsStatPayload, 'path' | 'size' | 'mtimeMs'>): string {
  const digest = createHash('sha256')
    .update(file.path)
    .update('\0')
    .update(String(file.size))
    .update('\0')
    .update(String(Math.trunc(file.mtimeMs)))
    .digest('hex')
    .slice(0, 24);
  return `W/"${digest}"`;
}

function setFsFileCacheHeaders(res: Response, etag: string, lastModified: string): void {
  res.setHeader('Accept-Ranges', 'bytes');
  res.setHeader('Cache-Control', 'private, max-age=300, must-revalidate');
  res.setHeader('ETag', etag);
  res.setHeader('Last-Modified', lastModified);
  res.setHeader('Vary', 'Range');
}

function fsFileNotModified(req: Request, file: Pick<FsStatPayload, 'mtimeMs'>, etag: string): boolean {
  const ifNoneMatch = firstHeaderValue(req.headers['if-none-match']);
  if (ifNoneMatch) return etagHeaderMatches(ifNoneMatch, etag);

  const ifModifiedSince = firstHeaderValue(req.headers['if-modified-since']);
  if (!ifModifiedSince) return false;
  const sinceMs = Date.parse(ifModifiedSince);
  if (!Number.isFinite(sinceMs)) return false;
  return Math.floor(file.mtimeMs / 1000) * 1000 <= sinceMs;
}

function etagHeaderMatches(headerValue: string, etag: string): boolean {
  const expected = normalizeEntityTag(etag);
  return headerValue
    .split(',')
    .map((value) => value.trim())
    .some((value) => value === '*' || normalizeEntityTag(value) === expected);
}

function normalizeEntityTag(value: string): string {
  return value.replace(/^W\//i, '').trim();
}

function fsListOptionsForRequest(req: Request, res: Response): FsListOptions {
  const session = res.locals.dashboardSession as DashboardSession | undefined;
  return { includeUnavailable: Boolean(session?.admin) && isLocalRequest(req) };
}

function unavailableMountEntry(mount: LocalFolderMountConfig): FsEntryPayload {
  return {
    mount: mount.name,
    name: mount.path.split('/').filter(Boolean).join('/') || mount.name,
    path: mount.path,
    relativePath: '',
    type: 'directory',
    size: 0,
    mtimeMs: 0,
    writeAccess: Boolean(mount.writeAccess),
    unavailable: true,
  };
}

function recordFsAudit(
  audit: AuditLogger | undefined,
  action: 'read' | 'write' | 'remove',
  inputPath: string,
  startedAt: number,
  isError: boolean,
  identity?: ClientIdentity,
  dashboardSession?: DashboardSession,
  deniedReason?: string,
): void {
  if (!audit) return;
  const { argKeys, argPreview } = summarizeArgs({ path: inputPath });
  const clientId = identity?.id ?? dashboardSession?.username;
  audit.record({
    ts: new Date().toISOString(),
    ...(identity && (identity.source === 'token' || identity.source === 'oauth')
      ? {
          event: 'token.use' as const,
          name: identity.id,
          result: isError ? 'error' as const : 'success' as const,
        }
      : {}),
    connectorId: 'mvmt',
    tool: `fs.${action}`,
    ...(clientId ? { clientId } : {}),
    argKeys,
    argPreview,
    isError,
    ...(deniedReason ? { deniedReason } : {}),
    durationMs: Date.now() - startedAt,
  });
}

function contentTypeForPath(inputPath: string): string {
  switch (path.extname(inputPath).toLowerCase()) {
    case '.avif':
      return 'image/avif';
    case '.gif':
      return 'image/gif';
    case '.heic':
      return 'image/heic';
    case '.heif':
      return 'image/heif';
    case '.jpg':
    case '.jpeg':
      return 'image/jpeg';
    case '.png':
      return 'image/png';
    case '.webp':
      return 'image/webp';
    case '.mov':
      return 'video/quicktime';
    case '.mp4':
      return 'video/mp4';
    case '.pdf':
      return 'application/pdf';
    case '.json':
      return 'application/json; charset=utf-8';
    case '.md':
    case '.text':
    case '.txt':
      return 'text/plain; charset=utf-8';
    default:
      return 'application/octet-stream';
  }
}

function resolveAllowLegacyDefaultClient(value: HttpServerOptions['allowLegacyDefaultClient']): boolean | undefined {
  return typeof value === 'function' ? value() : value;
}

function resolveClients(value: HttpServerOptions['clients']): readonly ClientConfig[] {
  return (typeof value === 'function' ? value() : value) ?? [];
}

function resolveLeaseMounts(value: HttpServerOptions['leaseMounts']): readonly LocalFolderMountConfig[] {
  return (typeof value === 'function' ? value() : value) ?? [];
}

function identityFromLease(lease: LeaseRecord): ClientIdentity {
  // Scope resolution is delegated to the Grant read-model so leases and
  // API tokens share one source of truth for path/action mapping.
  return {
    id: `lease:${lease.id}`,
    name: `Lease: ${lease.label}`,
    source: 'lease',
    rawToolsEnabled: false,
    permissions: leaseGrantScope(lease),
    ...(lease.published === undefined ? {} : { published: lease.published }),
  };
}

function logHttpRequest(
  requestLog: ((entry: HttpRequestLogEntry) => void) | undefined,
  req: Request,
  status: number,
  kind: string,
  detail?: string,
  clientId?: string,
): void {
  if (!requestLog) return;
  const ip = remoteAddressFor(req);
  const resolvedClientId = clientId ?? clientIdForRequestLog(req);
  requestLog({
    ts: new Date().toISOString(),
    kind,
    method: req.method,
    path: req.path,
    status,
    ...(detail ? { detail } : {}),
    ...(resolvedClientId ? { clientId: resolvedClientId } : {}),
    ...(ip ? { ip } : {}),
  });
}

function clientIdForRequestLog(req: Request): string | undefined {
  const identity = readClientIdentity(req);
  if (identity) return identity.id;
  const dashboardSession = req.res?.locals?.dashboardSession as Partial<DashboardSession> | undefined;
  return typeof dashboardSession?.username === 'string' ? dashboardSession.username : undefined;
}

function remoteAddressFor(req: Request): string | undefined {
  const relayAddress = relayRemoteAddressFor(req);
  if (relayAddress) return relayAddress;
  const raw = req.socket?.remoteAddress;
  if (!raw) return undefined;
  // Strip the IPv4-mapped IPv6 prefix so logs show 127.0.0.1 instead of ::ffff:127.0.0.1.
  return raw.startsWith('::ffff:') ? raw.slice(7) : raw;
}

function relayRemoteAddressFor(req: Request): string | undefined {
  // Trust these headers only after the configured relay marks the request
  // as relayed. The relay should stamp x-mvmt-relay-client-ip; Fly provides
  // fly-client-ip today, and x-forwarded-for is a compatibility fallback.
  if (!isRelayRequest(req)) return undefined;
  return firstForwardedAddress(
    firstHeaderValue(req.headers['x-mvmt-relay-client-ip'])
      ?? firstHeaderValue(req.headers['fly-client-ip'])
      ?? firstHeaderValue(req.headers['x-forwarded-for']),
  );
}

function firstForwardedAddress(value: string | undefined): string | undefined {
  const first = value?.split(',')[0]?.trim();
  if (!first) return undefined;
  const cleaned = first.replace(/[\u0000-\u001F\u007F]/g, '').slice(0, 128);
  return cleaned || undefined;
}

// True when the request was forwarded by the mvmt relay, which stamps
// this header on every request it proxies.
function isRelayRequest(req: Request): boolean {
  return firstHeaderValue(req.headers['x-mvmt-transport'])?.toLowerCase() === 'relay';
}

// True only for a request that genuinely originated on this machine.
// A tunnel (mvmt relay, Cloudflare, pinggy, localhost.run, custom)
// connects to 127.0.0.1 just like a local app, so the socket address is
// not a reliable signal. Two things distinguish a real local request:
// the mvmt relay stamps x-mvmt-transport, and every other public tunnel
// forwards the original public Host header rather than 127.0.0.1. A
// request missing a Host header is treated as non-local (stricter).
function isLocalRequest(req: Request): boolean {
  if (isRelayRequest(req)) return false;
  const hostHeader = firstHeaderValue(req.headers.host);
  if (!hostHeader) return false;
  const bracketEnd = hostHeader.indexOf(']');
  const host = hostHeader.startsWith('[') && bracketEnd > 0
    ? hostHeader.slice(1, bracketEnd)
    : hostHeader.split(':')[0];
  const normalized = host.toLowerCase();
  return normalized === 'localhost' || normalized === '127.0.0.1' || normalized === '::1';
}

function getSessionId(req: Request): string | undefined {
  const header = req.headers['mcp-session-id'];
  return Array.isArray(header) ? header[0] : header;
}

export function isStandaloneSseRequest(req: Request): boolean {
  return req.method === 'GET' && headerIncludes(req.headers.accept, 'text/event-stream');
}

export function isBenignDuplicateSseConflict(error: Error): boolean {
  return error.message === 'Conflict: Only one SSE stream is allowed per session';
}

function headerIncludes(value: string | string[] | undefined, needle: string): boolean {
  if (Array.isArray(value)) return value.some((entry) => entry.includes(needle));
  return value?.includes(needle) ?? false;
}

function closeHttpServer(server: HttpServer): Promise<void> {
  if (!server.listening) return Promise.resolve();

  return new Promise((resolve, reject) => {
    const forceCloseTimer = setTimeout(() => {
      server.closeAllConnections?.();
    }, 1000);
    forceCloseTimer.unref();

    server.close((err) => {
      clearTimeout(forceCloseTimer);
      if (err && (err as NodeJS.ErrnoException).code !== 'ERR_SERVER_NOT_RUNNING') {
        reject(err);
        return;
      }
      resolve();
    });
    server.closeIdleConnections?.();
  });
}

type AuthorizeParams = {
  responseType: string;
  clientId: string;
  redirectUri: string;
  resource?: string;
  state?: string;
  requestId?: string;
  scope?: string;
  codeChallenge: string;
  codeChallengeMethod: CodeChallengeMethod;
};

function parseAuthorizeParams(source: Record<string, unknown>): AuthorizeParams | { error: string } {
  const responseType = stringField(source.response_type);
  const clientId = stringField(source.client_id);
  const redirectUri = stringField(source.redirect_uri);
  const resource = stringField(source.resource);
  const codeChallenge = stringField(source.code_challenge);
  const codeChallengeMethodRaw = stringField(source.code_challenge_method) ?? 'S256';

  if (!responseType) return { error: 'Missing response_type' };
  if (responseType !== 'code') return { error: 'Only response_type=code is supported' };
  if (!clientId) return { error: 'Missing client_id' };
  if (!redirectUri) return { error: 'Missing redirect_uri' };
  if (!codeChallenge) return { error: 'Missing code_challenge (PKCE required)' };

  try {
    new URL(redirectUri);
  } catch {
    return { error: 'Invalid redirect_uri' };
  }
  if (codeChallengeMethodRaw !== 'S256') {
    return { error: 'Unsupported code_challenge_method (S256 required)' };
  }

  return {
    responseType,
    clientId,
    redirectUri,
    resource,
    state: stringField(source.state),
    requestId: stringField(source.request_id),
    scope: stringField(source.scope),
    codeChallenge,
    codeChallengeMethod: 'S256',
  };
}

function defaultAuthorizeResource(
  params: AuthorizeParams,
  canonicalResource: string,
): { params: AuthorizeParams; resourceDefaulted: boolean } {
  if (params.resource) return { params, resourceDefaulted: false };
  return {
    params: { ...params, resource: canonicalResource },
    resourceDefaulted: true,
  };
}

function stringField(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined;
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
}

type ResourceValidationError = {
  code: 'missing_resource' | 'invalid_resource';
  description: string;
};

function validateAuthorizeResource(resource: string | undefined, expectedResource: string): ResourceValidationError | undefined {
  if (!resource) return { code: 'missing_resource', description: 'Missing resource' };
  const normalized = normalizeResourceUrl(resource);
  const expected = normalizeResourceUrl(expectedResource);
  if (!normalized || !expected || normalized !== expected) {
    return { code: 'invalid_resource', description: 'Invalid resource' };
  }
  return undefined;
}

function normalizeResourceUrl(value: string): string | undefined {
  try {
    const parsed = new URL(value);
    if (parsed.username || parsed.password || parsed.hash) return undefined;
    const pathname = stripTrailingSlashes(parsed.pathname) || '/';
    return `${parsed.protocol.toLowerCase()}//${parsed.host.toLowerCase()}${pathname}${parsed.search}`;
  } catch {
    return undefined;
  }
}

function stripTrailingSlashes(value: string): string {
  let end = value.length;
  while (end > 0 && value[end - 1] === '/') end -= 1;
  return value.slice(0, end);
}

function authorizeErrorRedirect(
  params: Pick<AuthorizeParams, 'redirectUri' | 'state'>,
  error: string,
  description: string,
): string {
  const redirect = new URL(params.redirectUri);
  redirect.searchParams.set('error', error);
  redirect.searchParams.set('error_description', description);
  if (params.state) redirect.searchParams.set('state', params.state);
  return redirect.toString();
}

function firstHeaderValue(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) return value[0];
  return value;
}

function firstStringQuery(value: unknown): string | undefined {
  if (Array.isArray(value)) return firstStringQuery(value[0]);
  return typeof value === 'string' && value.length > 0 ? value : undefined;
}

function regexParam(req: Request, index: number): string | undefined {
  const value = (req.params as Record<string, string | undefined>)[String(index)];
  return value && value.length > 0 ? value : undefined;
}

function bearerToken(req: Request): string | undefined {
  const header = firstHeaderValue(req.headers.authorization);
  if (!header) return undefined;
  const scheme = 'bearer';
  if (header.slice(0, scheme.length).toLowerCase() !== scheme) return undefined;
  let index = scheme.length;
  if (!isHttpWhitespace(header[index])) return undefined;
  while (isHttpWhitespace(header[index])) index += 1;
  const token = header.slice(index).trim();
  return token.length > 0 ? token : undefined;
}

function isHttpWhitespace(value: string | undefined): boolean {
  return value === ' ' || value === '\t';
}

type ByteRange = { start: number; end: number };

function parseRangeHeader(value: string | undefined, size: number): ByteRange | 'invalid' | undefined {
  if (!value) return undefined;
  const match = /^bytes=(\d*)-(\d*)$/.exec(value.trim());
  if (!match) return 'invalid';
  if (size <= 0) return 'invalid';
  const [, startRaw, endRaw] = match;
  if (!startRaw && !endRaw) return 'invalid';
  if (!startRaw) {
    const suffixLength = Number(endRaw);
    if (!Number.isInteger(suffixLength) || suffixLength <= 0) return 'invalid';
    return { start: Math.max(0, size - suffixLength), end: size - 1 };
  }
  const start = Number(startRaw);
  const end = endRaw ? Number(endRaw) : size - 1;
  if (!Number.isInteger(start) || !Number.isInteger(end) || start < 0 || end < start || start >= size) return 'invalid';
  return { start, end: Math.min(end, size - 1) };
}

function escapeHeaderValue(value: string): string {
  return value.replace(/["\\\r\n]/g, '_');
}

function parseContentLength(value: string | undefined): number | 'invalid' | undefined {
  if (!value) return undefined;
  const parsed = Number(value);
  if (!Number.isSafeInteger(parsed) || parsed < 0) return 'invalid';
  return parsed;
}

class LeaseUploadTooLargeError extends Error {
  constructor(maxBytes: number) {
    super(`Lease upload exceeds ${maxBytes} bytes`);
  }
}

class FsUploadTooLargeError extends Error {
  constructor(maxBytes: number) {
    super(`FS upload exceeds ${maxBytes} bytes`);
  }
}

function limitFsUploadStream(input: NodeJS.ReadableStream, maxBytes: number): NodeJS.ReadableStream {
  let bytes = 0;
  const meter = new Transform({
    transform(chunk: Buffer, _encoding, callback) {
      bytes += chunk.length;
      if (bytes > maxBytes) {
        callback(new FsUploadTooLargeError(maxBytes));
        return;
      }
      callback(null, chunk);
    },
  });
  return input.pipe(meter);
}

async function writeLeaseUpload(
  req: Request,
  parentRealPath: string,
  targetRealPath: string,
  maxBytes: number,
  options: { overwrite: boolean },
): Promise<number> {
  const tempPath = path.join(parentRealPath, `.mvmt-upload-${process.pid}-${randomUUID()}.tmp`);
  let bytes = 0;
  const meter = new Transform({
    transform(chunk: Buffer, _encoding, callback) {
      bytes += chunk.length;
      if (bytes > maxBytes) {
        callback(new LeaseUploadTooLargeError(maxBytes));
        return;
      }
      callback(null, chunk);
    },
  });

  try {
    await pipeline(req, meter, createWriteStream(tempPath, { flags: 'wx', mode: 0o600 }));
    if (options.overwrite) {
      await fsp.rename(tempPath, targetRealPath);
    } else {
      await fsp.link(tempPath, targetRealPath);
    }
    return bytes;
  } finally {
    await fsp.unlink(tempPath).catch(() => undefined);
  }
}

function isNodeErrorCode(err: unknown, code: string): boolean {
  return typeof err === 'object' && err !== null && (err as NodeJS.ErrnoException).code === code;
}

function createDashboardSession(userId: string, username: string, admin: boolean): DashboardSession {
  return {
    id: randomUUID(),
    userId,
    username,
    expiresAt: Date.now() + DASHBOARD_SESSION_TTL_MS,
    admin,
  };
}

function dashboardSessionForRequest(
  req: Request,
  sessions: ReadonlyMap<string, DashboardSession>,
): DashboardSession | undefined {
  const cookies = parseCookieHeader(firstHeaderValue(req.headers.cookie));
  const id = cookies[DASHBOARD_SESSION_COOKIE];
  if (!id) return undefined;
  const session = sessions.get(id);
  if (!session || session.expiresAt <= Date.now()) return undefined;
  return session;
}

function dashboardSessionCookie(session: DashboardSession, req: Request): string {
  const maxAge = Math.max(0, Math.floor((session.expiresAt - Date.now()) / 1000));
  const secure = dashboardCookieSecureSuffix(req);
  return `${DASHBOARD_SESSION_COOKIE}=${encodeURIComponent(session.id)}; Path=/; HttpOnly; SameSite=Strict; Max-Age=${maxAge}${secure}`;
}

function clearDashboardSessionCookie(req: Request): string {
  const secure = dashboardCookieSecureSuffix(req);
  return `${DASHBOARD_SESSION_COOKIE}=; Path=/; HttpOnly; SameSite=Strict; Max-Age=0${secure}`;
}

// The Secure attribute must reflect the *browser's* scheme, not the
// configured publicBaseUrl. The same mvmt instance is reachable both
// locally over http://127.0.0.1 and via the relay over https://. If we
// always marked the cookie Secure based on the public URL, the local
// browser would silently drop it on every subsequent request and the
// dashboard would appear to log in then immediately bounce back to the
// sign-in screen.
function dashboardCookieSecureSuffix(req: Request): string {
  const origin = firstHeaderValue(req.headers.origin);
  return origin && origin.toLowerCase().startsWith('https://') ? '; Secure' : '';
}

function parseCookieHeader(header: string | undefined): Record<string, string> {
  const result: Record<string, string> = {};
  if (!header) return result;
  for (const part of header.split(';')) {
    const index = part.indexOf('=');
    if (index < 0) continue;
    const key = part.slice(0, index).trim();
    if (!key) continue;
    try {
      result[key] = decodeURIComponent(part.slice(index + 1).trim());
    } catch {
      result[key] = part.slice(index + 1).trim();
    }
  }
  return result;
}

function publicDashboardUser(user: PrivilegedUser): { id: string; username: string; createdAt: string; lastLoginAt?: string } {
  return {
    id: user.id,
    username: user.username,
    createdAt: user.createdAt,
    ...(user.lastLoginAt ? { lastLoginAt: user.lastLoginAt } : {}),
  };
}

function dashboardLeasePaths(body: Record<string, unknown>): string[] {
  if (Array.isArray(body.paths)) {
    return body.paths.filter((value): value is string => typeof value === 'string' && value.trim().length > 0);
  }
  return typeof body.path === 'string' && body.path.trim().length > 0 ? [body.path] : [];
}

function dashboardMaxDownloads(body: Record<string, unknown>): number | 'invalid' | undefined {
  const value = body.maxDownloads ?? body.downloads;
  if (value === undefined || value === null || value === '') return undefined;
  if (typeof value === 'number') {
    if (value === -1) return undefined;
    return Number.isSafeInteger(value) && value > 0 ? value : 'invalid';
  }
  if (typeof value !== 'string') return 'invalid';
  const trimmed = value.trim();
  if (trimmed === '-1') return undefined;
  if (!/^[1-9]\d*$/.test(trimmed)) return 'invalid';
  const limit = Number.parseInt(trimmed, 10);
  return Number.isSafeInteger(limit) ? limit : 'invalid';
}

function mountInputFromBody(
  body: Record<string, unknown>,
  existing: readonly { name: string }[],
): MountInput | string {
  const root = mountRootFromDashboardValue(body.root, 'root_required');
  if (root === 'root_required' || root === 'invalid_root') return root;
  const rawName = typeof body.name === 'string' ? body.name.trim() : '';
  const name = rawName || nextAvailableMountName(defaultMountNameFromRoot(root), existing);
  if (!/^[a-z0-9][a-z0-9_-]*$/.test(name)) return 'invalid_name';
  const mountPath = typeof body.path === 'string' && body.path.trim().length > 0 ? body.path.trim() : undefined;
  return {
    name,
    root,
    ...(mountPath ? { path: mountPath } : {}),
    writeAccess: Boolean(body.writeAccess),
    ...(typeof body.description === 'string' ? { description: body.description } : {}),
    ...(typeof body.guidance === 'string' ? { guidance: body.guidance } : {}),
    enabled: body.enabled === false ? false : true,
  };
}

function mountPatchFromBody(body: Record<string, unknown>): Partial<MountInput> | string {
  const patch: Partial<MountInput> = {};
  if (body.root !== undefined) {
    const root = mountRootFromDashboardValue(body.root, 'invalid_root');
    if (root === 'root_required' || root === 'invalid_root') return 'invalid_root';
    patch.root = root;
  }
  if (body.path !== undefined) {
    if (typeof body.path !== 'string' || body.path.trim().length === 0) return 'invalid_path';
    patch.path = body.path.trim();
  }
  if (body.writeAccess !== undefined) patch.writeAccess = Boolean(body.writeAccess);
  if (body.enabled !== undefined) patch.enabled = Boolean(body.enabled);
  if (typeof body.description === 'string') patch.description = body.description;
  if (typeof body.guidance === 'string') patch.guidance = body.guidance;
  return patch;
}

function mountRootFromDashboardValue(value: unknown, emptyError: 'root_required' | 'invalid_root'): string | 'root_required' | 'invalid_root' {
  if (typeof value !== 'string') return emptyError;
  const root = value.trim();
  if (!root) return emptyError;
  const resolved = resolveDashboardMountRoot(root);
  return resolved ?? 'invalid_root';
}

function resolveDashboardMountRoot(root: string): string | undefined {
  if (root.length > 4096 || root.includes('\0')) return undefined;
  const isHomeRelative = root === '~' || root.startsWith(`~${path.sep}`);
  if (!path.isAbsolute(root) && !isHomeRelative) return undefined;
  const resolved = resolveSetupPath(root);
  if (!path.isAbsolute(resolved) || resolved.includes('\0')) return undefined;
  return resolved;
}

async function assertMountRootUsable(root: string): Promise<void> {
  const resolvedRoot = resolveDashboardMountRoot(root);
  if (!resolvedRoot) throw new Error('invalid_root');
  if (process.platform === 'win32') throw new Error('invalid_root');
  if (await testLocalPath(resolvedRoot, '-d')) return;
  if (await testLocalPath(resolvedRoot, '-f')) return;
  throw new Error('invalid_root');
}

async function testLocalPath(resolvedRoot: string, testFlag: '-d' | '-f'): Promise<boolean> {
  try {
    await execFileAsync('/bin/test', [testFlag, resolvedRoot], { timeout: 5_000, maxBuffer: 1024 });
    return true;
  } catch {
    return false;
  }
}

async function pickLocalPathWithFinder(kind: 'file' | 'folder'): Promise<string | undefined> {
  if (process.platform !== 'darwin') throw new Error('finder_unavailable');
  const script = kind === 'file'
    ? 'POSIX path of (choose file with prompt "Choose a file to add to mvmt")'
    : 'POSIX path of (choose folder with prompt "Choose a folder to add to mvmt")';
  try {
    const { stdout } = await execFileAsync('osascript', ['-e', script], { timeout: 120_000, maxBuffer: 1024 * 1024 });
    const picked = String(stdout).trim();
    return picked || undefined;
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    if (message.includes('User canceled') || message.includes('-128')) return undefined;
    throw err;
  }
}

// Use the basename only so an auto-generated mount name doesn't leak full
// local paths or usernames (`/Users/alice/code/foo` → `foo`, not
// `users-alice-code-foo`). When the basename is empty or invalid, fall back
// to "mount" and let the collision-suffix step pick a free name.
function defaultMountNameFromRoot(root: string): string {
  const last = root.replaceAll('\\', '/').split('/').filter(Boolean).pop() ?? '';
  const slug = last
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '');
  if (!slug || !/^[a-z0-9]/.test(slug)) return 'mount';
  return slug;
}

function nextAvailableMountName(base: string, existing: readonly { name: string }[]): string {
  const taken = new Set(existing.map((entry) => entry.name));
  if (!taken.has(base)) return base;
  for (let i = 2; i < 10_000; i += 1) {
    const candidate = `${base}-${i}`;
    if (!taken.has(candidate)) return candidate;
  }
  return base;
}

async function dashboardLeaseResources(
  mounts: readonly LocalFolderMountConfig[],
  paths: string[],
): Promise<LeaseResource[]> {
  const resources: LeaseResource[] = [];
  for (const inputPath of paths) {
    const target = await resolveDashboardFileTarget(mounts, inputPath);
    resources.push({
      path: resourcePathForDashboardPath(target.virtualPath),
      sourcePath: target.virtualPath,
      type: target.type === 'file' ? 'file' : 'folder',
    });
  }
  return uniqueDashboardLeaseResources(resources);
}

async function dashboardLeaseResourcesAreWritable(
  mounts: readonly LocalFolderMountConfig[],
  paths: string[],
): Promise<boolean> {
  for (const inputPath of paths) {
    const target = await resolveDashboardFileTarget(mounts, inputPath);
    if (!target.writeAccess) return false;
  }
  return true;
}

function leasePermissionsForDashboardMode(mode: string): readonly ('read' | 'upload')[] | undefined {
  if (mode === 'read') return ['read'];
  if (mode === 'upload') return ['upload'];
  if (mode === 'two-way' || mode === 'two_way' || mode === 'read+upload') return ['read', 'upload'];
  return undefined;
}

function resourcePathForDashboardPath(inputPath: string): string {
  const segments = normalizeDashboardPath(inputPath).split('/').filter(Boolean);
  return `/${segments.join('-') || 'resource'}`;
}

function uniqueDashboardLeaseResources(resources: LeaseResource[]): LeaseResource[] {
  const seen = new Set<string>();
  const unique: LeaseResource[] = [];
  for (const resource of resources) {
    const key = `${resource.sourcePath}:${resource.type}`;
    if (seen.has(key)) continue;
    seen.add(key);
    unique.push(resource);
  }
  return unique;
}

function leasePublicUrl(baseUrl: string, id: string, token: string): string {
  // Concatenate paths instead of using `new URL('/lease/...', baseUrl)` —
  // the URL constructor treats an absolute-path reference as resetting the
  // path, which silently drops a relay workspace prefix like /t/demo.
  const trimmed = baseUrl.replace(/\/+$/, '');
  const url = new URL(`${trimmed}/lease/${encodeURIComponent(id)}`);
  url.searchParams.set('token', token);
  return url.toString();
}

const DASHBOARD_PAGE_HTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>mvmt</title>
<style>
:root{
  --bg:#f5f6f8;--panel:#fff;--text:#0f172a;--text-2:#475569;--muted:#94a3b8;
  --border:#e4e7ec;--border-strong:#cbd5e1;--hover:#f1f5f9;
  --accent:#0f766e;--accent-2:#115e59;--accent-soft:#ccfbf1;
  --danger:#b91c1c;--danger-soft:#fee2e2;
  --warn:#b45309;--warn-soft:#fef3c7;
  --ok:#15803d;--ok-soft:#dcfce7;
  --info:#0369a1;--info-soft:#e0f2fe;
  --shadow:0 1px 2px rgba(15,23,42,.04),0 1px 3px rgba(15,23,42,.04);
  --shadow-lg:0 10px 25px rgba(15,23,42,.12),0 4px 10px rgba(15,23,42,.06);
  --radius:10px;--radius-sm:6px;
}
*,*::before,*::after{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif;-webkit-font-smoothing:antialiased}
header.top{background:#fff;border-bottom:1px solid var(--border);padding:.85rem 1.5rem;display:flex;justify-content:space-between;align-items:center;position:sticky;top:0;z-index:5}
.brand{display:flex;align-items:center;gap:.5rem;font-weight:700;font-size:1.05rem;letter-spacing:-.01em}
.brand .dot{width:.55rem;height:.55rem;border-radius:50%;background:var(--accent)}
.brand-sub{color:var(--text-2);font-weight:500;font-size:.85rem;margin-left:.35rem}
.user-strip{display:flex;align-items:center;gap:.6rem}
.user-strip .who{display:inline-flex;align-items:center;gap:.4rem;padding:.25rem .6rem;border-radius:999px;background:var(--hover);color:var(--text-2);font-size:.85rem}
main{max-width:1280px;margin:0 auto;padding:1.5rem}
h1,h2,h3,h4{margin:0;font-weight:600;letter-spacing:-.01em;color:var(--text)}
h2{font-size:1.05rem}
h3{font-size:.95rem}
h4{font-size:.9rem}
a{color:var(--accent);text-decoration:none}
a:hover{text-decoration:underline}
.muted{color:var(--text-2)}
.subtle{color:var(--muted);font-size:.82rem}
.hidden{display:none !important}
/* Buttons */
.btn{font:inherit;display:inline-flex;align-items:center;gap:.4rem;padding:.45rem .85rem;background:#fff;color:var(--text);border:1px solid var(--border);border-radius:var(--radius-sm);cursor:pointer;font-weight:500;line-height:1.2;transition:background .12s,border-color .12s,color .12s}
.btn:hover{background:var(--hover);border-color:var(--border-strong)}
.btn:disabled{opacity:.55;cursor:not-allowed}
.btn-primary{background:var(--accent);color:#fff;border-color:var(--accent)}
.btn-primary:hover{background:var(--accent-2);border-color:var(--accent-2)}
.btn-danger{color:var(--danger);border-color:var(--border)}
.btn-danger:hover{background:var(--danger-soft);border-color:var(--danger);color:var(--danger)}
.btn-ghost{background:transparent;border-color:transparent;color:var(--text-2)}
.btn-ghost:hover{background:var(--hover);color:var(--text)}
.btn-sm{padding:.32rem .7rem;font-size:.82rem}
.btn-icon{padding:.4rem;width:2rem;height:2rem;justify-content:center}
.btn .icon{width:1rem;height:1rem;flex:none}
/* Inputs */
input,select,textarea{font:inherit;color:var(--text);padding:.5rem .65rem;border:1px solid var(--border);border-radius:var(--radius-sm);background:#fff;width:100%;outline:none;transition:border-color .12s,box-shadow .12s}
input:focus,select:focus,textarea:focus{border-color:var(--accent);box-shadow:0 0 0 3px rgba(15,118,110,.15)}
label{font-size:.82rem;color:var(--text-2);font-weight:500}
/* Panel / card */
.panel{background:var(--panel);border:1px solid var(--border);border-radius:var(--radius);box-shadow:var(--shadow);padding:1rem 1.25rem 1.1rem;margin-bottom:1rem}
.panel-head{display:flex;justify-content:space-between;align-items:center;gap:.75rem;margin-bottom:.85rem;flex-wrap:wrap}
.panel-title{display:flex;align-items:baseline;gap:.5rem}
.panel-title .count{color:var(--muted);font-size:.85rem;font-weight:500}
/* Finder shell */
.finder{display:grid;grid-template-columns:210px minmax(0,1fr);min-height:calc(100vh - 7rem);background:#fff;border:1px solid var(--border);border-radius:14px;box-shadow:var(--shadow);overflow:hidden}
.finder-sidebar{background:#f8fafc;border-right:1px solid var(--border);padding:.85rem}
.nav-btn{width:100%;display:flex;align-items:center;gap:.55rem;padding:.58rem .7rem;margin-bottom:.2rem;border:0;border-radius:8px;background:transparent;color:var(--text-2);font:inherit;font-weight:500;cursor:pointer;text-align:left}
.nav-btn:hover{background:#eef2f7;color:var(--text)}
.nav-btn.active{background:#e7f3f1;color:var(--accent)}
.nav-btn .icon{width:1rem;height:1rem;flex:none}
.nav-btn .nav-count{margin-left:auto;color:var(--muted);font-size:.78rem}
.finder-main{min-width:0;display:flex;flex-direction:column;background:#fff}
.finder-toolbar{display:flex;align-items:center;justify-content:space-between;gap:.75rem;padding:.7rem .85rem;border-bottom:1px solid var(--border);background:#fff;position:sticky;top:0;z-index:3}
.toolbar-left{display:flex;align-items:center;gap:.55rem;min-width:0;flex:1}
.toolbar-right{display:flex;align-items:center;gap:.35rem;flex:none}
.view-panel{padding:1rem}
.view-title{display:flex;align-items:baseline;gap:.45rem;margin-bottom:.8rem}
.view-title h2{font-size:1.08rem}
.view-title .count{color:var(--muted);font-size:.85rem}
.view-toggle.active{background:var(--hover);border-color:var(--border-strong);color:var(--accent)}
.file-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(112px,1fr));gap:.8rem}
.file-card{border:1px solid transparent;border-radius:10px;padding:.75rem .55rem;text-align:center;cursor:default;background:#fff;min-height:112px;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:.45rem}
.file-card:hover{background:var(--hover);border-color:var(--border)}
.file-card.selected{background:#e7f3f1;border-color:var(--accent)}
.file-card .icon{width:2.2rem;height:2.2rem;color:var(--muted)}
.file-card .file-card-name{font-weight:500;font-size:.84rem;line-height:1.25;word-break:break-word;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.file-card .file-card-meta{font-size:.72rem;color:var(--muted)}
.context-menu{position:fixed;z-index:80;min-width:180px;background:#fff;border:1px solid var(--border);border-radius:10px;box-shadow:var(--shadow-lg);padding:.25rem}
.context-menu button{width:100%;display:flex;align-items:center;gap:.55rem;border:0;background:transparent;border-radius:7px;padding:.5rem .6rem;text-align:left;font:inherit;color:var(--text);cursor:pointer}
.context-menu button:hover{background:#0a84ff;color:#fff}
.context-menu button .icon{width:1rem;height:1rem;flex:none}
.settings-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:1rem}
.settings-card{border:1px solid var(--border);border-radius:10px;padding:1rem;background:#fff}
.settings-card h3{margin-bottom:.65rem}
.apps-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:.8rem}
.app-card{display:block;text-decoration:none;color:inherit;border:1px solid var(--border);border-radius:10px;padding:1rem;background:#fff;transition:transform .1s ease, box-shadow .1s ease}
.app-card:hover{transform:translateY(-1px);box-shadow:var(--shadow-sm)}
.app-card-title{font-weight:600;margin-bottom:.25rem}
.app-card-desc{color:var(--muted);font-size:.85rem;line-height:1.35}
.kv{display:grid;grid-template-columns:7rem minmax(0,1fr);gap:.55rem;font-size:.86rem;margin:.45rem 0}
.kv span:first-child{color:var(--text-2);font-weight:500}
.kv code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.8rem;word-break:break-all}
.properties-grid{display:grid;grid-template-columns:8rem minmax(0,1fr);gap:.6rem 1rem;font-size:.9rem}
.properties-grid dt{color:var(--text-2);font-weight:500}
.properties-grid dd{margin:0;min-width:0;word-break:break-word}
.properties-grid code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.82rem}
.log-list{border:1px solid var(--border);border-radius:8px;overflow:hidden;background:#0b1020;color:#e5e7eb;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.75rem;max-height:260px;overflow-y:auto}
.log-row{display:grid;grid-template-columns:4.5rem 3rem 8rem minmax(0,1fr);gap:.5rem;padding:.38rem .55rem;border-bottom:1px solid rgba(255,255,255,.08)}
.log-row:last-child{border-bottom:0}
.log-row .path{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
/* Table */
.tablewrap{border:1px solid var(--border);border-radius:var(--radius-sm);overflow:hidden;background:#fff}
table.t{border-collapse:collapse;width:100%}
.t th{font-size:.7rem;font-weight:600;color:var(--text-2);background:#fafbfc;text-transform:uppercase;letter-spacing:.05em;padding:.55rem .85rem;text-align:left;border-bottom:1px solid var(--border)}
.t td{padding:.6rem .85rem;border-bottom:1px solid var(--border);font-size:.88rem;vertical-align:middle;color:var(--text)}
.t tbody tr:last-child td{border-bottom:none}
.t tbody tr{transition:background .1s}
.t tbody tr:hover{background:var(--hover)}
.t tbody tr.selected{background:var(--accent-soft)}
.cell-path{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.78rem;color:var(--text-2);word-break:break-all}
.cell-name{display:flex;align-items:center;gap:.5rem;min-width:0}
.cell-name .icon{color:var(--muted);flex:none}
.cell-name a{color:var(--text);font-weight:500}
.cell-name a:hover{color:var(--accent);text-decoration:none}
.cell-name .name-text{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.t .actions{display:flex;gap:.35rem;justify-content:flex-end;white-space:nowrap}
.t .actions .btn{padding:.3rem .65rem;font-size:.8rem}
.t .actions .btn-icon{padding:.32rem;width:1.8rem;height:1.8rem}
.cell-num{text-align:right;font-variant-numeric:tabular-nums;color:var(--text-2);font-size:.82rem}
/* Badges */
.badge{display:inline-flex;align-items:center;gap:.3rem;padding:.15rem .55rem;border-radius:999px;font-size:.7rem;font-weight:600;letter-spacing:.04em;text-transform:uppercase;white-space:nowrap;line-height:1.5;border:1px solid transparent}
.badge.read{background:var(--info-soft);color:var(--info);border-color:#bae6fd}
.badge.write,.badge.upload{background:var(--warn-soft);color:var(--warn);border-color:#fde68a}
.badge.readonly{background:#f1f5f9;color:var(--text-2);border-color:#e2e8f0}
.badge.active{background:var(--ok-soft);color:var(--ok);border-color:#bbf7d0}
.badge.expired,.badge.revoked{background:var(--danger-soft);color:var(--danger);border-color:#fecaca}
.badge .dot{width:.4rem;height:.4rem;border-radius:50%;background:currentColor}
/* Breadcrumbs */
.crumbs{display:flex;gap:.3rem;align-items:center;flex-wrap:wrap;font-size:.88rem;margin-bottom:.65rem;padding:.4rem .65rem;background:var(--hover);border-radius:var(--radius-sm)}
.crumbs a{color:var(--text-2);padding:.05rem .25rem;border-radius:4px}
.crumbs a:hover{background:#fff;color:var(--text);text-decoration:none}
.crumbs .sep{color:var(--muted);font-size:.85rem}
.crumbs .current{color:var(--text);font-weight:600;padding:.05rem .25rem}
/* Empty state */
.empty{padding:1.5rem .5rem;text-align:center;color:var(--text-2);font-size:.9rem}
.empty .icon{width:2rem;height:2rem;color:var(--muted);margin:0 auto .5rem;display:block}
.empty strong{display:block;color:var(--text);font-weight:600;margin-bottom:.2rem}
/* Tabs */
.tabs{display:flex;gap:.1rem;border-bottom:1px solid var(--border);margin-bottom:.85rem}
.tab{padding:.55rem .9rem;background:transparent;color:var(--text-2);border:none;border-bottom:2px solid transparent;cursor:pointer;font-weight:500;font-size:.85rem;border-radius:0;transition:color .1s,border-color .1s}
.tab:hover{color:var(--text)}
.tab.active{color:var(--accent);border-bottom-color:var(--accent);font-weight:600}
/* Login */
.login-card{max-width:380px;margin:5rem auto 0;padding:1.75rem 1.75rem 1.5rem}
.login-card h2{margin-bottom:1rem}
.login-card form{display:flex;flex-direction:column;gap:.65rem}
/* Form rows */
.form-grid{display:grid;grid-template-columns:140px 1fr;gap:.7rem 1rem;align-items:center}
.form-grid label{align-self:center}
.form-grid .field-help{grid-column:2;color:var(--text-2);font-size:.78rem;margin-top:-.35rem}
.form-grid .perm-note{grid-column:2;color:var(--danger);font-size:.8rem}
.path-picker-row{display:grid;grid-template-columns:minmax(0,1fr) auto auto;gap:.45rem;align-items:center}
.path-picker-row .btn{white-space:nowrap}
/* Modal */
.modal{position:fixed;inset:0;background:rgba(15,23,42,.45);display:flex;align-items:flex-start;justify-content:center;padding:3rem 1rem;z-index:50;overflow:auto;backdrop-filter:blur(2px)}
.modal-card{background:#fff;border-radius:12px;width:100%;max-width:540px;box-shadow:var(--shadow-lg);overflow:hidden;animation:pop .15s ease-out}
@keyframes pop{from{transform:translateY(8px);opacity:0}to{transform:none;opacity:1}}
.modal-head{display:flex;justify-content:space-between;align-items:center;padding:1rem 1.25rem;border-bottom:1px solid var(--border)}
.modal-body{padding:1.15rem 1.25rem}
.modal-foot{padding:.85rem 1.25rem;display:flex;gap:.5rem;justify-content:flex-end;background:var(--bg);border-top:1px solid var(--border)}
.modal-card.wide{max-width:720px}
.target-chip{display:flex;align-items:center;gap:.6rem;padding:.65rem .8rem;background:var(--hover);border:1px solid var(--border);border-radius:var(--radius-sm);margin-bottom:1rem}
.target-chip .icon{flex:none;color:var(--text-2);width:1.1rem;height:1.1rem}
.target-chip .target-path{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.85rem;font-weight:500;word-break:break-all;flex:1;min-width:0}
.success-state{text-align:center;padding:.5rem 0 1rem}
.success-state .check{width:3rem;height:3rem;border-radius:50%;background:var(--ok-soft);color:var(--ok);display:inline-flex;align-items:center;justify-content:center;margin-bottom:.85rem}
.success-state .check .icon{width:1.5rem;height:1.5rem}
.success-state h2{margin-bottom:.4rem}
.share-url{display:flex;align-items:center;gap:.5rem;padding:.5rem;background:var(--hover);border:1px solid var(--border);border-radius:var(--radius-sm);margin-top:1rem}
.share-url code{flex:1;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.78rem;word-break:break-all;color:var(--text);background:transparent;padding:.3rem .25rem}
.share-warning{font-size:.8rem;color:var(--warn);margin-top:.85rem;display:flex;gap:.4rem;align-items:flex-start}
.share-warning .icon{flex:none;width:1rem;height:1rem;margin-top:.1rem}
/* Toasts */
.toast-stack{position:fixed;top:1rem;right:1rem;display:flex;flex-direction:column;gap:.5rem;z-index:100;max-width:24rem}
.toast{background:#fff;border:1px solid var(--border);border-radius:var(--radius-sm);padding:.7rem .9rem;box-shadow:var(--shadow-lg);display:flex;align-items:flex-start;gap:.55rem;font-size:.85rem;animation:slide .15s ease-out}
.toast .icon{flex:none;width:1.05rem;height:1.05rem;margin-top:.05rem}
.toast.success{border-left:3px solid var(--ok)}
.toast.success .icon{color:var(--ok)}
.toast.error{border-left:3px solid var(--danger)}
.toast.error .icon{color:var(--danger)}
.toast.info{border-left:3px solid var(--info)}
.toast.info .icon{color:var(--info)}
@keyframes slide{from{transform:translateX(8px);opacity:0}to{transform:none;opacity:1}}
/* Misc */
.divider{height:1px;background:var(--border);margin:.85rem 0}
@media (max-width:760px){
  .form-grid{grid-template-columns:1fr}
  .form-grid label{margin-bottom:-.25rem}
  .form-grid .field-help{grid-column:1}
  .path-picker-row{grid-template-columns:1fr 1fr}
  .path-picker-row input{grid-column:1 / -1}
  main{padding:1rem}
}
@media (max-width:640px){
  body{font-size:15px}
  header.top{padding:.75rem 1rem;align-items:flex-start;flex-direction:column;gap:.65rem}
  .brand{width:100%;font-size:1rem}
  .brand-sub{font-size:.78rem}
  .user-strip{width:100%;justify-content:space-between;gap:.5rem}
  .user-strip .who{min-width:0;max-width:calc(100vw - 7rem)}
  .user-strip .who #who{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
  main{padding:.75rem}
  .panel{padding:.85rem;margin-bottom:.75rem;border-radius:8px}
  .panel-head{align-items:stretch}
  .panel-head .btn{min-height:2.4rem}
  .finder{display:block;min-height:auto;border-radius:8px}
  .finder-sidebar{display:flex;gap:.35rem;overflow-x:auto;border-right:0;border-bottom:1px solid var(--border);padding:.55rem}
  .nav-btn{width:auto;min-width:max-content;margin-bottom:0}
  .finder-toolbar{position:static;align-items:flex-start;flex-wrap:wrap}
  .toolbar-left{width:100%;flex-basis:100%}
  .toolbar-right{width:100%;justify-content:flex-end;flex-wrap:wrap}
  .view-panel{padding:.7rem}
  .file-grid{grid-template-columns:repeat(2,minmax(0,1fr))}
  .crumbs{overflow-x:auto;flex-wrap:nowrap;white-space:nowrap;-webkit-overflow-scrolling:touch}
  .tabs{display:grid;grid-template-columns:1fr 1fr;gap:0}
  .tab{padding:.7rem .5rem}
  .tablewrap{border:0;background:transparent;overflow:visible}
  table.t,.t thead,.t tbody,.t tr,.t td{display:block;width:100%}
  .t thead{display:none}
  .t tbody{display:flex;flex-direction:column;gap:.65rem}
  .t tbody tr{background:#fff;border:1px solid var(--border);border-radius:var(--radius-sm);padding:.75rem;box-shadow:var(--shadow)}
  .t tbody tr:hover{background:#fff}
  .t tbody tr.selected{background:#fff;border-color:var(--accent)}
  .t td{border-bottom:0;padding:.35rem 0;font-size:.88rem}
  .t td:first-child{padding-bottom:.6rem;margin-bottom:.25rem;border-bottom:1px solid var(--border)}
  .t td:not(:first-child){display:grid;grid-template-columns:6.5rem minmax(0,1fr);gap:.5rem;align-items:center}
  .t td:not(:first-child)::before{content:attr(data-label);color:var(--text-2);font-size:.72rem;font-weight:600;text-transform:uppercase;letter-spacing:.04em}
  .cell-name .name-text{white-space:normal;overflow:visible;text-overflow:clip;word-break:break-word}
  .cell-path{word-break:break-all}
  .cell-num{text-align:left}
  .t .actions{justify-content:stretch;white-space:normal;flex-wrap:wrap;gap:.5rem}
  .t .actions .btn{flex:1 1 auto;justify-content:center;min-height:2.4rem;font-size:.86rem}
  .t .actions .btn-icon{width:auto;min-width:2.6rem}
  .badge{white-space:normal;text-align:center}
  .modal{align-items:flex-end;padding:0}
  .modal-card{border-radius:16px 16px 0 0;max-height:92vh;overflow:auto}
  .modal-head,.modal-body,.modal-foot{padding-left:1rem;padding-right:1rem}
  .modal-foot{position:sticky;bottom:0;display:grid;grid-template-columns:1fr 1fr}
  .modal-foot .btn{justify-content:center;min-height:2.7rem}
  .target-chip{align-items:flex-start;flex-wrap:wrap}
  .target-chip .target-path{flex-basis:100%}
  .share-url{flex-direction:column;align-items:stretch}
  .share-url .btn{justify-content:center}
  .toast-stack{left:.75rem;right:.75rem;top:.75rem;max-width:none}
  .login-card{margin:2rem auto 0}
}
</style>
</head>
<body>
<header class="top">
  <div class="brand"><span class="dot"></span>mvmt<span class="brand-sub">share local folders</span></div>
  <div id="header-user" class="user-strip hidden">
    <span class="who"><svg class="icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></svg><span id="who"></span><span id="who-role" class="badge active hidden" style="margin-left:.4rem;">Admin</span></span>
    <button class="btn btn-ghost btn-sm" id="logout" type="button">Sign out</button>
  </div>
</header>
<main>
  <div id="toast-stack" class="toast-stack"></div>

  <section id="login" class="panel login-card">
    <h2>Sign in</h2>
    <p class="muted" style="margin:.25rem 0 1rem;font-size:.88rem;">Sign in to manage shared links for this device.</p>
    <form id="login-form" method="post" action="" autocomplete="on">
      <input id="username" name="username" autocomplete="username" placeholder="Username" required>
      <input id="password" name="password" type="password" autocomplete="current-password" placeholder="Password" required>
      <button class="btn btn-primary" type="submit" style="justify-content:center;margin-top:.25rem;">Sign in</button>
      <p id="login-status" class="muted" role="status" style="min-height:1.2em;margin:.15rem 0 0;font-size:.84rem;"></p>
    </form>
  </section>

  <section id="app" class="hidden">
    <div class="finder">
      <aside class="finder-sidebar">
        <button type="button" class="nav-btn active" data-view="files"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg><span>Files</span><span class="nav-count" id="sources-count"></span></button>
        <button type="button" class="nav-btn" data-view="shares"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/></svg><span>Shared links</span><span class="nav-count" id="leases-count"></span></button>
        <button type="button" class="nav-btn" data-view="mcp"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M4 17l6-6-6-6"/><path d="M12 19h8"/></svg><span>MCP access</span><span class="nav-count" id="grants-count"></span></button>
        <button type="button" class="nav-btn" data-view="apps"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg><span>Apps</span><span class="nav-count" id="apps-count"></span></button>
        <button type="button" class="nav-btn hidden" id="settings-nav" data-view="settings"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4z"/></svg><span>Settings</span></button>
      </aside>
      <section class="finder-main">
        <div class="finder-toolbar">
          <div class="toolbar-left">
            <div class="crumbs" id="crumbs" data-test="crumbs"></div>
            <div class="view-title hidden" id="shares-title"><h2>Shared links</h2></div>
            <div class="view-title hidden" id="mcp-title"><h2>MCP access</h2></div>
            <div class="view-title hidden" id="apps-title"><h2>Apps</h2></div>
            <div class="view-title hidden" id="settings-title"><h2>Local settings</h2></div>
          </div>
          <div class="toolbar-right">
            <button class="btn btn-primary btn-icon hidden" id="add-mount" type="button" title="Add local folder"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14M5 12h14"/></svg></button>
            <button class="btn btn-primary btn-sm hidden" id="new-grant" type="button"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M12 5v14M5 12h14"/></svg><span>New MCP token</span></button>
            <button class="btn btn-sm hidden" id="share-selected" type="button"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.59 13.51l6.83 3.98M15.41 6.51l-6.82 3.98"/></svg><span>Share</span></button>
            <button class="btn btn-ghost btn-icon view-toggle active" id="view-list" type="button" title="List view"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 6h13M8 12h13M8 18h13"/><path d="M3 6h.01M3 12h.01M3 18h.01"/></svg></button>
            <button class="btn btn-ghost btn-icon view-toggle" id="view-grid" type="button" title="Icon view"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg></button>
            <button class="btn btn-ghost btn-icon" id="refresh" type="button" title="Refresh"><svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M3 12a9 9 0 0 1 15.5-6.4L21 8M21 3v5h-5M21 12a9 9 0 0 1-15.5 6.4L3 16M3 21v-5h5"/></svg></button>
          </div>
        </div>

        <section class="view-panel" id="view-files-panel">
          <div id="empty-mounts" class="empty hidden">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>
            <strong>No folders mounted</strong>
            Open this dashboard locally to add folders from this device.
          </div>
          <div class="tablewrap hidden" id="files-wrap">
            <table class="t"><thead><tr><th>Name</th><th>Size</th><th>Kind</th><th>Modified</th></tr></thead><tbody id="files"></tbody></table>
          </div>
          <div class="file-grid hidden" id="files-grid"></div>
        </section>

        <section class="view-panel hidden" id="view-shares-panel">
          <div class="tabs" data-test="lease-tabs">
            <button type="button" class="tab active" data-tab="active">Active</button>
            <button type="button" class="tab" data-tab="inactive">Past</button>
          </div>
          <div id="leases-empty" class="empty hidden">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/></svg>
            <strong>No shared links</strong>
            Select a file or folder, then share it.
          </div>
          <div class="tablewrap hidden" id="leases-wrap">
            <table class="t"><thead><tr><th>Label</th><th>Path</th><th>Permission</th><th>Status</th><th>Activity</th><th></th></tr></thead><tbody id="leases"></tbody></table>
          </div>
        </section>

        <section class="view-panel hidden" id="view-mcp-panel">
          <p class="muted" id="mcp-intro" style="margin:0 0 .75rem;font-size:.88rem;">MCP tokens let an agent — Claude, your own apps — reach your mounts over the MCP protocol. Each token's access is set when you create it.</p>
          <div id="grants-empty" class="empty hidden">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M4 17l6-6-6-6"/><path d="M12 19h8"/></svg>
            <strong>No MCP tokens</strong>
            Create a token to give an agent access to your mounts.
          </div>
          <div class="tablewrap hidden" id="grants-wrap">
            <table class="t"><thead><tr><th>Label</th><th>Access</th><th>Reach</th><th>Expires</th><th>Last used</th><th></th></tr></thead><tbody id="grants"></tbody></table>
          </div>
        </section>

        <section class="view-panel hidden" id="view-apps-panel">
          <div class="apps-grid" id="apps-grid"></div>
          <div class="muted hidden" id="apps-empty">No apps installed yet.</div>
        </section>

        <section class="view-panel hidden" id="view-settings-panel">
          <div class="settings-grid">
            <div class="settings-card">
              <h3>Network</h3>
              <div id="network-status"></div>
            </div>
            <div class="settings-card" style="grid-column:1 / -1;">
              <h3>Live logs</h3>
              <div class="log-list" id="live-logs"></div>
            </div>
          </div>
        </section>
      </section>
    </div>
  </section>

  <div id="context-menu" class="context-menu hidden" role="menu">
    <button type="button" id="context-share" role="menuitem">
      <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.59 13.51l6.83 3.98M15.41 6.51l-6.82 3.98"/></svg>
      Share
    </button>
    <button type="button" id="context-properties" role="menuitem">
      <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>
      Properties
    </button>
    <button type="button" id="context-rename-source" class="hidden" role="menuitem">
      <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4z"/></svg>
      Rename source
    </button>
    <button type="button" id="context-remove-source" class="hidden danger-menu-item" role="menuitem">
      <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
      Remove source
    </button>
  </div>

  <div id="lease-modal" class="modal hidden" role="dialog" aria-modal="true">
    <div class="modal-card">
      <div id="lease-step-config">
        <div class="modal-head">
          <h2>Share a link</h2>
          <button class="btn btn-ghost btn-icon" type="button" id="lease-modal-close" title="Close">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="target-chip" id="lease-target"></div>
          <div class="form-grid">
            <label for="lease-mode">Permission</label>
            <select id="lease-mode">
              <option value="read">Read only</option>
              <option value="two-way">Read and upload</option>
              <option value="upload">Upload only</option>
            </select>
            <p id="mode-help" class="field-help">Recipients can browse and download.</p>
            <label for="lease-label">Label</label>
            <input id="lease-label" placeholder="e.g. Tax docs for accountant" required>
            <label for="lease-expires">Expires in</label>
            <select id="lease-expires">
              <option value="1h">1 hour</option>
              <option value="24h" selected>24 hours</option>
              <option value="7d">7 days</option>
              <option value="30d">30 days</option>
            </select>
            <label for="lease-downloads">Downloads</label>
            <input id="lease-downloads" type="number" step="1" min="-1" value="-1">
            <p class="field-help">Use -1 for unlimited.</p>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" type="button" id="lease-cancel">Cancel</button>
          <button class="btn btn-primary" type="button" id="lease-create">Create link</button>
        </div>
      </div>
      <div id="lease-step-success" class="hidden">
        <div class="modal-body">
          <div class="success-state">
            <span class="check">
              <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
            </span>
            <h2>Link ready to share</h2>
            <p class="muted" style="margin:.25rem 0 0;">Anyone with this URL can use the lease until it expires or is revoked.</p>
          </div>
          <div class="share-url" data-test="created-card">
            <code id="created-url"></code>
            <button class="btn btn-primary btn-sm" type="button" id="copy-url">
              <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
              Copy
            </button>
          </div>
          <p class="share-warning">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/></svg>
            You can copy this link again from the dashboard until it expires or is revoked.
          </p>
        </div>
        <div class="modal-foot">
          <button class="btn btn-primary" type="button" id="lease-done">Done</button>
        </div>
      </div>
    </div>
  </div>

  <div id="mcp-modal" class="modal hidden" role="dialog" aria-modal="true">
    <div class="modal-card">
      <div id="mcp-step-config">
        <div class="modal-head">
          <h2>New MCP token</h2>
          <button class="btn btn-ghost btn-icon" type="button" id="mcp-modal-close" title="Close">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-grid">
            <label for="mcp-label">Label</label>
            <input id="mcp-label" placeholder="e.g. Claude on this laptop" required>
            <label for="mcp-expires">Expires in</label>
            <select id="mcp-expires">
              <option value="never" selected>Never</option>
              <option value="30d">30 days</option>
              <option value="7d">7 days</option>
              <option value="24h">24 hours</option>
            </select>
            <label for="mcp-published">Public tunnel access</label>
            <label id="mcp-published-field" style="display:flex;align-items:center;gap:.4rem;font-size:.88rem;color:var(--text);"><input id="mcp-published" type="checkbox" style="width:auto;"> Reachable through public tunnels (leave off for apps on this machine only)</label>
            <label>Access</label>
            <div>
              <label style="display:flex;align-items:center;gap:.4rem;font-size:.88rem;color:var(--text);"><input type="radio" name="mcp-access-mode" value="all" checked style="width:auto;"> All mounts</label>
              <label style="display:flex;align-items:center;gap:.4rem;font-size:.88rem;color:var(--text);margin-top:.25rem;"><input type="radio" name="mcp-access-mode" value="custom" style="width:auto;"> Choose mounts</label>
            </div>
          </div>
          <div id="mcp-custom-scopes" class="hidden" style="margin-top:.6rem;border-top:1px solid var(--border);padding-top:.6rem;">
            <p class="muted" style="margin:0 0 .5rem;font-size:.84rem;">Set what this token can do per mount. Read + write is offered only for writable mounts.</p>
            <div id="mcp-scope-rows"></div>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" type="button" id="mcp-cancel">Cancel</button>
          <button class="btn btn-primary" type="button" id="mcp-create">Create token</button>
        </div>
      </div>
      <div id="mcp-step-success" class="hidden">
        <div class="modal-body">
          <div class="success-state">
            <span class="check">
              <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
            </span>
            <h2>MCP access ready</h2>
            <p class="muted" style="margin:.25rem 0 0;">Give the agent the endpoint and token below.</p>
          </div>
          <div class="form-grid" style="margin-top:.5rem;">
            <label>Endpoint</label>
            <div class="share-url" data-test="mcp-endpoint-card"><code id="mcp-created-endpoint"></code></div>
            <label>Token</label>
            <div class="share-url" data-test="mcp-token-card">
              <code id="mcp-created-token"></code>
              <button class="btn btn-primary btn-sm" type="button" id="mcp-copy-token">
                <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                Copy
              </button>
            </div>
          </div>
          <p class="share-warning">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/></svg>
            The token is shown once and is not stored in plaintext. If you lose it, rotate the grant with <code>mvmt token rotate</code>.
          </p>
        </div>
        <div class="modal-foot">
          <button class="btn btn-primary" type="button" id="mcp-done">Done</button>
        </div>
      </div>
    </div>
  </div>

  <div id="properties-modal" class="modal hidden" role="dialog" aria-modal="true">
    <div class="modal-card">
      <div class="modal-head">
        <h2>Properties</h2>
        <button class="btn btn-ghost btn-icon" type="button" id="properties-modal-close" title="Close">
          <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
        </button>
      </div>
      <div class="modal-body">
        <div class="target-chip" id="properties-target"></div>
        <dl class="properties-grid">
          <dt>Type</dt><dd id="properties-type"></dd>
          <dt>Size</dt><dd id="properties-size"></dd>
          <dt>Permission</dt><dd id="properties-permission"></dd>
          <dt>Modified</dt><dd id="properties-modified"></dd>
          <dt>Path</dt><dd><code id="properties-path"></code></dd>
        </dl>
      </div>
      <div class="modal-foot">
        <button class="btn btn-primary" type="button" id="properties-done">Done</button>
      </div>
    </div>
  </div>

  <div id="mount-modal" class="modal hidden" role="dialog" aria-modal="true">
    <div class="modal-card">
      <div class="modal-head">
        <h2 id="mount-modal-title">Add source</h2>
        <button class="btn btn-ghost btn-icon" type="button" id="mount-modal-close" title="Close">
          <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-grid">
          <label for="mount-root">Local path</label>
          <div class="path-picker-row">
            <input id="mount-root" placeholder="/Users/you/Documents">
            <button class="btn btn-sm" id="mount-pick-folder" type="button">Choose folder</button>
            <button class="btn btn-sm" id="mount-pick-file" type="button">Choose file</button>
          </div>
          <p class="field-help">Type a path, or use Finder on this Mac.</p>
          <label for="mount-name">Name</label>
          <input id="mount-name" placeholder="auto">
          <label for="mount-path">Shared as</label>
          <input id="mount-path" placeholder="/documents">
          <label for="mount-description">Description</label>
          <input id="mount-description" placeholder="optional">
          <label>Permission</label>
          <label style="display:flex;align-items:center;gap:.4rem;font-size:.88rem;color:var(--text);"><input id="mount-write" type="checkbox" style="width:auto;"> Allow recipients to upload &amp; modify</label>
          <label id="mount-enabled-label">State</label>
          <label id="mount-enabled-field" style="display:flex;align-items:center;gap:.4rem;font-size:.88rem;color:var(--text);"><input id="mount-enabled" type="checkbox" checked style="width:auto;"> Enabled</label>
        </div>
      </div>
      <div class="modal-foot">
        <button type="button" class="btn" id="mount-cancel">Cancel</button>
        <button type="button" class="btn btn-primary" id="mount-save">Save source</button>
      </div>
    </div>
  </div>

</main>

<script>
(function () {
  var $ = function (id) { return document.getElementById(id); };
  var DASHBOARD_API_PREFIX = '/dashboard/api/';
  var APP_API_PREFIX = '/api/';

  function scrubDashboardUrl() {
    if (!location.search && !location.hash) return;
    history.replaceState(null, '', location.pathname);
  }

  scrubDashboardUrl();

  var ICONS = {
    folder: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>',
    file: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/></svg>',
    copy: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
    link: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/></svg>',
    trash: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M3 6h18"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>',
    edit: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4z"/></svg>',
    check: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><circle cx="12" cy="12" r="10"/><path d="M9 12l2 2 4-4"/></svg>',
    alert: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/></svg>',
    share: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.59 13.51l6.83 3.98M15.41 6.51l-6.82 3.98"/></svg>',
    refresh: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M3 12a9 9 0 0 1 15.5-6.4L21 8M21 3v5h-5M21 12a9 9 0 0 1-15.5 6.4L3 16M3 21v-5h5"/></svg>',
    up: '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M12 19V5M5 12l7-7 7 7"/></svg>',
  };

  function dashboardBasePath() {
    var path = location.pathname.replace(/\\/+$/, '');
    var marker = '/dashboard';
    var index = path.lastIndexOf(marker);
    return index >= 0 ? path.slice(0, index + marker.length) : marker;
  }

  function appBasePath() {
    var base = dashboardBasePath();
    var marker = '/dashboard';
    return base.endsWith(marker) ? base.slice(0, -marker.length) : '';
  }

  function dashboardRequestUrl(url) {
    if (typeof url === 'string' && url.indexOf(DASHBOARD_API_PREFIX) === 0) {
      return dashboardBasePath() + '/api/' + url.slice(DASHBOARD_API_PREFIX.length);
    }
    if (typeof url === 'string' && url.indexOf(APP_API_PREFIX) === 0) {
      return appBasePath() + url;
    }
    return url;
  }

  var state = {
    view: 'files',
    viewMode: 'list',
    currentPath: '/',
    fileEntries: [],
    selectedEntry: null,
    leases: [],
    mounts: [],
    canManageMounts: false,
    grants: [],
    canManageGrants: false,
    apps: [],
    appsLoaded: false,
    localOwner: false,
    status: null,
    activeTab: 'active',
    editingMount: null,
    contextEntry: null,
  };

  // ---------- core helpers ----------
  async function api(url, options) {
    var opts = options || {};
    var fetchOptions = {};
    for (var key in opts) {
      if (Object.prototype.hasOwnProperty.call(opts, key)) fetchOptions[key] = opts[key];
    }
    var headers = { 'Content-Type': 'application/json' };
    if (opts.headers) {
      for (var headerName in opts.headers) {
        if (Object.prototype.hasOwnProperty.call(opts.headers, headerName)) headers[headerName] = opts.headers[headerName];
      }
    }
    fetchOptions.headers = headers;
    var response = await fetch(dashboardRequestUrl(url), fetchOptions);
    var body = null;
    try { body = await response.json(); } catch (_) { /* no body */ }
    if (!response.ok) {
      var err = new Error((body && body.error) || ('request_failed_' + response.status));
      err.status = response.status;
      err.detail = body && body.detail;
      throw err;
    }
    return body || {};
  }

  function toast(message, kind, ttlMs) {
    var stack = $('toast-stack');
    var el = document.createElement('div');
    el.className = 'toast ' + (kind || 'info');
    var iconHtml = kind === 'error' ? ICONS.alert : kind === 'success' ? ICONS.check : ICONS.alert;
    el.innerHTML = iconHtml + '<span></span>';
    el.querySelector('span').textContent = message;
    stack.appendChild(el);
    var ms = ttlMs || (kind === 'error' ? 6000 : 3500);
    setTimeout(function () {
      el.style.transition = 'opacity .2s';
      el.style.opacity = '0';
      setTimeout(function () { if (el.parentNode) el.parentNode.removeChild(el); }, 200);
    }, ms);
  }

  function showError(error) {
    var msg = (error && error.message) ? error.message : String(error);
    toast(humanizeError(msg), 'error');
  }

  function humanizeError(msg) {
    switch (msg) {
      case 'invalid_credentials': return 'Wrong username or password.';
      case 'dashboard_login_required': return 'Please sign in.';
      case 'label_required': return 'Add a label first.';
      case 'path_required': return 'Pick a file or folder.';
      case 'invalid_mode': return 'Pick a permission.';
      case 'mode_requires_folder': return 'Upload modes only work on folders.';
      case 'mount_read_only': return 'Source is read-only — enable write to allow uploads.';
      case 'mount_management_disabled': return 'Mount management is disabled on this server.';
      case 'admin_required': return 'Admin access required for that action.';
      case 'local_dashboard_required': return 'Open the dashboard on this device to change sources.';
      case 'root_required': return 'Choose a local file or folder first.';
      case 'invalid_root': return 'That local path does not exist or cannot be opened.';
      case 'invalid_name': return 'Use a source name with letters, numbers, dashes, or underscores.';
      case 'invalid_path': return 'Shared path must start with /, like /photos.';
      case 'invalid_picker_kind': return 'Choose file or folder.';
      case 'picker_unavailable': return 'Finder picker is only available from the local Mac dashboard. You can still type the path.';
      case 'mount_failed': return 'Source could not be saved.';
      case 'lease_not_found': return 'Lease not found.';
      case 'lease_revoked': return 'Lease was revoked.';
      default: return msg;
    }
  }

  async function copyToClipboard(text) {
    if (!text) return;
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
      return;
    }
    var tmp = document.createElement('textarea');
    tmp.value = text;
    tmp.setAttribute('readonly', '');
    tmp.style.position = 'absolute';
    tmp.style.left = '-9999px';
    document.body.appendChild(tmp);
    tmp.select();
    try { document.execCommand('copy'); } finally { document.body.removeChild(tmp); }
  }

  function formatDate(iso) {
    if (!iso) return '';
    try { return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }); } catch (_) { return iso; }
  }

  function relativeTime(iso) {
    if (!iso) return '';
    var ms = Date.now() - new Date(iso).getTime();
    if (isNaN(ms)) return iso;
    var abs = Math.abs(ms);
    var future = ms < 0;
    var min = 60000, hr = 3600000, day = 86400000;
    var phrase;
    if (abs < min) phrase = 'just now';
    else if (abs < hr) phrase = Math.round(abs / min) + 'm';
    else if (abs < day) phrase = Math.round(abs / hr) + 'h';
    else if (abs < day * 30) phrase = Math.round(abs / day) + 'd';
    else phrase = formatDate(iso);
    if (phrase === 'just now') return phrase;
    return future ? 'in ' + phrase : phrase + ' ago';
  }

  function setIcon(el, key) {
    el.innerHTML = ICONS[key] || '';
  }

  function formatBytes(value) {
    if (!value) return '0 bytes';
    var units = ['bytes', 'KB', 'MB', 'GB', 'TB'];
    var size = Number(value);
    var unit = 0;
    while (size >= 1024 && unit < units.length - 1) { size = size / 1024; unit += 1; }
    var display = unit === 0 ? String(Math.round(size)) : size.toFixed(size >= 10 ? 1 : 2).replace(/\.0+$/, '');
    return display + ' ' + units[unit];
  }

  function permissionText(entry) {
    if (entry.unavailable) return 'Source is unavailable';
    return entry.writeAccess ? 'Writable source' : 'Read-only source';
  }

  // ---------- app shell ----------
  function showApp(signedIn) {
    $('login').classList.toggle('hidden', signedIn);
    $('app').classList.toggle('hidden', !signedIn);
    $('header-user').classList.toggle('hidden', !signedIn);
    if (!signedIn) {
      state.currentPath = '/';
      state.fileEntries = [];
      state.selectedEntry = null;
      state.leases = [];
      state.mounts = [];
      state.localOwner = false;
      state.status = null;
      state.contextEntry = null;
      $('files').replaceChildren();
      $('files-grid').replaceChildren();
      $('leases').replaceChildren();
      $('crumbs').replaceChildren();
      $('settings-nav').classList.add('hidden');
    }
  }

  function applyDashboardSession(resp) {
    $('who').textContent = (resp.user && resp.user.username) || '';
    $('who-role').classList.toggle('hidden', !(resp.user && resp.user.admin));
    state.localOwner = !!resp.localOwner;
    $('settings-nav').classList.toggle('hidden', !state.localOwner);
    if (!state.localOwner && state.view === 'settings') setView('files');
  }

  function pathSegments(input) {
    var parts = input.split('/').filter(Boolean);
    var segs = [{ name: 'All sources', path: '/' }];
    var cur = '';
    for (var i = 0; i < parts.length; i += 1) {
      cur += '/' + parts[i];
      segs.push({ name: parts[i], path: cur });
    }
    return segs;
  }

  function renderCrumbs() {
    var segs = pathSegments(state.currentPath);
    var nodes = [];
    for (var i = 0; i < segs.length; i += 1) {
      var seg = segs[i];
      var last = i === segs.length - 1;
      var node;
      if (last) { node = document.createElement('span'); node.className = 'current'; }
      else {
        node = document.createElement('a');
        node.href = '#';
        (function (target) { node.addEventListener('click', function (event) { event.preventDefault(); loadFiles(target).catch(showError); }); })(seg.path);
      }
      node.textContent = seg.name;
      nodes.push(node);
      if (!last) { var sep = document.createElement('span'); sep.className = 'sep'; sep.textContent = '/'; nodes.push(sep); }
    }
    $('crumbs').replaceChildren.apply($('crumbs'), nodes);
  }

  function setView(view) {
    state.view = view;
    var panels = {
      files: $('view-files-panel'),
      shares: $('view-shares-panel'),
      mcp: $('view-mcp-panel'),
      apps: $('view-apps-panel'),
      settings: $('view-settings-panel'),
    };
    for (var key in panels) {
      if (Object.prototype.hasOwnProperty.call(panels, key)) panels[key].classList.toggle('hidden', key !== view);
    }
    var nav = document.querySelectorAll('.nav-btn[data-view]');
    for (var i = 0; i < nav.length; i += 1) nav[i].classList.toggle('active', nav[i].getAttribute('data-view') === view);
    renderToolbar();
    if (view === 'apps' && !state.appsLoaded) loadApps();
  }

  function renderToolbar() {
    var files = state.view === 'files';
    var shares = state.view === 'shares';
    var mcp = state.view === 'mcp';
    var apps = state.view === 'apps';
    var settings = state.view === 'settings';
    $('crumbs').classList.toggle('hidden', !files);
    $('shares-title').classList.toggle('hidden', !shares);
    $('mcp-title').classList.toggle('hidden', !mcp);
    $('apps-title').classList.toggle('hidden', !apps);
    $('settings-title').classList.toggle('hidden', !settings);
    $('add-mount').classList.toggle('hidden', !(files && state.canManageMounts));
    $('new-grant').classList.toggle('hidden', !(mcp && state.canManageGrants));
    $('share-selected').classList.toggle('hidden', !(files && state.selectedEntry));
    $('view-list').classList.toggle('hidden', !files);
    $('view-grid').classList.toggle('hidden', !files);
  }

  function sourceMountForEntry(entry) {
    if (!entry || state.currentPath !== '/') return null;
    for (var i = 0; i < state.mounts.length; i += 1) {
      if (state.mounts[i].path === entry.path) return state.mounts[i];
    }
    return null;
  }

  function canManageSourceEntry(entry) {
    return !!(state.canManageMounts && sourceMountForEntry(entry));
  }

  function setFileViewMode(mode) {
    state.viewMode = mode;
    $('view-list').classList.toggle('active', mode === 'list');
    $('view-grid').classList.toggle('active', mode === 'grid');
    renderFileEntries();
  }

  // ---------- files ----------
  function selectEntry(entry) {
    state.selectedEntry = entry;
    var nodes = document.querySelectorAll('[data-file-path]');
    for (var i = 0; i < nodes.length; i += 1) nodes[i].classList.toggle('selected', nodes[i].getAttribute('data-file-path') === entry.path);
    renderToolbar();
  }

  function clearSelectedEntry() {
    state.selectedEntry = null;
    var nodes = document.querySelectorAll('[data-file-path]');
    for (var i = 0; i < nodes.length; i += 1) nodes[i].classList.remove('selected');
    renderToolbar();
  }

  function showContextMenu(entry, event) {
    event.preventDefault();
    state.contextEntry = entry;
    selectEntry(entry);
    var sourceEditable = canManageSourceEntry(entry);
    $('context-share').classList.toggle('hidden', !!entry.unavailable);
    $('context-rename-source').classList.toggle('hidden', !sourceEditable);
    $('context-remove-source').classList.toggle('hidden', !sourceEditable);
    var menu = $('context-menu');
    menu.classList.remove('hidden');
    var width = menu.offsetWidth || 180;
    var height = menu.offsetHeight || 92;
    var left = Math.min(event.clientX, window.innerWidth - width - 8);
    var top = Math.min(event.clientY, window.innerHeight - height - 8);
    menu.style.left = Math.max(8, left) + 'px';
    menu.style.top = Math.max(8, top) + 'px';
  }

  function hideContextMenu() {
    $('context-menu').classList.add('hidden');
    state.contextEntry = null;
  }

  function fileKind(entry) {
    return entry.type === 'directory' ? 'Folder' : 'File';
  }

  function entryModified(entry) {
    if (!entry.mtimeMs) return '';
    var iso = new Date(entry.mtimeMs).toISOString();
    return { short: relativeTime(iso), full: formatDate(iso) };
  }

  function attachFileEntryHandlers(element, entry) {
    element.addEventListener('click', function () { selectEntry(entry); });
    element.addEventListener('dblclick', function () {
      if (entry.type === 'directory' && !entry.unavailable) loadFiles(entry.path).catch(showError);
    });
    element.addEventListener('contextmenu', function (event) {
      showContextMenu(entry, event);
    });
  }

  function renderFileRow(entry) {
    var row = document.createElement('tr');
    row.setAttribute('data-path', entry.path);
    row.setAttribute('data-file-path', entry.path);
    row.classList.toggle('selected', state.selectedEntry && state.selectedEntry.path === entry.path);

    var nameCell = document.createElement('td');
    nameCell.setAttribute('data-label', 'Name');
    var nameWrap = document.createElement('div');
    nameWrap.className = 'cell-name';
    var iconSpan = document.createElement('span');
    iconSpan.innerHTML = entry.type === 'directory' ? ICONS.folder : ICONS.file;
    var nameText = document.createElement('span');
    nameText.className = 'name-text';
    nameText.title = entry.path;
    nameText.textContent = entry.name;
    nameWrap.append(iconSpan, nameText);
    nameCell.append(nameWrap);

    var kindCell = document.createElement('td');
    kindCell.setAttribute('data-label', 'Kind');
    kindCell.textContent = entry.unavailable ? 'Missing source' : fileKind(entry);

    var sizeCell = document.createElement('td');
    sizeCell.setAttribute('data-label', 'Size');
    sizeCell.className = 'cell-num';
    sizeCell.textContent = entry.type === 'directory' ? '--' : formatBytes(entry.size);

    var modCell = document.createElement('td');
    modCell.setAttribute('data-label', 'Modified');
    modCell.className = 'cell-num';
    var modified = entryModified(entry);
    if (entry.unavailable) modCell.textContent = 'Unavailable';
    else if (modified) { modCell.title = modified.full; modCell.textContent = modified.short; }

    attachFileEntryHandlers(row, entry);
    row.append(nameCell, sizeCell, kindCell, modCell);
    return row;
  }

  function renderFileCard(entry) {
    var card = document.createElement('button');
    card.type = 'button';
    card.className = 'file-card';
    card.setAttribute('data-file-path', entry.path);
    card.classList.toggle('selected', state.selectedEntry && state.selectedEntry.path === entry.path);
    card.title = entry.unavailable ? entry.path + ' is unavailable. Right-click to remove source.' : entry.path + ' (right-click for options)';
    card.innerHTML = (entry.type === 'directory' ? ICONS.folder : ICONS.file)
      + '<span class="file-card-name"></span><span class="file-card-meta"></span>';
    card.querySelector('.file-card-name').textContent = entry.name;
    card.querySelector('.file-card-meta').textContent = entry.unavailable ? 'Missing source' : entry.type === 'directory' ? 'Folder' : formatBytes(entry.size);
    attachFileEntryHandlers(card, entry);
    return card;
  }

  function renderFileEntries() {
    var entries = state.fileEntries || [];
    var hasEntries = entries.length > 0;
    var atRoot = state.currentPath === '/';
    if (atRoot) $('sources-count').textContent = entries.length ? '(' + entries.length + ')' : '';
    $('empty-mounts').classList.toggle('hidden', hasEntries || !atRoot);
    $('files-wrap').classList.toggle('hidden', !hasEntries || state.viewMode !== 'list');
    $('files-grid').classList.toggle('hidden', !hasEntries || state.viewMode !== 'grid');
    if (!hasEntries) {
      $('files').replaceChildren();
      $('files-grid').replaceChildren();
      return;
    }
    var rows = [];
    var cards = [];
    for (var i = 0; i < entries.length; i += 1) {
      rows.push(renderFileRow(entries[i]));
      cards.push(renderFileCard(entries[i]));
    }
    $('files').replaceChildren.apply($('files'), rows);
    $('files-grid').replaceChildren.apply($('files-grid'), cards);
  }

  async function loadFiles(targetPath) {
    hideContextMenu();
    var requested = targetPath || state.currentPath;
    var listing;
    try {
      listing = await api('/api/fs/list?path=' + encodeURIComponent(requested));
    } catch (err) {
      if (err && err.status === 404 && requested !== '/') {
        toast('Path is unavailable. Returned to top.', 'error');
        return loadFiles('/');
      }
      throw err;
    }
    state.currentPath = listing.path || requested;
    state.selectedEntry = null;
    state.fileEntries = listing.entries || [];
    renderCrumbs();
    renderFileEntries();
    renderToolbar();
  }

  // ---------- leases ----------
  var LEASE_URLS_KEY = 'mvmt_lease_urls_v1';
  function readLeaseUrls() { try { return JSON.parse(window.localStorage.getItem(LEASE_URLS_KEY) || '{}'); } catch (_) { return {}; } }
  function writeLeaseUrls(value) { try { window.localStorage.setItem(LEASE_URLS_KEY, JSON.stringify(value)); } catch (_) { /* ignore */ } }
  function getStoredLeaseUrl(id) { return readLeaseUrls()[id] || null; }
  function storeLeaseUrl(id, url) { var m = readLeaseUrls(); m[id] = url; writeLeaseUrls(m); }
  function forgetLeaseUrl(id) { var m = readLeaseUrls(); delete m[id]; writeLeaseUrls(m); }

  function leaseStatus(lease) {
    if (lease.revokedAt) return { state: 'revoked', label: 'Revoked' };
    if (lease.expiresAt && new Date(lease.expiresAt).getTime() <= Date.now()) return { state: 'expired', label: 'Expired' };
    if (lease.maxDownloads && (lease.downloadCount || 0) >= lease.maxDownloads) return { state: 'expired', label: 'Downloaded' };
    return { state: 'active', label: lease.expiresAt ? 'Expires in ' + relativeTime(lease.expiresAt).replace('in ', '').replace(' ago', '') : 'No expiry' };
  }

  function permissionLabel(perms) {
    var p = perms || ['read'];
    var hasRead = p.indexOf('read') !== -1;
    var hasUpload = p.indexOf('upload') !== -1;
    var hasWrite = p.indexOf('write') !== -1;
    if (hasWrite) return 'Read + write';
    if (hasRead && hasUpload) return 'Read + upload';
    if (hasUpload) return 'Upload only';
    return 'Read only';
  }

  function leaseActivityLabel(lease) {
    var perms = lease.permissions || ['read'];
    var hasRead = perms.indexOf('read') !== -1;
    var hasUpload = perms.indexOf('upload') !== -1;
    var downloads = lease.downloadCount || 0;
    var uploads = lease.uploadCount || 0;
    if (lease.maxDownloads) return downloads + '/' + lease.maxDownloads + ' downloads';
    if (hasUpload && !hasRead) return uploads + ' uploads';
    if (hasRead && hasUpload) return downloads + ' downloads, ' + uploads + ' uploads';
    return downloads + ' downloads';
  }

  function renderLeaseRow(lease) {
    var s = leaseStatus(lease);
    var row = document.createElement('tr');

    var labelCell = document.createElement('td');
    labelCell.setAttribute('data-label', 'Label');
    labelCell.textContent = lease.label || '(unlabeled)';
    labelCell.style.fontWeight = '500';

    var pathCell = document.createElement('td');
    pathCell.setAttribute('data-label', 'Path');
    pathCell.className = 'cell-path';
    var paths = (lease.resources || []).map(function (r) { return r.path; });
    pathCell.textContent = paths.join(', ') || lease.path || '';
    pathCell.title = pathCell.textContent;

    var permCell = document.createElement('td');
    permCell.setAttribute('data-label', 'Permission');
    var permBadge = document.createElement('span');
    var hasUpload = (lease.permissions || []).indexOf('upload') !== -1;
    var hasWrite = (lease.permissions || []).indexOf('write') !== -1;
    permBadge.className = 'badge ' + (hasWrite || hasUpload ? 'write' : 'read');
    permBadge.textContent = permissionLabel(lease.permissions);
    permCell.append(permBadge);

    var statusCell = document.createElement('td');
    statusCell.setAttribute('data-label', 'Status');
    var statusBadge = document.createElement('span');
    statusBadge.className = 'badge ' + s.state;
    statusBadge.textContent = s.label;
    statusBadge.title = lease.expiresAt ? formatDate(lease.expiresAt) : '';
    statusCell.append(statusBadge);
    if (lease.published === false) {
      var localBadge = document.createElement('span');
      localBadge.className = 'badge';
      localBadge.textContent = 'Local only';
      localBadge.title = 'Capability-only: reachable by local apps, not over the relay';
      localBadge.style.marginLeft = '.35rem';
      statusCell.append(localBadge);
    }

    var usedCell = document.createElement('td');
    usedCell.setAttribute('data-label', 'Activity');
    usedCell.className = 'cell-num';
    usedCell.textContent = leaseActivityLabel(lease);
    if (lease.lastUsedAt) usedCell.title = 'Last used ' + formatDate(lease.lastUsedAt);

    var actionCell = document.createElement('td');
    actionCell.setAttribute('data-label', 'Actions');
    var actions = document.createElement('div');
    actions.className = 'actions';
    if (s.state === 'active') {
      var copyBtn = document.createElement('button');
      copyBtn.className = 'btn btn-sm';
      copyBtn.type = 'button';
      var knownUrl = lease.url || getStoredLeaseUrl(lease.id);
      if (lease.url) storeLeaseUrl(lease.id, lease.url);
      copyBtn.innerHTML = (knownUrl ? ICONS.copy : ICONS.link) + '<span>' + (knownUrl ? 'Copy link' : 'Replace link') + '</span>';
      copyBtn.title = knownUrl ? 'Copy this lease URL' : 'No recoverable URL is saved. Replace this link to copy a new URL.';
      (function (id, serverUrl) {
        copyBtn.addEventListener('click', async function () {
          try {
            var url = serverUrl || getStoredLeaseUrl(id);
            if (!url) {
              if (!confirm('No recoverable URL is saved for this lease. Replace it now? Any previously shared URL for this lease will stop working.')) return;
              var rotated = await api('/dashboard/api/leases/' + encodeURIComponent(id) + '/rotate', { method: 'POST', body: '{}' });
              url = rotated.lease.url;
              storeLeaseUrl(id, url);
            }
            await copyToClipboard(url);
            toast('Link copied to clipboard.', 'success');
            renderLeases();
          } catch (err) { showError(err); }
        });
      })(lease.id, lease.url || null);
      actions.append(copyBtn);

      var publishBtn = document.createElement('button');
      publishBtn.className = 'btn btn-sm';
      publishBtn.type = 'button';
      var isPublished = lease.published !== false;
      publishBtn.textContent = isPublished ? 'Unpublish' : 'Publish';
      publishBtn.title = isPublished
        ? 'Block public tunnels; apps on this machine keep access'
        : 'Allow this lease through public tunnels';
      (function (id, publish) {
        publishBtn.addEventListener('click', async function () {
          try {
            await api('/dashboard/api/leases/' + encodeURIComponent(id) + '/publish', {
              method: 'POST',
              body: JSON.stringify({ published: publish }),
            });
            await loadLeases();
            toast(publish ? 'Lease published.' : 'Lease unpublished - apps on this machine keep access.', 'success');
          } catch (err) { showError(err); }
        });
      })(lease.id, !isPublished);
      actions.append(publishBtn);

      var revoke = document.createElement('button');
      revoke.className = 'btn btn-danger btn-sm btn-icon';
      revoke.type = 'button';
      revoke.title = 'Revoke';
      revoke.innerHTML = ICONS.trash;
      (function (id) {
        revoke.addEventListener('click', async function () {
          if (!confirm('Revoke this link? Anyone holding the URL will lose access immediately.')) return;
          try {
            await api('/dashboard/api/leases/' + encodeURIComponent(id) + '/revoke', { method: 'POST', body: '{}' });
            forgetLeaseUrl(id);
            await loadLeases();
            toast('Link revoked.', 'success');
          } catch (err) { showError(err); }
        });
      })(lease.id);
      actions.append(revoke);
    }
    actionCell.append(actions);

    row.append(labelCell, pathCell, permCell, statusCell, usedCell, actionCell);
    return row;
  }

  function renderLeases() {
    var items = [];
    for (var i = 0; i < state.leases.length; i += 1) {
      var lease = state.leases[i];
      var st = leaseStatus(lease).state;
      if (state.activeTab === 'active' ? st === 'active' : st !== 'active') items.push(lease);
    }
    $('leases-count').textContent = items.length ? '(' + items.length + ')' : '';
    var hasItems = items.length > 0;
    $('leases-empty').classList.toggle('hidden', hasItems);
    $('leases-wrap').classList.toggle('hidden', !hasItems);
    if (!hasItems) { $('leases').replaceChildren(); return; }
    var rows = [];
    for (var j = 0; j < items.length; j += 1) rows.push(renderLeaseRow(items[j]));
    $('leases').replaceChildren.apply($('leases'), rows);
  }

  async function loadLeases() {
    var payload = await api('/dashboard/api/leases');
    state.leases = payload.leases || [];
    renderLeases();
  }

  // ---------- properties modal ----------
  function openPropertiesModal(entry) {
    hideContextMenu();
    selectEntry(entry);
    $('properties-target').innerHTML = '<span>' + (entry.type === 'directory' ? ICONS.folder : ICONS.file) + '</span><div class="target-path">' + escapeHtml(entry.path) + '</div>';
    $('properties-type').textContent = entry.unavailable ? 'Missing source' : fileKind(entry);
    $('properties-size').textContent = entry.type === 'directory' ? 'Not calculated' : formatBytes(entry.size);
    $('properties-permission').textContent = permissionText(entry);
    $('properties-modified').textContent = entry.mtimeMs ? formatDate(new Date(entry.mtimeMs).toISOString()) : 'Unknown';
    $('properties-path').textContent = entry.path;
    $('properties-modal').classList.remove('hidden');
  }

  function closePropertiesModal() {
    $('properties-modal').classList.add('hidden');
  }

  // ---------- lease modal ----------
  function openLeaseModal(entry) {
    hideContextMenu();
    state.selectedEntry = entry;
    $('lease-step-config').classList.remove('hidden');
    $('lease-step-success').classList.add('hidden');
    var target = $('lease-target');
    target.innerHTML = '<span>' + (entry.type === 'directory' ? ICONS.folder : ICONS.file) + '</span><div class="target-path">' + escapeHtml(entry.path) + '</div><span class="badge ' + (entry.writeAccess ? 'write' : 'readonly') + '">' + (entry.type === 'directory' ? 'Folder' : 'File') + ' · ' + (entry.writeAccess ? 'writable' : 'read-only') + '</span>';
    $('lease-label').value = entry.name || entry.path;
    $('lease-expires').value = '24h';
    $('lease-downloads').value = '-1';
    configureLeaseModeSelect(entry);
    $('lease-modal').classList.remove('hidden');
    setTimeout(function () { $('lease-label').focus(); $('lease-label').select(); }, 0);
  }

  function closeLeaseModal() {
    $('lease-modal').classList.add('hidden');
    clearSelectedEntry();
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, function (c) { return ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]; });
  }

  function configureLeaseModeSelect(entry) {
    var select = $('lease-mode');
    var options = select.options;
    for (var i = 0; i < options.length; i += 1) {
      var option = options[i];
      option.disabled = option.value !== 'read' && (entry.type === 'file' || !entry.writeAccess);
    }
    select.value = 'read';
    updateModeHelp();
  }

  function updateModeHelp() {
    var mode = $('lease-mode').value;
    var modeNotes = {
      'read': 'Recipients can view and download. No uploads.',
      'upload': 'Recipients can drop files into the folder. Existing files are not visible or modifiable.',
      'two-way': 'Recipients can view, download, and add new files. Existing files are never overwritten.',
    };
    $('mode-help').textContent = modeNotes[mode] || '';
    $('lease-downloads').disabled = mode === 'upload';
    if (mode === 'upload') $('lease-downloads').value = '-1';
  }

  async function submitLease() {
    if (!state.selectedEntry) return;
    var mode = $('lease-mode').value;
    if (!mode) { toast('Pick a permission.', 'error'); return; }
    if (!$('lease-label').value.trim()) { toast('Add a label first.', 'error'); $('lease-label').focus(); return; }
    var downloads = $('lease-downloads').value.trim();
    if (!downloads) downloads = '-1';
    var btn = $('lease-create');
    btn.disabled = true;
    try {
      var payload = await api('/dashboard/api/leases', {
        method: 'POST',
        body: JSON.stringify({
          path: state.selectedEntry.path,
          label: $('lease-label').value.trim(),
          mode: mode,
          expires: $('lease-expires').value,
          maxDownloads: Number(downloads),
        }),
      });
      storeLeaseUrl(payload.lease.id, payload.lease.url);
      $('created-url').textContent = payload.lease.url;
      $('lease-step-config').classList.add('hidden');
      $('lease-step-success').classList.remove('hidden');
      await loadLeases();
    } catch (err) { showError(err); }
    finally { btn.disabled = false; }
  }

  // ---------- MCP access (clients[] API-token grants) ----------
  async function loadGrants() {
    var payload = await api('/dashboard/api/grants');
    state.grants = payload.grants || [];
    state.canManageGrants = !!payload.canManage;
    renderGrants();
  }

  async function loadApps() {
    try {
      var payload = await api('/dashboard/api/apps');
      state.apps = payload.apps || [];
      state.appsLoaded = true;
      renderApps();
    } catch (err) {
      state.apps = [];
      state.appsLoaded = false;
      renderApps();
      showError(err);
    }
  }

  function renderApps() {
    var grid = $('apps-grid');
    var empty = $('apps-empty');
    $('apps-count').textContent = state.apps.length ? '(' + state.apps.length + ')' : '';
    grid.innerHTML = '';
    if (state.apps.length === 0) {
      empty.textContent = state.appsLoaded ? 'No apps installed yet.' : 'Apps failed to load. Click Apps again to retry.';
      empty.classList.remove('hidden');
      return;
    }
    empty.classList.add('hidden');
    for (var i = 0; i < state.apps.length; i += 1) {
      var manifest = state.apps[i];
      var card = document.createElement('a');
      card.className = 'app-card';
      card.href = appBasePath() + '/apps/' + manifest.id;
      var title = document.createElement('div');
      title.className = 'app-card-title';
      title.textContent = manifest.label;
      var desc = document.createElement('div');
      desc.className = 'app-card-desc';
      desc.textContent = manifest.description || '';
      card.appendChild(title);
      card.appendChild(desc);
      grid.appendChild(card);
    }
  }

  function renderGrants() {
    $('grants-count').textContent = state.grants.length ? '(' + state.grants.length + ')' : '';
    $('new-grant').classList.toggle('hidden', !(state.view === 'mcp' && state.canManageGrants));
    var hasGrants = state.grants.length > 0;
    $('grants-empty').classList.toggle('hidden', hasGrants);
    $('grants-wrap').classList.toggle('hidden', !hasGrants);
    var body = $('grants');
    body.innerHTML = '';
    for (var i = 0; i < state.grants.length; i += 1) body.append(renderGrantRow(state.grants[i]));
  }

  function renderGrantRow(grant) {
    var row = document.createElement('tr');

    var labelCell = document.createElement('td');
    labelCell.setAttribute('data-label', 'Label');
    labelCell.textContent = grant.label || grant.id;
    labelCell.style.fontWeight = '500';

    var accessCell = document.createElement('td');
    accessCell.setAttribute('data-label', 'Access');
    accessCell.textContent = grant.scope || '';

    var reachCell = document.createElement('td');
    reachCell.setAttribute('data-label', 'Reach');
    var reachBadge = document.createElement('span');
    reachBadge.className = 'badge ' + (grant.published ? 'active' : '');
    reachBadge.textContent = grant.published ? 'Published' : 'Local only';
    reachBadge.title = grant.published
      ? 'Reachable through public tunnels'
      : 'Reachable by apps on this machine only';
    reachCell.append(reachBadge);

    var expiresCell = document.createElement('td');
    expiresCell.setAttribute('data-label', 'Expires');
    expiresCell.textContent = grant.expiresAt ? formatDate(grant.expiresAt) : 'Never';

    var usedCell = document.createElement('td');
    usedCell.setAttribute('data-label', 'Last used');
    usedCell.className = 'cell-num';
    if (grant.lastUsedAt) { usedCell.textContent = relativeTime(grant.lastUsedAt); usedCell.title = formatDate(grant.lastUsedAt); }
    else { usedCell.textContent = '—'; }

    var actionCell = document.createElement('td');
    actionCell.setAttribute('data-label', 'Actions');
    if (state.canManageGrants) {
      var actions = document.createElement('div');
      actions.className = 'actions';

      var publishBtn = document.createElement('button');
      publishBtn.className = 'btn btn-sm';
      publishBtn.type = 'button';
      publishBtn.textContent = grant.published ? 'Unpublish' : 'Publish';
      publishBtn.title = grant.published
        ? 'Block public tunnels; apps on this machine keep access'
        : 'Allow this token through public tunnels';
      (function (id, publish) {
        publishBtn.addEventListener('click', async function () {
          try {
            await api('/dashboard/api/grants/' + encodeURIComponent(id) + '/publish', {
              method: 'POST',
              body: JSON.stringify({ published: publish }),
            });
            await loadGrants();
            toast(publish ? 'Token published.' : 'Token unpublished - apps on this machine keep access.', 'success');
          } catch (err) { showError(err); }
        });
      })(grant.id, !grant.published);
      actions.append(publishBtn);

      var revokeBtn = document.createElement('button');
      revokeBtn.className = 'btn btn-danger btn-sm btn-icon';
      revokeBtn.type = 'button';
      revokeBtn.title = 'Revoke';
      revokeBtn.innerHTML = ICONS.trash;
      (function (id) {
        revokeBtn.addEventListener('click', async function () {
          if (!confirm('Revoke this MCP token? Any agent using it loses access immediately.')) return;
          try {
            await api('/dashboard/api/grants/' + encodeURIComponent(id), { method: 'DELETE', body: '{}' });
            await loadGrants();
            toast('Token revoked.', 'success');
          } catch (err) { showError(err); }
        });
      })(grant.id);
      actions.append(revokeBtn);

      actionCell.append(actions);
    }

    row.append(labelCell, accessCell, reachCell, expiresCell, usedCell, actionCell);
    return row;
  }

  function buildMcpScopeRows() {
    var container = $('mcp-scope-rows');
    container.innerHTML = '';
    for (var i = 0; i < state.mounts.length; i += 1) {
      (function (mount) {
        var row = document.createElement('div');
        row.className = 'mcp-scope-row';
        row.setAttribute('data-path', mount.path);
        row.style.display = 'flex';
        row.style.alignItems = 'center';
        row.style.gap = '.5rem';
        row.style.marginBottom = '.35rem';

        var name = document.createElement('span');
        name.textContent = mount.path;
        name.style.flex = '1';
        name.style.fontSize = '.88rem';

        var select = document.createElement('select');
        select.className = 'mcp-scope-mode';
        select.style.width = 'auto';
        var noneOpt = document.createElement('option');
        noneOpt.value = 'none';
        noneOpt.textContent = 'No access';
        var readOpt = document.createElement('option');
        readOpt.value = 'read';
        readOpt.textContent = 'Read';
        select.append(noneOpt, readOpt);
        if (mount.writeAccess) {
          var writeOpt = document.createElement('option');
          writeOpt.value = 'write';
          writeOpt.textContent = 'Read + write';
          select.append(writeOpt);
        }
        // Pre-fill with the mount's natural level.
        select.value = mount.writeAccess ? 'write' : 'read';

        row.append(name, select);
        container.append(row);
      })(state.mounts[i]);
    }
  }

  function syncMcpAccessMode() {
    var custom = document.querySelector('input[name=mcp-access-mode]:checked');
    $('mcp-custom-scopes').classList.toggle('hidden', !custom || custom.value !== 'custom');
  }

  function openMcpModal() {
    $('mcp-step-config').classList.remove('hidden');
    $('mcp-step-success').classList.add('hidden');
    $('mcp-label').value = '';
    $('mcp-expires').value = 'never';
    $('mcp-published').checked = false;
    var allRadio = document.querySelector('input[name=mcp-access-mode][value=all]');
    if (allRadio) allRadio.checked = true;
    buildMcpScopeRows();
    syncMcpAccessMode();
    $('mcp-modal').classList.remove('hidden');
    setTimeout(function () { $('mcp-label').focus(); }, 0);
  }

  function closeMcpModal() { $('mcp-modal').classList.add('hidden'); }

  async function submitMcpGrant() {
    if (!$('mcp-label').value.trim()) { toast('Add a label first.', 'error'); $('mcp-label').focus(); return; }
    var selected = document.querySelector('input[name=mcp-access-mode]:checked');
    var allMounts = !selected || selected.value === 'all';
    var requestBody = {
      label: $('mcp-label').value.trim(),
      expires: $('mcp-expires').value,
      published: $('mcp-published').checked,
      allMounts: allMounts,
    };
    if (!allMounts) {
      var rows = $('mcp-scope-rows').querySelectorAll('.mcp-scope-row');
      var scopes = [];
      for (var i = 0; i < rows.length; i += 1) {
        var mode = rows[i].querySelector('.mcp-scope-mode').value;
        if (mode === 'none') continue;
        scopes.push({ path: rows[i].getAttribute('data-path'), mode: mode });
      }
      if (scopes.length === 0) { toast('Give the token at least one mount.', 'error'); return; }
      requestBody.scopes = scopes;
    }
    var btn = $('mcp-create');
    btn.disabled = true;
    try {
      var payload = await api('/dashboard/api/grants', { method: 'POST', body: JSON.stringify(requestBody) });
      $('mcp-created-endpoint').textContent = payload.endpoint || '';
      $('mcp-created-token').textContent = payload.token || '';
      $('mcp-step-config').classList.add('hidden');
      $('mcp-step-success').classList.remove('hidden');
      await loadGrants();
    } catch (err) { showError(err); }
    finally { btn.disabled = false; }
  }

  // ---------- settings ----------
  function renderStatus() {
    if (!state.localOwner) return;
    var status = state.status || {};
    var server = status.server || {};
    var tunnel = status.tunnel || {};
    var rows = [];
    function kv(label, value) {
      rows.push('<div class="kv"><span>' + escapeHtml(label) + '</span><code>' + escapeHtml(value || 'Not configured') + '</code></div>');
    }
    kv('Local UI', server.localUrl || '');
    kv('Public UI', server.publicUrl || '');
    kv('Tunnel', tunnel.configured ? (tunnel.provider || 'configured') : 'Not configured');
    rows.push('<p class="muted" style="margin:.65rem 0 0;font-size:.84rem;">Tunnel reconfiguration is local-only. Use the interactive terminal menu for now: <code>tunnel config</code>.</p>');
    $('network-status').innerHTML = rows.join('');

    var logs = status.logs || [];
    if (!logs.length) {
      $('live-logs').innerHTML = '<div class="log-row"><span></span><span></span><span></span><span class="path">No dashboard events yet</span></div>';
      return;
    }
    var logRows = [];
    for (var i = logs.length - 1; i >= 0; i -= 1) {
      var entry = logs[i];
      var time = entry.ts ? new Date(entry.ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }) : '';
      logRows.push('<div class="log-row"><span>' + escapeHtml(time) + '</span><span>' + escapeHtml(String(entry.status || '')) + '</span><span>' + escapeHtml(entry.kind || '') + '</span><span class="path">' + escapeHtml((entry.method || '') + ' ' + (entry.path || '') + (entry.detail ? ' · ' + entry.detail : '')) + '</span></div>');
    }
    $('live-logs').innerHTML = logRows.join('');
  }

  async function loadStatus() {
    if (!state.localOwner) return;
    state.status = await api('/dashboard/api/status');
    renderStatus();
  }

  async function reloadMounts() {
    var payload = await api('/dashboard/api/mounts');
    state.mounts = payload.mounts || [];
    state.canManageMounts = !!payload.canManage;
    $('sources-count').textContent = state.mounts.length ? '(' + state.mounts.length + ')' : '';
    $('add-mount').classList.toggle('hidden', !(state.view === 'files' && state.canManageMounts));
    renderToolbar();
  }

  function openMountModal(existing) {
    state.editingMount = existing || null;
    $('mount-modal-title').textContent = existing ? 'Edit source: ' + existing.name : 'Add source';
    $('mount-root').value = existing ? (existing.root || '') : '';
    $('mount-name').value = existing ? (existing.name || '') : '';
    $('mount-name').disabled = !!existing;
    $('mount-path').value = existing ? (existing.path || '') : '';
    $('mount-description').value = existing ? (existing.description || '') : '';
    $('mount-write').checked = existing ? !!existing.writeAccess : false;
    $('mount-enabled').checked = existing ? existing.enabled !== false : true;
    $('mount-enabled-label').classList.toggle('hidden', !existing);
    $('mount-enabled-field').classList.toggle('hidden', !existing);
    $('mount-modal').classList.remove('hidden');
    setTimeout(function () { $('mount-root').focus(); }, 0);
  }

  function closeMountModal() { $('mount-modal').classList.add('hidden'); state.editingMount = null; }

  async function pickMountRoot(kind) {
    try {
      var payload = await api('/dashboard/api/local-path-picker', {
        method: 'POST',
        body: JSON.stringify({ kind: kind }),
      });
      if (payload.path) {
        $('mount-root').value = payload.path;
        $('mount-root').focus();
      }
    } catch (err) { showError(err); }
  }

  function normalizeSourceRename(value) {
    var trimmed = String(value || '').trim().replace(/^\\/+|\\/+$/g, '');
    if (!trimmed || trimmed.indexOf('/') !== -1) return null;
    if (!/^[A-Za-z0-9._~-]+$/.test(trimmed)) return null;
    return '/' + trimmed;
  }

  async function renameSource(entry) {
    hideContextMenu();
    var mount = sourceMountForEntry(entry);
    if (!mount) return;
    var current = (mount.path || entry.path || '').replace(/^\\/+/, '');
    var nextName = prompt('Rename source:', current);
    if (nextName === null) return;
    var nextPath = normalizeSourceRename(nextName);
    if (!nextPath) { toast('Use one source name, like photos or tax-docs.', 'error'); return; }
    if (nextPath === mount.path) return;
    for (var i = 0; i < state.mounts.length; i += 1) {
      if (state.mounts[i].name !== mount.name && state.mounts[i].path === nextPath) {
        toast('A source with that name already exists.', 'error');
        return;
      }
    }
    if (!confirm('Rename this source to ' + nextPath + '? Existing links that use the old path may stop working.')) return;
    try {
      await api('/dashboard/api/mounts/' + encodeURIComponent(mount.name), {
        method: 'PATCH',
        body: JSON.stringify({ path: nextPath }),
      });
      await reloadMounts();
      await loadFiles('/');
      toast('Source renamed.', 'success');
    } catch (err) { showError(err); }
  }

  async function removeSource(entry) {
    hideContextMenu();
    var mount = sourceMountForEntry(entry);
    if (!mount) return;
    if (!confirm('Remove source "' + (entry.name || mount.path) + '" from mvmt? This does not delete local files. Existing links that reference it become unusable.')) return;
    try {
      await api('/dashboard/api/mounts/' + encodeURIComponent(mount.name), { method: 'DELETE', body: '{}' });
      await reloadMounts();
      await loadFiles('/');
      toast('Source removed.', 'success');
    } catch (err) { showError(err); }
  }

  async function saveMount() {
    var root = $('mount-root').value.trim();
    if (!root) { toast('Choose a local file or folder first.', 'error'); $('mount-root').focus(); return; }
    var body = { root: root, writeAccess: $('mount-write').checked };
    if ($('mount-path').value.trim()) body.path = $('mount-path').value.trim();
    if ($('mount-description').value.trim()) body.description = $('mount-description').value.trim();
    var url = '/dashboard/api/mounts'; var method = 'POST';
    if (state.editingMount) {
      url = '/dashboard/api/mounts/' + encodeURIComponent(state.editingMount.name);
      method = 'PATCH';
      body.enabled = $('mount-enabled').checked;
    } else if ($('mount-name').value.trim()) body.name = $('mount-name').value.trim();
    var save = $('mount-save');
    save.disabled = true;
    try {
      await api(url, { method: method, body: JSON.stringify(body) });
      closeMountModal();
      await reloadMounts();
      await loadFiles(state.currentPath);
      toast(state.editingMount ? 'Source updated.' : 'Source added.', 'success');
    } catch (err) { showError(err); }
    finally { save.disabled = false; }
  }

  // ---------- refresh / boot ----------
  async function refresh() {
    await reloadMounts();
    await loadFiles(state.currentPath);
    await loadLeases();
    await loadGrants();
    await loadStatus();
  }

  // ---------- wiring ----------
  $('login-form').addEventListener('submit', async function (event) {
    event.preventDefault();
    var submit = event.target.querySelector('button[type=submit]');
    $('login-status').textContent = 'Signing in...';
    if (submit) submit.disabled = true;
    try {
      var resp = await api('/dashboard/api/login', {
        method: 'POST',
        body: JSON.stringify({ username: $('username').value, password: $('password').value }),
      });
      applyDashboardSession(resp);
      $('password').value = '';
      $('login-status').textContent = '';
      showApp(true);
      await refresh();
    } catch (err) {
      $('login-status').textContent = humanizeError((err && err.message) ? err.message : String(err));
      showError(err);
    }
    finally { if (submit) submit.disabled = false; }
  });

  $('logout').addEventListener('click', async function () {
    try { await api('/dashboard/api/logout', { method: 'POST', body: '{}' }); } catch (_) { /* ignore */ }
    $('who').textContent = '';
    showApp(false);
    toast('Signed out.', 'info');
  });

  $('refresh').addEventListener('click', function () { refresh().catch(showError); });
  $('share-selected').addEventListener('click', function () {
    if (state.selectedEntry) openLeaseModal(state.selectedEntry);
  });
  $('view-list').addEventListener('click', function () { setFileViewMode('list'); });
  $('view-grid').addEventListener('click', function () { setFileViewMode('grid'); });

  var navButtons = document.querySelectorAll('.nav-btn[data-view]');
  for (var n = 0; n < navButtons.length; n += 1) {
    (function (btn) {
      btn.addEventListener('click', function () {
        var view = btn.getAttribute('data-view') || 'files';
        if (view === 'settings' && !state.localOwner) return;
        setView(view);
      });
    })(navButtons[n]);
  }

  var tabButtons = document.querySelectorAll('.tab');
  for (var t = 0; t < tabButtons.length; t += 1) {
    (function (btn) {
      btn.addEventListener('click', function () {
        state.activeTab = btn.getAttribute('data-tab');
        for (var k = 0; k < tabButtons.length; k += 1) tabButtons[k].classList.toggle('active', tabButtons[k] === btn);
        renderLeases();
      });
    })(tabButtons[t]);
  }

  $('add-mount').addEventListener('click', function () { openMountModal(null); });
  $('mount-modal-close').addEventListener('click', closeMountModal);
  $('mount-cancel').addEventListener('click', closeMountModal);
  $('mount-save').addEventListener('click', function () { saveMount().catch(showError); });
  $('mount-pick-folder').addEventListener('click', function () { pickMountRoot('folder').catch(showError); });
  $('mount-pick-file').addEventListener('click', function () { pickMountRoot('file').catch(showError); });
  $('mount-modal').addEventListener('click', function (event) { if (event.target === $('mount-modal')) closeMountModal(); });

  $('context-menu').addEventListener('click', function (event) { event.stopPropagation(); });
  $('context-share').addEventListener('click', function () {
    var entry = state.contextEntry || state.selectedEntry;
    if (entry) openLeaseModal(entry);
  });
  $('context-properties').addEventListener('click', function () {
    var entry = state.contextEntry || state.selectedEntry;
    if (entry) openPropertiesModal(entry);
  });
  $('context-rename-source').addEventListener('click', function () {
    var entry = state.contextEntry || state.selectedEntry;
    if (entry) renameSource(entry).catch(showError);
  });
  $('context-remove-source').addEventListener('click', function () {
    var entry = state.contextEntry || state.selectedEntry;
    if (entry) removeSource(entry).catch(showError);
  });
  document.addEventListener('click', function (event) {
    if (!event.target.closest('#context-menu')) hideContextMenu();
  });
  $('lease-mode').addEventListener('change', updateModeHelp);
  $('lease-modal-close').addEventListener('click', closeLeaseModal);
  $('lease-cancel').addEventListener('click', closeLeaseModal);
  $('lease-done').addEventListener('click', closeLeaseModal);
  $('lease-create').addEventListener('click', function () { submitLease().catch(showError); });
  $('copy-url').addEventListener('click', async function () {
    try { await copyToClipboard($('created-url').textContent || ''); toast('Link copied to clipboard.', 'success'); }
    catch (err) { showError(err); }
  });
  $('lease-modal').addEventListener('click', function (event) { if (event.target === $('lease-modal')) closeLeaseModal(); });
  $('new-grant').addEventListener('click', function () { openMcpModal(); });
  $('mcp-modal-close').addEventListener('click', closeMcpModal);
  $('mcp-cancel').addEventListener('click', closeMcpModal);
  $('mcp-done').addEventListener('click', closeMcpModal);
  $('mcp-create').addEventListener('click', function () { submitMcpGrant().catch(showError); });
  var mcpAccessRadios = document.querySelectorAll('input[name=mcp-access-mode]');
  for (var mr = 0; mr < mcpAccessRadios.length; mr += 1) {
    mcpAccessRadios[mr].addEventListener('change', syncMcpAccessMode);
  }
  $('mcp-copy-token').addEventListener('click', async function () {
    try { await copyToClipboard($('mcp-created-token').textContent || ''); toast('Token copied to clipboard.', 'success'); }
    catch (err) { showError(err); }
  });
  $('mcp-modal').addEventListener('click', function (event) { if (event.target === $('mcp-modal')) closeMcpModal(); });
  $('properties-modal-close').addEventListener('click', closePropertiesModal);
  $('properties-done').addEventListener('click', closePropertiesModal);
  $('properties-modal').addEventListener('click', function (event) { if (event.target === $('properties-modal')) closePropertiesModal(); });
  document.addEventListener('keydown', function (event) {
    if (event.key === 'Escape') {
      if (!$('context-menu').classList.contains('hidden')) hideContextMenu();
      else if (!$('lease-modal').classList.contains('hidden')) closeLeaseModal();
      else if (!$('mcp-modal').classList.contains('hidden')) closeMcpModal();
      else if (!$('properties-modal').classList.contains('hidden')) closePropertiesModal();
      else if (!$('mount-modal').classList.contains('hidden')) closeMountModal();
    }
  });

  // boot
  api('/dashboard/api/me')
    .then(async function (resp) {
      applyDashboardSession(resp);
      showApp(true);
      await refresh();
    })
    .catch(function () { showApp(false); $('login').classList.remove('hidden'); });
})();
</script>
</body>
</html>`;

const LEASE_BROWSER_PAGE_HTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>mvmt folder lease</title>
<style>
:root{--bg:#f5f6f8;--panel:#fff;--text:#0f172a;--text-2:#475569;--muted:#94a3b8;--border:#e4e7ec;--border-strong:#cbd5e1;--hover:#f1f5f9;--accent:#0f766e;--accent-2:#115e59;--accent-soft:#ccfbf1;--ok:#15803d;--ok-soft:#dcfce7;--warn:#b45309;--warn-soft:#fef3c7;--danger:#b91c1c;--shadow:0 1px 2px rgba(15,23,42,.04),0 1px 3px rgba(15,23,42,.04)}
*,*::before,*::after{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif;-webkit-font-smoothing:antialiased}
.top{height:52px;background:#fff;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between;padding:0 1.25rem;position:sticky;top:0;z-index:5}
.brand{display:flex;align-items:center;gap:.5rem;font-weight:700}
.brand .dot{width:.55rem;height:.55rem;border-radius:50%;background:var(--accent)}
.badge{display:inline-flex;align-items:center;gap:.3rem;padding:.16rem .55rem;border-radius:999px;font-size:.7rem;font-weight:650;letter-spacing:.04em;text-transform:uppercase;border:1px solid #bbf7d0;background:var(--ok-soft);color:var(--ok)}
main{max-width:1120px;margin:0 auto;padding:1.25rem}
.hero{display:flex;align-items:flex-end;justify-content:space-between;gap:1rem;margin:.3rem 0 1rem}
h1{font-size:1.6rem;line-height:1.1;margin:0;font-weight:700;letter-spacing:-.02em}
.meta{margin:.45rem 0 0;color:var(--text-2)}
.shell{background:#fff;border:1px solid var(--border);border-radius:14px;box-shadow:var(--shadow);overflow:hidden}
.toolbar{display:flex;align-items:center;justify-content:space-between;gap:.75rem;padding:.7rem .85rem;border-bottom:1px solid var(--border);background:#fff}
.crumbs{display:flex;align-items:center;gap:.35rem;min-width:0;overflow-x:auto;white-space:nowrap;color:var(--text-2)}
.crumbs a{color:var(--text-2);text-decoration:none;border-radius:6px;padding:.18rem .35rem}
.crumbs a:hover{background:var(--hover);color:var(--text)}
.crumbs .current{font-weight:650;color:var(--text);padding:.18rem .35rem}
.crumbs .sep{color:var(--muted)}
.status{color:var(--text-2);font-size:.88rem;min-height:1.25rem}
.upload{border:1px dashed #94a3b8;border-radius:12px;margin:1rem;padding:1rem 1.1rem;background:#f8fafc;display:flex;align-items:center;justify-content:space-between;gap:1rem}
.upload.drag{border-color:var(--accent);background:#ecfdf5}
.upload.uploading{border-color:var(--accent);background:#ecfdf5;box-shadow:inset 0 0 0 1px rgba(15,118,110,.15)}
.upload strong{display:block;font-size:.95rem}
.hint{margin:.18rem 0 0;color:var(--text-2)}
.upload input{max-width:18rem}
.upload-progress{flex:1 1 16rem;min-width:14rem}
.upload-status{display:flex;align-items:center;gap:.55rem;font-weight:700;color:var(--text);margin-bottom:.55rem}
.spinner{width:1rem;height:1rem;border:2px solid #cbd5e1;border-top-color:var(--accent);border-radius:50%;animation:spin .8s linear infinite;flex:none}
.progress{height:.55rem;border-radius:999px;background:#e2e8f0;overflow:hidden}
.bar{height:100%;width:0%;border-radius:999px;background:var(--accent);transition:width .15s ease}
.progress.indeterminate .bar{width:35%;animation:progress-slide 1s ease-in-out infinite}
.upload-progress.done .spinner,.upload-progress.error .spinner{display:none}
.upload-progress.error .bar{background:var(--danger)}
@keyframes spin{to{transform:rotate(360deg)}}
@keyframes progress-slide{0%{transform:translateX(-120%)}100%{transform:translateX(320%)}}
.hidden{display:none!important}
.tablewrap{overflow:auto}
table{border-collapse:collapse;width:100%;min-width:640px}
th{font-size:.7rem;font-weight:700;color:var(--text-2);background:#fafbfc;text-transform:uppercase;letter-spacing:.05em;padding:.58rem .9rem;text-align:left;border-bottom:1px solid var(--border)}
td{padding:.68rem .9rem;border-bottom:1px solid var(--border);vertical-align:middle}
tbody tr:hover{background:var(--hover)}
.name{display:flex;align-items:center;gap:.55rem;min-width:0}
.icon{width:1.05rem;height:1.05rem;color:var(--muted);flex:none}
.name a{color:var(--text);font-weight:550;text-decoration:none;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.name a:hover{color:var(--accent)}
.kind,.bytes{color:var(--text-2)}
.bytes{text-align:right;font-variant-numeric:tabular-nums}
.empty{padding:2.5rem 1rem;text-align:center;color:var(--text-2)}
@media (max-width:640px){.top{padding:0 .85rem}.brand span:last-child{display:none}main{padding:.75rem}.hero{display:block}.shell{border-radius:10px}.toolbar{align-items:flex-start;flex-direction:column}.upload{align-items:flex-start;flex-direction:column}.upload input{max-width:100%}table{min-width:0}table,thead,tbody,tr,td{display:block;width:100%}thead{display:none}tbody{display:flex;flex-direction:column;gap:.55rem;padding:.7rem;background:var(--bg)}tr{background:#fff;border:1px solid var(--border);border-radius:10px;padding:.7rem}td{border:0;padding:.25rem 0}.bytes{text-align:left}.kind::before,.bytes::before{content:attr(data-label);display:inline-block;width:4.5rem;color:var(--text-2);font-size:.72rem;font-weight:700;text-transform:uppercase;letter-spacing:.04em}}
</style>
</head>
<body>
<header class="top">
  <div class="brand"><span class="dot"></span><span>mvmt</span></div>
  <span class="badge" id="lease-kind">Folder lease</span>
</header>
<main>
  <section class="hero">
    <div>
      <h1 id="title">Folder lease</h1>
      <p class="meta" id="meta"></p>
    </div>
    <p class="status" id="status"></p>
  </section>
  <section class="shell">
    <div class="toolbar">
      <nav class="crumbs" id="crumbs" aria-label="Folder path"></nav>
    </div>
  <div class="upload hidden" id="upload">
    <div><strong>Upload to this folder</strong><p class="hint">Choose files or drop them here. Existing filenames are saved with a suffix.</p></div>
    <input id="upload-files" type="file" multiple>
    <div class="upload-progress hidden" id="upload-progress" aria-live="polite">
      <div class="upload-status"><span class="spinner"></span><span id="upload-status-text">Uploading...</span></div>
      <div class="progress" id="upload-progress-track"><div class="bar" id="upload-progress-bar"></div></div>
    </div>
  </div>
    <div class="tablewrap">
      <table><thead><tr><th>Name</th><th>Type</th><th style="text-align:right">Size</th></tr></thead><tbody id="entries"></tbody></table>
    </div>
    <div class="empty hidden" id="empty">No files in this folder.</div>
  </section>
</main>
  <script>
const pathParts = location.pathname.split('/').filter(Boolean);
// When served through a relay the path is /t/{slug}/lease/{id}; locally
// it is just /lease/{id}. Anchor on 'lease' so the id and the prefix
// before it stay correct in both shapes.
const leaseIdx = pathParts.indexOf('lease');
const leaseId = leaseIdx >= 0 ? decodeURIComponent(pathParts[leaseIdx + 1] || '') : '';
const basePrefix = leaseIdx > 0 ? '/' + pathParts.slice(0, leaseIdx).join('/') : '';
const params = new URLSearchParams(location.search);
const tokenKey = 'mvmt_lease_token_' + leaseId;
let token = params.get('token') || params.get('t') || sessionStorage.getItem(tokenKey) || '';
if (token) sessionStorage.setItem(tokenKey, token);
const requestedPath = params.get('path') || '';
if (params.has('token') || params.has('t')) {
  const clean = new URL(location.href);
  clean.searchParams.delete('token');
  clean.searchParams.delete('t');
  history.replaceState(null, '', clean.pathname + clean.search);
}
const title = document.getElementById('title');
const meta = document.getElementById('meta');
  const crumbs = document.getElementById('crumbs');
  const status = document.getElementById('status');
  const entries = document.getElementById('entries');
  const empty = document.getElementById('empty');
  const upload = document.getElementById('upload');
  const uploadInput = document.getElementById('upload-files');
  const uploadProgress = document.getElementById('upload-progress');
  const uploadStatusText = document.getElementById('upload-status-text');
  const uploadProgressTrack = document.getElementById('upload-progress-track');
  const uploadProgressBar = document.getElementById('upload-progress-bar');
  let currentListingPath = '/';

function pageUrl(nextPath) {
  const url = new URL(basePrefix + '/lease/' + encodeURIComponent(leaseId), location.origin);
  if (nextPath && nextPath !== '/') url.searchParams.set('path', nextPath);
  return url.pathname + url.search;
}

  function fileUrl(entryPath) {
  const encodedPath = entryPath.split('/').filter(Boolean).map(encodeURIComponent).join('/');
  const url = new URL(basePrefix + '/lease/' + encodeURIComponent(leaseId) + '/files/' + encodedPath, location.origin);
  if (token) url.searchParams.set('token', token);
  return url.pathname + url.search;
}

function parentPath(inputPath) {
  const parts = inputPath.split('/').filter(Boolean);
  parts.pop();
  return parts.length === 0 ? '/' : '/' + parts.join('/');
}

function renderCrumbs(inputPath) {
  const parts = (inputPath || '/').split('/').filter(Boolean);
  const nodes = [];
  const root = document.createElement(parts.length ? 'a' : 'span');
  root.textContent = 'Shared folder';
  root.className = parts.length ? '' : 'current';
  if (parts.length) root.href = pageUrl('/');
  nodes.push(root);
  let current = '';
  for (const part of parts) {
    current += '/' + part;
    const sep = document.createElement('span');
    sep.className = 'sep';
    sep.textContent = '/';
    const node = document.createElement(current === inputPath ? 'span' : 'a');
    node.textContent = part;
    node.className = current === inputPath ? 'current' : '';
    if (current !== inputPath) node.href = pageUrl(current);
    nodes.push(sep, node);
  }
  crumbs.replaceChildren(...nodes);
}

function uploadUrl(fileName) {
  const parts = currentListingPath.split('/').filter(Boolean);
  parts.push(fileName);
  const encodedPath = parts.map(encodeURIComponent).join('/');
  const url = new URL(basePrefix + '/lease/' + encodeURIComponent(leaseId) + '/files/' + encodedPath, location.origin);
  if (token) url.searchParams.set('token', token);
  return url.pathname + url.search;
}

function formatBytes(value) {
  if (!value) return '0 bytes';
  const units = ['bytes', 'KB', 'MB', 'GB', 'TB'];
  let size = Number(value);
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size = size / 1024;
    unit += 1;
  }
  const display = unit === 0 ? String(Math.round(size)) : size.toFixed(size >= 10 ? 1 : 2).replace(/\\.0+$/, '');
  return display + ' ' + units[unit];
}

function iconFor(type) {
  if (type === 'directory') {
    return '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>';
  }
  return '<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/></svg>';
}

async function uploadFailureMessage(response, fileName) {
  let detail = '';
  try { detail = await response.text(); } catch (_) { /* no response body */ }
  if (response.status === 413) return 'Upload failed: ' + fileName + ' is too large for this public relay.';
  if (response.status === 401) return 'Upload failed: this link token is missing or invalid.';
  if (response.status === 403) return 'Upload failed: this link does not allow uploads.';
  if (response.status === 404) return 'Upload failed: the shared folder is unavailable.';
  if (response.status === 409) return 'Upload failed: a file with that name already exists.';
  return 'Upload failed: ' + fileName + ' (HTTP ' + response.status + (detail ? ' - ' + detail.slice(0, 120) : '') + ')';
}

function uploadNetworkFailureMessage() {
  return 'Upload failed before reaching mvmt. Check the relay or network connection.';
}

function showUploadProgress(message, percent) {
  upload.classList.add('uploading');
  uploadProgress.classList.remove('hidden', 'done', 'error');
  uploadStatusText.textContent = message;
  status.textContent = '';
  if (typeof percent === 'number' && Number.isFinite(percent)) {
    uploadProgressTrack.classList.remove('indeterminate');
    uploadProgressBar.style.width = Math.max(4, Math.min(100, percent)) + '%';
  } else {
    uploadProgressTrack.classList.add('indeterminate');
    uploadProgressBar.style.width = '';
  }
}

function showUploadDone(message) {
  upload.classList.remove('uploading');
  uploadProgress.classList.remove('hidden', 'error');
  uploadProgress.classList.add('done');
  uploadProgressTrack.classList.remove('indeterminate');
  uploadProgressBar.style.width = '100%';
  uploadStatusText.textContent = message;
  status.textContent = '';
}

function showUploadError(message) {
  upload.classList.remove('uploading');
  uploadProgress.classList.remove('hidden', 'done');
  uploadProgress.classList.add('error');
  uploadProgressTrack.classList.remove('indeterminate');
  uploadProgressBar.style.width = '100%';
  uploadStatusText.textContent = message;
  status.textContent = '';
}

function xhrResponse(xhr) {
  return {
    status: xhr.status,
    text: async function() { return xhr.responseText || ''; },
  };
}

async function loadListing() {
  const url = new URL(basePrefix + '/lease/' + encodeURIComponent(leaseId) + '/files', location.origin);
  if (requestedPath) url.searchParams.set('path', requestedPath);
  if (token) url.searchParams.set('token', token);
  const response = await fetch(url);
  if (!response.ok) {
    status.textContent = response.status === 401 ? 'Invalid or missing lease token.' : 'Folder is unavailable.';
    return;
  }

	  const listing = await response.json();
	  currentListingPath = listing.path || '/';
	  title.textContent = listing.label || 'Folder lease';
	  meta.textContent = (listing.path || '/') + (listing.expiresAt ? ' - expires ' + new Date(listing.expiresAt).toLocaleString() : '');
	  upload.classList.toggle('hidden', !listing.canUpload);
  renderCrumbs(listing.path || '/');
  empty.classList.toggle('hidden', listing.entries.length > 0);
  entries.replaceChildren(...listing.entries.map((entry) => {
    const row = document.createElement('tr');
    const nameCell = document.createElement('td');
    const nameWrap = document.createElement('div');
    nameWrap.className = 'name';
    nameWrap.innerHTML = iconFor(entry.type);
    const link = document.createElement('a');
    link.href = entry.type === 'directory' ? pageUrl(entry.path) : fileUrl(entry.path);
    link.textContent = entry.name;
    nameWrap.append(link);
    nameCell.append(nameWrap);
    const typeCell = document.createElement('td');
    typeCell.className = 'kind';
    typeCell.setAttribute('data-label', 'Type');
    typeCell.textContent = entry.type === 'directory' ? 'Folder' : 'File';
    const sizeCell = document.createElement('td');
    sizeCell.className = 'bytes';
    sizeCell.setAttribute('data-label', 'Size');
    sizeCell.textContent = entry.type === 'directory' ? '--' : formatBytes(entry.size);
    row.append(nameCell, typeCell, sizeCell);
    return row;
	  }));
	}

	async function uploadFiles(fileList) {
	  const files = Array.from(fileList || []);
	  if (files.length === 0) return;
	  let uploaded = 0;
	  for (const file of files) {
	    try {
	      await uploadFile(file);
	      uploaded += 1;
	    } catch (error) {
	      showUploadError(error instanceof Error ? error.message : uploadNetworkFailureMessage());
	      return;
	    }
	  }
	  uploadInput.value = '';
	  showUploadDone(uploaded === 1 ? 'Upload complete.' : 'Uploaded ' + uploaded + ' files.');
	  await loadListing();
	}

	function uploadFile(file) {
	  return new Promise((resolve, reject) => {
	    const xhr = new XMLHttpRequest();
	    xhr.open('PUT', uploadUrl(file.name));
	    xhr.upload.addEventListener('loadstart', () => {
	      showUploadProgress('Uploading ' + file.name + '...', undefined);
	    });
	    xhr.upload.addEventListener('progress', (event) => {
	      if (event.lengthComputable && event.total > 0) {
	        const percent = Math.round((event.loaded / event.total) * 100);
	        showUploadProgress('Uploading ' + file.name + ' (' + percent + '%)', percent);
	      } else {
	        showUploadProgress('Uploading ' + file.name + '...', undefined);
	      }
	    });
	    xhr.addEventListener('load', () => {
	      if (xhr.status >= 200 && xhr.status < 300) {
	        showUploadProgress('Finishing ' + file.name + '...', 100);
	        resolve();
	        return;
	      }
	      uploadFailureMessage(xhrResponse(xhr), file.name).then((message) => {
	        reject(new Error(message));
	      }, () => {
	        reject(new Error('Upload failed: ' + file.name));
	      });
	    });
	    xhr.addEventListener('error', () => reject(new Error(uploadNetworkFailureMessage())));
	    xhr.addEventListener('abort', () => reject(new Error('Upload canceled: ' + file.name)));
	    showUploadProgress('Uploading ' + file.name + '...', undefined);
	    xhr.send(file);
	  });
	}

	uploadInput.addEventListener('change', () => uploadFiles(uploadInput.files).catch(() => {
	  showUploadError(uploadNetworkFailureMessage());
	}));
	upload.addEventListener('dragover', (event) => {
	  event.preventDefault();
	  upload.classList.add('drag');
	});
	upload.addEventListener('dragleave', () => {
	  upload.classList.remove('drag');
	});
	upload.addEventListener('drop', (event) => {
	  event.preventDefault();
	  upload.classList.remove('drag');
	  uploadFiles(event.dataTransfer.files).catch(() => {
	    showUploadError(uploadNetworkFailureMessage());
	  });
	});

loadListing().catch(() => {
  status.textContent = 'Folder is unavailable.';
});
</script>
</body>
</html>`;

const LEASE_UPLOAD_PAGE_HTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>mvmt upload lease</title>
<style>
:root{--bg:#f5f6f8;--panel:#fff;--text:#0f172a;--text-2:#475569;--muted:#94a3b8;--border:#e4e7ec;--accent:#0f766e;--accent-2:#115e59;--ok:#15803d;--ok-soft:#dcfce7;--danger:#b91c1c;--shadow:0 1px 2px rgba(15,23,42,.04),0 1px 3px rgba(15,23,42,.04)}
*,*::before,*::after{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif;-webkit-font-smoothing:antialiased}
.top{height:52px;background:#fff;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between;padding:0 1.25rem;position:sticky;top:0;z-index:5}
.brand{display:flex;align-items:center;gap:.5rem;font-weight:700}
.brand .dot{width:.55rem;height:.55rem;border-radius:50%;background:var(--accent)}
.badge{display:inline-flex;align-items:center;padding:.16rem .55rem;border-radius:999px;font-size:.7rem;font-weight:650;letter-spacing:.04em;text-transform:uppercase;border:1px solid #bbf7d0;background:var(--ok-soft);color:var(--ok)}
main{max-width:720px;margin:0 auto;padding:1.25rem}
.hero{margin:.3rem 0 1rem}
h1{font-size:1.6rem;line-height:1.1;margin:0;font-weight:700;letter-spacing:-.02em}
.meta{margin:.45rem 0 0;color:var(--text-2)}
.card{background:#fff;border:1px solid var(--border);border-radius:14px;box-shadow:var(--shadow);overflow:hidden}
.drop{border:1px dashed #94a3b8;border-radius:12px;margin:1rem;padding:2rem;text-align:center;background:#f8fafc}
.drop.drag{border-color:var(--accent);background:#ecfdf5}
.drop.uploading{border-color:var(--accent);background:#ecfdf5;box-shadow:inset 0 0 0 1px rgba(15,118,110,.15)}
.drop strong{display:block;font-size:1rem;margin-bottom:.35rem}
.status{color:var(--text-2);margin:.65rem 0 0}
.upload-progress{max-width:28rem;margin:1rem auto 0;text-align:left}
.upload-status{display:flex;align-items:center;gap:.55rem;font-weight:700;color:var(--text);margin-bottom:.55rem}
.spinner{width:1rem;height:1rem;border:2px solid #cbd5e1;border-top-color:var(--accent);border-radius:50%;animation:spin .8s linear infinite;flex:none}
.progress{height:.55rem;border-radius:999px;background:#e2e8f0;overflow:hidden}
.bar{height:100%;width:0%;border-radius:999px;background:var(--accent);transition:width .15s ease}
.progress.indeterminate .bar{width:35%;animation:progress-slide 1s ease-in-out infinite}
.upload-progress.done .spinner,.upload-progress.error .spinner{display:none}
.upload-progress.error .bar{background:var(--danger)}
@keyframes spin{to{transform:rotate(360deg)}}
@keyframes progress-slide{0%{transform:translateX(-120%)}100%{transform:translateX(320%)}}
input{max-width:100%}
@media (max-width:640px){.top{padding:0 .85rem}.brand span:last-child{display:none}main{padding:.75rem}.card{border-radius:10px}.drop{margin:.75rem;padding:1.25rem}}
</style>
</head>
<body>
<header class="top">
  <div class="brand"><span class="dot"></span><span>mvmt</span></div>
  <span class="badge">Upload only</span>
</header>
<main>
  <section class="hero">
    <h1>Upload files</h1>
    <p class="meta">Upload-only lease</p>
  </section>
  <section class="card">
    <div class="drop">
      <strong>Upload to this folder</strong>
      <input id="files" type="file" multiple>
      <div class="upload-progress hidden" id="upload-progress" aria-live="polite">
        <div class="upload-status"><span class="spinner"></span><span id="upload-status-text">Uploading...</span></div>
        <div class="progress" id="upload-progress-track"><div class="bar" id="upload-progress-bar"></div></div>
      </div>
      <p class="status" id="status">Choose files or drop them here.</p>
    </div>
  </section>
</main>
<script>
const pathParts = location.pathname.split('/').filter(Boolean);
// Mirror the lease browser page: anchor on 'lease' so the id and any
// relay prefix in front of it (/t/{slug}) are both captured correctly.
const leaseIdx = pathParts.indexOf('lease');
const leaseId = leaseIdx >= 0 ? decodeURIComponent(pathParts[leaseIdx + 1] || '') : '';
const basePrefix = leaseIdx > 0 ? '/' + pathParts.slice(0, leaseIdx).join('/') : '';
const params = new URLSearchParams(location.search);
const tokenKey = 'mvmt_lease_token_' + leaseId;
let token = params.get('token') || params.get('t') || sessionStorage.getItem(tokenKey) || '';
if (token) sessionStorage.setItem(tokenKey, token);
if (params.has('token') || params.has('t')) {
  const clean = new URL(location.href);
  clean.searchParams.delete('token');
  clean.searchParams.delete('t');
  history.replaceState(null, '', clean.pathname + clean.search);
}
const drop = document.querySelector('.drop');
const input = document.getElementById('files');
const status = document.getElementById('status');
const uploadProgress = document.getElementById('upload-progress');
const uploadStatusText = document.getElementById('upload-status-text');
const uploadProgressTrack = document.getElementById('upload-progress-track');
const uploadProgressBar = document.getElementById('upload-progress-bar');
function uploadPath(name) {
  const url = new URL(basePrefix + '/lease/' + encodeURIComponent(leaseId) + '/files/' + encodeURIComponent(name), location.origin);
  if (token) url.searchParams.set('token', token);
  return url.pathname + url.search;
}
async function uploadFailureMessage(response, fileName) {
  let detail = '';
  try { detail = await response.text(); } catch (_) { /* no response body */ }
  if (response.status === 413) return 'Upload failed: ' + fileName + ' is too large for this public relay.';
  if (response.status === 401) return 'Upload failed: this link token is missing or invalid.';
  if (response.status === 403) return 'Upload failed: this link does not allow uploads.';
  if (response.status === 404) return 'Upload failed: the shared folder is unavailable.';
  if (response.status === 409) return 'Upload failed: a file with that name already exists.';
  return 'Upload failed: ' + fileName + ' (HTTP ' + response.status + (detail ? ' - ' + detail.slice(0, 120) : '') + ')';
}
function uploadNetworkFailureMessage() {
  return 'Upload failed before reaching mvmt. Check the relay or network connection.';
}
function showUploadProgress(message, percent) {
  drop.classList.add('uploading');
  uploadProgress.classList.remove('hidden', 'done', 'error');
  uploadStatusText.textContent = message;
  status.textContent = '';
  if (typeof percent === 'number' && Number.isFinite(percent)) {
    uploadProgressTrack.classList.remove('indeterminate');
    uploadProgressBar.style.width = Math.max(4, Math.min(100, percent)) + '%';
  } else {
    uploadProgressTrack.classList.add('indeterminate');
    uploadProgressBar.style.width = '';
  }
}
function showUploadDone(message) {
  drop.classList.remove('uploading');
  uploadProgress.classList.remove('hidden', 'error');
  uploadProgress.classList.add('done');
  uploadProgressTrack.classList.remove('indeterminate');
  uploadProgressBar.style.width = '100%';
  uploadStatusText.textContent = message;
  status.textContent = '';
}
function showUploadError(message) {
  drop.classList.remove('uploading');
  uploadProgress.classList.remove('hidden', 'done');
  uploadProgress.classList.add('error');
  uploadProgressTrack.classList.remove('indeterminate');
  uploadProgressBar.style.width = '100%';
  uploadStatusText.textContent = message;
  status.textContent = '';
}
function xhrResponse(xhr) {
  return {
    status: xhr.status,
    text: async function() { return xhr.responseText || ''; },
  };
}
async function uploadFiles(files) {
  const selectedFiles = Array.from(files || []);
  if (selectedFiles.length === 0) return;
  let uploaded = 0;
  for (const file of selectedFiles) {
    try {
      await uploadFile(file);
      uploaded += 1;
    } catch (error) {
      showUploadError(error instanceof Error ? error.message : uploadNetworkFailureMessage());
      return;
    }
  }
  showUploadDone(uploaded === 1 ? 'Upload complete.' : 'Uploaded ' + uploaded + ' files.');
  input.value = '';
}
function uploadFile(file) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('PUT', uploadPath(file.name));
    xhr.upload.addEventListener('loadstart', () => {
      showUploadProgress('Uploading ' + file.name + '...', undefined);
    });
    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable && event.total > 0) {
        const percent = Math.round((event.loaded / event.total) * 100);
        showUploadProgress('Uploading ' + file.name + ' (' + percent + '%)', percent);
      } else {
        showUploadProgress('Uploading ' + file.name + '...', undefined);
      }
    });
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        showUploadProgress('Finishing ' + file.name + '...', 100);
        resolve();
        return;
      }
      uploadFailureMessage(xhrResponse(xhr), file.name).then((message) => {
        reject(new Error(message));
      }, () => {
        reject(new Error('Upload failed: ' + file.name));
      });
    });
    xhr.addEventListener('error', () => reject(new Error(uploadNetworkFailureMessage())));
    xhr.addEventListener('abort', () => reject(new Error('Upload canceled: ' + file.name)));
    showUploadProgress('Uploading ' + file.name + '...', undefined);
    xhr.send(file);
  });
}
input.addEventListener('change', () => uploadFiles(input.files).catch(() => {
  showUploadError(uploadNetworkFailureMessage());
}));
drop.addEventListener('dragover', (event) => {
  event.preventDefault();
  drop.classList.add('drag');
});
drop.addEventListener('dragleave', () => {
  drop.classList.remove('drag');
});
drop.addEventListener('drop', (event) => {
  event.preventDefault();
  drop.classList.remove('drag');
  uploadFiles(event.dataTransfer.files).catch(() => {
    showUploadError(uploadNetworkFailureMessage());
  });
});
</script>
</body>
</html>`;

function requestClientHint(req: Request, oauthClientId?: string): string | undefined {
  const values = [
    oauthClientId,
    firstHeaderValue(req.headers['user-agent']),
  ].filter((value): value is string => Boolean(value));
  return values.length > 0 ? values.join(' ') : undefined;
}

type AuthorizeApproval =
  | { ok: true; mvmtClientId?: string; mvmtClientCredentialVersion?: number }
  | { ok: false; phase: string; message: string };

function resolveAuthorizeApproval(
  body: Record<string, unknown>,
  clients: readonly ClientConfig[],
  tokenPath: string | undefined,
  clientHint?: string,
): AuthorizeApproval {
  const apiTokenRaw = stringField(body.api_token);
  if (apiTokenRaw) {
    const client = findApiTokenClient(apiTokenRaw, clients, clientHint);
    if (client) {
      return {
        ok: true,
        mvmtClientId: client.id,
        mvmtClientCredentialVersion: clientCredentialVersion(client),
      };
    }
    return {
      ok: false,
      phase: 'deny_invalid_api_token',
      message: 'Invalid API token. Try again.',
    };
  }

  const hasPolicy = clients.length > 0;
  if (hasPolicy) {
    return {
      ok: false,
      phase: 'deny_missing_api_token',
      message: 'Enter a scoped API token. Create one with mvmt token add.',
    };
  }

  const sessionTokenRaw = stringField(body.session_token);
  if (sessionTokenRaw && validateSessionToken(`Bearer ${sessionTokenRaw}`, tokenPath)) {
    return { ok: true };
  }
  return {
    ok: false,
    phase: 'deny_invalid_session_token',
    message: 'Invalid API token. Try again.',
  };
}

function findApiTokenClient(token: string, clients: readonly ClientConfig[], clientHint?: string): ClientConfig | undefined {
  for (const client of clients) {
    if (client.auth.type !== 'token') continue;
    if (isExpired(client.expiresAt)) continue;
    if (!clientBindingMatches(client.clientBinding, clientHint)) continue;
    if (verifyApiToken(token, client.auth.tokenHash)) return client;
  }
  return undefined;
}

function oauthGrantMatchesCurrentClient(token: AccessToken, clients: readonly ClientConfig[], clientHint?: string): boolean {
  if (!token.mvmtClientId) return true;
  const client = clients.find((entry) => entry.id === token.mvmtClientId && entry.auth.type === 'token');
  return Boolean(
    client
      && !isExpired(client.expiresAt)
      && clientCredentialVersion(client) === (token.mvmtClientCredentialVersion ?? 1)
      && clientBindingMatches(client.clientBinding, clientHint),
  );
}

// Used in request logs for redirect/resource URIs. We log the host only
// (not the full path) because full redirect URLs can include
// client-supplied identifiers that belong in an audit trail but not in
// a general request log.
function safeHost(uri: string): string {
  try {
    return new URL(uri).host || '(unknown)';
  } catch {
    return '(invalid)';
  }
}

function formatAuthorizeLogDetail(input: {
  phase: string;
  requestId: string;
  redirectUri: string;
  resource?: string;
  resourceDefaulted?: boolean;
  state?: string;
  authorizedClientId?: string;
}): string {
  const authorizedClient = input.authorizedClientId ? ` authorized_client=${input.authorizedClientId}` : '';
  return `${input.phase} rid=${input.requestId} redirect_host=${safeHost(input.redirectUri)} resource_host=${input.resource ? safeHost(input.resource) : '(none)'} resource_defaulted=${input.resourceDefaulted ? 'true' : 'false'} state_hash=${hashForLog(input.state)}${authorizedClient}`;
}

function hashForLog(value: string | undefined): string {
  if (!value) return '(none)';
  return createHash('sha256').update(value).digest('hex').slice(0, 12);
}

export function buildOriginCheck(extraAllowed: string[]): (req: Request) => boolean {
  const allowed = new Set<string>(extraAllowed.map((origin) => origin.toLowerCase()));

  return (req) => {
    const rawOrigin = req.headers.origin;
    const origin = Array.isArray(rawOrigin) ? rawOrigin[0] : rawOrigin;
    if (!origin) return true;

    if (allowed.has(origin.toLowerCase())) return true;

    try {
      const parsed = new URL(origin);
      const host = parsed.hostname.toLowerCase();
      if (host === 'localhost' || host === '127.0.0.1' || host === '::1' || host === '[::1]') {
        return true;
      }
    } catch {
      return false;
    }
    return false;
  };
}

function originMatchesBaseUrl(req: Request, baseUrl: string): boolean {
  const origin = firstHeaderValue(req.headers.origin);
  if (!origin) return true;
  const requestOrigin = normalizedOrigin(origin);
  const baseOrigin = normalizedOrigin(baseUrl);
  return requestOrigin !== undefined && requestOrigin === baseOrigin;
}

function normalizedOrigin(value: string): string | undefined {
  try {
    const parsed = new URL(value);
    return `${parsed.protocol}//${parsed.host}`.toLowerCase();
  } catch {
    return undefined;
  }
}
