/**
 * Vitest equivalents of the previous inline `runTests` self-checks in
 * src/daemon/agents/parsers.ts. Same fixtures, same assertions; lifted
 * into a proper test harness as part of the parsers/ split.
 */
import { describe, expect, it } from 'vitest';
import {
  parseClaude,
  parseGemini,
  parseCodexExit,
  parseOpencodeExit,
} from '../src/daemon/agents/parsers/index.js';
import type { AgentEvent } from '../src/daemon/agents/types.js';

describe('parseClaude — real fixture (Claude Code 2.1.123, captured 2026-04-30)', () => {
  const fixture = [
    '{"type":"system","subtype":"init","cwd":"/tmp","session_id":"abc"}',
    '{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_01"}},"session_id":"abc"}',
    '{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}},"session_id":"abc"}',
    '{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hi there"}},"session_id":"abc"}',
    '{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" friend"}},"session_id":"abc"}',
    '{"type":"stream_event","event":{"type":"content_block_stop","index":0},"session_id":"abc"}',
    '{"type":"stream_event","event":{"type":"message_stop"},"session_id":"abc"}',
    '{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"}}',
    '{"type":"result","subtype":"success","is_error":false,"result":"Hi there friend","total_cost_usd":0.26}',
  ];

  it('emits 2 text_deltas + 1 message_done with assembled finalText', () => {
    const events: AgentEvent[] = [];
    for (const line of fixture) events.push(...parseClaude(line));

    const textDeltas = events.filter((e) => e.type === 'text_delta');
    expect(textDeltas).toHaveLength(2);
    expect((textDeltas[0] as { text: string }).text).toBe('Hi there');
    expect((textDeltas[1] as { text: string }).text).toBe(' friend');

    const dones = events.filter((e) => e.type === 'message_done');
    expect(dones).toHaveLength(1);
    expect((dones[0] as { finalText: string }).finalText).toBe('Hi there friend');
  });

  it('extracts tool_use as tool_call_start with name + input', () => {
    const toolLine =
      '{"type":"stream_event","event":{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","name":"Read","input":{"file_path":"/tmp/x.ts"}}},"session_id":"abc"}';
    const toolEvents = parseClaude(toolLine);
    expect(toolEvents).toHaveLength(1);
    expect(toolEvents[0].type).toBe('tool_call_start');
    expect((toolEvents[0] as { tool: string }).tool).toBe('Read');
  });

  it('emits error event on result.subtype=error', () => {
    const errEvents = parseClaude(
      '{"type":"result","subtype":"error","is_error":true,"result":"rate limited"}',
    );
    expect(errEvents).toHaveLength(1);
    expect(errEvents[0].type).toBe('error');
  });

  it('returns [] for malformed/system/empty lines', () => {
    expect(parseClaude('not json')).toHaveLength(0);
    expect(parseClaude('')).toHaveLength(0);
    expect(parseClaude('{"type":"system","subtype":"hook_started"}')).toHaveLength(0);
  });

  // Plan-equivalent spend on the home page silently undercounted by the
  // largest contributor (Claude is most users' primary voice) because
  // the result-line `total_cost_usd` + `usage` block weren't extracted.
  // Lock in the lift so a future refactor can't drop it again.
  it('lifts total_cost_usd + usage into the message_done event', () => {
    const events = parseClaude(
      '{"type":"result","subtype":"success","is_error":false,"result":"hi",' +
        '"total_cost_usd":0.42,' +
        '"usage":{"input_tokens":1500,"output_tokens":300,"cache_read_input_tokens":12000}}',
    );
    expect(events).toHaveLength(1);
    const ev = events[0] as {
      type: string;
      finalText: string;
      usage?: {
        inputTokens?: number;
        outputTokens?: number;
        cachedInputTokens?: number;
        costUsd?: number;
      };
    };
    expect(ev.type).toBe('message_done');
    expect(ev.usage?.costUsd).toBe(0.42);
    expect(ev.usage?.inputTokens).toBe(1500);
    expect(ev.usage?.outputTokens).toBe(300);
    expect(ev.usage?.cachedInputTokens).toBe(12000);
  });
});

