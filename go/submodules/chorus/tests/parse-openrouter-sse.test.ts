/**
 * Coverage for parseOpenRouterSSE — the OpenAI-compatible chat-completions
 * SSE chunk parser that powers the openrouter shim's runHeadless.
 *
 * Architecture: caller (the shim) splits the SSE stream on `\n\n` event
 * boundaries and strips the `data: ` prefix; this parser receives ONE
 * payload at a time. We assert it correctly emits text_delta + usage on
 * normal flow, swallows control payloads (`[DONE]`, comment lines, empty
 * lines), and forwards error envelopes as AgentEvents.
 */

import { describe, expect, it } from "vitest";
import { parseOpenRouterSSE } from "../src/daemon/agents/parsers/index.js";

describe("parseOpenRouterSSE", () => {
  it("emits text_delta for delta.content", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}',
    );
    expect(events).toEqual([{ type: "text_delta", text: "Hello" }]);
  });

  it("emits nothing for empty delta (finish chunk before usage)", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}',
    );
    expect(events).toEqual([]);
  });

  it("emits message_done with usage on usage-bearing chunk", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[],"usage":{"prompt_tokens":12,"completion_tokens":48,"total_tokens":60,"cost":0.00012}}',
    );
    expect(events).toHaveLength(1);
    const ev = events[0];
    expect(ev.type).toBe("message_done");
    if (ev.type !== "message_done") throw new Error("type guard");
    expect(ev.finalText).toBe("");
    expect(ev.usage).toEqual({
      inputTokens: 12,
      outputTokens: 48,
      costUsd: 0.00012,
    });
  });

  it("usage chunk without cost still emits tokens", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[],"usage":{"prompt_tokens":3,"completion_tokens":7}}',
    );
    expect(events).toHaveLength(1);
    const ev = events[0];
    if (ev.type !== "message_done") throw new Error("type guard");
    expect(ev.usage).toEqual({ inputTokens: 3, outputTokens: 7 });
  });

  it("[DONE] sentinel returns no events", () => {
    expect(parseOpenRouterSSE("[DONE]")).toEqual([]);
    expect(parseOpenRouterSSE(" [DONE] ")).toEqual([]);
  });

  it("empty / comment / non-JSON lines return no events", () => {
    expect(parseOpenRouterSSE("")).toEqual([]);
    expect(parseOpenRouterSSE("   ")).toEqual([]);
    expect(parseOpenRouterSSE(": OPENROUTER PROCESSING")).toEqual([]);
    expect(parseOpenRouterSSE("not even json")).toEqual([]);
  });

  it("forwards stream-error envelope as AgentEvent error", () => {
    const events = parseOpenRouterSSE(
      '{"error":{"message":"insufficient credits","code":402}}',
    );
    expect(events).toHaveLength(1);
    const ev = events[0];
    expect(ev.type).toBe("error");
    if (ev.type !== "error") throw new Error("type guard");
    expect(ev.kind).toBe("402");
    expect(ev.message).toBe("insufficient credits");
  });

  it("error envelope without code falls back to openrouter_error kind", () => {
    const events = parseOpenRouterSSE('{"error":{"message":"unspecified"}}');
    if (events[0].type !== "error") throw new Error("type guard");
    expect(events[0].kind).toBe("openrouter_error");
  });

  it("handles multi-choice chunk by concatenating deltas (defensive)", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[{"index":0,"delta":{"content":"A"}},{"index":1,"delta":{"content":"B"}}]}',
    );
    expect(events).toEqual([
      { type: "text_delta", text: "A" },
      { type: "text_delta", text: "B" },
    ]);
  });

  it("ignores delta.content of empty string (no spurious text_delta)", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[{"index":0,"delta":{"content":""}}]}',
    );
    expect(events).toEqual([]);
  });

  it("usage with non-numeric fields is gracefully ignored", () => {
    const events = parseOpenRouterSSE(
      '{"id":"gen-1","choices":[],"usage":{"prompt_tokens":"oops","completion_tokens":5}}',
    );
    if (events[0]?.type !== "message_done") throw new Error("type guard");
    expect(events[0].usage).toEqual({ outputTokens: 5 });
  });
});
