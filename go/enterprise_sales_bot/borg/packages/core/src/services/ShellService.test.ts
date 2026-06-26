import { describe, expect, it } from 'vitest';

import { buildShellInvocation } from './ShellService.ts';

describe('buildShellInvocation', () => {
    it('builds a PowerShell invocation when requested', () => {
        expect(buildShellInvocation('Get-ChildItem', {
            shellFamily: 'powershell',
            shellPath: 'C:\\Program Files\\PowerShell\\7\\pwsh.exe',
        })).toEqual({
            command: 'C:\\Program Files\\PowerShell\\7\\pwsh.exe',
            args: ['-NoLogo', '-NoProfile', '-Command', 'Get-ChildItem'],
        });
    });

    it('builds a cmd invocation for compatibility mode', () => {
        expect(buildShellInvocation('dir', {
            shellFamily: 'cmd',
            shellPath: 'C:\\Windows\\System32\\cmd.exe',
        })).toEqual({
            command: 'C:\\Windows\\System32\\cmd.exe',
            args: ['/d', '/s', '/c', 'dir'],
        });
    });

    it('builds a WSL invocation for Linux-native commands', () => {
        expect(buildShellInvocation('pwd', {
            shellFamily: 'wsl',
            shellPath: 'wsl',
        })).toEqual({
            command: 'wsl',
            args: ['-e', 'sh', '-lc', 'pwd'],
        });
    });
});