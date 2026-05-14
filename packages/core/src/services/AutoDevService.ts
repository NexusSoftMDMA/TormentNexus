/**
 * AutoDevService - "Fix until Pass" loops for tests and linters
 * Automatically retries fixing code until tests/lints pass or max attempts reached
 */

import { spawnAsync } from '../utils/exec.js';

export interface LoopConfig {
    maxAttempts: number;
    type: 'test' | 'lint' | 'build';
    target?: string; // Specific file or pattern
    command?: string; // Custom command override
}

export interface LoopResult {
    success: boolean;
    attempts: number;
    output: string;
    errors: string[];
    duration: number;
}

export interface ActiveLoop {
    id: string;
    config: LoopConfig;
    status: 'running' | 'success' | 'failed' | 'cancelled';
    currentAttempt: number;
    startTime: number;
    lastOutput: string;
}

interface AutoDevDirector {
    executeTask(goal: string, maxSteps?: number): Promise<unknown>;
}

export class AutoDevService {
    private activeLoops: Map<string, ActiveLoop> = new Map();
    private loopCounter = 0;
    private director?: AutoDevDirector;
    private rootDir: string;

    constructor(rootDir: string, director?: AutoDevDirector) {
        this.rootDir = rootDir;
        this.director = director;
    }

    /**
     * Start a "Fix until Pass" loop
     */
    async startLoop(config: LoopConfig): Promise<string> {
        const id = `loop-${++this.loopCounter}`;
        const loop: ActiveLoop = {
            id,
            config,
            status: 'running',
            currentAttempt: 0,
            startTime: Date.now(),
            lastOutput: ''
        };

        this.activeLoops.set(id, loop);
        console.log(`[AutoDev] 🔄 Starting ${config.type} loop (max ${config.maxAttempts} attempts)`);

        // Run the loop asynchronously
        this.runLoop(id).catch(e => {
            console.error(`[AutoDev] Loop ${id} error:`, e);
            const l = this.activeLoops.get(id);
            if (l) l.status = 'failed';
        });

        return id;
    }

    /**
     * Cancel an active loop
     */
    cancelLoop(id: string): boolean {
        const loop = this.activeLoops.get(id);
        if (loop && loop.status === 'running') {
            loop.status = 'cancelled';
            console.log(`[AutoDev] 🛑 Cancelled loop ${id}`);
            return true;
        }
        return false;
    }

    /**
     * Get status of all loops
     */
    getLoops(): ActiveLoop[] {
        return Array.from(this.activeLoops.values());
    }

    /**
     * Get a specific loop
     */
    getLoop(id: string): ActiveLoop | undefined {
        return this.activeLoops.get(id);
    }

    private async runLoop(id: string): Promise<void> {
        const loop = this.activeLoops.get(id);
        if (!loop) return;

        const { config } = loop;
        const { cmd, args } = this.getCommandParts(config);

        while (loop.currentAttempt < config.maxAttempts && loop.status === 'running') {
            loop.currentAttempt++;
            console.log(`[AutoDev] Attempt ${loop.currentAttempt}/${config.maxAttempts}`);

            try {
                const result = await spawnAsync(cmd, args, {
                    cwd: this.rootDir,
                    timeout: 120000 // 2 minute timeout per attempt
                });

                loop.lastOutput = result.stdout || result.stderr;

                if (result.exitCode === 0) {
                    loop.status = 'success';
                    console.log(`[AutoDev] ✅ ${config.type} passed on attempt ${loop.currentAttempt}`);
                    return;
                }

                throw new Error(loop.lastOutput);

            } catch (error: unknown) {
                const message = error instanceof Error ? error.message : String(error);
                loop.lastOutput = message;

                if (loop.status !== 'running') {
                    return;
                }

                console.log(`[AutoDev] ❌ Attempt ${loop.currentAttempt} failed`);

                if (loop.currentAttempt >= config.maxAttempts) {
                    loop.status = 'failed';
                    console.log(`[AutoDev] 💀 Max attempts reached. Loop failed.`);
                    return;
                }

                // exponential backoff
                const delay = Math.min(1000 * Math.pow(2, loop.currentAttempt - 1), 30000);

                // AUTONOMOUS REPAIR
                if (this.director && loop.status === 'running') {
                    console.log(`[AutoDev] 🔧 Requesting Director fix...`);
                    const goal = `Fix the following ${config.type} error in ${config.target || 'the project'}. 
Output:
${loop.lastOutput.substring(0, 2000)}

Please analyze the file, fix the code, and ensure it passes.`;

                    try {
                        await this.director.executeTask(goal, 5);
                    } catch (e) {
                        console.error(`[AutoDev] Director fix failed:`, e);
                    }
                } else {
                    await new Promise(r => setTimeout(r, delay));
                }
            }
        }
    }

    private getCommandParts(config: LoopConfig): { cmd: string, args: string[] } {
        if (config.command) {
            const parts = config.command.split(' ');
            return { cmd: parts[0], args: parts.slice(1) };
        }

        const isWin = process.platform === 'win32';
        const npx = isWin ? 'npx.cmd' : 'npx';
        const npm = isWin ? 'npm.cmd' : 'npm';

        switch (config.type) {
            case 'test':
                return config.target
                    ? { cmd: npx, args: ['vitest', 'run', config.target] }
                    : { cmd: npm, args: ['test'] };
            case 'lint':
                return config.target
                    ? { cmd: npx, args: ['eslint', '--fix', config.target] }
                    : { cmd: npm, args: ['run', 'lint', '--', '--fix'] };
            case 'build':
                return { cmd: npm, args: ['run', 'build'] };
            default:
                return { cmd: npm, args: ['test'] };
        }
    }

    /**
     * Clear completed loops
     */
    clearCompleted(): number {
        let count = 0;
        for (const [id, loop] of this.activeLoops) {
            if (loop.status !== 'running') {
                this.activeLoops.delete(id);
                count++;
            }
        }
        return count;
    }
}
