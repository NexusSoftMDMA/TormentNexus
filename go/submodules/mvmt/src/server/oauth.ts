import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import type { Request } from 'express';

export type CodeChallengeMethod = 'S256';

export interface AuthorizationCode {
  code: string;
  clientId: string;
  mvmtClientId?: string;
  mvmtClientCredentialVersion?: number;
  redirectUri: string;
  resource?: string;
  codeChallenge: string;
  codeChallengeMethod: CodeChallengeMethod;
  scope?: string;
  expiresAt: number;
  consumed: boolean;
}

export interface AccessToken {
  token: string;
  clientId: string;
  mvmtClientId?: string;
  mvmtClientCredentialVersion?: number;
  scope?: string;
  // RFC 8707 audience — the resource this token was minted for. Set
  // to the `resource` parameter the client sent during the authorize
  // flow (copied from the auth code to the access token). Validated
  // on every incoming request; a token minted for resource A cannot
  // be replayed against resource B even if both share the signing key.
  audience?: string;
  expiresAt: number;
}

export interface RefreshToken {
  token: string;
  id: string;
  familyId: string;
  clientId: string;
  mvmtClientId?: string;
  mvmtClientCredentialVersion?: number;
  scope?: string;
  audience?: string;
  expiresAt: number;
}

export interface TokenGrant {
  accessToken: AccessToken;
  refreshToken: RefreshToken;
}

export interface IssueCodeInput {
  clientId: string;
  mvmtClientId?: string;
  mvmtClientCredentialVersion?: number;
  redirectUri: string;
  resource?: string;
  codeChallenge: string;
  codeChallengeMethod: CodeChallengeMethod;
  scope?: string;
}

export interface ConsumeCodeInput {
  code: string;
  clientId: string;
  redirectUri: string;
  resource?: string;
  codeVerifier: string;
}

export interface ConsumeRefreshTokenInput {
  refreshToken: string;
  clientId: string;
  scope?: string;
}

export interface OAuthStoreOptions {
  codeTtlMs?: number;
  tokenTtlMs?: number;
  refreshTokenTtlMs?: number;
  maxRegisteredClients?: number;
  maxRedirectUrisPerClient?: number;
  // Key material used to sign and validate self-contained access tokens.
  // Pass a function when the key can change at runtime (e.g. backed by a
  // file that internal session-token rotation rewrites) so rotation invalidates
  // outstanding tokens without requiring a server restart.
  signingKey?: string | (() => string);
  clientsPath?: string;
  refreshTokensPath?: string;
  now?: () => number;
}

interface ActiveRefreshTokenRecord {
  familyId: string;
  tokenId: string;
  clientId: string;
  expiresAt: number;
}

const DEFAULT_CODE_TTL_MS = 10 * 60 * 1000;
const DEFAULT_TOKEN_TTL_MS = 24 * 60 * 60 * 1000;
const DEFAULT_REFRESH_TOKEN_TTL_MS = 30 * 24 * 60 * 60 * 1000;
const DEFAULT_MAX_REGISTERED_CLIENTS = 100;
const DEFAULT_MAX_REDIRECT_URIS_PER_CLIENT = 10;
const MAX_CLIENT_ID_LENGTH = 128;
const MAX_REDIRECT_URI_LENGTH = 2048;
const ACCESS_TOKEN_PREFIX = 'mvmtv1';
const REFRESH_TOKEN_PREFIX = 'mvmtr1';

export interface RegisteredClient {
  clientId: string;
  redirectUris: string[];
}

export class OAuthClientAlreadyRegisteredError extends Error {}
export class OAuthClientPersistenceError extends Error {}
export class OAuthClientRegistryLimitError extends Error {}

