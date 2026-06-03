import fs from 'fs';
import path from 'path';
import os from 'os';
import { exec, spawn, type ChildProcessWithoutNullStreams, type SpawnOptionsWithoutStdio } from 'child_process';

export interface ShellCommandEntry {
    id: string;
    command: string;
    cwd: string;
    timestamp: number;
    exitCode?: number;
    duration?: number;
    outputSnippet?: string;
    session: string;
}

export interface ShellExecutionPolicyLike {
    shellFamily?: 'powershell' | 'cmd' | 'posix' | 'wsl' | null;
    shellPath?: string | null;
    shellLabel?: string | null;
}

export interface ShellInvocation {
    command: string;
    args: string[];
}

export interface ShellExecutionOptions {
    cwd?: string;
    env?: Record<string, string>;
    session?: string;
    executionPolicy?: ShellExecutionPolicyLike | null;
    timeoutMs?: number;
}

export interface ShellExecutionResult {
    command: string;
    cwd: string;
    shellFamily: ShellExecutionPolicyLike['shellFamily'] | 'default';
    shellPath: string | null;
    stdout: string;
    stderr: string;
    output: string;
    exitCode: number;
    durationMs: number;
    succeeded: boolean;
}

type ShellExecutionRuntimeResult = Omit<ShellExecutionResult, 'command' | 'cwd' | 'shellFamily' | 'shellPath'>;

export function buildShellInvocation(command: string, policy?: ShellExecutionPolicyLike | null): ShellInvocation | null {
    if (!policy?.shellFamily) {
        return null;
    }

    switch (policy.shellFamily) {
        case 'powershell':
            return {
                command: policy.shellPath || 'pwsh',
                args: ['-NoLogo', '-NoProfile', '-Command', command],
            };
        case 'cmd':
            return {
                command: policy.shellPath || process.env.COMSPEC || 'cmd.exe',
                args: ['/d', '/s', '/c', command],
            };
        case 'posix':
            return {
                command: policy.shellPath || process.env.SHELL || 'sh',
                args: ['-lc', command],
            };
        case 'wsl':
            return {
                command: policy.shellPath || 'wsl',
                args: ['-e', 'sh', '-lc', command],
            };
        default:
            return null;
    }
}

export class ShellService {
    private historyPath: string;
    private enrichedHistoryPath: string;
    private historyCache: ShellCommandEntry[] = [];

    constructor() {
        // Platform specific history file
        this.historyPath = process.platform === 'win32'
            ? path.join(os.homedir(), 'AppData', 'Roaming', 'Microsoft', 'Windows', 'PowerShell', 'PSReadLine', 'ConsoleHost_history.txt')
            : path.join(os.homedir(), '.bash_history'); // Simple fallback

        this.enrichedHistoryPath = path.join(process.cwd(), '.tormentnexus', 'shell_history.json');
        this.loadEnrichedHistory();
    }

    private loadEnrichedHistory() {
        try {
            if (fs.existsSync(this.enrichedHistoryPath)) {
                const data = fs.readFileSync(this.enrichedHistoryPath, 'utf-8');
                this.historyCache = JSON.parse(data);
            }
        } catch (e) {
            console.error('[ShellService] Failed to load enriched history', e);
            this.historyCache = [];
        }
    }

    private saveEnrichedHistory() {
        try {
            const dir = path.dirname(this.enrichedHistoryPath);
            if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

            // Limit cache size
            if (this.historyCache.length > 1000) {
                this.historyCache = this.historyCache.slice(-1000);
            }

            fs.writeFileSync(this.enrichedHistoryPath, JSON.stringify(this.historyCache, null, 2));
        } catch (e) {
            console.error('[ShellService] Failed to save history', e);
        }
    }

    /**
     * Log a command execution
     */
    async logCommand(entry: Omit<ShellCommandEntry, 'id' | 'timestamp'>): Promise<string> {
        const id = Math.random().toString(36).substring(2, 11);
        const fullEntry: ShellCommandEntry = {
            ...entry,
            id,
            timestamp: Date.now()
        };

        this.historyCache.push(fullEntry);
        this.saveEnrichedHistory();
        return id;
    }

    /**
     * Search history with semantic-ish filter
     */
    async queryHistory(query: string, limit: number = 20): Promise<ShellCommandEntry[]> {
        const lowerQuery = query.toLowerCase();
        return this.historyCache
            .filter(entry =>
                entry.command.toLowerCase().includes(lowerQuery) ||
                entry.outputSnippet?.toLowerCase().includes(lowerQuery)
            )
            .sort((a, b) => b.timestamp - a.timestamp) // Newest first
            .slice(0, limit);
    }

