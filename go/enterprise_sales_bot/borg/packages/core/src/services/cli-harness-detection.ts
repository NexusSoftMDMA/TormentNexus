import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

import { CLI_HARNESS_CATALOG, type CliHarnessCatalogEntry } from '../supervisor/cliHarnessCatalog.js';

const execFileAsync = promisify(execFile);

export interface CliHarnessDetectionResult extends CliHarnessCatalogEntry {
    installed: boolean;
    resolvedPath: string | null;
    version: string | null;
    detectionError: string | null;
}

async function runCommand(command: string, args: string[]): Promise<string | null> {
    try {
        const { stdout, stderr } = await execFileAsync(command, args, {
            timeout: 3_000,
            windowsHide: true,
            maxBuffer: 512 * 1024,
        });

        const output = `${stdout ?? ''}\n${stderr ?? ''}`.trim();
        return output || null;
    } catch {
        return null;
    }
}

async function resolveCommandPath(command: string): Promise<string | null> {
    const locator = process.platform === 'win32' ? 'where' : 'which';
    const output = await runCommand(locator, [command]);

    if (!output) {
        return null;
    }

    return output
        .split(/\r?\n/)
        .map((line) => line.trim())
        .find(Boolean) ?? null;
}

function formatVersion(output: string | null): string | null {
    if (!output) {
        return null;
    }

    return output
        .split(/\r?\n/)
        .map((line) => line.trim())
        .find(Boolean)
        ?.slice(0, 120) ?? null;
}

export async function detectCliHarnesses(
    catalog: CliHarnessCatalogEntry[] = CLI_HARNESS_CATALOG,
): Promise<CliHarnessDetectionResult[]> {
    return Promise.all(catalog.map(async (entry) => {
        if (entry.detectionMode === 'manual') {
            return {
                ...entry,
                installed: false,
                resolvedPath: null,
                version: null,
                detectionError: 'Manual install surface; no PATH-detectable command is currently modeled for this harness.',
            } satisfies CliHarnessDetectionResult;
        }

        const resolvedPath = await resolveCommandPath(entry.command);

        if (!resolvedPath) {
            return {
                ...entry,
                installed: false,
                resolvedPath: null,
                version: null,
                detectionError: 'Command not found on PATH.',
            } satisfies CliHarnessDetectionResult;
        }

        const version = formatVersion(await runCommand(entry.command, entry.versionArgs));

        return {
            ...entry,
            installed: true,
            resolvedPath,
            version,
            detectionError: null,
        } satisfies CliHarnessDetectionResult;
    }));
}
