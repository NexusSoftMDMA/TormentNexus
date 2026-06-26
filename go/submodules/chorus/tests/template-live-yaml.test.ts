/**
 * Tests for selectLiveYaml — the source-of-truth precedence chain that
 * decides which YAML the TemplateDialog hands to the daemon on save.
 *
 * The bug this guards against: a builtin template opened in Form mode and
 * saved without any edit was being byte-shifted by buildYamlFromForm
 * (comment loss, whitespace normalisation), which tripped the server's
 * source-promotion guard and silently flipped the row from 'builtin' to
 * 'user'. Round 3 of PR #10 review specifically called out the lack of a
 * test for Fix C — this is that test.
 */

import { describe, expect, it } from "vitest";
import { selectLiveYaml, type LiveYamlState } from "../src/lib/template-live-yaml";

const baseState: LiveYamlState = {
  yamlDirty: false,
  tab: "form",
  formDirty: false,
  yamlText: "yaml: pane-edited",
  formYaml: "yaml: form-restringified",
  initialYaml: "yaml: original\n# preserved comment",
};

describe("selectLiveYaml", () => {
  it("returns initialYaml when nothing is dirty and tab=form (the no-edit-on-builtin guard)", () => {
    expect(selectLiveYaml(baseState)).toBe(baseState.initialYaml);
  });

  it("returns yamlText when the user typed in the YAML pane (yamlDirty wins regardless of tab)", () => {
    const state: LiveYamlState = { ...baseState, yamlDirty: true };
    expect(selectLiveYaml(state)).toBe(state.yamlText);
  });

  it("returns yamlText when the user is on the YAML tab even if they haven't typed", () => {
    // Reading raw YAML on the YAML tab should not silently restringify on save.
    const state: LiveYamlState = { ...baseState, tab: "yaml" };
    expect(selectLiveYaml(state)).toBe(state.yamlText);
  });

  it("returns formYaml when only formDirty is set (the genuine form-edit path)", () => {
    const state: LiveYamlState = { ...baseState, formDirty: true };
    expect(selectLiveYaml(state)).toBe(state.formYaml);
  });

  it("yamlDirty beats formDirty (most-specific-first precedence)", () => {
    const state: LiveYamlState = {
      ...baseState,
      yamlDirty: true,
      formDirty: true,
    };
    expect(selectLiveYaml(state)).toBe(state.yamlText);
  });

  it("yamlDirty beats tab=yaml (both pick yamlText, but the rule order matters for clarity)", () => {
    const state: LiveYamlState = {
      ...baseState,
      yamlDirty: true,
      tab: "yaml",
    };
    expect(selectLiveYaml(state)).toBe(state.yamlText);
  });

  it("tab=yaml beats formDirty when the user has switched panes without confirming form edits", () => {
    const state: LiveYamlState = {
      ...baseState,
      tab: "yaml",
      formDirty: true,
    };
    expect(selectLiveYaml(state)).toBe(state.yamlText);
  });

  it("preserves comments byte-for-byte when nothing is dirty (the bug-Fix-C contract)", () => {
    const initialYaml =
      "id: review-only\n# Important: do not edit\nphases:\n  - kind: review_only\n";
    const state: LiveYamlState = {
      ...baseState,
      initialYaml,
      formYaml: "id: review-only\nphases:\n  - kind: review_only\n", // form-restringify drops the comment
    };
    const result = selectLiveYaml(state);
    expect(result).toBe(initialYaml);
    expect(result).toContain("# Important: do not edit");
  });
});