export class OAuthStore {
  private readonly codes = new Map<string, AuthorizationCode>();
  private readonly clients = new Map<string, RegisteredClient>();
  private readonly codeTtlMs: number;
  private readonly tokenTtlMs: number;
  private readonly refreshTokenTtlMs: number;
  private readonly maxRegisteredClients: number;
  private readonly maxRedirectUrisPerClient: number;
  private readonly resolveSigningKey: () => string;
  private readonly clientsPath?: string;
  private readonly refreshTokensPath?: string;
  private readonly now: () => number;
  private readonly activeRefreshTokens = new Map<string, ActiveRefreshTokenRecord>();
  private readonly revokedRefreshTokenFamilies = new Set<string>();

  constructor(options: OAuthStoreOptions = {}) {
    this.codeTtlMs = options.codeTtlMs ?? DEFAULT_CODE_TTL_MS;
    this.tokenTtlMs = options.tokenTtlMs ?? DEFAULT_TOKEN_TTL_MS;
    this.refreshTokenTtlMs = options.refreshTokenTtlMs ?? DEFAULT_REFRESH_TOKEN_TTL_MS;
    this.maxRegisteredClients = options.maxRegisteredClients ?? DEFAULT_MAX_REGISTERED_CLIENTS;
    this.maxRedirectUrisPerClient = options.maxRedirectUrisPerClient ?? DEFAULT_MAX_REDIRECT_URIS_PER_CLIENT;
    const keyOption = options.signingKey;
    if (typeof keyOption === 'function') {
      this.resolveSigningKey = keyOption;
    } else if (typeof keyOption === 'string') {
      const staticKey = keyOption;
      this.resolveSigningKey = () => staticKey;
    } else {
      const ephemeral = crypto.randomBytes(32).toString('base64url');
      this.resolveSigningKey = () => ephemeral;
    }
    this.clientsPath = options.clientsPath;
    this.refreshTokensPath = options.refreshTokensPath;
    this.now = options.now ?? Date.now;
    this.loadClientsFromDisk();
    this.loadRefreshTokensFromDisk();
  }

  get tokenTtlSeconds(): number {
    return Math.floor(this.tokenTtlMs / 1000);
  }

  get refreshTokenTtlSeconds(): number {
    return Math.floor(this.refreshTokenTtlMs / 1000);
  }

  registerClient(client: RegisteredClient): RegisteredClient {
    const normalized = this.normalizeRegisteredClient(client);
    this.assertClientCanBeRegistered(normalized);
    this.clients.set(normalized.clientId, normalized);
    try {
      this.persistClients();
    } catch (err) {
      this.clients.delete(normalized.clientId);
      throw err;
    }
    return normalized;
  }

  getClient(clientId: string): RegisteredClient | undefined {
    return this.clients.get(clientId);
  }

  isRedirectUriAllowed(clientId: string, redirectUri: string): boolean {
    const client = this.clients.get(clientId);
    if (!client) return false;
    return client.redirectUris.includes(redirectUri);
  }

  private normalizeRegisteredClient(client: RegisteredClient): RegisteredClient {
    return {
      clientId: client.clientId,
      redirectUris: [...new Set(client.redirectUris.filter((uri) => typeof uri === 'string' && uri.length > 0))],
    };
  }

  private assertClientCanBeRegistered(client: RegisteredClient): void {
    if (client.clientId.length > MAX_CLIENT_ID_LENGTH) {
      throw new OAuthClientRegistryLimitError(`client_id exceeds ${MAX_CLIENT_ID_LENGTH} characters`);
    }
    if (client.redirectUris.length > this.maxRedirectUrisPerClient) {
      throw new OAuthClientRegistryLimitError(`redirect_uris exceeds ${this.maxRedirectUrisPerClient} entries`);
    }
    if (client.redirectUris.some((uri) => uri.length > MAX_REDIRECT_URI_LENGTH)) {
      throw new OAuthClientRegistryLimitError(`redirect_uri exceeds ${MAX_REDIRECT_URI_LENGTH} characters`);
    }
    if (this.clients.has(client.clientId)) {
      throw new OAuthClientAlreadyRegisteredError('OAuth client_id is already registered');
    }
    if (this.clients.size >= this.maxRegisteredClients) {
      throw new OAuthClientRegistryLimitError(`OAuth client registry is limited to ${this.maxRegisteredClients} clients`);
    }
  }

