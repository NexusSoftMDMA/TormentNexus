import crypto from 'crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { OAuthClientRegistryLimitError, OAuthError, OAuthStore, verifyPkce } from '../src/server/oauth.js';

function pkcePair(verifier = crypto.randomBytes(32).toString('base64url')): {
  verifier: string;
  challenge: string;
} {
  const challenge = crypto.createHash('sha256').update(verifier).digest('base64url');
  return { verifier, challenge };
}

describe('verifyPkce', () => {
  it('accepts a matching S256 verifier', () => {
    const { verifier, challenge } = pkcePair();
    expect(verifyPkce(challenge, 'S256', verifier)).toBe(true);
  });

  it('rejects an incorrect S256 verifier', () => {
    const { challenge } = pkcePair();
    expect(verifyPkce(challenge, 'S256', 'not-the-verifier')).toBe(false);
  });

  it('rejects non-S256 methods (plain is no longer supported)', () => {
    expect(verifyPkce('abc', 'plain' as unknown as 'S256', 'abc')).toBe(false);
  });
});

describe('OAuthStore client registration', () => {
  it('tracks registered redirect_uris and rejects unregistered ones', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    store.registerClient({
      clientId: 'claude',
      redirectUris: ['https://claude.ai/cb', 'https://claude.ai/cb2'],
    });

    expect(store.isRedirectUriAllowed('claude', 'https://claude.ai/cb')).toBe(true);
    expect(store.isRedirectUriAllowed('claude', 'https://claude.ai/cb2')).toBe(true);
    expect(store.isRedirectUriAllowed('claude', 'https://attacker.example/cb')).toBe(false);
    expect(store.isRedirectUriAllowed('unknown-client', 'https://claude.ai/cb')).toBe(false);
  });

  it('persists registered clients to disk and reloads them on reconstruction', () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-oauth-clients-'));
    const clientsPath = path.join(tmp, '.clients.json');
    try {
      const first = new OAuthStore({ signingKey: 'k', clientsPath });
      first.registerClient({
        clientId: 'claude',
        redirectUris: ['https://claude.ai/cb'],
      });

      const second = new OAuthStore({ signingKey: 'k', clientsPath });
      expect(second.isRedirectUriAllowed('claude', 'https://claude.ai/cb')).toBe(true);
      expect(second.isRedirectUriAllowed('claude', 'https://attacker.example/cb')).toBe(false);
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('fails closed when client registrations cannot be persisted', () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-oauth-clients-'));
    const lockedDir = path.join(tmp, 'locked');
    const clientsPath = path.join(lockedDir, '.clients.json');
    fs.mkdirSync(lockedDir, { recursive: true });
    fs.chmodSync(lockedDir, 0o500);
    try {
      const store = new OAuthStore({ signingKey: 'k', clientsPath });
      expect(() =>
        store.registerClient({
          clientId: 'claude',
          redirectUris: ['https://claude.ai/cb'],
        }),
      ).toThrow();
      expect(store.isRedirectUriAllowed('claude', 'https://claude.ai/cb')).toBe(false);
    } finally {
      fs.chmodSync(lockedDir, 0o700);
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('limits registered clients and redirect URI fan-out', () => {
    const store = new OAuthStore({
      signingKey: 'k',
      maxRegisteredClients: 1,
      maxRedirectUrisPerClient: 1,
    });
    store.registerClient({
      clientId: 'first',
      redirectUris: ['https://client.example/cb'],
    });

    expect(() =>
      store.registerClient({
        clientId: 'second',
        redirectUris: ['https://client.example/cb'],
      }),
    ).toThrow(OAuthClientRegistryLimitError);

    const fanoutStore = new OAuthStore({
      signingKey: 'k',
      maxRedirectUrisPerClient: 1,
    });
    expect(() =>
      fanoutStore.registerClient({
        clientId: 'fanout',
        redirectUris: ['https://client.example/cb', 'https://client.example/other'],
      }),
    ).toThrow(OAuthClientRegistryLimitError);
  });
});

describe('OAuthStore signing key rotation', () => {
  it('invalidates outstanding access tokens when the signing key callback returns a new value', () => {
    let currentKey = 'key-one';
    const store = new OAuthStore({ signingKey: () => currentKey });
    const token = store.issueAccessToken({ clientId: 'claude' });

    expect(store.validateAccessToken(`Bearer ${token.token}`)?.token).toBe(token.token);

    currentKey = 'key-two';
    expect(store.validateAccessToken(`Bearer ${token.token}`)).toBeUndefined();
  });
});

describe('OAuthStore', () => {
  it('issues an auth code and exchanges it for a bearer token', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();

    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/api/mcp/auth_callback',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    const token = store.consumeCode({
      code: code.code,
      clientId: 'claude',
      redirectUri: 'https://claude.ai/api/mcp/auth_callback',
      codeVerifier: verifier,
    });

    expect(token.token).toBeTypeOf('string');
    expect(token.token.length).toBeGreaterThan(20);
    expect(store.validateAccessToken(`Bearer ${token.token}`)?.token).toBe(token.token);
  });

  it('rejects code reuse', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/cb',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    store.consumeCode({
      code: code.code,
      clientId: 'claude',
      redirectUri: 'https://claude.ai/cb',
      codeVerifier: verifier,
    });

    expect(() =>
      store.consumeCode({
        code: code.code,
        clientId: 'claude',
        redirectUri: 'https://claude.ai/cb',
        codeVerifier: verifier,
      }),
    ).toThrow(OAuthError);
  });

  it('rejects a PKCE mismatch', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/cb',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    expect(() =>
      store.consumeCode({
        code: code.code,
        clientId: 'claude',
        redirectUri: 'https://claude.ai/cb',
        codeVerifier: 'wrong-verifier',
      }),
    ).toThrow(/PKCE/);
  });

  it('rejects a redirect_uri mismatch', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/cb',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    expect(() =>
      store.consumeCode({
        code: code.code,
        clientId: 'claude',
        redirectUri: 'https://evil.example/cb',
        codeVerifier: verifier,
      }),
    ).toThrow(/Redirect URI/);
  });

  it('binds authorization codes to the requested resource', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'chatgpt',
      redirectUri: 'https://chatgpt.com/connector/oauth/callback',
      resource: 'https://mcp.example.com/mcp',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    expect(() =>
      store.consumeCode({
        code: code.code,
        clientId: 'chatgpt',
        redirectUri: 'https://chatgpt.com/connector/oauth/callback',
        resource: 'https://other.example.com/mcp',
        codeVerifier: verifier,
      }),
    ).toThrow(/Resource mismatch/);
  });

  it('allows the token exchange to omit resource when the code is already resource-bound', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/api/mcp/auth_callback',
      resource: 'https://mcp.example.com/mcp',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    const token = store.consumeCode({
      code: code.code,
      clientId: 'claude',
      redirectUri: 'https://claude.ai/api/mcp/auth_callback',
      codeVerifier: verifier,
    });

    expect(token.audience).toBe('https://mcp.example.com/mcp');
  });

  it('exchanges authorization codes into both access and refresh tokens', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'chatgpt',
      redirectUri: 'https://chatgpt.com/connector/oauth/callback',
      scope: 'mcp offline_access',
      resource: 'https://mcp.example.com/mcp',
      mvmtClientId: 'codex',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    const tokens = store.exchangeCode({
      code: code.code,
      clientId: 'chatgpt',
      redirectUri: 'https://chatgpt.com/connector/oauth/callback',
      resource: 'https://mcp.example.com/mcp',
      codeVerifier: verifier,
    });

    expect(tokens.accessToken.scope).toBe('mcp offline_access');
    expect(tokens.accessToken.mvmtClientId).toBe('codex');
    expect(tokens.refreshToken.scope).toBe('mcp offline_access');
    expect(tokens.refreshToken.mvmtClientId).toBe('codex');
    expect(tokens.refreshToken.audience).toBe('https://mcp.example.com/mcp');
    expect(tokens.refreshToken.token).toBeTypeOf('string');

    const validated = store.validateAccessToken(`Bearer ${tokens.accessToken.token}`);
    expect(validated?.mvmtClientId).toBe('codex');
  });

  it('exchanges refresh tokens for a new access token and preserves selected mvmt client identity', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const refreshToken = store.issueRefreshToken({
      clientId: 'chatgpt',
      scope: 'mcp offline_access',
      audience: 'https://mcp.example.com/mcp',
      mvmtClientId: 'codex',
    });

    const tokens = store.exchangeRefreshToken({
      refreshToken: refreshToken.token,
      clientId: 'chatgpt',
    });

    expect(tokens.accessToken.token).not.toBe(tokens.refreshToken.token);
    expect(tokens.accessToken.scope).toBe('mcp offline_access');
    expect(tokens.accessToken.audience).toBe('https://mcp.example.com/mcp');
    expect(tokens.accessToken.mvmtClientId).toBe('codex');
    expect(tokens.refreshToken.scope).toBe('mcp offline_access');
    expect(tokens.refreshToken.mvmtClientId).toBe('codex');
    expect(tokens.refreshToken.token).not.toBe(refreshToken.token);
  });

  it('rejects replayed refresh tokens and revokes the replacement family', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const refreshToken = store.issueRefreshToken({
      clientId: 'chatgpt',
      scope: 'mcp offline_access',
      audience: 'https://mcp.example.com/mcp',
    });

    const replacement = store.exchangeRefreshToken({
      refreshToken: refreshToken.token,
      clientId: 'chatgpt',
    });

    expect(() =>
      store.exchangeRefreshToken({
        refreshToken: refreshToken.token,
        clientId: 'chatgpt',
      }),
    ).toThrow(/already used/i);
    expect(() =>
      store.exchangeRefreshToken({
        refreshToken: replacement.refreshToken.token,
        clientId: 'chatgpt',
      }),
    ).toThrow(/revoked/i);
  });

  it('persists refresh-token rotation state across OAuthStore instances', () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-oauth-refresh-'));
    const refreshTokensPath = path.join(tmp, 'refresh-tokens.json');
    try {
      const issuer = new OAuthStore({ signingKey: 'test-secret', refreshTokensPath });
      const refreshToken = issuer.issueRefreshToken({
        clientId: 'chatgpt',
        scope: 'mcp offline_access',
        audience: 'https://mcp.example.com/mcp',
      });

      const firstConsumer = new OAuthStore({ signingKey: 'test-secret', refreshTokensPath });
      const replacement = firstConsumer.exchangeRefreshToken({
        refreshToken: refreshToken.token,
        clientId: 'chatgpt',
      });

      const secondConsumer = new OAuthStore({ signingKey: 'test-secret', refreshTokensPath });
      expect(() =>
        secondConsumer.exchangeRefreshToken({
          refreshToken: refreshToken.token,
          clientId: 'chatgpt',
        }),
      ).toThrow(/already used/i);
      expect(() =>
        secondConsumer.exchangeRefreshToken({
          refreshToken: replacement.refreshToken.token,
          clientId: 'chatgpt',
        }),
      ).toThrow(/revoked/i);
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('keeps the old refresh token active when rotation persistence fails', () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-oauth-refresh-'));
    const refreshTokensPath = path.join(tmp, 'refresh-tokens.json');
    try {
      const store = new OAuthStore({ signingKey: 'test-secret', refreshTokensPath });
      const refreshToken = store.issueRefreshToken({
        clientId: 'chatgpt',
        scope: 'mcp offline_access',
        audience: 'https://mcp.example.com/mcp',
      });
      fs.rmSync(refreshTokensPath);
      fs.mkdirSync(refreshTokensPath);

      expect(() =>
        store.exchangeRefreshToken({
          refreshToken: refreshToken.token,
          clientId: 'chatgpt',
        }),
      ).toThrow();

      fs.rmSync(refreshTokensPath, { recursive: true, force: true });
      const replacement = store.exchangeRefreshToken({
        refreshToken: refreshToken.token,
        clientId: 'chatgpt',
      });
      expect(replacement.refreshToken.token).not.toBe(refreshToken.token);
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('does not throw from cleanup when refresh-token persistence fails', () => {
    let now = 1_000_000;
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-oauth-refresh-'));
    const refreshTokensPath = path.join(tmp, 'refresh-tokens.json');
    try {
      const store = new OAuthStore({
        signingKey: 'test-secret',
        refreshTokenTtlMs: 1000,
        refreshTokensPath,
        now: () => now,
      });
      store.issueRefreshToken({
        clientId: 'chatgpt',
        scope: 'mcp offline_access',
        audience: 'https://mcp.example.com/mcp',
      });
      now += 2000;
      fs.rmSync(refreshTokensPath);
      fs.mkdirSync(refreshTokensPath);

      expect(() => store.cleanup()).not.toThrow();
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects refresh token scope widening', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const refreshToken = store.issueRefreshToken({
      clientId: 'chatgpt',
      scope: 'mcp',
      audience: 'https://mcp.example.com/mcp',
    });

    expect(() =>
      store.exchangeRefreshToken({
        refreshToken: refreshToken.token,
        clientId: 'chatgpt',
        scope: 'mcp offline_access',
      }),
    ).toThrow(/scope/i);
  });

  it('rejects introducing a resource at token time when the authorization code had none', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/api/mcp/auth_callback',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    expect(() =>
      store.consumeCode({
        code: code.code,
        clientId: 'claude',
        redirectUri: 'https://claude.ai/api/mcp/auth_callback',
        resource: 'https://mcp.example.com/mcp',
        codeVerifier: verifier,
      }),
    ).toThrow(/Resource mismatch/);
  });

  it('expires access tokens', () => {
    let now = 1_000_000;
    const store = new OAuthStore({ tokenTtlMs: 1000, signingKey: 'test-secret', now: () => now });
    const token = store.issueAccessToken({ clientId: 'claude' });

    expect(store.validateAccessToken(`Bearer ${token.token}`)).toBeDefined();
    now += 2000;
    expect(store.validateAccessToken(`Bearer ${token.token}`)).toBeUndefined();
  });

  it('expires authorization codes', () => {
    let now = 1_000_000;
    const store = new OAuthStore({ codeTtlMs: 1000, signingKey: 'test-secret', now: () => now });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      redirectUri: 'https://claude.ai/cb',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });
    now += 2000;

    expect(() =>
      store.consumeCode({
        code: code.code,
        clientId: 'claude',
        redirectUri: 'https://claude.ai/cb',
        codeVerifier: verifier,
      }),
    ).toThrow(/expired/);
  });

  it('validateAccessToken ignores malformed or missing headers', () => {
    const store = new OAuthStore({ signingKey: 'test-secret' });
    expect(store.validateAccessToken(undefined)).toBeUndefined();
    expect(store.validateAccessToken('Basic xxx')).toBeUndefined();
    expect(store.validateAccessToken('Bearer nope')).toBeUndefined();
  });

  it('validates access tokens across store instances that share the same secret', () => {
    const tokenIssuer = new OAuthStore({ signingKey: 'shared-secret' });
    const tokenValidator = new OAuthStore({ signingKey: 'shared-secret' });
    const token = tokenIssuer.issueAccessToken({ clientId: 'codex' });

    const validated = tokenValidator.validateAccessToken(`Bearer ${token.token}`);

    expect(validated?.token).toBe(token.token);
    expect(validated?.clientId).toBe('codex');
  });

  it('preserves selected mvmt token credential version across OAuth grants', () => {
    const store = new OAuthStore({ signingKey: 'shared-secret' });
    const { verifier, challenge } = pkcePair();
    const code = store.issueCode({
      clientId: 'claude',
      mvmtClientId: 'codex',
      mvmtClientCredentialVersion: 3,
      redirectUri: 'https://claude.ai/cb',
      resource: 'https://mvmt.example.com/mcp',
      codeChallenge: challenge,
      codeChallengeMethod: 'S256',
    });

    const grant = store.exchangeCode({
      code: code.code,
      clientId: 'claude',
      redirectUri: 'https://claude.ai/cb',
      resource: 'https://mvmt.example.com/mcp',
      codeVerifier: verifier,
    });
    const validated = store.validateAccessToken(`Bearer ${grant.accessToken.token}`);
    const refreshed = store.exchangeRefreshToken({
      refreshToken: grant.refreshToken.token,
      clientId: 'claude',
    });

    expect(validated?.mvmtClientId).toBe('codex');
    expect(validated?.mvmtClientCredentialVersion).toBe(3);
    expect(refreshed.accessToken.mvmtClientCredentialVersion).toBe(3);
  });

  it('rejects access tokens whose audience does not match the expected resource', () => {
    const store = new OAuthStore({ signingKey: 'shared-secret' });
    const token = store.issueAccessToken({
      clientId: 'codex',
      audience: 'https://mvmt.example.com/mcp',
    });

    expect(
      store.validateAccessToken(`Bearer ${token.token}`, {
        expectedAudience: 'https://mvmt.example.com/mcp',
      })?.token,
    ).toBe(token.token);
    expect(
      store.validateAccessToken(`Bearer ${token.token}`, {
        expectedAudience: 'https://other.example.com/mcp',
      }),
    ).toBeUndefined();
  });

  it('allows legacy audience-less tokens only when explicitly requested', () => {
    const store = new OAuthStore({ signingKey: 'shared-secret' });
    const token = store.issueAccessToken({ clientId: 'codex' });

    expect(
      store.validateAccessToken(`Bearer ${token.token}`, {
        expectedAudience: 'https://mvmt.example.com/mcp',
      }),
    ).toBeUndefined();
    expect(
      store.validateAccessToken(`Bearer ${token.token}`, {
        expectedAudience: 'https://mvmt.example.com/mcp',
        allowLegacyNoAudience: true,
      })?.token,
    ).toBe(token.token);
  });

  it('rejects access tokens signed with a different secret', () => {
    const tokenIssuer = new OAuthStore({ signingKey: 'issuer-secret' });
    const tokenValidator = new OAuthStore({ signingKey: 'validator-secret' });
    const token = tokenIssuer.issueAccessToken({ clientId: 'codex' });

    expect(tokenValidator.validateAccessToken(`Bearer ${token.token}`)).toBeUndefined();
  });
});
