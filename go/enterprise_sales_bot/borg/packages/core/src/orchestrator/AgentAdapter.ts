import { spawn, ChildProcess } from 'child_process';
import { EventEmitter } from 'events';

export interface AgentConfig {
    name: string;
    command: string;
    args: string[];
    cwd?: string;
    env?: NodeJS.ProcessEnv;
}

export abstract class AgentAdapter extends EventEmitter {
    protected process: ChildProcess | null = null;
    protected config: AgentConfig;
    protected isRunning: boolean = false;

    constructor(config: AgentConfig) {
        super();
        this.config = config;
    }

    public async start(): Promise<void> {
        if (this.isRunning) return;

        console.log(`[AgentAdapter] Spawning ${this.config.name} (${this.config.command})...`);

        this.process = spawn(this.config.command, this.config.args, {
            cwd: this.config.cwd || process.cwd(),
            env: { ...process.env, ...this.config.env },
            stdio: ['pipe', 'pipe', 'pipe'],
            shell: true // Useful for windows compatibility usually
        });

        this.isRunning = true;

        this.process.stdout?.on('data', (data) => {
            this.emit('output', data.toString());
        });

        this.process.stderr?.on('data', (data) => {
            this.emit('error_output', data.toString());
        });

        this.process.on('close', (code) => {
            console.log(`[AgentAdapter] ${this.config.name} exited with code ${code}`);
            this.isRunning = false;
            this.process = null;
            this.emit('exit', code);
        });

        this.process.on('error', (err) => {
            console.error(`[AgentAdapter] ${this.config.name} error:`, err);
            this.emit('error', err);
        });
    }

    public async stop(): Promise<void> {
        if (!this.process || !this.isRunning) return;

        console.log(`[AgentAdapter] Stopping ${this.config.name}...`);
        this.process.kill();
        this.isRunning = false;
        this.process = null;
    }

    public async send(input: string): Promise<void> {
        if (!this.process || !this.isRunning) {
            throw new Error(`Agent ${this.config.name} is not running.`);
        }

        // Write input to stdin
        this.process.stdin?.write(input + '\n');
    }

    public isActive(): boolean {
        return this.isRunning;
    }
}
