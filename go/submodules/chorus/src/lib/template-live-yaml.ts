/**
 * Source-of-truth precedence for the YAML the TemplateDialog hands to the
 * server on save. Pure function so the precedence rules can be unit-tested
 * without spinning up React.
 *
 * Precedence (most-specific first):
 *   1. yamlDirty → user typed in the YAML pane → that wins regardless of tab
 *   2. tab === "yaml" → YAML pane is the surface, even if untouched (the
 *      user is reading raw YAML and we shouldn't restringify behind their
 *      back)
 *   3. formDirty → user changed a form field → emit from form
 *   4. otherwise → preserve the original YAML verbatim
 *
 * Step 4 is what prevents builtin templates from being silently promoted to
 * 'user' source via stringify-induced byte drift. Opening a builtin in Form
 * mode without touching anything must produce byte-identical YAML on save.
 */
export interface LiveYamlState {
  yamlDirty: boolean;
  tab: "form" | "yaml";
  formDirty: boolean;
  /** Current contents of the YAML editor pane. */
  yamlText: string;
  /** Result of stringifying the form state. */
  formYaml: string;
  /** Original YAML the dialog was opened with. Comments + whitespace intact. */
  initialYaml: string;
}

export function selectLiveYaml(state: LiveYamlState): string {
  if (state.yamlDirty) return state.yamlText;
  if (state.tab === "yaml") return state.yamlText;
  if (state.formDirty) return state.formYaml;
  return state.initialYaml;
}
