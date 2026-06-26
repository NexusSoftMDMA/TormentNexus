export const DEFAULT_MAX_RECONNECT_ATTEMPTS = 3;
export const DEFAULT_RECONNECT_BASE_DELAY_MS = 3000;
export const DEFAULT_RECONNECT_MAX_DELAY_MS = 15000;

export interface ReconnectPolicy {
  maxAttempts: number;
  baseDelayMs: number;
  maxDelayMs: number;
}

export function createReconnectPolicy(partial?: Partial<ReconnectPolicy>): ReconnectPolicy {
  return {
    maxAttempts: partial?.maxAttempts ?? DEFAULT_MAX_RECONNECT_ATTEMPTS,
    baseDelayMs: partial?.baseDelayMs ?? DEFAULT_RECONNECT_BASE_DELAY_MS,
    maxDelayMs: partial?.maxDelayMs ?? DEFAULT_RECONNECT_MAX_DELAY_MS,
  };
}

export function shouldRetryReconnect(attemptsSoFar: number, policy: ReconnectPolicy): boolean {
  return attemptsSoFar < policy.maxAttempts;
}

export function getReconnectDelayMs(nextAttempt: number, policy: ReconnectPolicy): number {
  const safeAttempt = Math.max(1, nextAttempt);
  const delay = policy.baseDelayMs * Math.pow(2, safeAttempt - 1);
  return Math.min(delay, policy.maxDelayMs);
}

export function normalizeNumericInput(
  value: string,
  fallback: number,
  min: number,
  max: number,
): number {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }

  return Math.min(max, Math.max(min, parsed));
}
