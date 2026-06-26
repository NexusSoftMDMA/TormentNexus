import * as pty from 'node-pty';
import { SessionSupervisor } from './SessionSupervisor.js';
import type { 
    SessionSupervisorOptions, 
    SupervisedProcessHandle, 
    SpawnProcessOptions,
    MinimalReadableLike
} from './types.js';

/**
 * SupervisedProcessHandle implementation for node-pty
 */
class PtyProcessHandle implements SupervisedProcessHandle {
    private ptyProcess: pty.IPty;
    public pid: number;
    public stdout: MinimalReadableLike;
    public stderr: MinimalReadableLike;

    constructor(command: string, args: string[], options: SpawnProcessOptions) {
        this.ptyProcess = pty.spawn(command, args, {
            name: 'xterm-color',
            cols: 80,
            rows: 30,
            cwd: options.cwd,
            env: options.env as Record<string, string>,
        });
        
        this.pid = this.ptyProcess.pid;
        
        // node-pty merges stdout and stderr into onData
        this.stdout = {
            on: (event: 'data', listener: (data: any) => void) => {
                if (event === 'data') {
                    return this.ptyProcess.onData(listener);
                }
                return { dispose: () => {} };
            }
        };
        
        this.stderr = {
            on: () => ({ dispose: () => {} }) // Not used by pty
        };
    }

    on(event: 'exit', listener: (code: number | null, signal: NodeJS.Signals | null) => void) {
        if (event === 'exit') {
            return this.ptyProcess.onExit(({ exitCode, signal }) => {
                listener(exitCode, typeof signal === 'string' ? signal as NodeJS.Signals : null);
            });
        }
        return { dispose: () => {} };
    }

    kill(signal?: NodeJS.Signals | number): boolean {
        this.ptyProcess.kill(signal as string);
        return true;
    }

    write(data: string) {
        this.ptyProcess.write(data);
    }

    resize(cols: number, rows: number) {
        this.ptyProcess.resize(cols, rows);
    }
}

/**
 * PtySupervisor - A SessionSupervisor that supports interactive PTY processes
 */
export class PtySupervisor extends SessionSupervisor {
    constructor(options: SessionSupervisorOptions = {}) {
        const spawnProcess = (command: string, args: string[], opts: SpawnProcessOptions) => {
            return new PtyProcessHandle(command, args, opts);
        };

        super({
            ...options,
            spawnProcess,
        });
    }

    /**
     * Get the raw PTY handle for writing data
     */
    public getPtyHandle(id: string): PtyProcessHandle | undefined {
        const runtime = (this as any).runtimes.get(id);
        if (runtime?.process instanceof PtyProcessHandle) {
            return runtime.process;
        }
        return undefined;
    }

    /**
     * Send input to a running PTY session
     */
    public async sendInput(id: string, data: string): Promise<void> {
        const handle = this.getPtyHandle(id);
        if (!handle) {
            throw new Error(`Session '${id}' is not an active PTY session.`);
        }
        handle.write(data);
    }
}