  private loadClientsFromDisk(): void {
    if (!this.clientsPath) return;
    let raw: string;
    try {
      raw = fs.readFileSync(this.clientsPath, 'utf-8');
    } catch {
      return;
    }
    try {
      const parsed = JSON.parse(raw) as unknown;
      if (!Array.isArray(parsed)) return;
      for (const entry of parsed) {
        if (!entry || typeof entry !== 'object') continue;
        const clientId = (entry as RegisteredClient).clientId;
        const redirectUris = (entry as RegisteredClient).redirectUris;
        if (typeof clientId !== 'string' || !Array.isArray(redirectUris)) continue;
        const cleaned = redirectUris.filter((uri) => typeof uri === 'string' && uri.length > 0);
        if (cleaned.length === 0) continue;
        this.clients.set(clientId, { clientId, redirectUris: [...new Set(cleaned)] });
      }
    } catch {
      // Ignore corrupt registry; next registration rewrites the file.
    }
  }

  private persistClients(): void {
    if (!this.clientsPath) return;
    const data = JSON.stringify([...this.clients.values()]);
    try {
      fs.mkdirSync(path.dirname(this.clientsPath), { recursive: true });
      fs.writeFileSync(this.clientsPath, data, { mode: 0o600 });
      if (process.platform !== 'win32') {
        fs.chmodSync(this.clientsPath, 0o600);
      }
    } catch (err) {
      throw new OAuthClientPersistenceError(
        err instanceof Error ? err.message : 'Failed to persist OAuth client registry',
      );
    }
  }

  private loadRefreshTokensFromDisk(): void {
    if (!this.refreshTokensPath) return;
    let raw: string;
    try {
      raw = fs.readFileSync(this.refreshTokensPath, 'utf-8');
    } catch {
      return;
    }
    try {
      const parsed = JSON.parse(raw) as unknown;
      if (!parsed || typeof parsed !== 'object') return;
      const active = (parsed as { active?: unknown }).active;
      if (Array.isArray(active)) {
        for (const entry of active) {
          if (!entry || typeof entry !== 'object') continue;
          const candidate = entry as Partial<ActiveRefreshTokenRecord>;
          if (
            typeof candidate.familyId !== 'string'
            || typeof candidate.tokenId !== 'string'
            || typeof candidate.clientId !== 'string'
            || typeof candidate.expiresAt !== 'number'
            || !Number.isFinite(candidate.expiresAt)
          ) {
            continue;
          }
          if (candidate.expiresAt >= this.now()) {
            this.activeRefreshTokens.set(candidate.familyId, {
              familyId: candidate.familyId,
              tokenId: candidate.tokenId,
              clientId: candidate.clientId,
              expiresAt: candidate.expiresAt,
            });
          }
        }
      }
      const revokedFamilies = (parsed as { revokedFamilies?: unknown }).revokedFamilies;
      if (Array.isArray(revokedFamilies)) {
        for (const familyId of revokedFamilies) {
          if (typeof familyId === 'string' && familyId.length > 0) {
            this.revokedRefreshTokenFamilies.add(familyId);
          }
        }
      }
    } catch {
      // Ignore corrupt refresh state; replayed refresh tokens fail closed.
    }
  }

  private persistRefreshTokens(): void {
    if (!this.refreshTokensPath) return;
    const data = JSON.stringify({
      active: [...this.activeRefreshTokens.values()],
      revokedFamilies: [...this.revokedRefreshTokenFamilies],
    });
    try {
      fs.mkdirSync(path.dirname(this.refreshTokensPath), { recursive: true });
      const tempPath = `${this.refreshTokensPath}.${process.pid}.${crypto.randomBytes(6).toString('hex')}.tmp`;
      fs.writeFileSync(tempPath, data, { mode: 0o600 });
      if (process.platform !== 'win32') {
        fs.chmodSync(tempPath, 0o600);
      }
      fs.renameSync(tempPath, this.refreshTokensPath);
    } catch (err) {
      throw new OAuthError(
        'server_error',
        err instanceof Error ? err.message : 'Failed to persist OAuth refresh-token state',
      );
    }
  }

