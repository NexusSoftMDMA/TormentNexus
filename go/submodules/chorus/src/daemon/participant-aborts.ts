/**
 * Per-participant abort registry.
 *
 * Chat-level cancel goes through the chat-wide `AbortController` in
 * `runWithMultiplex` and tears down every runner. This module covers the
 * narrower case: cancelling a single reviewer (or doer) without
 * collapsing the whole chat. The runner registers a child controller
 * here when a participant subprocess starts; the cockpit hits a
 * dedicated route to fire it.
 *
 * Why a Map keyed by chatId → participantKey:
 *   - O(1) lookup from the route handler
 *   - Lets `cleanupChat(chatId)` wipe stale entries when a chat ends
 *   - Naturally scopes participant keys (a reviewer-1 in chat A doesn't
 *     collide with reviewer-1 in chat B)
 *
 * Combining with the parent signal: each `register` returns a
 * `combinedSignal` that fires when EITHER the parent (chat-level) signal
 * fires OR `abortParticipant` is called for this key. We use
 * AbortSignal.any (Node 20+) to compose. This way chat-cancel still
 * propagates to every runner without duplicating wiring at every call
 * site.
 */

const registry = new Map<string, Map<string, AbortController>>();

export interface ParticipantAbortHandle {
  /** Combined signal — passes to runDoerHeadless / runReviewerHeadless. */
  signal: AbortSignal;
  /** Call in a `finally` block so an exited runner doesn't leave a stale controller. */
  release: () => void;
}

/**
 * Build the participant key. Doers carry `agentName`; reviewers carry
 * `agentName` + index. Format kept stable so the route handler can
 * accept the same shape from the URL path.
 */
export function participantKey(
  role: 'doer' | 'reviewer',
  agentName: string,
  reviewerIdx?: number,
): string {
  return role === 'reviewer'
    ? `reviewer-${agentName}-${reviewerIdx ?? 0}`
    : `doer-${agentName}`;
}

/**
 * Register a runner for per-participant cancellation.
 *
 * @param chatId  Chat the runner belongs to.
 * @param key     Output of `participantKey(...)`.
 * @param parent  The chat-wide signal — its abort fires the combined signal too.
 */
export function register(
  chatId: string,
  key: string,
  parent: AbortSignal,
): ParticipantAbortHandle {
  let chatMap = registry.get(chatId);
  if (!chatMap) {
    chatMap = new Map();
    registry.set(chatId, chatMap);
  }
  // Dedup: if a previous runner with the same key didn't release (crash,
  // race), aborting it now is harmless and avoids leaking the prior
  // controller. The new runner gets a fresh controller below.
  const previous = chatMap.get(key);
  if (previous) {
    try { previous.abort(); } catch { /* ignore */ }
  }

  const child = new AbortController();
  chatMap.set(key, child);

  // AbortSignal.any composes — fires when either parent or child aborts.
  // Standardised in Node 20.x; engines pin in package.json gates lower.
  const signal = AbortSignal.any([parent, child.signal]);

  return {
    signal,
    release: () => {
      const cur = registry.get(chatId);
      if (!cur) return;
      // Only remove if the entry still points at THIS controller — a
      // re-registered key from a retried run shouldn't be wiped here.
      if (cur.get(key) === child) {
        cur.delete(key);
        if (cur.size === 0) registry.delete(chatId);
      }
    },
  };
}

/**
 * Fire the controller for a single participant. Returns true if a
 * controller was found and aborted, false if the participant isn't
 * tracked (already finished, never started, wrong key).
 */
export function abortParticipant(chatId: string, key: string): boolean {
  const chatMap = registry.get(chatId);
  if (!chatMap) return false;
  const controller = chatMap.get(key);
  if (!controller) return false;
  try {
    controller.abort();
    return true;
  } catch {
    return false;
  }
}

/**
 * Drop all controllers for a chat — called when the chat ends so the
 * registry doesn't accrete stale entries across long daemon uptime.
 * Doesn't fire any controllers: the chat is already terminating, the
 * underlying processes are exiting through their own signals.
 */
export function cleanupChat(chatId: string): void {
  registry.delete(chatId);
}

/**
 * Test/debug: peek at the keys for a chat. Production code never reads
 * this; exported so vitest can assert registry hygiene.
 */
export function _activeKeys(chatId: string): string[] {
  const chatMap = registry.get(chatId);
  return chatMap ? Array.from(chatMap.keys()) : [];
}
