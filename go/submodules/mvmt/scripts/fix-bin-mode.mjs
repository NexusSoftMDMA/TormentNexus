#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

const binPath = path.join(process.cwd(), 'dist', 'bin', 'mvmt.js');

if (!fs.existsSync(binPath)) {
  throw new Error(`Built CLI not found at ${binPath}`);
}

if (process.platform !== 'win32') {
  fs.chmodSync(binPath, 0o755);
}
