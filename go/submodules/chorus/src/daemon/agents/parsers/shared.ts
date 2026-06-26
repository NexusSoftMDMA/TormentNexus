/** JSON.parse → undefined on malformed lines (CLIs sometimes emit blank
 *  lines, log lines, or partial frames in error paths). */
export function tryJson(line: string): unknown {
  const trimmed = line.trim();
  if (trimmed.length === 0) return undefined;
  try {
    return JSON.parse(trimmed);
  } catch {
    return undefined;
  }
}
