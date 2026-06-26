// Compact number formatters used across the dashboard / analytics / hermes pages.
// Token counts can range from a single digit up to trillions for power users
// across long timespans, so we scale through K → M → B → T.
//
// Examples:
//   formatTokens(0)            → "0"
//   formatTokens(847)          → "847"
//   formatTokens(12_400)       → "12.4k"
//   formatTokens(921_600_000)  → "921.6M"
//   formatTokens(1_234_000_000) → "1.23B"
//   formatTokens(2_500_000_000_000) → "2.5T"

const TRILLION = 1_000_000_000_000;
const BILLION  =     1_000_000_000;
const MILLION  =         1_000_000;
const THOUSAND =             1_000;

/**
 * Compact number formatting with K/M/B/T suffixes. Returns "0" for falsy input
 * so callers don't have to guard the `0 | undefined | null` case.
 */
export function formatTokens(n: number | null | undefined): string {
  if (!n || !Number.isFinite(n)) return "0";
  const abs = Math.abs(n);
  if (abs >= TRILLION) return `${trim(n / TRILLION)}T`;
  if (abs >= BILLION)  return `${trim(n / BILLION)}B`;
  if (abs >= MILLION)  return `${trim(n / MILLION)}M`;
  if (abs >= THOUSAND) return `${(n / THOUSAND).toFixed(1)}k`;
  return n.toLocaleString();
}

/**
 * Two decimals for values < 10 (e.g. "1.23B"), one decimal otherwise
 * (e.g. "921.6M"). Strips trailing ".0" so "1.0M" → "1M".
 */
function trim(v: number): string {
  const fixed = Math.abs(v) >= 10 ? v.toFixed(1) : v.toFixed(2);
  return fixed.replace(/\.?0+$/, "");
}

/**
 * Cost formatting — keeps small spends visible (4 decimals under $0.01,
 * 2 decimals otherwise) and scales big cumulative spends to K/M.
 */
export function formatCost(usd: number | null | undefined): string {
  if (usd == null || !Number.isFinite(usd)) return "$0.00";
  const abs = Math.abs(usd);
  if (abs >= MILLION) return `$${trim(usd / MILLION)}M`;
  if (abs >= THOUSAND) return `$${(usd / THOUSAND).toFixed(1)}k`;
  if (abs < 0.01 && abs > 0) return `$${usd.toFixed(4)}`;
  return `$${usd.toFixed(2)}`;
}
