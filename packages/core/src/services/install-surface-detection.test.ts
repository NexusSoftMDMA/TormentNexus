import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';

import { afterEach, describe, expect, it } from 'vitest';

import { detectInstallSurfaceArtifacts } from './install-surface-detection.js';

const tempDirs: string[] = [];

async function createWorkspace(): Promise<string> {
    const workspaceRoot = await fs.mkdtemp(path.join(os.tmpdir(), 'tormentnexus-install-surfaces-'));
    tempDirs.push(workspaceRoot);
    await fs.writeFile(path.join(workspaceRoot, 'pnpm-workspace.yaml'), 'packages:\n  - packages/*\n', 'utf-8');
    return workspaceRoot;
}

afterEach(async () => {
    await Promise.all(tempDirs.splice(0).map((dir) => fs.rm(dir, { recursive: true, force: true })));
});

describe('detectInstallSurfaceArtifacts', () => {
    it('reports ready artifacts when installable outputs already exist', async () => {
        const workspaceRoot = await createWorkspace();

        await fs.mkdir(path.join(workspaceRoot, 'apps', 'tormentnexus-extension'), { recursive: true });
        await fs.mkdir(path.join(workspaceRoot, 'apps', 'tormentnexus-extension', 'dist-chromium'), { recursive: true });
        await fs.mkdir(path.join(workspaceRoot, 'apps', 'tormentnexus-extension', 'dist-firefox'), { recursive: true });
        await fs.writeFile(path.join(workspaceRoot, 'apps', 'tormentnexus-extension', 'package.json'), JSON.stringify({ version: '0.7.3' }), 'utf-8');
        await fs.mkdir(path.join(workspaceRoot, 'packages', 'vscode', 'dist'), { recursive: true });
        await fs.writeFile(path.join(workspaceRoot, 'packages', 'vscode', 'package.json'), JSON.stringify({ version: '0.2.0' }), 'utf-8');
        await fs.writeFile(path.join(workspaceRoot, 'packages', 'vscode', 'tormentnexus-vscode-extension-0.2.0.vsix'), '', 'utf-8');
        await fs.writeFile(path.join(workspaceRoot, 'mcp.jsonc'), '{"mcpServers":{}}', 'utf-8');

        const statuses = await detectInstallSurfaceArtifacts(workspaceRoot);

        expect(statuses).toEqual([
            expect.objectContaining({
                id: 'browser-extension-chromium',
                status: 'ready',
                artifactPath: path.join('apps', 'tormentnexus-extension', 'dist-chromium'),
                artifactKind: 'Chromium unpacked bundle',
                declaredVersion: '0.7.3',
                lastModifiedAt: expect.any(String),
            }),
            expect.objectContaining({
                id: 'browser-extension-firefox',
                status: 'ready',
                artifactPath: path.join('apps', 'tormentnexus-extension', 'dist-firefox'),
                artifactKind: 'Firefox unpacked bundle',
                declaredVersion: '0.7.3',
                lastModifiedAt: expect.any(String),
            }),
            expect.objectContaining({
                id: 'vscode-extension',
                status: 'ready',
                artifactPath: path.join('packages', 'vscode', 'tormentnexus-vscode-extension-0.2.0.vsix'),
                artifactKind: 'VSIX package',
                declaredVersion: '0.2.0',
                lastModifiedAt: expect.any(String),
            }),
            expect.objectContaining({
                id: 'mcp-client-sync',
                status: 'ready',
                artifactPath: 'mcp.jsonc',
                artifactKind: 'JSONC config source',
                declaredVersion: null,
                lastModifiedAt: expect.any(String),
            }),
        ]);
    });

    it('reports partial readiness when only source-level install assets exist', async () => {
        const workspaceRoot = await createWorkspace();

        await fs.mkdir(path.join(workspaceRoot, 'apps', 'extension'), { recursive: true });
        await fs.mkdir(path.join(workspaceRoot, 'apps', 'tormentnexus-extension'), { recursive: true });
        await fs.writeFile(path.join(workspaceRoot, 'apps', 'tormentnexus-extension', 'package.json'), JSON.stringify({ version: '0.7.3' }), 'utf-8');
        await fs.writeFile(path.join(workspaceRoot, 'apps', 'extension', 'manifest.firefox.json'), '{}', 'utf-8');
        await fs.mkdir(path.join(workspaceRoot, 'packages', 'vscode', 'dist'), { recursive: true });
        await fs.writeFile(path.join(workspaceRoot, 'packages', 'vscode', 'package.json'), JSON.stringify({ version: '0.2.0' }), 'utf-8');
        await fs.writeFile(path.join(workspaceRoot, 'packages', 'vscode', 'dist', 'extension.js'), '', 'utf-8');

        const statuses = await detectInstallSurfaceArtifacts(workspaceRoot);
        const firefox = statuses.find((status) => status.id === 'browser-extension-firefox');
        const vscode = statuses.find((status) => status.id === 'vscode-extension');

        expect(firefox).toMatchObject({
            status: 'partial',
            artifactPath: path.join('apps', 'extension', 'manifest.firefox.json'),
            artifactKind: 'Firefox manifest source',
            declaredVersion: '0.7.3',
            lastModifiedAt: expect.any(String),
        });
        expect(vscode).toMatchObject({
            status: 'partial',
            artifactPath: path.join('packages', 'vscode', 'dist', 'extension.js'),
            artifactKind: 'Compiled extension output',
            declaredVersion: '0.2.0',
            lastModifiedAt: expect.any(String),
        });
    });

    it('reports missing artifacts when no install outputs are present', async () => {
        const workspaceRoot = await createWorkspace();

        const statuses = await detectInstallSurfaceArtifacts(workspaceRoot);

        expect(statuses).toEqual([
            expect.objectContaining({ id: 'browser-extension-chromium', status: 'missing', artifactPath: null, artifactKind: null, declaredVersion: null, lastModifiedAt: null }),
            expect.objectContaining({ id: 'browser-extension-firefox', status: 'missing', artifactPath: null, artifactKind: null, declaredVersion: null, lastModifiedAt: null }),
            expect.objectContaining({ id: 'vscode-extension', status: 'missing', artifactPath: null, artifactKind: null, declaredVersion: null, lastModifiedAt: null }),
            expect.objectContaining({ id: 'mcp-client-sync', status: 'missing', artifactPath: null, artifactKind: null, declaredVersion: null, lastModifiedAt: null }),
        ]);
    });
});