  issueCode(input: IssueCodeInput): AuthorizationCode {
    const code = crypto.randomBytes(32).toString('base64url');
    const entry: AuthorizationCode = {
      code,
      clientId: input.clientId,
      mvmtClientId: input.mvmtClientId,
      mvmtClientCredentialVersion: input.mvmtClientCredentialVersion,
      redirectUri: input.redirectUri,
      resource: input.resource,
      codeChallenge: input.codeChallenge,
      codeChallengeMethod: input.codeChallengeMethod,
      scope: input.scope,
      expiresAt: this.now() + this.codeTtlMs,
      consumed: false,
    };
    this.codes.set(code, entry);
    return entry;
  }

  consumeCode(input: ConsumeCodeInput): AccessToken {
    return this.exchangeCode(input).accessToken;
  }

  exchangeCode(input: ConsumeCodeInput): TokenGrant {
    const entry = this.codes.get(input.code);
    if (!entry) throw new OAuthError('invalid_grant', 'Authorization code not found');

    if (entry.consumed) {
      this.codes.delete(input.code);
      throw new OAuthError('invalid_grant', 'Authorization code already used');
    }
    if (entry.expiresAt < this.now()) {
      this.codes.delete(input.code);
      throw new OAuthError('invalid_grant', 'Authorization code expired');
    }
    if (entry.clientId !== input.clientId) {
      throw new OAuthError('invalid_grant', 'Client mismatch for authorization code');
    }
    if (entry.redirectUri !== input.redirectUri) {
      throw new OAuthError('invalid_grant', 'Redirect URI mismatch for authorization code');
    }
    // The authorization code is already bound to the resource chosen at
    // /authorize time. Clients may repeat that same resource at /token,
    // but they do not need to. We only reject when the token request
    // tries to introduce a different resource than the one already bound
    // into the code.
    if (input.resource !== undefined && entry.resource !== input.resource) {
      throw new OAuthError('invalid_grant', 'Resource mismatch for authorization code');
    }
    if (!verifyPkce(entry.codeChallenge, entry.codeChallengeMethod, input.codeVerifier)) {
      throw new OAuthError('invalid_grant', 'PKCE verification failed');
    }

    entry.consumed = true;
    this.codes.delete(input.code);

    return this.issueTokenGrant({
      clientId: entry.clientId,
      mvmtClientId: entry.mvmtClientId,
      mvmtClientCredentialVersion: entry.mvmtClientCredentialVersion,
      scope: entry.scope,
      audience: entry.resource,
    });
  }

  exchangeRefreshToken(input: ConsumeRefreshTokenInput): TokenGrant {
    const refreshToken = this.parseRefreshToken(input.refreshToken);
    if (!refreshToken) {
      throw new OAuthError('invalid_grant', 'Refresh token not found');
    }
    if (refreshToken.clientId !== input.clientId) {
      throw new OAuthError('invalid_grant', 'Client mismatch for refresh token');
    }
    this.validateActiveRefreshToken(refreshToken);

    return this.issueTokenGrant({
      clientId: refreshToken.clientId,
      mvmtClientId: refreshToken.mvmtClientId,
      mvmtClientCredentialVersion: refreshToken.mvmtClientCredentialVersion,
      scope: resolveRefreshScope(refreshToken.scope, input.scope),
      audience: refreshToken.audience,
      refreshFamilyId: refreshToken.familyId,
    });
  }

  issueTokenGrant(input: { clientId: string; mvmtClientId?: string; mvmtClientCredentialVersion?: number; scope?: string; audience?: string; refreshFamilyId?: string }): TokenGrant {
    return {
      accessToken: this.issueAccessToken(input),
      refreshToken: this.issueRefreshToken(input),
    };
  }

