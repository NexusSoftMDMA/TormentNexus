import { describe, expect, it } from 'vitest';
import { ClientConfig } from '../src/config/schema.js';
import {
  isQuarantined,
  quarantineIdentity,
  resolveClientIdentity,
  synthesizeDefaultClient,
} from '../src/server/client-identity.js';
import { AccessToken } from '../src/server/oauth.js';
import { hashApiToken } from '../src/utils/api-token-hash.js';

const ALWAYS_FALSE = () => false;
const ALWAYS_TRUE = () => true;

function tokenClient(id: string, plaintext: string, overrides: Partial<ClientConfig> = {}): ClientConfig {
  return {
    id,
    name: id,
    auth: { type: 'token', tokenHash: hashApiToken(plaintext) },
    rawToolsEnabled: false,
    permissions: [],
    ...overrides,
  };
}

function oauthClient(id: string, oauthClientIds: string[], overrides: Partial<ClientConfig> = {}): ClientConfig {
  return {
    id,
    name: id,
    auth: { type: 'oauth', oauthClientIds },
    rawToolsEnabled: false,
    permissions: [],
    ...overrides,
  };
}

function fakeOauthAccessToken(clientId: string, mvmtClientId?: string, mvmtClientCredentialVersion?: number): AccessToken {
  return {
    token: 'access-token-stub',
    clientId,
    mvmtClientId,
    mvmtClientCredentialVersion,
    audience: 'http://127.0.0.1:4141/mcp',
    expiresAt: Date.now() + 60_000,
  };
}

describe('synthesizeDefaultClient', () => {
  it('returns a legacy default identity with raw tools enabled', () => {
    const id = synthesizeDefaultClient();
    expect(id.id).toBe('default');
    expect(id.source).toBe('session');
    expect(id.rawToolsEnabled).toBe(true);
    expect(id.isLegacyDefault).toBe(true);
    expect(id.permissions).toEqual([]);
  });
});

describe('quarantineIdentity', () => {
  it('produces a zero-permission identity prefixed with quarantine:', () => {
    const id = quarantineIdentity('mystery-oauth-client');
    expect(id.id).toBe('quarantine:mystery-oauth-client');
    expect(id.source).toBe('quarantine');
    expect(id.rawToolsEnabled).toBe(false);
    expect(id.permissions).toEqual([]);
    expect(id.oauthClientId).toBe('mystery-oauth-client');
    expect(isQuarantined(id)).toBe(true);
  });
});

