import { existsSync } from 'node:fs';
import fs from 'node:fs/promises';
import path from 'node:path';

export type InstallSurfaceArtifactStatus = {
    id: string;
    status: 'ready' | 'partial' | 'missing';
    artifactPath: string | null;
    artifactKind: string | null;
    detail: string;
    declaredVersion: string | null;
    lastModifiedAt: string | null;
};

function findMonorepoRoot(startDir: string = process.cwd()): string {
    let current = startDir;
    const root = path.parse(current).root;

    while (current !== root) {
        if (existsSync(path.join(current, 'pnpm-workspace.yaml')) || existsSync(path.join(current, 'turbo.json'))) {
            return current;
        }

        current = path.dirname(current);
    }

    return startDir;
}

function resolveExistingRelativePath(workspaceRoot: string, candidates: string[]): string | null {
    for (const candidate of candidates) {
        if (existsSync(path.join(workspaceRoot, candidate))) {
            return candidate;
        }
    }

    return null;
}

async function resolveLatestVsixRelativePath(workspaceRoot: string): Promise<string | null> {
    const vscodeDir = path.join(workspaceRoot, 'packages', 'vscode');
    if (!existsSync(vscodeDir)) {
        return null;
    }

    try {
        const entries = await fs.readdir(vscodeDir, { withFileTypes: true });
        const vsixFiles = entries
            .filter((entry) => entry.isFile() && entry.name.toLowerCase().endsWith('.vsix'))
            .map((entry) => entry.name)
            .sort((left, right) => left.localeCompare(right));

        const latest = vsixFiles.at(-1);
        return latest ? path.join('packages', 'vscode', latest) : null;
    } catch {
        return null;
    }
}

async function readPackageVersion(workspaceRoot: string, packageJsonRelativePath: string): Promise<string | null> {
    const packageJsonPath = path.join(workspaceRoot, packageJsonRelativePath);
    if (!existsSync(packageJsonPath)) {
        return null;
    }

    try {
        const raw = await fs.readFile(packageJsonPath, 'utf-8');
        const parsed = JSON.parse(raw) as { version?: unknown };
        return typeof parsed.version === 'string' ? parsed.version : null;
    } catch {
        return null;
    }
}

async function resolveLastModifiedAt(workspaceRoot: string, relativePath: string | null): Promise<string | null> {
    if (!relativePath) {
        return null;
    }

    try {
        const stats = await fs.stat(path.join(workspaceRoot, relativePath));
        return stats.mtime.toISOString();
    } catch {
        return null;
    }
}