  issueAccessToken(input: { clientId: string; mvmtClientId?: string; mvmtClientCredentialVersion?: number; scope?: string; audience?: string }): AccessToken {
    const expiresAt = this.now() + this.tokenTtlMs;
    const token = this.issueSignedToken(ACCESS_TOKEN_PREFIX, {
      clientId: input.clientId,
      mvmtClientId: input.mvmtClientId,
      mvmtClientCredentialVersion: input.mvmtClientCredentialVersion,
      scope: input.scope,
      aud: input.audience,
      expiresAt,
    });

    return {
      token,
      clientId: input.clientId,
      mvmtClientId: input.mvmtClientId,
      mvmtClientCredentialVersion: input.mvmtClientCredentialVersion,
      scope: input.scope,
      audience: input.audience,
      expiresAt,
    };
  }

  issueRefreshToken(input: { clientId: string; mvmtClientId?: string; mvmtClientCredentialVersion?: number; scope?: string; audience?: string; refreshFamilyId?: string }): RefreshToken {
    const expiresAt = this.now() + this.refreshTokenTtlMs;
    const id = crypto.randomBytes(16).toString('base64url');
    const familyId = input.refreshFamilyId ?? crypto.randomBytes(16).toString('base64url');
    const token = this.issueSignedToken(REFRESH_TOKEN_PREFIX, {
      jti: id,
      familyId,
      clientId: input.clientId,
      mvmtClientId: input.mvmtClientId,
      mvmtClientCredentialVersion: input.mvmtClientCredentialVersion,
      scope: input.scope,
      aud: input.audience,
      expiresAt,
    });
    const previousActive = this.activeRefreshTokens.get(familyId);
    this.activeRefreshTokens.set(familyId, {
      familyId,
      tokenId: id,
      clientId: input.clientId,
      expiresAt,
    });
    try {
      this.persistRefreshTokens();
    } catch (err) {
      if (previousActive) {
        this.activeRefreshTokens.set(familyId, previousActive);
      } else {
        this.activeRefreshTokens.delete(familyId);
      }
      throw err;
    }

    return {
      token,
      id,
      familyId,
      clientId: input.clientId,
      mvmtClientId: input.mvmtClientId,
      mvmtClientCredentialVersion: input.mvmtClientCredentialVersion,
      scope: input.scope,
      audience: input.audience,
      expiresAt,
    };
  }

  validateAccessToken(
    authHeader: string | undefined,
    options: { expectedAudience?: string; allowLegacyNoAudience?: boolean } = {},
  ): AccessToken | undefined {
    if (!authHeader) return undefined;
    const parts = authHeader.split(' ');
    if (parts.length !== 2 || parts[0] !== 'Bearer') return undefined;
    const provided = parts[1];
    const parsed = this.parseAccessToken(provided);
    if (!parsed) return undefined;
    if (options.expectedAudience !== undefined) {
      // Reject tokens whose audience does not match the resource this
      // request is for. Per RFC 8707 / OpenAI Apps SDK, the resource
      // parameter supplied at /authorize time is bound to the token
      // via its `aud` claim.
      //
      // For compatibility, callers can explicitly allow legacy tokens
      // minted before `aud` binding existed. That keeps existing OAuth
      // sessions working across upgrades while still enforcing audience
      // checks for all newly issued tokens.
      if (parsed.audience === undefined) {
        if (!options.allowLegacyNoAudience) return undefined;
      } else if (parsed.audience !== options.expectedAudience) {
        return undefined;
      }
    }
    return parsed;
  }

  cleanup(): void {
    const now = this.now();
    for (const [code, entry] of this.codes) {
      if (entry.expiresAt < now || entry.consumed) this.codes.delete(code);
    }
    let refreshStateChanged = false;
    for (const [familyId, entry] of this.activeRefreshTokens) {
      if (entry.expiresAt < now) {
        this.activeRefreshTokens.delete(familyId);
        refreshStateChanged = true;
      }
    }
    if (refreshStateChanged) {
      try {
        this.persistRefreshTokens();
      } catch {
        // Expired entries were already removed in memory; future writes retry persistence.
      }
    }
  }

