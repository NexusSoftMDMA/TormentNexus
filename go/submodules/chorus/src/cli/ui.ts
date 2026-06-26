/**
 * CLI output helpers — colors, symbols, formatters.
 * Zero deps — raw ANSI escapes. Respects NO_COLOR env var.
 */

const useColor = !process.env.NO_COLOR && process.stdout.isTTY !== false;

function ansi(code: string): (s: string) => string {
  return (s: string) => (useColor ? `\x1b[${code}m${s}\x1b[0m` : s);
}

export const c = {
  bold: ansi('1'),
  dim: ansi('2'),
  red: ansi('31'),
  green: ansi('32'),
  yellow: ansi('33'),
  blue: ansi('34'),
  magenta: ansi('35'),
  cyan: ansi('36'),
  gray: ansi('90'),
};

export const sym = {
  ok: c.green('✓'),
  err: c.red('✗'),
  bullet: c.cyan('●'),
  arrow: c.gray('→'),
  pointer: c.cyan('▶'),
  rocket: '🚀',
  info: c.blue('ⓘ'),
};

/**
 * Header line with bold title, optional dim subtitle.
 *   header(sym.ok, 'Chorus started')  →  ✓ Chorus started
 *   header(sym.ok, 'Started', 'PID 1234')  →  ✓ Started  (PID 1234)
 */
export function header(icon: string, title: string, sub?: string): string {
  const main = `${icon} ${c.bold(title)}`;
  return sub ? `${main}  ${c.dim(sub)}` : main;
}

/**
 * Aligned key/value rows. Pads keys to the longest key width.
 *   kv([['Cockpit', 'http://...'], ['Daemon', 'http://...']])
 */
export function kv(rows: Array<[string, string]>, indent = '   '): string {
  const width = Math.max(...rows.map(([k]) => k.length));
  return rows
    .map(([k, v]) => `${indent}${c.gray(k.padEnd(width))}  ${v}`)
    .join('\n');
}

/**
 * A boxed tip line. Used for one-time hints/footnotes.
 */
export function tip(text: string, indent = '   '): string {
  return `${indent}${sym.info} ${c.dim(text)}`;
}
