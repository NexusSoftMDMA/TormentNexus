import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { afterEach, describe, expect, it } from 'vitest';

import { readLocalStartupProvenance } from './startup-provenance.js';

const tempDirs: string[] = [];

afterEach(() => {
    while (tempDirs.length > 0) {
        const dir = tempDirs.pop();
        if (dir) {
            rmSync(dir, { recursive: true, force: true });
        }
    }
});

function createTempDir(): string {
    const dir = mkdtempSync(path.join(os.tmpdir(), 'tormentnexus-startup-provenance-'));
    tempDirs.push(dir);
    return dir;
}

describe('readLocalStartupProvenance', () => {
    it('derives control-plane port provenance from legacy lock files', () => {
        const dataDir = createTempDir();
        writeFileSync(path.join(dataDir, 'lock'), JSON.stringify({
            instanceId: 'tormentnexus-legacy',
            pid: 1234,
            port: 4555,
            host: '127.0.0.1',
            createdAt: '2026-04-06T09:00:00.000Z',
        }), 'utf8');

        expect(readLocalStartupProvenance(dataDir)).toEqual({
            requestedPort: 4555,
            activePort: 4555,
            portDecision: 'derived from lock record',
            portReason: 'Detailed startup port provenance was unavailable; using the current control-plane lock port.',
            updatedAt: '2026-04-06T09:00:00.000Z',
        });
    });
});
