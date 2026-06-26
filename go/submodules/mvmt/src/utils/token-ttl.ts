const TTL_RE = /^(\d+)([mhd])$/;
const DEFAULT_TOKEN_TTL = '30d';

export interface ParsedTokenTtl {
  label: string;
  expiresAt?: string;
}

export function defaultTokenTtl(): string {
  return DEFAULT_TOKEN_TTL;
}

export function parseTokenTtl(value: string | undefined, now = Date.now()): ParsedTokenTtl {
  const raw = (value ?? DEFAULT_TOKEN_TTL).trim().toLowerCase();
  if (raw === 'never') return { label: 'never', expiresAt: undefined };

  const match = TTL_RE.exec(raw);
  if (!match) throw new Error('TTL must look like 30m, 7d, or never.');

  const amount = Number(match[1]);
  if (!Number.isInteger(amount) || amount <= 0) throw new Error('TTL must be greater than zero.');

  const unit = match[2];
  const multiplier = unit === 'm'
    ? 60_000
    : unit === 'h'
      ? 60 * 60_000
      : 24 * 60 * 60_000;
  return {
    label: raw,
    expiresAt: new Date(now + amount * multiplier).toISOString(),
  };
}

export function formatTokenExpiry(expiresAt: string | undefined): string {
  return expiresAt ?? 'never';
}

export function isExpired(expiresAt: string | undefined, now = Date.now()): boolean {
  if (!expiresAt) return false;
  const timestamp = Date.parse(expiresAt);
  return !Number.isFinite(timestamp) || timestamp <= now;
}
