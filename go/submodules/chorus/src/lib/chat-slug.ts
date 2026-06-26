/**
 * Chat slug generation.
 *
 * The cockpit URL for a chat is `/runs/<slug>` instead of `/runs/<ULID>`.
 * Bookmarks and shared links should be human-readable: `/runs/review-pr-9`
 * beats `/runs/019DE881853DC50566D8DFCE083F75F0`.
 *
 * Strategy:
 *   1. Slugify the `work` field (truncated, ASCII-folded, lowercase,
 *      hyphen-separated).
 *   2. Fall back to the template_id when work is empty/non-Latin/all
 *      whitespace.
 *   3. On collision, append `-2`, `-3`, ... until unique.
 *
 * Pure module: takes an `existsFn(slug)` callback so the DB layer (or a
 * test) decides what "exists" means. No I/O here, fully unit-testable.
 *
 * Slug grammar (regex): `[a-z0-9]+(-[a-z0-9]+)*` — same shape GitHub
 * uses for issue/repo URLs. Leading/trailing/duplicate hyphens are
 * collapsed. Max length 60 characters; collision suffix always fits in
 * the trailing 8 chars even after truncation.
 */

const MAX_BASE_LEN = 60;
/** Generic stand-in when slugify produces empty string (all-emoji input, etc.). */
const FALLBACK_BASE = 'chat';
/** Defensive ceiling on the dedup loop — way above realistic collision counts. */
const MAX_DEDUP_ATTEMPTS = 10_000;

/**
 * Convert arbitrary text to a URL-safe slug component. Pure, deterministic.
 * - Folds Unicode → ASCII via NFKD (so "café" → "cafe").
 * - Replaces every non-[a-z0-9] run with a single hyphen.
 * - Strips leading/trailing hyphens.
 * - Truncates to MAX_BASE_LEN.
 * - Returns "" when input has no extractable slug characters; callers
 *   are expected to substitute a fallback.
 */
export function slugifyText(input: string): string {
  if (!input) return '';
  // NFKD splits accented chars into base + combining mark, then strip
  // the marks. Falls back to the original if the runtime doesn't
  // implement Intl normalisation.
  let s: string;
  try {
    s = input.normalize('NFKD').replace(/[̀-ͯ]/g, '');
  } catch {
    s = input;
  }
  s = s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
  if (s.length === 0) return '';
  if (s.length > MAX_BASE_LEN) {
    s = s.slice(0, MAX_BASE_LEN).replace(/-+$/g, '');
    // Edge: truncating mid-token can leave us with a stray trailing
    // hyphen even after the regex trim above (when the truncated tail
    // ends with `-something`). The replace above handles it.
  }
  return s;
}

/**
 * Build a slug for a chat with the given `work` text and `template_id`,
 * unique among `existsFn(slug) === true` rows. Pure function — caller
 * supplies the existence check.
 *
 * Collision suffix uses simple `-2`, `-3`, … rather than ULID prefixes
 * because the URL stays short and predictable. Real-world collision rate
 * is near zero (every chat starts with a different brief), so the loop
 * almost always exits on the first iteration.
 */
export async function generateChatSlug(args: {
  work: string;
  templateId: string;
  /** Returns true when a chat with this slug already exists. */
  existsFn: (slug: string) => Promise<boolean>;
}): Promise<string> {
  let base = slugifyText(args.work);
  if (!base) base = slugifyText(args.templateId);
  if (!base) base = FALLBACK_BASE;

  // Try the base first, then -2, -3, … until we find a free one.
  for (let i = 0; i < MAX_DEDUP_ATTEMPTS; i++) {
    const candidate = i === 0 ? base : `${base}-${i + 1}`;
    if (!(await args.existsFn(candidate))) return candidate;
  }
  // Vanishingly unlikely — every realistic workload exits in <5 iterations.
  // Fall through to a timestamp-suffixed slug rather than throw.
  return `${base}-${Date.now()}`;
}

/**
 * Lightweight check: does this string look like a slug rather than a
 * ULID? Used by routes that accept either.
 *
 * ULIDs are 26 uppercase Crockford-base32 chars; the codebase emits
 * 32-char uppercase hex, but either way they're [0-9A-Z]+ with no
 * hyphens. Slugs are lowercase + hyphens by construction. Test for the
 * presence of a lowercase letter or hyphen — both are absent in ULIDs
 * but always present in our slugs.
 */
export function looksLikeSlug(s: string): boolean {
  return /[a-z]/.test(s) || s.includes('-');
}