  private parseAccessToken(token: string): AccessToken | undefined {
    const parsed = this.parseSignedToken(token, ACCESS_TOKEN_PREFIX);
    if (!parsed) return undefined;
    return {
      token,
      clientId: parsed.clientId,
      mvmtClientId: parsed.mvmtClientId,
      mvmtClientCredentialVersion: parsed.mvmtClientCredentialVersion,
      scope: parsed.scope,
      audience: parsed.aud,
      expiresAt: parsed.expiresAt,
    };
  }

  private parseRefreshToken(token: string): RefreshToken | undefined {
    const parsed = this.parseSignedToken(token, REFRESH_TOKEN_PREFIX);
    if (!parsed) return undefined;
    if (!parsed.jti || !parsed.familyId) return undefined;
    return {
      token,
      id: parsed.jti,
      familyId: parsed.familyId,
      clientId: parsed.clientId,
      mvmtClientId: parsed.mvmtClientId,
      mvmtClientCredentialVersion: parsed.mvmtClientCredentialVersion,
      scope: parsed.scope,
      audience: parsed.aud,
      expiresAt: parsed.expiresAt,
    };
  }

  private validateActiveRefreshToken(refreshToken: RefreshToken): void {
    const active = this.activeRefreshTokens.get(refreshToken.familyId);
    if (this.revokedRefreshTokenFamilies.has(refreshToken.familyId)) {
      throw new OAuthError('invalid_grant', 'Refresh token family was revoked');
    }
    if (!active) {
      throw new OAuthError('invalid_grant', 'Refresh token not found');
    }
    if (active.expiresAt < this.now()) {
      this.activeRefreshTokens.delete(refreshToken.familyId);
      this.persistRefreshTokens();
      throw new OAuthError('invalid_grant', 'Refresh token expired');
    }
    if (active.clientId !== refreshToken.clientId) {
      this.revokeRefreshTokenFamily(refreshToken.familyId);
      throw new OAuthError('invalid_grant', 'Client mismatch for refresh token');
    }
    if (active.tokenId !== refreshToken.id) {
      this.revokeRefreshTokenFamily(refreshToken.familyId);
      throw new OAuthError('invalid_grant', 'Refresh token already used');
    }
  }

  private revokeRefreshTokenFamily(familyId: string): void {
    this.activeRefreshTokens.delete(familyId);
    this.revokedRefreshTokenFamilies.add(familyId);
    this.persistRefreshTokens();
  }

  private issueSignedToken(
    prefix: string,
    payloadObject: { jti?: string; familyId?: string; clientId: string; mvmtClientId?: string; mvmtClientCredentialVersion?: number; scope?: string; aud?: string; expiresAt: number },
  ): string {
    const payload = Buffer.from(
      JSON.stringify({
        ...payloadObject,
        jti: payloadObject.jti ?? crypto.randomBytes(16).toString('base64url'),
      }),
      'utf-8',
    ).toString('base64url');
    // HMAC-SHA256 MACs the token payload with a 256-bit random signing key.
    // This is JWT-style token signing, not password hashing.
    const signature = crypto.createHmac('sha256', this.resolveSigningKey()).update(payload).digest('base64url');
    return `${prefix}.${payload}.${signature}`;
  }

