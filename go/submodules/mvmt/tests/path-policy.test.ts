import path from 'path';
import { describe, expect, it } from 'vitest';
import {
  isGloballyDeniedPath,
  matchesConfiguredOrGlobalPattern,
  matchesGlobalSecretPattern,
  matchesPathPatterns,
} from '../src/context/path-policy.js';

describe('path policy helpers', () => {
  it('matches literal, single-star, and subtree patterns', () => {
    expect(matchesPathPatterns('docs/guide.md', ['docs/*.md'])).toBe(true);
    expect(matchesPathPatterns('docs/nested/guide.md', ['docs/*.md'])).toBe(false);
    expect(matchesPathPatterns('docs/nested/guide.md', ['docs/**'])).toBe(true);
    expect(matchesPathPatterns('exact.env', ['exact.env'])).toBe(true);
  });

  it('matches global secret paths even when config patterns omit them', () => {
    expect(matchesGlobalSecretPattern('.mvmt/.session-token')).toBe(true);
    expect(matchesGlobalSecretPattern('.ssh/id_ed25519')).toBe(true);
    expect(matchesConfiguredOrGlobalPattern('.aws/credentials', [])).toBe(true);
  });

  it('denies real paths rooted inside sensitive directories', () => {
    expect(isGloballyDeniedPath('config.yaml', path.join('/tmp/demo/.mvmt', 'config.yaml'))).toBe(true);
    expect(isGloballyDeniedPath('hosts.yaml', path.join('/Users/me/.config/gh', 'hosts.yaml'))).toBe(true);
    expect(isGloballyDeniedPath('notes/today.md', path.join('/Users/me/notes', 'today.md'))).toBe(false);
  });
});