describe('resolveClientIdentity', () => {
  describe('OAuth path', () => {
    it('maps a known OAuth client_id to the named client', () => {
      const chatgpt = oauthClient('chatgpt', ['chatgpt-mvmt', 'chatgpt-mvmt-v2']);
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [chatgpt],
        oauthAccessToken: fakeOauthAccessToken('chatgpt-mvmt-v2'),
        validateSession: ALWAYS_FALSE,
      });

      expect(identity?.source).toBe('oauth');
      expect(identity?.id).toBe('chatgpt');
      expect(identity?.oauthClientId).toBe('chatgpt-mvmt-v2');
    });

    it('quarantines an unknown OAuth client_id when per-client policy is configured', () => {
      // With a non-empty clients[], an unmapped OAuth client_id is
      // quarantined (zero permissions, raw tools off). authMiddleware
      // returns 403 before the request reaches any tool surface.
      const someTokenClient = tokenClient('codex', 'codex-token');
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [someTokenClient],
        oauthAccessToken: fakeOauthAccessToken('unknown-dcr-client'),
        validateSession: ALWAYS_TRUE,
      });

      expect(identity?.source).toBe('quarantine');
      expect(identity?.id).toBe('quarantine:unknown-dcr-client');
      expect(identity?.rawToolsEnabled).toBe(false);
      expect(identity?.permissions).toEqual([]);
    });

    it('maps an OAuth access token to the scoped API-token client selected at authorization time', () => {
      const codex = tokenClient('codex', 'codex-token', {
        credentialVersion: 2,
        permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
      });
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [codex],
        oauthAccessToken: fakeOauthAccessToken('unknown-dcr-client', 'codex', 2),
        validateSession: ALWAYS_FALSE,
      });

      expect(identity?.source).toBe('oauth');
      expect(identity?.id).toBe('codex');
      expect(identity?.oauthClientId).toBe('unknown-dcr-client');
      expect(identity?.permissions).toEqual([{ path: '/workspace/**', actions: ['search', 'read'] }]);
    });

    it('rejects an OAuth access token minted before scoped API-token rotation', () => {
      const codex = tokenClient('codex', 'codex-token', {
        credentialVersion: 2,
      });
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [codex],
        oauthAccessToken: fakeOauthAccessToken('unknown-dcr-client', 'codex', 1),
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });

    it('treats legacy OAuth tokens and token clients as credential version 1', () => {
      const codex = tokenClient('codex', 'codex-token');
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [codex],
        oauthAccessToken: fakeOauthAccessToken('unknown-dcr-client', 'codex'),
        validateSession: ALWAYS_FALSE,
      });

      expect(identity?.id).toBe('codex');
    });

    it('rejects an OAuth access token when the selected scoped token has expired', () => {
      const codex = tokenClient('codex', 'codex-token', {
        expiresAt: '2000-01-01T00:00:00.000Z',
      });
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [codex],
        oauthAccessToken: fakeOauthAccessToken('dcr-client', 'codex'),
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });

    it('rejects an OAuth access token when the selected scoped token binding does not match', () => {
      const codex = tokenClient('codex', 'codex-token', {
        clientBinding: 'chatgpt',
      });
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [codex],
        oauthAccessToken: fakeOauthAccessToken('claude', 'codex'),
        validateSession: ALWAYS_FALSE,
        clientHint: 'claude',
      });

      expect(identity).toBeUndefined();
    });

    it('falls back to the legacy default identity for OAuth tokens when clients[] is empty', () => {
      // Pre-PR OAuth flows keep working in legacy mode (no clients[]
      // configured). The operator opts into the strict quarantine model
      // by adding entries to clients[].
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [],
        oauthAccessToken: fakeOauthAccessToken('legacy-oauth-client'),
        validateSession: ALWAYS_FALSE,
      });

      expect(identity?.source).toBe('session');
      expect(identity?.isLegacyDefault).toBe(true);
    });

    it('quarantines OAuth tokens when legacy default access is disabled', () => {
      const identity = resolveClientIdentity({
        authHeader: 'Bearer mvmtv1.something',
        clients: [],
        oauthAccessToken: fakeOauthAccessToken('legacy-oauth-client'),
        validateSession: ALWAYS_FALSE,
        allowLegacyDefault: false,
      });

      expect(identity?.source).toBe('quarantine');
      expect(identity?.id).toBe('quarantine:legacy-oauth-client');
    });

    it('does not fall through to client-token or session paths when OAuth token is present (and policy is configured)', () => {
      // Even if a token client matches the bearer string, the OAuth path
      // takes precedence for OAuth-authenticated requests.
      const codex = tokenClient('codex', 'plaintext');
      const identity = resolveClientIdentity({
        authHeader: 'Bearer plaintext',
        clients: [codex],
        oauthAccessToken: fakeOauthAccessToken('chatgpt-mvmt'),
        validateSession: ALWAYS_TRUE,
      });

      expect(identity?.source).toBe('quarantine');
      expect(identity?.oauthClientId).toBe('chatgpt-mvmt');
    });
  });

  describe('client token path', () => {
    it('matches a configured token client by scrypt verifier', () => {
      const codex = tokenClient('codex', 'codex-plaintext-token', { rawToolsEnabled: true });
      const identity = resolveClientIdentity({
        authHeader: 'Bearer codex-plaintext-token',
        clients: [codex],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
      });

      expect(identity?.source).toBe('token');
      expect(identity?.id).toBe('codex');
      expect(identity?.rawToolsEnabled).toBe(true);
    });

    it('requires configured token client bindings to match the request hint', () => {
      const codex = tokenClient('codex', 'codex-plaintext-token', {
        clientBinding: 'claude-desktop',
      });

      const rejected = resolveClientIdentity({
        authHeader: 'Bearer codex-plaintext-token',
        clients: [codex],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
        clientHint: 'curl/8.0',
      });
      const accepted = resolveClientIdentity({
        authHeader: 'Bearer codex-plaintext-token',
        clients: [codex],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
        clientHint: 'Claude-Desktop/1.0',
      });

      expect(rejected).toBeUndefined();
      expect(accepted?.id).toBe('codex');
    });

    it('rejects an expired configured token client even when the bearer matches', () => {
      const codex = tokenClient('codex', 'codex-plaintext-token', {
        expiresAt: '2000-01-01T00:00:00.000Z',
      });
      const identity = resolveClientIdentity({
        authHeader: 'Bearer codex-plaintext-token',
        clients: [codex],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });

    it('does not match when bearer hash differs', () => {
      const codex = tokenClient('codex', 'expected-token');
      const identity = resolveClientIdentity({
        authHeader: 'Bearer wrong-token',
        clients: [codex],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });

    it('skips oauth-typed clients when matching token bearers', () => {
      const oauth = oauthClient('chatgpt', ['chatgpt-mvmt']);
      const identity = resolveClientIdentity({
        authHeader: 'Bearer some-token',
        clients: [oauth],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });

  });

  describe('session token path', () => {
    it('synthesizes default identity when clients[] is empty (legacy compatibility)', () => {
      const identity = resolveClientIdentity({
        authHeader: 'Bearer session-token',
        clients: [],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_TRUE,
      });

      expect(identity?.source).toBe('session');
      expect(identity?.id).toBe('default');
      expect(identity?.isLegacyDefault).toBe(true);
    });

    it('rejects the legacy session token when legacy default access is disabled', () => {
      const identity = resolveClientIdentity({
        authHeader: 'Bearer session-token',
        clients: [],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_TRUE,
        allowLegacyDefault: false,
      });

      expect(identity).toBeUndefined();
    });

    it('does NOT fall back to default when clients[] is non-empty (session token is admin-only once policy is configured)', () => {
      const codex = tokenClient('codex', 'real-codex-token');
      const identity = resolveClientIdentity({
        authHeader: 'Bearer session-token',
        clients: [codex],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_TRUE,
      });

      // Session token is valid but no client token matched and clients[] is configured;
      // the resolver returns undefined so the caller can 401. This prevents the session
      // token from being a parallel /mcp credential that bypasses per-client policy.
      expect(identity).toBeUndefined();
    });

    it('returns undefined when nothing matches', () => {
      const identity = resolveClientIdentity({
        authHeader: 'Bearer unknown',
        clients: [],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });
  });

  describe('bearer extraction', () => {
    it('returns undefined for missing Bearer prefix when no other path matches', () => {
      const identity = resolveClientIdentity({
        authHeader: 'Basic abc',
        clients: [tokenClient('codex', 'codex-token')],
        oauthAccessToken: undefined,
        validateSession: ALWAYS_FALSE,
      });

      expect(identity).toBeUndefined();
    });
  });
});
