/**
 * Shell-quoting and validation helpers.
 * Used by all agent shims to safely quote values for tmux launch commands.
 */

const SHELL_METACHARACTERS = /[$\`;|&<>()\\"']/;

class InvalidValueError extends Error {
  constructor(field: string, value: string) {
    super(
      `Invalid ${field}: contains shell metacharacters. ` +
      `Got: ${value}`
    );
    this.name = 'InvalidValueError';
  }
}

/**
 * Quote a string for safe use in shell commands (bash %q equivalent).
 * Uses single-quote wrapping with escaped embedded quotes.
 * Safe to pass to `tmux new-session -d -s <name> "..."`.
 */
export function quoteValue(s: string): string {
  // Wrap in single quotes, escape any embedded single quotes
  return `'${s.replace(/'/g, `'\\''`)}'`;
}

/**
 * Quote a filesystem path for safe use in shell commands.
 * Identical to quoteValue; kept separate for semantic clarity.
 */
export const quotePath = quoteValue;

/**
 * Validate a value against shell metacharacters.
 * Throws InvalidValueError if validation fails.
 * Use BEFORE building launch commands to catch bad input early.
 */
export function validateValue(field: string, value: string | undefined): void {
  if (value === undefined || value === '') {
    return; // empty/undefined is OK
  }
  if (SHELL_METACHARACTERS.test(value)) {
    throw new InvalidValueError(field, value);
  }
}
