/**
 * Tests for the chat slug generator.
 *
 * Pins the contracts /runs/<slug> URL routing depends on:
 *   - slugifyText: ASCII-fold, lowercase, hyphenated, length-capped, empty on no-input
 *   - generateChatSlug: collision dedup, fallback chain (work → templateId → "chat")
 *   - looksLikeSlug: distinguishes slugs from ULIDs at the route layer
 */
import { describe, it, expect } from 'vitest';
import {
  slugifyText,
  generateChatSlug,
  looksLikeSlug,
} from '../src/lib/chat-slug';

describe('slugifyText', () => {
  it('lowercases + collapses non-alphanumeric to hyphens', () => {
    expect(slugifyText('Review PR #10')).toBe('review-pr-10');
    expect(slugifyText('Fix Bug 123')).toBe('fix-bug-123');
  });

  it('strips leading/trailing/duplicate hyphens', () => {
    expect(slugifyText('  hello  world  ')).toBe('hello-world');
    expect(slugifyText('---test---')).toBe('test');
    expect(slugifyText('a___b')).toBe('a-b');
  });

  it('returns empty string when no slug chars survive', () => {
    expect(slugifyText('')).toBe('');
    expect(slugifyText('   ')).toBe('');
    expect(slugifyText('!!!')).toBe('');
    expect(slugifyText('🤖🚀')).toBe('');
  });

  it('truncates to 60 chars and trims trailing hyphen', () => {
    const long = 'a'.repeat(80);
    expect(slugifyText(long)).toBe('a'.repeat(60));
    // Truncation that lands mid-token shouldn't leave a trailing hyphen
    const tokenized = `${'a'.repeat(58)}-bbb-ccc`;
    const result = slugifyText(tokenized);
    expect(result.length).toBeLessThanOrEqual(60);
    expect(result.endsWith('-')).toBe(false);
  });

  it('folds accented characters to plain ASCII', () => {
    expect(slugifyText('Café Résumé')).toBe('cafe-resume');
  });

  it('preserves digits', () => {
    expect(slugifyText('PR #42 review v2')).toBe('pr-42-review-v2');
  });
});

describe('generateChatSlug', () => {
  it('uses work field by default', async () => {
    const slug = await generateChatSlug({
      work: 'Review PR #10',
      templateId: 'code-review',
      existsFn: async () => false,
    });
    expect(slug).toBe('review-pr-10');
  });

  it('falls back to templateId when work yields no slug chars', async () => {
    const slug = await generateChatSlug({
      work: '🤖🚀',
      templateId: 'code-review',
      existsFn: async () => false,
    });
    expect(slug).toBe('code-review');
  });

  it('falls back to "chat" when both work and templateId are unsluggable', async () => {
    const slug = await generateChatSlug({
      work: '   ',
      templateId: '!!!',
      existsFn: async () => false,
    });
    expect(slug).toBe('chat');
  });

  it('appends -2, -3 on collision', async () => {
    const taken = new Set(['review-pr-10', 'review-pr-10-2']);
    const slug = await generateChatSlug({
      work: 'Review PR #10',
      templateId: 'code-review',
      existsFn: async (s) => taken.has(s),
    });
    expect(slug).toBe('review-pr-10-3');
  });

  it('starts collision suffix at -2 (not -1) to keep the base slug clean', async () => {
    const slug = await generateChatSlug({
      work: 'foo',
      templateId: 'x',
      existsFn: async (s) => s === 'foo',
    });
    expect(slug).toBe('foo-2');
  });
});

describe('looksLikeSlug', () => {
  it('flags lowercase + hyphenated strings as slugs', () => {
    expect(looksLikeSlug('review-pr-10')).toBe(true);
    expect(looksLikeSlug('foo')).toBe(true);
    expect(looksLikeSlug('foo-bar')).toBe(true);
  });

  it('rejects all-uppercase ULIDs', () => {
    // 32-char uppercase hex (chorus emits these for chat ids)
    expect(looksLikeSlug('019DE881853DC50566D8DFCE083F75F0')).toBe(false);
    // 26-char Crockford ULID
    expect(looksLikeSlug('01ARZ3NDEKTSV4RRFFQ69G5FAV')).toBe(false);
  });

  it('treats hyphenated all-uppercase as a slug (defensive)', () => {
    // Not a real ULID shape but safe to treat as slug — getBySlugOrId
    // will miss the slug lookup and fall back to id, so worst case is
    // an extra DB query, never a wrong answer.
    expect(looksLikeSlug('FOO-BAR')).toBe(true);
  });
});
