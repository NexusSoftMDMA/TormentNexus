/**
 * Coverage for opencode token-usage extraction.
 *
 * Why: parseOpencode used to emit only text_delta events; token counts
 * lived in step_finish but the runner never saw them, so the cockpit
 * time/tokens chip was empty for opencode + kimi-via-opencode-go.
 *
 * Architecture (post retroactive PR #25 review): parseOpencode stays
 * line-stateless and only emits text_delta. parseOpencodeExit walks the
 * whole stdout, sums tokens across EVERY step_finish event, and emits
 * exactly one message_done with finalText="" + aggregated usage. This
 * fixes the multi-step-overwrite bug all 3 reviewers flagged.
 */

import { describe, expect, it } from "vitest";
import {
  parseOpencode,
  parseOpencodeExit,
} from "../src/daemon/agents/parsers/index.js";

const TEXT_LINE =
  '{"type":"text","timestamp":1,"sessionID":"ses_1","part":{"type":"text","text":"Hello!","time":{"start":0,"end":1}}}';

const STEP_FINISH_LINE =
  '{"type":"step_finish","timestamp":2,"sessionID":"ses_1","part":{"type":"step-finish","tokens":{"total":14655,"input":12712,"output":3,"reasoning":20,"cache":{"write":0,"read":1920}},"cost":0.022}}';

describe("parseOpencode (line-stateless)", () => {
  it("emits text_delta for text events", () => {
    expect(parseOpencode(TEXT_LINE)).toEqual([
      { type: "text_delta", text: "Hello!" },
    ]);
  });

  it("ignores empty text bodies", () => {
    const empty = JSON.stringify({
      type: "text",
      part: { type: "text", text: "" },
    });
    expect(parseOpencode(empty)).toEqual([]);
  });

  it("does NOT emit message_done on step_finish (aggregated by parseOpencodeExit instead)", () => {
    // This is the fix from PR #25 round-2 review: per-line message_done
    // would clobber the runner's finalText and fire participant_done
    // multiple times. Aggregation lives in parseOpencodeExit now.
    expect(parseOpencode(STEP_FINISH_LINE)).toEqual([]);
  });

  it("returns [] for unknown event types and malformed lines", () => {
    expect(parseOpencode('{"type":"step_start","part":{}}')).toEqual([]);
    expect(parseOpencode("not json")).toEqual([]);
    expect(parseOpencode("")).toEqual([]);
  });
});

describe("parseOpencodeExit — JSON-Lines aggregation", () => {
  it("emits a single message_done with summed tokens across multiple step_finish events", () => {
    const stream = [
      '{"type":"step_start","part":{}}',
      TEXT_LINE,
      '{"type":"step_finish","part":{"tokens":{"input":1000,"output":50,"cache":{"read":200}}}}',
      '{"type":"text","part":{"type":"text","text":" then more"}}',
      '{"type":"step_finish","part":{"tokens":{"input":500,"output":25,"cache":{"read":100}}}}',
    ].join("\n");

    const events = parseOpencodeExit(stream);
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe("message_done");
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].finalText).toBe("");
    expect(events[0].usage).toEqual({
      inputTokens: 1500,
      outputTokens: 75,
      cachedInputTokens: 300,
    });
  });

  it("emits a single message_done for a single step_finish (degenerate aggregation)", () => {
    const stream = [TEXT_LINE, STEP_FINISH_LINE].join("\n");
    const events = parseOpencodeExit(stream);
    expect(events).toHaveLength(1);
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].usage).toEqual({
      inputTokens: 12712,
      outputTokens: 3,
      cachedInputTokens: 1920,
      costUsd: 0.022,
    });
  });

  it("sums cost across multiple step_finish events", () => {
    const stream = [
      '{"type":"step_finish","part":{"tokens":{"input":1000,"output":50},"cost":0.015}}',
      '{"type":"step_finish","part":{"tokens":{"input":500,"output":25},"cost":0.007}}',
    ].join("\n");
    const events = parseOpencodeExit(stream);
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].usage?.costUsd).toBeCloseTo(0.022, 4);
  });

  it("aggregates cost even when one step_finish has malformed tokens", () => {
    const stream = [
      // Real cost, malformed tokens — still surface the cost.
      '{"type":"step_finish","part":{"tokens":{"input":"twelve"},"cost":0.0042}}',
    ].join("\n");
    const events = parseOpencodeExit(stream);
    expect(events).toHaveLength(1);
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].usage?.costUsd).toBeCloseTo(0.0042, 4);
    expect(events[0].usage?.inputTokens).toBeUndefined();
  });

  it("emits message_done with no usage when JSON-Lines stream has no step_finish events", () => {
    // Updated post-launch-eve review (gemini + deepseek): previously this
    // returned [], which dropped the terminal event entirely. The runner's
    // `for await` loop would exit without firing participant_done, leaving
    // the phase sitting in `working` until the watchdog timeout. The fix
    // emits exactly one message_done whenever JSON-Lines is detected;
    // usage is attached only when step_finish provided it.
    const stream = [TEXT_LINE, '{"type":"step_start","part":{}}'].join("\n");
    const events = parseOpencodeExit(stream);
    expect(events).toHaveLength(1);
    expect(events[0]).toEqual({ type: "message_done", finalText: "" });
  });

  it("emits message_done with no usage when step_finish events have no usable tokens (malformed)", () => {
    const stream = [
      TEXT_LINE,
      '{"type":"step_finish","part":{"reason":"stop"}}',
      '{"type":"step_finish","part":{"tokens":{"input":"twelve","output":null}}}',
    ].join("\n");
    const events = parseOpencodeExit(stream);
    expect(events).toHaveLength(1);
    expect(events[0]).toEqual({ type: "message_done", finalText: "" });
  });

  it("partially aggregates when some step_finish events lack a field", () => {
    const stream = [
      '{"type":"step_finish","part":{"tokens":{"input":1000,"output":10}}}',
      // No cache.read on second step — sum should still produce input + output.
      '{"type":"step_finish","part":{"tokens":{"input":500,"output":5}}}',
    ].join("\n");
    const events = parseOpencodeExit(stream);
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].usage).toEqual({
      inputTokens: 1500,
      outputTokens: 15,
    });
    expect(events[0].usage?.cachedInputTokens).toBeUndefined();
  });
});

describe("parseOpencodeExit — single-blob fallback", () => {
  it("emits message_done with text from `message` field on older opencode builds", () => {
    const blob = '{"message":"Reviewed: looks good","cost":0.0}';
    const events = parseOpencodeExit(blob);
    expect(events).toHaveLength(1);
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].finalText).toBe("Reviewed: looks good");
    expect(events[0].usage).toBeUndefined();
  });

  it("falls back to raw stdout when JSON parse fails", () => {
    const blob = "not json at all";
    const events = parseOpencodeExit(blob);
    if (events[0].type !== "message_done") throw new Error("type guard");
    expect(events[0].finalText).toBe("not json at all");
  });

  it("returns [] for empty stdout", () => {
    expect(parseOpencodeExit("")).toEqual([]);
    expect(parseOpencodeExit("   \n\n  ")).toEqual([]);
  });
});