  private parseSignedToken(
    token: string,
    expectedPrefix: string,
  ): { jti?: string; familyId?: string; clientId: string; mvmtClientId?: string; mvmtClientCredentialVersion?: number; scope?: string; aud?: string; expiresAt: number } | undefined {
    const parts = token.split('.');
    if (parts.length !== 3 || parts[0] !== expectedPrefix) return undefined;

    const payload = parts[1];
    const signature = parts[2];
    const expectedSignature = crypto.createHmac('sha256', this.resolveSigningKey()).update(payload).digest('base64url');
    if (!timingSafeStringEquals(signature, expectedSignature)) return undefined;

    try {
      const decoded = JSON.parse(Buffer.from(payload, 'base64url').toString('utf-8')) as {
        clientId?: unknown;
        mvmtClientId?: unknown;
        mvmtClientCredentialVersion?: unknown;
        scope?: unknown;
        aud?: unknown;
        jti?: unknown;
        familyId?: unknown;
        expiresAt?: unknown;
      };
      if (typeof decoded.clientId !== 'string') return undefined;
      if (decoded.jti !== undefined && typeof decoded.jti !== 'string') return undefined;
      if (decoded.familyId !== undefined && typeof decoded.familyId !== 'string') return undefined;
      if (decoded.mvmtClientId !== undefined && typeof decoded.mvmtClientId !== 'string') return undefined;
      const mvmtClientCredentialVersion = decoded.mvmtClientCredentialVersion;
      if (
        mvmtClientCredentialVersion !== undefined
        && (typeof mvmtClientCredentialVersion !== 'number'
          || !Number.isInteger(mvmtClientCredentialVersion)
          || mvmtClientCredentialVersion < 1)
      ) {
        return undefined;
      }
      if (decoded.scope !== undefined && typeof decoded.scope !== 'string') return undefined;
      if (decoded.aud !== undefined && typeof decoded.aud !== 'string') return undefined;
      if (typeof decoded.expiresAt !== 'number' || !Number.isFinite(decoded.expiresAt)) return undefined;
      if (decoded.expiresAt < this.now()) return undefined;
      return {
        jti: decoded.jti,
        familyId: decoded.familyId,
        clientId: decoded.clientId,
        mvmtClientId: decoded.mvmtClientId,
        mvmtClientCredentialVersion,
        scope: decoded.scope,
        aud: decoded.aud,
        expiresAt: decoded.expiresAt,
      };
    } catch {
      return undefined;
    }
  }
}

export class OAuthError extends Error {
  constructor(
    public readonly code:
      | 'invalid_request'
      | 'invalid_scope'
      | 'invalid_grant'
      | 'unauthorized_client'
      | 'unsupported_grant_type'
      | 'invalid_client'
      | 'access_denied'
      | 'server_error',
    message: string,
  ) {
    super(message);
  }
}

export function verifyPkce(
  challenge: string,
  method: CodeChallengeMethod,
  verifier: string,
): boolean {
  if (!verifier) return false;
  if (method !== 'S256') return false;
  const hashed = crypto.createHash('sha256').update(verifier).digest();
  const expected = hashed.toString('base64url');
  return timingSafeStringEquals(expected, challenge);
}

// Per-process random key used to equalize input lengths before compare.
// HMACing both sides produces fixed-size 32-byte digests, so the
// subsequent timing-safe compare never short-circuits on length and
// cannot leak the length of the expected value through timing.
const COMPARE_KEY = crypto.randomBytes(32);

function timingSafeStringEquals(a: string, b: string): boolean {
  const hashA = crypto.createHmac('sha256', COMPARE_KEY).update(a, 'utf8').digest();
  const hashB = crypto.createHmac('sha256', COMPARE_KEY).update(b, 'utf8').digest();
  return crypto.timingSafeEqual(hashA, hashB);
}

function resolveRefreshScope(grantedScope: string | undefined, requestedScope: string | undefined): string | undefined {
  if (requestedScope === undefined) return grantedScope;
  if (grantedScope === undefined) {
    throw new OAuthError('invalid_scope', 'Refresh token does not allow scoped access');
  }
  if (!isScopeSubset(requestedScope, grantedScope)) {
    throw new OAuthError('invalid_scope', 'Requested scope exceeds refresh token scope');
  }
  return requestedScope;
}

function isScopeSubset(requestedScope: string, grantedScope: string): boolean {
  const granted = new Set(splitScope(grantedScope));
  return splitScope(requestedScope).every((scope) => granted.has(scope));
}

function splitScope(scope: string): string[] {
  return [...new Set(scope.split(/\s+/).map((entry) => entry.trim()).filter(Boolean))];
}

