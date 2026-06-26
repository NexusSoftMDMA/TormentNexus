/**
 * Extract a display-worthy title from a chat's `work` field.
 *
 * Why this exists: `mcp__chorus__invoke_persona` composes the chat's `work`
 * as `# Persona: <Label>\n\n<system_prompt>\n\n---\n\n# User request\n\n<brief>`.
 * Without unwrapping, every Cartographer / Sentinel / Profiler chat looks
 * identical in the sidebar — the first 200 chars are always the persona's
 * worldview prompt, never the user's actual request. See ROADMAP #13.
 *
 * The helper matches the exact structure invoke_persona writes today, returns
 * `[Persona] <user request>` when it can extract both, and falls back to the
 * raw work string otherwise so non-persona chats stay unaffected.
 */

// Bound length so a 1MB unwrapped work string can't trigger pathological
// backtracking on the inner `[\s\S]+?` quantifier; the persona separator
// always lives in the first ~5KB of any well-formed invoke_persona payload.
const MAX_PARSE_LEN = 8 * 1024;
const PERSONA_PREFIX_RE = /^# Persona: (.+?)\n[\s\S]+?\n---\n\n# User request\n\n([\s\S]+)$/;

export function chatDisplayTitle(work: string): string {
  if (!work) return "";
  // Cheap structural pre-check — if the marker isn't in the first MAX_PARSE_LEN
  // bytes, the regex won't find it either, and skipping the regex avoids the
  // worst-case O(n²) backtracking on adversarial input.
  const slice = work.length > MAX_PARSE_LEN ? work.slice(0, MAX_PARSE_LEN) : work;
  if (!slice.startsWith("# Persona:") || !slice.includes("\n---\n\n# User request\n\n")) {
    return work;
  }
  const m = PERSONA_PREFIX_RE.exec(work);
  if (m) {
    const persona = m[1].trim();
    const request = m[2].trim();
    return `[${persona}] ${request}`;
  }
  return work;
}
