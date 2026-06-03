import { access } from 'node:fs/promises';
import { constants as fsConstants } from 'node:fs';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

import { detectCliHarnesses, type CliHarnessDetectionResult } from './cli-harness-detection.js';

const execFileAsync = promisify(execFile);

export type ShellFamily = 'powershell' | 'cmd' | 'posix' | 'wsl';

interface ShellCatalogEntry {
    id: string;
    name: string;
    command: string;
    family: ShellFamily;
    versionArgs: string[];
    smokeTestArgs: string[];
    executableHints?: string[];
    pathPattern?: RegExp;
    notes?: string[];
}

interface BinaryCatalogEntry {
    id: string;
    name: string;
    command: string;
    versionArgs: string[];
    capabilities: string[];
    notes?: string[];
}

export interface DetectedExecutionShell {
    id: string;
    name: string;
    family: ShellFamily;
    installed: boolean;
    verified: boolean;
    resolvedPath: string | null;
    version: string | null;
    preferred: boolean;
    notes: string[];
}

export interface DetectedExecutionTool {
    id: string;
    name: string;
    installed: boolean;
    verified: boolean;
    resolvedPath: string | null;
    version: string | null;
    capabilities: string[];
    notes: string[];
}

export interface ExecutionEnvironmentSummary {
    ready: boolean;
    preferredShellId: string | null;
    preferredShellLabel: string | null;
    shellCount: number;
    verifiedShellCount: number;
    toolCount: number;
    verifiedToolCount: number;
    harnessCount: number;
    verifiedHarnessCount: number;
    supportsPowerShell: boolean;
    supportsPosixShell: boolean;
    notes: string[];
}

export interface LocalExecutionEnvironment {
    os: NodeJS.Platform;
    summary: ExecutionEnvironmentSummary;
    shells: DetectedExecutionShell[];
    tools: DetectedExecutionTool[];
    harnesses: CliHarnessDetectionResult[];
}

export interface ExecutionEnvironmentDeps {
    runCommand: (command: string, args: string[]) => Promise<string | null>;
    resolveCommandPath: (command: string) => Promise<string | null>;
    pathExists: (filePath: string) => Promise<boolean>;
    detectHarnesses: () => Promise<CliHarnessDetectionResult[]>;
    platform: NodeJS.Platform;
}

const SHELL_CATALOG: ShellCatalogEntry[] = [
    {
        id: 'pwsh',
        name: 'PowerShell 7',
        command: 'pwsh',
        family: 'powershell',
        versionArgs: ['-NoLogo', '-NoProfile', '-Command', '$PSVersionTable.PSVersion.ToString()'],
        smokeTestArgs: ['-NoLogo', '-NoProfile', '-Command', 'Write-Output tormentnexus-ready'],
        notes: ['Preferred TormentNexus shell on Windows for general command execution and structured scripting.'],
    },
    {
        id: 'powershell',
        name: 'Windows PowerShell',
        command: 'powershell',
        family: 'powershell',
        versionArgs: ['-NoLogo', '-NoProfile', '-Command', '$PSVersionTable.PSVersion.ToString()'],
        smokeTestArgs: ['-NoLogo', '-NoProfile', '-Command', 'Write-Output tormentnexus-ready'],
        notes: ['Useful legacy fallback when PowerShell 7 is unavailable.'],
    },
    {
        id: 'cmd',
        name: 'Command Prompt',
        command: 'cmd',
        family: 'cmd',
        versionArgs: ['/d', '/c', 'ver'],
        smokeTestArgs: ['/d', '/c', 'echo tormentnexus-ready'],
        notes: ['Lowest-common-denominator Windows shell for compatibility-only flows.'],
    },
    {
        id: 'cygwin-bash',
        name: 'Cygwin Bash',
        command: 'bash',
        family: 'posix',
        versionArgs: ['-lc', 'printf "%s" "$(uname -s) $(uname -r)"'],
        smokeTestArgs: ['-lc', 'printf tormentnexus-ready'],
        executableHints: [
            'C:\\cygwin64\\bin\\bash.exe',
            'C:\\cygwin\\bin\\bash.exe',
        ],
        pathPattern: /cygwin/i,
        notes: ['POSIX shell with Cygwin path semantics; useful for Unix-style pipelines on Windows.'],
    },
    {
        id: 'git-bash',
        name: 'Git Bash',
        command: 'bash',
        family: 'posix',
        versionArgs: ['-lc', 'printf "%s" "$(uname -s) $(uname -r)"'],
        smokeTestArgs: ['-lc', 'printf tormentnexus-ready'],
        executableHints: [
            'C:\\Program Files\\Git\\bin\\bash.exe',
            'C:\\Program Files\\Git\\usr\\bin\\bash.exe',
            'C:\\Program Files (x86)\\Git\\bin\\bash.exe',
        ],
        pathPattern: /[\\/]git[\\/].*bash\.exe$/i,
        notes: ['POSIX-friendly shell for lightweight Unix tooling without a full Cygwin install.'],
    },
    {
        id: 'wsl',
        name: 'Windows Subsystem for Linux',
        command: 'wsl',
        family: 'wsl',
        versionArgs: ['--version'],
        smokeTestArgs: ['-e', 'sh', '-lc', 'printf tormentnexus-ready'],
        notes: ['Best fit for Linux-native commands when WSL is installed and configured.'],
    },
];

