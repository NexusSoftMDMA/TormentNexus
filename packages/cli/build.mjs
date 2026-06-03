#!/usr/bin/env node
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
  console.error('[@tormentnexus/cli] Fatal build errors:\n' + fatalLines.join('\n'));
}

process.exit(0);
