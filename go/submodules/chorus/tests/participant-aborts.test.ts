/**
 * Tests for the per-participant abort registry. These cover the
 * registry hygiene + signal-composition properties the route handler
 * depends on — chat-cancel still aborts everything, per-participant
 * cancel only fires the targeted controller, double-register doesn't
 * leak.
 */

import { afterEach, describe, expect, it } from "vitest";
import {
  abortParticipant,
  cleanupChat,
  participantKey,
  register,
  _activeKeys,
} from "../src/daemon/participant-aborts";

afterEach(() => {
  // Belt-and-braces: each test cleans up its own chats, but if a test
  // throws mid-flight we still want the registry empty for the next.
  cleanupChat("chat-a");
  cleanupChat("chat-b");
});

describe("participantKey", () => {
  it("formats reviewer keys with index", () => {
    expect(participantKey("reviewer", "codex-cli", 0)).toBe("reviewer-codex-cli-0");
    expect(participantKey("reviewer", "gemini-cli", 2)).toBe("reviewer-gemini-cli-2");
  });

  it("defaults reviewer index to 0 when omitted", () => {
    expect(participantKey("reviewer", "codex-cli")).toBe("reviewer-codex-cli-0");
  });

  it("formats doer keys without index", () => {
    expect(participantKey("doer", "claude-code")).toBe("doer-claude-code");
  });
});

describe("register / abortParticipant", () => {
  it("aborts ONLY the targeted participant — not siblings", () => {
    const parent = new AbortController();
    const a = register("chat-a", "reviewer-codex-cli-0", parent.signal);
    const b = register("chat-a", "reviewer-gemini-cli-1", parent.signal);

    abortParticipant("chat-a", "reviewer-codex-cli-0");

    expect(a.signal.aborted).toBe(true);
    expect(b.signal.aborted).toBe(false);

    a.release();
    b.release();
  });

  it("returns false when the participant key isn't tracked", () => {
    expect(abortParticipant("chat-a", "reviewer-not-here-0")).toBe(false);
  });

  it("propagates parent (chat-level) abort to ALL combined signals", () => {
    const parent = new AbortController();
    const a = register("chat-a", "reviewer-codex-cli-0", parent.signal);
    const b = register("chat-a", "reviewer-gemini-cli-1", parent.signal);

    parent.abort();

    expect(a.signal.aborted).toBe(true);
    expect(b.signal.aborted).toBe(true);

    a.release();
    b.release();
  });
});

// The route in daemon/index.ts validates the URL `key` param against this
// regex before touching the registry. Keep the source of truth here so a
// regression in the regex shows up alongside the registry tests.
const KEY_REGEX = /^(doer-|reviewer-)[A-Za-z0-9][A-Za-z0-9_-]*(?:-\d+)?$/;

describe("participant-key URL regex (route guard)", () => {
  it("accepts well-formed keys", () => {
    expect(KEY_REGEX.test("doer-claude-code")).toBe(true);
    expect(KEY_REGEX.test("reviewer-codex-cli-0")).toBe(true);
    expect(KEY_REGEX.test("reviewer-gemini-cli-2")).toBe(true);
    expect(KEY_REGEX.test("reviewer-opencode-cli-15")).toBe(true);
  });

  it("rejects empty agent name (the bug retroactive PR #24 review caught)", () => {
    expect(KEY_REGEX.test("reviewer--0")).toBe(false);
    expect(KEY_REGEX.test("doer-")).toBe(false);
    expect(KEY_REGEX.test("reviewer-")).toBe(false);
  });

  it("rejects path-traversal-shaped probes", () => {
    expect(KEY_REGEX.test("../etc/passwd")).toBe(false);
    expect(KEY_REGEX.test("doer-..")).toBe(false);
    expect(KEY_REGEX.test("reviewer-../-0")).toBe(false);
  });

  it("rejects unknown role prefixes", () => {
    expect(KEY_REGEX.test("admin-codex-cli-0")).toBe(false);
    expect(KEY_REGEX.test("codex-cli-0")).toBe(false);
    expect(KEY_REGEX.test("")).toBe(false);
  });

  it("rejects keys whose agent name starts with a separator", () => {
    expect(KEY_REGEX.test("doer--")).toBe(false);
    expect(KEY_REGEX.test("doer-_secret")).toBe(false);
    expect(KEY_REGEX.test("reviewer-_x-0")).toBe(false);
  });
});

describe("registry hygiene", () => {
  it("release() removes the entry so cleanup-after-finish works", () => {
    const parent = new AbortController();
    const handle = register("chat-a", "reviewer-codex-cli-0", parent.signal);
    expect(_activeKeys("chat-a")).toContain("reviewer-codex-cli-0");

    handle.release();
    expect(_activeKeys("chat-a")).not.toContain("reviewer-codex-cli-0");
  });

  it("cleanupChat drops every controller for a chat", () => {
    const parent = new AbortController();
    register("chat-a", "reviewer-codex-cli-0", parent.signal);
    register("chat-a", "reviewer-gemini-cli-1", parent.signal);
    register("chat-b", "doer-claude-code", parent.signal);

    cleanupChat("chat-a");

    expect(_activeKeys("chat-a")).toEqual([]);
    expect(_activeKeys("chat-b")).toContain("doer-claude-code");
  });

  it("re-registering the same key aborts the previous controller (no leak across retries)", () => {
    const parent = new AbortController();
    const first = register("chat-a", "reviewer-codex-cli-0", parent.signal);
    const second = register("chat-a", "reviewer-codex-cli-0", parent.signal);

    // The previous controller fired so any in-flight runner from the
    // earlier retry exits cleanly.
    expect(first.signal.aborted).toBe(true);
    // The new one is still live.
    expect(second.signal.aborted).toBe(false);

    second.release();
  });

  it("release() is a no-op when the entry has already been overwritten by a re-register", () => {
    // This guards against the "stale runner finally-block deletes the
    // freshly-replaced controller" race.
    const parent = new AbortController();
    const first = register("chat-a", "reviewer-codex-cli-0", parent.signal);
    const second = register("chat-a", "reviewer-codex-cli-0", parent.signal);

    first.release();

    // The second controller's entry must still be tracked.
    expect(_activeKeys("chat-a")).toContain("reviewer-codex-cli-0");
    second.release();
  });
});