const TOOL_CATALOG: BinaryCatalogEntry[] = [
    {
        id: 'git',
        name: 'Git',
        command: 'git',
        versionArgs: ['--version'],
        capabilities: ['source control', 'worktrees', 'diffs'],
        notes: ['Required for session isolation, repo inspection, and worktree-based supervision.'],
    },
    {
        id: 'rg',
        name: 'ripgrep',
        command: 'rg',
        versionArgs: ['--version'],
        capabilities: ['fast text search', 'code discovery'],
        notes: ['Primary fast search binary for code and text inventories.'],
    },
    {
        id: 'jq',
        name: 'jq',
        command: 'jq',
        versionArgs: ['--version'],
        capabilities: ['json shaping', 'pipeline transforms'],
        notes: ['Handy for JSON-heavy CLI workflows and operator debugging.'],
    },
    {
        id: 'python',
        name: 'Python',
        command: 'python',
        versionArgs: ['--version'],
        capabilities: ['scripting', 'ecosystem tools'],
        notes: ['Useful for external tool ecosystems and glue scripts.'],
    },
    {
        id: 'node',
        name: 'Node.js',
        command: 'node',
        versionArgs: ['--version'],
        capabilities: ['javascript runtime', 'tool execution'],
        notes: ['Required for large parts of the TormentNexus stack and many MCP servers.'],
    },
    {
        id: 'pnpm',
        name: 'pnpm',
        command: 'pnpm',
        versionArgs: ['--version'],
        capabilities: ['workspace package management', 'build orchestration'],
        notes: ['Primary package manager for this TormentNexus workspace.'],
    },
    {
        id: 'docker',
        name: 'Docker',
        command: 'docker',
        versionArgs: ['--version'],
        capabilities: ['containers', 'compose stacks'],
        notes: ['Useful for local services, MCP servers, and reproducible dev environments.'],
    },
];

