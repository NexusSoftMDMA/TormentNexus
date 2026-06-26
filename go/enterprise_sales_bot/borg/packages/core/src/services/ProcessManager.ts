import { spawn, ChildProcess, execSync } from 'child_process';
import { EventEmitter } from 'events';
import * as fs from 'fs';
import * as path from 'path';

export interface ProcessConfig {
    sessionId: string;
    command: string;
    args: string[];
    cwd: string;
    env?: Record<string, string>;
}

export interface ProcessOutput {
    sessionId: string;
    data: string;
    type: 'stdout' | 'stderr';
}

function findTerminalEmulator(): 'tabby' | 'warp' | null {
    if (process.env.TERM_PROGRAM === 'Tabby') return 'tabby';
    if (process.env.TERM_PROGRAM === 'Warp') return 'warp';

    const localAppData = process.env.LOCALAPPDATA || '';
    const programFiles = process.env.PROGRAMFILES || '';

    const tabbyPaths = [
        path.join(localAppData, 'Programs', 'Tabby', 'Tabby.exe'),
        path.join(localAppData, 'Programs', 'Tabby', 'tabby.exe')
    ];
    for (const p of tabbyPaths) {
        if (fs.existsSync(p)) return 'tabby';
    }

    const warpPaths = [
        path.join(localAppData, 'Warp', 'Warp.exe'),
        path.join(programFiles, 'Warp', 'Warp.exe')
    ];
    for (const p of warpPaths) {
        if (fs.existsSync(p)) return 'warp';
    }

    try {
        execSync(process.platform === 'win32' ? 'where tabby' : 'which tabby', { stdio: 'ignore' });
        return 'tabby';
    } catch {}

    try {
        execSync(process.platform === 'win32' ? 'where warp-terminal' : 'which warp-terminal', { stdio: 'ignore' });
        return 'warp';
    } catch {}

    return null;
}

export class ProcessManager extends EventEmitter {
    private activeProcesses: Map<string, ChildProcess> = new Map();

    /**
     * Spawns a new process and tracks its output.
     */
    async spawn(config: ProcessConfig): Promise<{ pid: number; success: boolean }> {
        const emulator = findTerminalEmulator();
        let cmd = config.command;
        let args = config.args;

        if (emulator === 'tabby') {
            console.log(`[Core:Process] Detected Tabby: wrapping command`);
            cmd = 'tabby';
            args = ['run', config.command, ...config.args];
        } else if (emulator === 'warp') {
            console.log(`[Core:Process] Detected Warp: wrapping command`);
            cmd = 'warp-terminal';
            args = ['--', config.command, ...config.args];
        }

        console.log(`[Core:Process] Spawning: ${cmd} ${args.join(' ')}`);
        
        try {
            const child = spawn(cmd, args, {
                cwd: config.cwd,
                env: { ...process.env, ...config.env },
                shell: true
            });

            if (!child.pid) {
                return { pid: -1, success: false };
            }

            this.activeProcesses.set(config.sessionId, child);

            child.stdout?.on('data', (data) => {
                this.emit('output', {
                    sessionId: config.sessionId,
                    data: data.toString(),
                    type: 'stdout'
                });
            });

            child.stderr?.on('data', (data) => {
                this.emit('output', {
                    sessionId: config.sessionId,
                    data: data.toString(),
                    type: 'stderr'
                });
            });

            child.on('close', (code) => {
                console.log(`[Core:Process] Process ${config.sessionId} exited with code ${code}`);
                this.activeProcesses.delete(config.sessionId);
                this.emit('exit', { sessionId: config.sessionId, code });
            });

            return { pid: child.pid, success: true };
        } catch (error) {
            console.error(`[Core:Process] Failed to spawn ${config.command}:`, error);
            return { pid: -1, success: false };
        }
    }

    /**
     * Writes data to a process's stdin.
     */
    write(sessionId: string, data: string): boolean {
        const child = this.activeProcesses.get(sessionId);
        if (child && child.stdin && child.stdin.writable) {
            child.stdin.write(data);
            return true;
        }
        return false;
    }

    /**
     * Kills an active process.
     */
    kill(sessionId: string): boolean {
        const child = this.activeProcesses.get(sessionId);
        if (child) {
            child.kill();
            this.activeProcesses.delete(sessionId);
            return true;
        }
        return false;
    }

    /**
     * Lists all active process session IDs.
     */
    listActiveSessions(): string[] {
        return Array.from(this.activeProcesses.keys());
    }
}