export async function detectInstallSurfaceArtifacts(workspaceRoot: string = findMonorepoRoot()): Promise<InstallSurfaceArtifactStatus[]> {
    const chromiumBundlePath = resolveExistingRelativePath(workspaceRoot, [
        path.join('apps', 'tormentnexus-extension', 'dist-chromium'),
        path.join('apps', 'extension', 'dist'),
        path.join('apps', 'tormentnexus-extension', 'dist'),
    ]);

    const firefoxBundlePath = resolveExistingRelativePath(workspaceRoot, [
        path.join('apps', 'tormentnexus-extension', 'dist-firefox'),
    ]);
    const firefoxManifestPath = resolveExistingRelativePath(workspaceRoot, [
        path.join('apps', 'extension', 'manifest.firefox.json'),
    ]);

    const vscodeVsixPath = await resolveLatestVsixRelativePath(workspaceRoot);
    const vscodeBuildPath = resolveExistingRelativePath(workspaceRoot, [
        path.join('packages', 'vscode', 'dist', 'extension.js'),
        path.join('packages', 'vscode', 'dist'),
    ]);

    const mcpConfigPath = resolveExistingRelativePath(workspaceRoot, ['mcp.jsonc', 'mcp.json']);

    const [browserExtensionVersion, vscodeExtensionVersion] = await Promise.all([
        readPackageVersion(workspaceRoot, path.join('apps', 'tormentnexus-extension', 'package.json')),
        readPackageVersion(workspaceRoot, path.join('packages', 'vscode', 'package.json')),
    ]);

    const [
        chromiumLastModifiedAt,
        firefoxLastModifiedAt,
        vscodeLastModifiedAt,
        mcpConfigLastModifiedAt,
    ] = await Promise.all([
        resolveLastModifiedAt(workspaceRoot, chromiumBundlePath),
        resolveLastModifiedAt(workspaceRoot, firefoxBundlePath ?? firefoxManifestPath),
        resolveLastModifiedAt(workspaceRoot, vscodeVsixPath ?? vscodeBuildPath),
        resolveLastModifiedAt(workspaceRoot, mcpConfigPath),
    ]);

    return [
        {
            id: 'browser-extension-chromium',
            status: chromiumBundlePath ? 'ready' : 'missing',
            artifactPath: chromiumBundlePath,
            artifactKind: chromiumBundlePath === path.join('apps', 'tormentnexus-extension', 'dist-chromium')
                ? 'Chromium unpacked bundle'
                : chromiumBundlePath === path.join('apps', 'extension', 'dist')
                    ? 'Legacy extension dist bundle'
                    : chromiumBundlePath === path.join('apps', 'tormentnexus-extension', 'dist')
                        ? 'Generic tormentnexus-extension dist bundle'
                        : null,
            detail: chromiumBundlePath
                ? 'Unpacked Chromium-compatible browser extension output is available.'
                : 'Build the browser extension to generate a Chromium/Edge unpacked bundle.',
            declaredVersion: browserExtensionVersion,
            lastModifiedAt: chromiumLastModifiedAt,
        },
        {
            id: 'browser-extension-firefox',
            status: firefoxBundlePath ? 'ready' : firefoxManifestPath ? 'partial' : 'missing',
            artifactPath: firefoxBundlePath ?? firefoxManifestPath,
            artifactKind: firefoxBundlePath
                ? 'Firefox unpacked bundle'
                : firefoxManifestPath
                    ? 'Firefox manifest source'
                    : null,
            detail: firefoxBundlePath
                ? 'Firefox-specific browser extension output is available.'
                : firefoxManifestPath
                    ? 'Firefox manifest source is present, but no packaged Firefox bundle was detected yet.'
                    : 'No Firefox-ready browser extension artifact was detected yet.',
            declaredVersion: browserExtensionVersion,
            lastModifiedAt: firefoxLastModifiedAt,
        },
        {
            id: 'vscode-extension',
            status: vscodeVsixPath ? 'ready' : vscodeBuildPath ? 'partial' : 'missing',
            artifactPath: vscodeVsixPath ?? vscodeBuildPath,
            artifactKind: vscodeVsixPath
                ? 'VSIX package'
                : vscodeBuildPath
                    ? 'Compiled extension output'
                    : null,
            detail: vscodeVsixPath
                ? 'Packaged VS Code extension artifact is ready to install.'
                : vscodeBuildPath
                    ? 'VS Code extension is compiled, but no `.vsix` package was detected yet.'
                    : 'Build and package the VS Code extension to generate an installable `.vsix`.',
            declaredVersion: vscodeExtensionVersion,
            lastModifiedAt: vscodeLastModifiedAt,
        },
        {
            id: 'mcp-client-sync',
            status: mcpConfigPath ? 'ready' : 'missing',
            artifactPath: mcpConfigPath,
            artifactKind: mcpConfigPath === 'mcp.jsonc'
                ? 'JSONC config source'
                : mcpConfigPath === 'mcp.json'
                    ? 'JSON config source'
                    : null,
            detail: mcpConfigPath
                ? 'TormentNexus-managed MCP config source is present for dashboard sync and preview flows.'
                : 'No TormentNexus MCP config source file was detected yet.',
            declaredVersion: null,
            lastModifiedAt: mcpConfigLastModifiedAt,
        },
    ];
}