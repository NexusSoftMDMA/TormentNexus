/**
 * Coverage for pickShimForVoice — the dispatch hook that routes a (lineage,
 * model) pair to either a CLI shim or the OpenRouter HTTP shim.
 *
 * Why: templates declare lineage + an optional model id. When the model
 * starts with `openrouter:`, dispatch must go through the HTTP shim
 * regardless of the declared lineage (lineage stays accurate for diversity
 * scoring). Other model strings — including bare CLI model ids — should
 * resolve via the existing lineage→shim map.
 */

import { describe, expect, it } from "vitest";
import {
  pickShimForVoice,
  isHttpDispatchedShim,
  openrouterShim,
  claudeShim,
  codexShim,
  geminiShim,
  opencodeShim,
  kimiShim,
} from "../src/daemon/agents/index.js";

describe("pickShimForVoice", () => {
  it("returns openrouterShim for openrouter:* model regardless of lineage", () => {
    expect(pickShimForVoice("anthropic", "openrouter:anthropic/claude-3.5-sonnet")).toBe(
      openrouterShim,
    );
    expect(pickShimForVoice("openai", "openrouter:openai/gpt-4")).toBe(openrouterShim);
    expect(pickShimForVoice("google", "openrouter:google/gemini-pro")).toBe(
      openrouterShim,
    );
    expect(pickShimForVoice("opencode", "openrouter:moonshotai/kimi-k2")).toBe(
      openrouterShim,
    );
    expect(pickShimForVoice("any", "openrouter:meta/llama-3.1-405b")).toBe(
      openrouterShim,
    );
  });

  it("returns lineage-mapped CLI shim when no openrouter prefix", () => {
    expect(pickShimForVoice("anthropic", "claude-opus-4-7")).toBe(claudeShim);
    expect(pickShimForVoice("openai", "gpt-5.5")).toBe(codexShim);
    expect(pickShimForVoice("google", "gemini-3.1-pro-preview")).toBe(geminiShim);
    expect(pickShimForVoice("opencode", "deepseek-v4-pro")).toBe(opencodeShim);
    expect(pickShimForVoice("moonshot", "kimi-k2.6")).toBe(kimiShim);
  });

  it("returns lineage-mapped CLI shim when model is undefined", () => {
    expect(pickShimForVoice("anthropic", undefined)).toBe(claudeShim);
    expect(pickShimForVoice("openai")).toBe(codexShim);
  });

  it("openrouter prefix wins over CLI lineage even when CLI lineage names match", () => {
    // Catch the failure mode: a template declares lineage='anthropic' AND
    // model='openrouter:anthropic/claude-3.5-sonnet'. The runner must NOT
    // pick claudeShim (CLI) — it must pick openrouterShim (HTTP).
    expect(pickShimForVoice("anthropic", "openrouter:anthropic/claude-3.5-sonnet")).toBe(
      openrouterShim,
    );
    expect(
      pickShimForVoice("anthropic", "openrouter:anthropic/claude-3.5-sonnet").name,
    ).toBe("openrouter");
  });

  it("does not match a model id that merely contains 'openrouter:' substring", () => {
    // Defensive: a hypothetical CLI model named "claude-openrouter:variant"
    // (unlikely but safe) should still go to claudeShim because the prefix
    // check is anchored at position 0.
    expect(pickShimForVoice("anthropic", "claude-openrouter:weird")).toBe(claudeShim);
  });
});

describe("isHttpDispatchedShim", () => {
  it("is true only for openrouterShim", () => {
    expect(isHttpDispatchedShim(openrouterShim)).toBe(true);
    expect(isHttpDispatchedShim(claudeShim)).toBe(false);
    expect(isHttpDispatchedShim(codexShim)).toBe(false);
    expect(isHttpDispatchedShim(geminiShim)).toBe(false);
    expect(isHttpDispatchedShim(opencodeShim)).toBe(false);
    expect(isHttpDispatchedShim(kimiShim)).toBe(false);
  });
});