    /**
     * Reads the last N lines from the system history file.
     * Efficiently reads from the end of the file.
     */
    async getSystemHistory(limit: number = 50): Promise<string[]> {
        if (!fs.existsSync(this.historyPath)) {
            return [];
        }

        try {
            const stats = fs.statSync(this.historyPath);
            const fileSize = stats.size;
            if (fileSize === 0) return [];

            // Simple read if small
            if (fileSize < 1024 * 1024) { // 1MB
                const content = fs.readFileSync(this.historyPath, 'utf-8');
                const allLines = content.split(/\r?\n/).filter(line => line.trim() !== '');
                return allLines.slice(-limit);
            }

            // Fallback for large files
            const readSize = Math.min(50 * 1024, fileSize);
            const buffer = Buffer.alloc(readSize);
            const fd = fs.openSync(this.historyPath, 'r');
            fs.readSync(fd, buffer, 0, readSize, fileSize - readSize);
            fs.closeSync(fd);

            const content = buffer.toString('utf-8');
            const allLines = content.split(/\r?\n/).slice(1).filter(line => line.trim() !== '');
            return allLines.slice(-limit);

        } catch (error) {
            console.error('[ShellService] Error reading history:', error);
            return [];
        }
    }

    /**
     * Executes a command on the host shell.
     */
    async execute(command: string, cwd: string = process.cwd()): Promise<string> {
        const result = await this.executeWithContext(command, { cwd, session: 'api-call' });
        if (!result.succeeded) {
            const error = new Error(result.output || `Command failed with exit code ${result.exitCode}`) as Error & { code?: number };
            error.code = result.exitCode;
            throw error;
        }
        return result.output;
    }

    async executeWithContext(command: string, options: ShellExecutionOptions = {}): Promise<ShellExecutionResult> {
        const cwd = options.cwd ?? process.cwd();
        const env = {
            ...process.env,
            ...(options.env ?? {}),
        };
        const timeoutMs = options.timeoutMs ?? 15_000;
        const session = options.session ?? 'api-call';
        const policy = options.executionPolicy ?? null;
        const shellInvocation = buildShellInvocation(command, policy);
        const startedAt = Date.now();

        const result = shellInvocation
            ? await this.executeViaSpawn(shellInvocation, cwd, env, timeoutMs)
            : await this.executeViaExec(command, cwd, env, timeoutMs);

        await this.logCommand({
            command,
            cwd,
            outputSnippet: result.output.substring(0, 200),
            session,
            exitCode: result.exitCode,
            duration: result.durationMs,
        }).catch(() => { });

        return {
            ...result,
            command,
            cwd,
            shellFamily: policy?.shellFamily ?? 'default',
            shellPath: policy?.shellPath ?? null,
            durationMs: Date.now() - startedAt,
        };
    }

    private executeViaExec(command: string, cwd: string, env: NodeJS.ProcessEnv, timeoutMs: number): Promise<ShellExecutionRuntimeResult> {
        return new Promise((resolve) => {
            const startedAt = Date.now();
            exec(command, { cwd, env, timeout: timeoutMs }, (error, stdout, stderr) => {
                const resolvedStdout = stdout || '';
                const resolvedStderr = stderr || '';
                const output = resolvedStdout || resolvedStderr || '';
                resolve({
                    stdout: resolvedStdout,
                    stderr: resolvedStderr,
                    output,
                    exitCode: error ? 1 : 0,
                    durationMs: Date.now() - startedAt,
                    succeeded: !error,
                });
            });
        });
    }

    private executeViaSpawn(
        invocation: ShellInvocation,
        cwd: string,
        env: NodeJS.ProcessEnv,
        timeoutMs: number,
    ): Promise<ShellExecutionRuntimeResult> {
        return new Promise((resolve) => {
            const startedAt = Date.now();
            let stdout = '';
            let stderr = '';
            let settled = false;
            let timedOut = false;

            const spawnOptions: SpawnOptionsWithoutStdio = {
                cwd,
                env,
                windowsHide: true,
            };

            let child: ChildProcessWithoutNullStreams;
            try {
                child = spawn(invocation.command, invocation.args, spawnOptions);
            } catch (error) {
                resolve({
                    stdout: '',
                    stderr: error instanceof Error ? error.message : 'Failed to spawn shell process',
                    output: error instanceof Error ? error.message : 'Failed to spawn shell process',
                    exitCode: 1,
                    durationMs: Date.now() - startedAt,
                    succeeded: false,
                });
                return;
            }

            const timer = setTimeout(() => {
                timedOut = true;
                child.kill();
            }, timeoutMs);

            child.stdout.on('data', (chunk) => {
                stdout += chunk.toString();
            });
            child.stderr.on('data', (chunk) => {
                stderr += chunk.toString();
            });
            child.on('error', (error) => {
                if (settled) {
                    return;
                }
                settled = true;
                clearTimeout(timer);
                const message = error instanceof Error ? error.message : 'Shell process error';
                resolve({
                    stdout,
                    stderr: stderr || message,
                    output: stdout || stderr || message,
                    exitCode: 1,
                    durationMs: Date.now() - startedAt,
                    succeeded: false,
                });
            });
            child.on('close', (code) => {
                if (settled) {
                    return;
                }
                settled = true;
                clearTimeout(timer);
                const exitCode = timedOut ? 124 : (code ?? 1);
                const timeoutSuffix = timedOut ? `${stderr ? '\n' : ''}Command timed out after ${timeoutMs}ms.` : '';
                const finalStderr = `${stderr}${timeoutSuffix}`;
                resolve({
                    stdout,
                    stderr: finalStderr,
                    output: stdout || finalStderr || '',
                    exitCode,
                    durationMs: Date.now() - startedAt,
                    succeeded: exitCode === 0,
                });
            });
        });
    }
}
