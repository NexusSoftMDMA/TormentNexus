import { describe, expect, it } from 'vitest';
import {
  Grant,
  clientConfigToGrant,
  isGrantPublished,
  leaseRecordToGrant,
} from '../src/grant/model.js';
import { ClientConfig } from '../src/config/schema.js';
import { LeaseRecord } from '../src/core/leases/store.js';

const TOKEN_VERIFIER = 'scrypt:v1:c2FsdHNhbHRzYWx0c2FsdA:aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g';

function client(overrides: Partial<ClientConfig> = {}): ClientConfig {
  return {
    id: 'codex',
    name: 'Codex CLI',
    description: '',
    auth: { type: 'token', tokenHash: TOKEN_VERIFIER },
    rawToolsEnabled: false,
    permissions: [
      { path: '/workspace/**', actions: ['search', 'read', 'write'] },
      { path: '/notes/**', actions: ['read'] },
    ],
    ...overrides,
  } as ClientConfig;
}

function lease(overrides: Partial<LeaseRecord> = {}): LeaseRecord {
  return {
    id: 'lease-1',
    label: 'Sarah - tax docs',
    path: '/taxes',
    resources: [
      { path: '/taxes', sourcePath: '/taxes', type: 'folder' },
      { path: '/taxes/w2.pdf', sourcePath: '/taxes/w2.pdf', type: 'file' },
    ],
    permissions: ['read'],
    tokenHash: TOKEN_VERIFIER,
    createdAt: new Date().toISOString(),
    downloadCount: 0,
    uploadCount: 0,
    ...overrides,
  };
}

describe('isGrantPublished — grandfather rule', () => {
  it('treats an absent published value as published', () => {
    expect(isGrantPublished(undefined)).toBe(true);
  });

  it('treats an explicit false as capability-only', () => {
    expect(isGrantPublished(false)).toBe(false);
  });

  it('treats an explicit true as published', () => {
    expect(isGrantPublished(true)).toBe(true);
  });
});

describe('clientConfigToGrant', () => {
  it('projects a token client into a grant, preserving scope', () => {
    const grant = clientConfigToGrant(client());
    expect(grant).toMatchObject<Partial<Grant>>({
      id: 'codex',
      label: 'Codex CLI',
      kind: 'token',
      published: true,
    });
    expect(grant.scope).toEqual([
      { path: '/workspace/**', actions: ['search', 'read', 'write'] },
      { path: '/notes/**', actions: ['read'] },
    ]);
  });

  it('marks oauth-auth clients with kind oauth', () => {
    const grant = clientConfigToGrant(
      client({ auth: { type: 'oauth', oauthClientIds: ['abc'] } }),
    );
    expect(grant.kind).toBe('oauth');
  });

  it('carries an explicit published:false through as capability-only', () => {
    expect(clientConfigToGrant(client({ published: false })).published).toBe(false);
  });

  it('carries expiresAt when present', () => {
    const expiresAt = '2030-01-01T00:00:00.000Z';
    expect(clientConfigToGrant(client({ expiresAt })).expiresAt).toBe(expiresAt);
  });

  it('does not let scope mutation leak back into the client config', () => {
    const source = client();
    const grant = clientConfigToGrant(source);
    grant.scope[0]!.actions.push('write');
    expect(source.permissions[0]!.actions).toEqual(['search', 'read', 'write']);
  });
});

describe('leaseRecordToGrant', () => {
  it('expands folders to a /** subtree and keeps files exact, with search on read', () => {
    const grant = leaseRecordToGrant(lease());
    expect(grant).toMatchObject<Partial<Grant>>({
      id: 'lease-1',
      label: 'Sarah - tax docs',
      kind: 'lease',
      published: true,
    });
    expect(grant.scope).toEqual([
      { path: '/taxes/**', actions: ['search', 'read'] },
      { path: '/taxes/w2.pdf', actions: ['search', 'read'] },
    ]);
  });

  it('grants search + read + write for a two-way lease', () => {
    const grant = leaseRecordToGrant(lease({ permissions: ['read', 'write'] }));
    for (const entry of grant.scope) {
      expect(entry.actions).toEqual(['search', 'read', 'write']);
    }
  });

  it('projects an upload-only lease to an empty scope (no MCP-resolvable access)', () => {
    expect(leaseRecordToGrant(lease({ permissions: ['upload'] })).scope).toEqual([]);
  });

  it('projects a write-without-read lease to an empty scope', () => {
    expect(leaseRecordToGrant(lease({ permissions: ['write'] })).scope).toEqual([]);
  });

  it('carries an explicit published:false through as capability-only', () => {
    expect(leaseRecordToGrant(lease({ published: false })).published).toBe(false);
  });

  it('carries expiresAt and revokedAt when present', () => {
    const expiresAt = '2030-01-01T00:00:00.000Z';
    const revokedAt = '2026-01-01T00:00:00.000Z';
    const grant = leaseRecordToGrant(lease({ expiresAt, revokedAt }));
    expect(grant.expiresAt).toBe(expiresAt);
    expect(grant.revokedAt).toBe(revokedAt);
  });
});