describe('parseGemini — real fixture (gemini-cli, captured 2026-04-30)', () => {
  const fixture = [
    '{"type":"init","timestamp":"2026-04-30T12:16:50.412Z","session_id":"sess-1","model":"gemini-3.1-pro-preview"}',
    '{"type":"message","timestamp":"2026-04-30T12:16:50.416Z","role":"user","content":"say hi in 3 words"}',
    '{"type":"message","timestamp":"2026-04-30T12:16:54.358Z","role":"assistant","content":"Hi there friend!","delta":true}',
    '{"type":"result","timestamp":"2026-04-30T12:16:54.419Z","status":"success","stats":{"total_tokens":11638}}',
  ];

  it('emits 1 text_delta + 1 message_done with empty finalText', () => {
    const events: AgentEvent[] = [];
    for (const line of fixture) events.push(...parseGemini(line));

    const deltas = events.filter((e) => e.type === 'text_delta');
    expect(deltas).toHaveLength(1);
    expect((deltas[0] as { text: string }).text).toBe('Hi there friend!');

    const dones = events.filter((e) => e.type === 'message_done');
    expect(dones).toHaveLength(1);
    // Gemini's result line carries no text — runner falls back to its
    // accumulator. The empty finalText is the contract.
    expect((dones[0] as { finalText: string }).finalText).toBe('');
  });

  it('emits error event on result.status=error', () => {
    const events = parseGemini('{"type":"result","status":"error","error":"rate limit"}');
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('error');
  });

  it('detects quota exhaustion and surfaces the reset window', () => {
    const events = parseGemini(
      '{"type":"result","status":"error","error":{"message":"You have exhausted your capacity on this model. Your quota will reset after 6h23m52s."}}',
    );
    expect(events).toHaveLength(1);
    expect((events[0] as { kind: string }).kind).toBe('quota_exhausted');
    const message = (events[0] as { message: string }).message;
    expect(message).toMatch(/quota/i);
    expect(message).toMatch(/6h23m52s/);
  });

  it('digs nested error.cause.message when result.error is an object', () => {
    const events = parseGemini(
      '{"type":"result","status":"error","error":{"cause":{"message":"upstream is on fire"}}}',
    );
    expect(events).toHaveLength(1);
    expect((events[0] as { message: string }).message).toBe('upstream is on fire');
  });
});

describe('parseGeminiExit — stderr-driven quota detection', () => {
  it('emits quota_exhausted with reset window from stderr', async () => {
    const { parseGeminiExit } = await import('@/daemon/agents/parsers/gemini');
    const stderr =
      'Error: You have exhausted your capacity on this model. Your quota will reset after 8h14m16s.\n' +
      '    at handle ()  cause: { code: 429, reason: "QUOTA_EXHAUSTED" }';
    const events = parseGeminiExit('', stderr, 1);
    expect(events).toHaveLength(1);
    expect((events[0] as { kind: string }).kind).toBe('quota_exhausted');
    expect((events[0] as { message: string }).message).toMatch(/8h14m16s/);
  });

  it('returns no events when stderr does not look like a quota error', async () => {
    const { parseGeminiExit } = await import('@/daemon/agents/parsers/gemini');
    expect(parseGeminiExit('', 'random stderr noise', 1)).toEqual([]);
    expect(parseGeminiExit('', '', 0)).toEqual([]);
  });

  it('emits auth_error when GEMINI_API_KEY env var is missing', async () => {
    const { parseGeminiExit } = await import('@/daemon/agents/parsers/gemini');
    const stderr =
      '(node:105820) [DEP0040] DeprecationWarning: punycode...\n' +
      "Error connecting to MCP server 'chorus': MCP error -32000: Connection closed\n" +
      'GEMINI_API_KEY environment variable not found. Add that to your environment and try again (no reload needed if using .env)!';
    const events = parseGeminiExit('', stderr, 1);
    expect(events).toHaveLength(1);
    expect((events[0] as { kind: string }).kind).toBe('auth_error');
    expect((events[0] as { message: string }).message).toMatch(
      /GEMINI_API_KEY/,
    );
    expect((events[0] as { message: string }).message).toMatch(
      /aistudio\.google\.com/,
    );
  });

  it('also catches GOOGLE_API_KEY variant (legacy alias)', async () => {
    const { parseGeminiExit } = await import('@/daemon/agents/parsers/gemini');
    const stderr = 'GOOGLE_API_KEY not set. Configure your environment.';
    const events = parseGeminiExit('', stderr, 1);
    expect(events).toHaveLength(1);
    expect((events[0] as { kind: string }).kind).toBe('auth_error');
  });

  // gemini-cli 0.40.x changed the auth error wording — the prior regex
  // matched only "_API_KEY ... not found/set" which slipped past this
  // newer format. User dogfood (chorus-codes 2026-05-06) caught it: a
  // gemini reviewer fell through to lineage_fallback with no clear CTA.
  it('catches gemini-cli 0.40.x "set an Auth method" wording', async () => {
    const { parseGeminiExit } = await import('@/daemon/agents/parsers/gemini');
    const stderr =
      'Please set an Auth method in your /root/.gemini/settings.json or ' +
      'specify one of the following environment variables before running: ' +
      'GEMINI_API_KEY, GOOGLE_GENAI_USE_VERTEXAI, GOOGLE_GENAI_USE_GCA';
    const events = parseGeminiExit('', stderr, 41);
    expect(events).toHaveLength(1);
    expect((events[0] as { kind: string }).kind).toBe('auth_error');
    expect((events[0] as { message: string }).message).toMatch(/GEMINI_API_KEY/);
  });
});