async function defaultRunCommand(command: string, args: string[]): Promise<string | null> {
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

async function defaultResolveCommandPath(command: string): Promise<string | null> {
    const locator = process.platform === 'win32' ? 'where' : 'which';
    const output = await defaultRunCommand(locator, [command]);

    if (!output) {
        return null;
    }

    return output
        .split(/\r?\n/)
        .map((line) => line.trim())
        .find(Boolean) ?? null;
}

async function defaultPathExists(filePath: string): Promise<boolean> {
    try {
        await access(filePath, fsConstants.F_OK);
        return true;
    } catch {
        return false;
    }
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

async function resolveCatalogPath(
    entry: Pick<ShellCatalogEntry, 'command' | 'executableHints' | 'pathPattern'>,
    deps: ExecutionEnvironmentDeps,
): Promise<string | null> {
    for (const hint of entry.executableHints ?? []) {
        if (await deps.pathExists(hint)) {
            return hint;
        }
    }

    const resolvedPath = await deps.resolveCommandPath(entry.command);
    if (!resolvedPath) {
        return null;
    }

    if (entry.pathPattern && !entry.pathPattern.test(resolvedPath)) {
        return null;
    }

    return resolvedPath;
}

async function detectShell(entry: ShellCatalogEntry, deps: ExecutionEnvironmentDeps): Promise<DetectedExecutionShell> {
    const resolvedPath = await resolveCatalogPath(entry, deps);

    if (!resolvedPath) {
        return {
            id: entry.id,
            name: entry.name,
            family: entry.family,
            installed: false,
            verified: false,
            resolvedPath: null,
            version: null,
            preferred: false,
            notes: [...(entry.notes ?? []), 'Shell not detected on this machine.'],
        };
    }

    const version = formatVersion(await deps.runCommand(resolvedPath, entry.versionArgs));
    const smokeOutput = await deps.runCommand(resolvedPath, entry.smokeTestArgs);
    const verified = Boolean(smokeOutput?.toLowerCase().includes('tormentnexus-ready'));

    return {
        id: entry.id,
        name: entry.name,
        family: entry.family,
        installed: true,
        verified,
        resolvedPath,
        version,
        preferred: false,
        notes: verified
            ? [...(entry.notes ?? [])]
            : [...(entry.notes ?? []), 'Shell was found, but TormentNexus could not verify a simple smoke test.'],
    };
}

async function detectBinary(entry: BinaryCatalogEntry, deps: ExecutionEnvironmentDeps): Promise<DetectedExecutionTool> {
    const resolvedPath = await deps.resolveCommandPath(entry.command);

    if (!resolvedPath) {
        return {
            id: entry.id,
            name: entry.name,
            installed: false,
            verified: false,
            resolvedPath: null,
            version: null,
            capabilities: [...entry.capabilities],
            notes: [...(entry.notes ?? []), 'Binary not detected on PATH.'],
        };
    }

    const versionOutput = await deps.runCommand(resolvedPath, entry.versionArgs);
    return {
        id: entry.id,
        name: entry.name,
        installed: true,
        verified: Boolean(versionOutput),
        resolvedPath,
        version: formatVersion(versionOutput),
        capabilities: [...entry.capabilities],
        notes: versionOutput
            ? [...(entry.notes ?? [])]
            : [...(entry.notes ?? []), 'Binary resolved on PATH, but version verification failed.'],
    };
}

function selectPreferredShell(shells: DetectedExecutionShell[], platform: NodeJS.Platform): string | null {
    const verifiedShells = shells.filter((shell) => shell.verified);
    const preferenceOrder = platform === 'win32'
        ? ['pwsh', 'powershell', 'cmd', 'cygwin-bash', 'git-bash', 'wsl']
        : ['bash', 'zsh', 'sh'];

    for (const shellId of preferenceOrder) {
        const match = verifiedShells.find((shell) => shell.id === shellId);
        if (match) {
            return match.id;
        }
    }

    return verifiedShells[0]?.id ?? null;
}

function buildSummary(
    shells: DetectedExecutionShell[],
    tools: DetectedExecutionTool[],
    harnesses: CliHarnessDetectionResult[],
    platform: NodeJS.Platform,
): ExecutionEnvironmentSummary {
    const preferredShellId = selectPreferredShell(shells, platform);
    const preferredShell = shells.find((shell) => shell.id === preferredShellId) ?? null;

    for (const shell of shells) {
        shell.preferred = shell.id === preferredShellId;
    }

    const verifiedShellCount = shells.filter((shell) => shell.verified).length;
    const verifiedToolCount = tools.filter((tool) => tool.verified).length;
    const verifiedHarnessCount = harnesses.filter((harness) => harness.installed && !harness.detectionError).length;
    const supportsPowerShell = shells.some((shell) => shell.verified && shell.family === 'powershell');
    const supportsPosixShell = shells.some((shell) => shell.verified && (shell.family === 'posix' || shell.family === 'wsl'));
    const notes: string[] = [];

    if (preferredShell) {
        notes.push(`Prefer ${preferredShell.name} for default TormentNexus shell execution on this host.`);
    }

    if (supportsPosixShell) {
        const posixShell = shells.find((shell) => shell.verified && (shell.family === 'posix' || shell.family === 'wsl'));
        if (posixShell) {
            notes.push(`${posixShell.name} is available for POSIX-style pipelines and Unix-first tooling.`);
        }
    }

    if (!supportsPosixShell && platform === 'win32') {
        notes.push('No verified POSIX shell detected yet; TormentNexus should prefer PowerShell-compatible command plans.');
    }

    return {
        ready: verifiedShellCount > 0,
        preferredShellId: preferredShell?.id ?? null,
        preferredShellLabel: preferredShell?.name ?? null,
        shellCount: shells.filter((shell) => shell.installed).length,
        verifiedShellCount,
        toolCount: tools.filter((tool) => tool.installed).length,
        verifiedToolCount,
        harnessCount: harnesses.filter((harness) => harness.installed).length,
        verifiedHarnessCount,
        supportsPowerShell,
        supportsPosixShell,
        notes,
    };
}

export async function detectLocalExecutionEnvironment(overrides: Partial<ExecutionEnvironmentDeps> = {}): Promise<LocalExecutionEnvironment> {
    const deps: ExecutionEnvironmentDeps = {
        runCommand: overrides.runCommand ?? defaultRunCommand,
        resolveCommandPath: overrides.resolveCommandPath ?? defaultResolveCommandPath,
        pathExists: overrides.pathExists ?? defaultPathExists,
        detectHarnesses: overrides.detectHarnesses ?? (() => detectCliHarnesses()),
        platform: overrides.platform ?? process.platform,
    };

    const [shells, tools, harnesses] = await Promise.all([
        Promise.all(SHELL_CATALOG.map((entry) => detectShell(entry, deps))),
        Promise.all(TOOL_CATALOG.map((entry) => detectBinary(entry, deps))),
        deps.detectHarnesses(),
    ]);

    return {
        os: deps.platform,
        summary: buildSummary(shells, tools, harnesses, deps.platform),
        shells,
        tools,
        harnesses,
    };
}