// Resolves the public-facing base URL for metadata and redirect responses.
//
// If publicBaseUrl is provided (set by the operator to the configured
// tunnel URL at startup), it is used verbatim. Otherwise the server falls
// back to the request's Host header. X-Forwarded-* headers are never
// trusted, because the server binds to 127.0.0.1 and those headers can
// be spoofed by any remote client behind a tunnel proxy — honoring them
// would let a remote attacker redirect OAuth issuer metadata to a host
// they control.
export function getBaseUrl(req: Request, publicBaseUrl?: string): string {
  if (publicBaseUrl) {
    try {
      const parsed = new URL(publicBaseUrl);
      return `${parsed.protocol}//${parsed.host}`;
    } catch {
      // fall through to host-header fallback
    }
  }
  const host = pickHeader(req.headers.host) ?? 'localhost';
  const proto = isLocalHost(host) ? 'http' : 'https';
  return `${proto}://${host}`;
}

function pickHeader(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) return value[0];
  if (typeof value === 'string' && value.length > 0) return value.split(',')[0].trim();
  return undefined;
}

function isLocalHost(host: string): boolean {
  const hostname = host.split(':')[0].toLowerCase();
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '::1';
}

export interface AuthorizePageParams {
  clientId: string;
  redirectUri: string;
  resource?: string;
  state?: string;
  requestId?: string;
  scope?: string;
  codeChallenge: string;
  codeChallengeMethod: CodeChallengeMethod;
  responseType: string;
  error?: string;
}

export function renderAuthorizePage(params: AuthorizePageParams): string {
  const hidden = (name: string, value: string | undefined) =>
    value === undefined ? '' : `<input type="hidden" name="${escapeHtml(name)}" value="${escapeHtml(value)}" />`;

  const error = params.error
    ? `<p class="error">${escapeHtml(params.error)}</p>`
    : '';

  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>mvmt – authorize connector</title>
  <style>
    body { font-family: -apple-system, system-ui, sans-serif; max-width: 420px; margin: 4rem auto; padding: 0 1rem; color: #111; }
    h1 { font-size: 1.25rem; }
    p { color: #444; line-height: 1.5; }
    label { display: block; font-size: 0.875rem; margin-top: 1rem; }
    input[type=password] { width: 100%; padding: 0.5rem; font-size: 1rem; box-sizing: border-box; }
    button { margin-top: 1rem; padding: 0.6rem 1rem; font-size: 1rem; background: #111; color: #fff; border: none; border-radius: 4px; cursor: pointer; }
    code { background: #f4f4f4; padding: 0.1rem 0.3rem; border-radius: 3px; }
    .error { color: #b00020; }
    .muted { color: #666; font-size: 0.85rem; }
  </style>
</head>
<body>
  <h1>Authorize connector</h1>
  <p>
    A client (<code>${escapeHtml(params.clientId)}</code>) is requesting access to your local mvmt instance.
    Paste a scoped mvmt API token to approve. Run <code>mvmt token add</code> on the host machine to create one.
  </p>
  ${error}
  <!-- Relative action: the page may be served behind a relay path prefix
       (e.g. /t/demo/authorize). An absolute "/authorize" would POST to the
       relay root and 404. "authorize" resolves against the current URL. -->
  <form method="POST" action="authorize">
    ${hidden('response_type', params.responseType)}
    ${hidden('client_id', params.clientId)}
    ${hidden('redirect_uri', params.redirectUri)}
    ${hidden('resource', params.resource)}
    ${hidden('state', params.state)}
    ${hidden('request_id', params.requestId)}
    ${hidden('scope', params.scope)}
    ${hidden('code_challenge', params.codeChallenge)}
    ${hidden('code_challenge_method', params.codeChallengeMethod)}
    <label for="api_token">mvmt API token</label>
    <input id="api_token" name="api_token" type="password" autocomplete="off" required />
    <button type="submit">Approve</button>
  </form>
  <p class="muted">You will be redirected back to <code>${escapeHtml(new URL(params.redirectUri).host)}</code>.</p>
</body>
</html>`;
}

export function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
