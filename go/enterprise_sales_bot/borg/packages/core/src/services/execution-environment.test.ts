import { describe, expect, it } from 'vitest';

import { detectLocalExecutionEnvironment } from './execution-environment.js';

describe('detectLocalExecutionEnvironment', () => {
    it('detects verified PowerShell and Cygwin shells plus common binaries', async () => {
        const existingPaths = new Set([
            'C:\\Program Files\\PowerShell\\7\\pwsh.exe',
            'C:\\cygwin64\\bin\\bash.exe',
        ]);

        const environment = await detectLocalExecutionEnvironment({
            platform: 'win32',
            pathExists: async (filePath) => existingPaths.has(filePath),
            resolveCommandPath: async (command) => {
                switch (command) {
                    case 'pwsh':
                        return 'C:\\Program Files\\PowerShell\\7\\pwsh.exe';
                    case 'git':
                        return 'C:\\Program Files\\Git\\cmd\\git.exe';
                    case 'rg':
                        return 'C:\\tools\\rg.exe';
                    case 'python':
                        return 'C:\\Python313\\python.exe';
                    case 'node':
                        return 'C:\\Program Files\\nodejs\\node.exe';
                    case 'pnpm':
                        return 'C:\\Users\\hyper\\AppData\\Roaming\\npm\\pnpm.cmd';
                    default:
                        return null;
                }
            },
            runCommand: async (command, args) => {
                const joined = `${command} ${args.join(' ')}`;
                if (joined.includes('tormentnexus-ready')) {
                    return 'tormentnexus-ready';
                }
                if (joined.includes('PSVersionTable')) {
                    return '7.5.0';
                }
                if (joined.includes('uname -s')) {
                    return 'CYGWIN_NT-10.0 3.5.4';
                }
                if (joined.includes('git.exe')) {
                    return 'git version 2.48.1.windows.1';
                }
                if (joined.includes('rg.exe')) {
                    return 'ripgrep 14.1.0';
                }
                if (joined.includes('python.exe')) {
                    return 'Python 3.13.2';
                }
                if (joined.includes('node.exe')) {
                    return 'v22.14.0';
                }
                if (joined.includes('pnpm.cmd')) {
                    return '10.6.1';
                }
                return null;
            },
            detectHarnesses: async () => ([
                {
                    id: 'claude',
                    name: 'Claude Code',
                    command: 'claude',
                    args: [],
                    homepage: 'https://example.com',
                    docsUrl: 'https://example.com/docs',
                    installHint: 'install',
                    category: 'cli',
                    sessionCapable: true,
                    versionArgs: ['--version'],
                    installed: true,
                    resolvedPath: 'C:\\Users\\hyper\\AppData\\Roaming\\npm\\claude.cmd',
                    version: '1.0.0',
                    detectionError: null,
                },
            ]),
        });

        expect(environment.summary.ready).toBe(true);
        expect(environment.summary.preferredShellId).toBe('pwsh');
        expect(environment.summary.supportsPowerShell).toBe(true);
        expect(environment.summary.supportsPosixShell).toBe(true);
        expect(environment.summary.verifiedToolCount).toBe(5);
        expect(environment.summary.verifiedHarnessCount).toBe(1);
        expect(environment.shells.find((shell) => shell.id === 'cygwin-bash')).toEqual(
            expect.objectContaining({ installed: true, verified: true }),
        );
    });

    it('marks the environment partial when no verified shell is available', async () => {
        const environment = await detectLocalExecutionEnvironment({
            platform: 'win32',
            pathExists: async () => false,
            resolveCommandPath: async (command) => (command === 'cmd' ? 'C:\\Windows\\System32\\cmd.exe' : null),
            runCommand: async () => null,
            detectHarnesses: async () => [],
        });

        expect(environment.summary.ready).toBe(false);
        expect(environment.summary.preferredShellId).toBeNull();
        expect(environment.summary.verifiedShellCount).toBe(0);
        expect(environment.summary.notes).toContain('No verified POSIX shell detected yet; TormentNexus should prefer PowerShell-compatible command plans.');
    });
});
