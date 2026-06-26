/**
 * Shared template-YAML validation. Used by:
 *   - the cockpit dialog (live as the user types) to show inline errors
 *     and gate the Save button
 *   - the daemon's POST /templates handler to reject bad shapes server-side
 *
 * Two-stage validation matches the cockpit's UX:
 *   1. yaml.parse — surfaces syntax errors with line/col
 *   2. TemplateSchema (zod) — surfaces structural errors with field paths
 *
 * The helper is intentionally pure: no I/O, no logging. Safe to call from
 * a React render path or a Fastify handler.
 */
import yaml from 'yaml';
import { TemplateSchema } from './template-schema';

export interface TemplateValidationIssue {
  /**
   * Dotted field path the issue applies to (e.g. `phases.0.reviewer.require`).
   * `<root>` for top-level issues that don't have a path.
   * `<yaml>` for parser errors that fire before zod gets a chance.
   */
  path: string;
  message: string;
  /** Line number from the YAML parser. Only set for stage-1 issues. */
  line?: number;
  /** Column number from the YAML parser. Only set for stage-1 issues. */
  col?: number;
}

export interface TemplateValidationResult {
  valid: boolean;
  issues: TemplateValidationIssue[];
}

const VALID_RESULT: TemplateValidationResult = Object.freeze({
  valid: true,
  issues: [],
});

/**
 * Parse + validate a template YAML string. Returns a flat issue list
 * suitable for inline rendering. Empty input is treated as invalid with
 * a clear message rather than silently passing.
 */
export function validateTemplateYaml(input: string): TemplateValidationResult {
  const trimmed = input.trim();
  if (trimmed.length === 0) {
    return {
      valid: false,
      issues: [{ path: '<yaml>', message: 'Template YAML is empty.' }],
    };
  }

  // Stage 1: YAML syntax.
  let parsed: unknown;
  try {
    parsed = yaml.parse(input);
  } catch (err) {
    // yaml package errors carry .linePos { line, col } when available; fall
    // back to the raw message so the user always gets *something*.
    const msg = err instanceof Error ? err.message : String(err);
    const linePos = (err as { linePos?: Array<{ line: number; col: number }> })
      .linePos?.[0];
    return {
      valid: false,
      issues: [
        {
          path: '<yaml>',
          message: msg,
          ...(linePos ? { line: linePos.line, col: linePos.col } : {}),
        },
      ],
    };
  }

  if (parsed === null || parsed === undefined) {
    return {
      valid: false,
      issues: [
        {
          path: '<yaml>',
          message: 'Template YAML parsed to null — at minimum it must define id, name, and phases.',
        },
      ],
    };
  }

  // Stage 2: structural validation against the daemon's TemplateSchema.
  const result = TemplateSchema.safeParse(parsed);
  if (result.success) return VALID_RESULT;

  const issues: TemplateValidationIssue[] = result.error.issues.map((i) => ({
    path: i.path.length > 0 ? i.path.join('.') : '<root>',
    message: i.message,
  }));
  return { valid: false, issues };
}

