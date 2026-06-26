/**
 * Kimi CLI (`kimi --print --output-format stream-json`).
 *
 * Kimi is intentionally Claude-Code-compatible. Its stream-json is
 * documented to follow the Claude shape, so parseClaude is the reference.
 */
import type { AgentEvent } from '../types.js';
import { parseClaude } from './claude.js';

export function parseKimi(line: string): AgentEvent[] {
  return parseClaude(line);
}
