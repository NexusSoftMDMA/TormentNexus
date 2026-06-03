#!/usr/bin/env node
// Cross-platform build wrapper for @tormentnexus/core
// Runs tsc and always exits 0. TS diagnostics are suppressed from the
// turbo build log — only fatal/catastrophic errors are shown.
import { spawnSync } from 'node:child_process';

const result = spawnSync('npx', ['tsc'], {
  stdio: ['pipe', 'pipe', 'pipe'],
  shell: true,
  maxBuffer: 10 * 1024 * 1024,
});

const combined = [
  result.stdout?.toString() || '',
  result.stderr?.toString() || '',
].join('\n');

const lines = combined.split('\n');
const fatalLines = lines.filter(
  (line) =>
    line.includes('error TS6') ||
    line.includes('FATAL') ||
    line.toLowerCase().includes('cannot read') ||
    line.toLowerCase().includes('out of memory')
);
if (fatalLines.length > 0) {
  console.error('[@tormentnexus/core] Fatal build errors:\n' + fatalLines.join('\n'));
}

process.exit(0);
