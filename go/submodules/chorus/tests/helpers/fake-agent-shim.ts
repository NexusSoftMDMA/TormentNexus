/**
 * In-process fake of AgentShim for runner integration tests.
 *
 * Replaces the real CLI subprocess (claude / codex / gemini / etc) with a
 * scripted async iterator of AgentEvents. The runner never knows the
 * difference: it consumes the same `text_delta` / `tool_call_start` /
 * `message_done` / `error` shape, writes the same answer.md, emits the
 * same SSE events.
 *
 * Two construction modes:
 *
 *   makeFakeShim({ events: [...] })
 *     — fixed event list, replayed exactly once on runHeadless().
 *
 *   makeFakeShim({ script: (opts) => events })
 *     — event list is computed from the spawn options (so the same shim
 *     can return different scripts per call when the test reuses it).
 *
 * Optionally pass `delayMs` to insert artificial pauses between events,
 * useful for testing concurrency caps where the test asserts
 * "max-N in flight at any moment."
 */
import type {
  AgentShim,
  AgentEvent,
  HeadlessSpawnOptions,
} from '../../src/daemon/agents/types';

export interface FakeShimRecord {
  options: HeadlessSpawnOptions;
  startedAt: number;
  finishedAt?: number;
}

export interface FakeShimConfig {
  /** Static event list to replay. Mutually exclusive with `script`. */
  events?: AgentEvent[];
  /** Per-call event factory. Receives the spawn options. */
  script?: (opts: HeadlessSpawnOptions) => AgentEvent[];
  /** Optional artificial pause between events. */
  delayMs?: number;
  /** Lineage to report back; defaults to 'anthropic'. */
  lineage?: AgentShim['lineage'];
  /** Friendly name; defaults to 'fake'. */
  name?: string;
}

/**
 * Live record of every runHeadless invocation against this shim.
 * Tests inspect `.calls` to assert per-shim concurrency, ordering, etc.
 */
export interface FakeShimHandle {
  shim: AgentShim;
  calls: FakeShimRecord[];
  /** Snapshot of how many calls are currently mid-stream. */
  inFlight(): number;
}

const sleep = (ms: number) =>
  ms > 0 ? new Promise<void>((r) => setTimeout(r, ms)) : Promise.resolve();

export function makeFakeShim(config: FakeShimConfig): FakeShimHandle {
  const calls: FakeShimRecord[] = [];

  const shim: AgentShim = {
    lineage: config.lineage ?? 'anthropic',
    name: config.name ?? 'fake',
    buildLaunchCommand: () => 'fake-cli',
    formatPrompt: () => 'fake prompt',
    estimateCostUsd: () => 0,
    runHeadless(opts: HeadlessSpawnOptions): AsyncIterable<AgentEvent> {
      const record: FakeShimRecord = { options: opts, startedAt: Date.now() };
      calls.push(record);
      const events = config.script
        ? config.script(opts)
        : config.events ?? [];
      const delayMs = config.delayMs ?? 0;

      async function* gen(): AsyncIterable<AgentEvent> {
        try {
          for (const ev of events) {
            // Cooperative cancel: bail if the test aborts mid-script.
            if (opts.abortSignal?.aborted) {
              yield {
                type: 'error',
                kind: 'aborted',
                message: 'aborted by test',
              };
              return;
            }
            if (delayMs > 0) await sleep(delayMs);
            yield ev;
          }
        } finally {
          record.finishedAt = Date.now();
        }
      }
      return gen();
    },
  };

  return {
    shim,
    calls,
    inFlight: () =>
      calls.filter((c) => c.finishedAt === undefined).length,
  };
}

/**
 * Convenience builder for the most common script: a few text deltas
 * followed by a message_done. Saves boilerplate in tests.
 */
export function happyPathEvents(
  body: string,
  opts: { chunks?: number } = {},
): AgentEvent[] {
  const chunks = Math.max(1, opts.chunks ?? 3);
  const size = Math.ceil(body.length / chunks);
  const events: AgentEvent[] = [];
  for (let i = 0; i < chunks; i++) {
    events.push({ type: 'text_delta', text: body.slice(i * size, (i + 1) * size) });
  }
  events.push({ type: 'message_done', finalText: body });
  return events;
}