describe('parseOpencodeExit — single-blob fallback (older opencode builds)', () => {
  it('lifts message field into finalText', () => {
    const events = parseOpencodeExit('{"message":"Reviewed: looks good","cost":0.0}');
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('message_done');
    expect((events[0] as { finalText: string }).finalText).toBe('Reviewed: looks good');
  });

  it('falls back to raw stdout when JSON parse fails', () => {
    const raw = 'Reviewed in plain text\n';
    const events = parseOpencodeExit(raw);
    expect(events).toHaveLength(1);
    expect((events[0] as { finalText: string }).finalText).toBe(raw);
  });
});

describe('parseCodexExit', () => {
  it('happy path — emits message_done with stdout when code=0', () => {
    const events = parseCodexExit('Code reviewed.\nVerdict: approve.', '', 0);
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('message_done');
  });

  it('detects quota_exhausted from stderr ERROR line (real ChatGPT-plan codex output)', () => {
    const stderr =
      "OpenAI Codex v0.128.0 (research preview)\n" +
      "session id: 019de827-...\n" +
      "user\nsay hello\n" +
      "ERROR: You've hit your usage limit. Upgrade to Plus to continue using Codex (https://chatgpt.com/explore/plus), or try again at May 8th, 2026 9:05 PM.\n";
    const events = parseCodexExit('', stderr, 1);
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('error');
    expect((events[0] as { kind: string }).kind).toBe('quota_exhausted');
    expect((events[0] as { message: string }).message).toMatch(/usage limit/i);
  });

  it('emits cli_error (not quota) for generic non-zero exit', () => {
    const events = parseCodexExit('', 'panic: something\nbye\n', 134);
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('error');
    expect((events[0] as { kind: string }).kind).toBe('cli_error');
  });

  it('returns [] for silent zero-exit (legacy "no output" no-op)', () => {
    expect(parseCodexExit('', '', 0)).toHaveLength(0);
  });

  // Round-1 review (PR #9) caught the unanchored alternation. Codex
  // `exec` echoes the user prompt back into stderr; a legitimate review
  // brief containing "try again at midnight" or "upgrade to Plus" used
  // to false-positive into quota_exhausted, dropping the real crash
  // diagnostic. The gate is now anchored on the literal `ERROR:` prefix.
  it('does NOT misclassify echoed prompt phrases as quota when the literal ERROR: prefix is missing', () => {
    const echoedStderr =
      "user\nReview the doc that says 'try again at midnight' and 'upgrade to Plus'.\n" +
      "panic: codex worker crashed\n";
    const events = parseCodexExit('', echoedStderr, 1);
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('error');
    expect((events[0] as { kind: string }).kind).toBe('cli_error');
  });
});
