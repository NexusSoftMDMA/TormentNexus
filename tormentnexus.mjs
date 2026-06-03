#!/usr/bin/env node
/**
 * TormentNexus TORMENTNEXUS - Top-level CLI wrapper
 * Runs the compiled CLI from packages/cli/dist
 */
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { createRequire } from 'module';

const __dirname = dirname(fileURLToPath(import.meta.url));
const cliEntry = resolve(__dirname, 'packages/cli/dist/cli/src/index.js');

const require = createRequire(import.meta.url);
require('module').runMain();
import(cliEntry);
