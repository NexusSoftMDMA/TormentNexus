import { execFile } from 'child_process';
import util from 'util';

const execFileAsync = util.promisify(execFile);

export function toPowerShellString(value: string): string {
    return `'${value.replace(/'/g, "''")}'`;
}

export async function runPowerShell(script: string): Promise<string> {
    const { stdout, stderr } = await execFileAsync(
        'powershell',
        [
            '-NoLogo',
            '-NoProfile',
            '-NonInteractive',
            '-ExecutionPolicy',
            'Bypass',
            '-Command',
            script
        ],
        {
            windowsHide: true,
            maxBuffer: 4 * 1024 * 1024
        }
    );

    if (stderr?.trim()) {
        throw new Error(stderr.trim());
    }

    return stdout.trim();
}

export async function runPowerShellJson<T>(script: string): Promise<T> {
    const output = await runPowerShell(script);

    if (!output) {
        throw new Error('PowerShell command returned no JSON output');
    }

    return JSON.parse(output) as T;
}
