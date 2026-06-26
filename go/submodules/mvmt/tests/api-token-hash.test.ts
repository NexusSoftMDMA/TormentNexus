import { describe, expect, it } from 'vitest';
import {
  hashApiToken,
  isApiTokenVerifier,
  verifyApiToken,
} from '../src/utils/api-token-hash.js';

const LEGACY_SHA256_FOR_CASE_TOKEN = '68422947E634BBE26EEACDFC99FDAD68E4E5FD75DA9EB717B1DA0F710EC7B0AB';

describe('API token verifiers', () => {
  it('stores new token verifiers with scrypt', () => {
    const verifier = hashApiToken('plain-token');

    expect(verifier).toMatch(/^scrypt:v1:/);
    expect(isApiTokenVerifier(verifier)).toBe(true);
    expect(verifyApiToken('plain-token', verifier)).toBe(true);
    expect(verifyApiToken('wrong-token', verifier)).toBe(false);
  });

  it('rejects SHA-256-looking verifiers', () => {
    expect(isApiTokenVerifier(LEGACY_SHA256_FOR_CASE_TOKEN)).toBe(false);
    expect(verifyApiToken('case-token', LEGACY_SHA256_FOR_CASE_TOKEN)).toBe(false);
  });

  it('rejects malformed scrypt verifiers', () => {
    const shortSalt = 'scrypt:v1:AAAA:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';
    const shortHash = 'scrypt:v1:AAAAAAAAAAAAAAAAAAAAAA:AAAA';

    expect(isApiTokenVerifier(shortSalt)).toBe(false);
    expect(isApiTokenVerifier(shortHash)).toBe(false);
    expect(verifyApiToken('plain-token', shortSalt)).toBe(false);
    expect(verifyApiToken('plain-token', shortHash)).toBe(false);
  });
});
