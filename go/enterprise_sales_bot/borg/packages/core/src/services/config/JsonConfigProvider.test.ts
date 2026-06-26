import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';

import { afterEach, describe, expect, it } from 'vitest';

import { JsonConfigProvider } from './JsonConfigProvider.js';
import { loadTormentNexusMcpConfig } from '../../mcp/mcpJsonConfig.js';

const tempDirs: string[] = [];

async function createTempWorkspace(): Promise<string> {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'tormentnexus-json-config-provider-'));
    tempDirs.push(dir);
    return dir;
}

afterEach(async () => {
    await Promise.all(tempDirs.splice(0).map(async (dir) => {
        await fs.rm(dir, { recursive: true, force: true });
    }));
});

describe('JsonConfigProvider always-visible tools compatibility', () => {
    it('prefers toolSelection.alwaysLoadedTools over legacy alwaysVisibleTools', async () => {
        const workspace = await createTempWorkspace();
        await fs.writeFile(path.join(workspace, 'mcp.jsonc'), JSON.stringify({
            mcpServers: {},
            alwaysVisibleTools: ['legacy__tool'],
            settings: {
                toolSelection: {
                    alwaysLoadedTools: ['modern__tool'],
                },
            },
        }, null, 2));

        const provider = new JsonConfigProvider(workspace);

        await expect(provider.loadAlwaysVisibleTools()).resolves.toEqual(['modern__tool']);
    });

    it('saves always-visible tools into both legacy and toolSelection settings shapes', async () => {
        const workspace = await createTempWorkspace();
        const provider = new JsonConfigProvider(workspace);

        await expect(provider.saveAlwaysVisibleTools([' browser__open ', 'browser__open', ''])).resolves.toEqual(['browser__open']);

        const config = await loadTormentNexusMcpConfig(workspace);
        expect(config.alwaysVisibleTools).toEqual(['browser__open']);
        expect(config.settings?.toolSelection).toMatchObject({
            importantTools: [],
            alwaysLoadedTools: ['browser__open'],
            autoLoadMinConfidence: 0.85,
            // Session working-set capacity fields
            maxLoadedTools: expect.any(Number),
            maxHydratedSchemas: expect.any(Number),
        });
    });
});