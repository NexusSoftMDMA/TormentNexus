/**
 * AutoDevService - "Fix until Pass" loops for tests and linters
 * Automatically retries fixing code until tests/lints pass or max attempts reached
 */

import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

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
        const command = this.getCommand(config);

        while (loop.currentAttempt < config.maxAttempts && loop.status === 'running') {
            loop.currentAttempt++;
            console.log(`[AutoDev] Attempt ${loop.currentAttempt}/${config.maxAttempts}`);

            try {
                const { stdout, stderr } = await execAsync(command, {
                    cwd: this.rootDir,
                    timeout: 120000 // 2 minute timeout per attempt
                });

                loop.lastOutput = stdout || stderr;

                // Success!
                loop.status = 'success';
                console.log(`[AutoDev] ✅ ${config.type} passed on attempt ${loop.currentAttempt}`);
                return;

            } catch (error: unknown) {
                const errorRecord = error && typeof error === 'object'
                    ? (error as Record<string, unknown>)
                    : null;

                const stdout = typeof errorRecord?.stdout === 'string' ? errorRecord.stdout : '';
                const stderr = typeof errorRecord?.stderr === 'string' ? errorRecord.stderr : '';
                const message = error instanceof Error ? error.message : String(error);

                loop.lastOutput = stdout || stderr || message;

                if (loop.status !== 'running') {
                    // Cancelled
                    return;
                }

                console.log(`[AutoDev] ❌ Attempt ${loop.currentAttempt} failed`);

                // Don't retry on last attempt
                if (loop.currentAttempt >= config.maxAttempts) {
                    loop.status = 'failed';
                    console.log(`[AutoDev] 💀 Max attempts reached. Loop failed.`);
                    return;
                }

                // Wait before retry (exponential backoff)
                const delay = Math.min(1000 * Math.pow(2, loop.currentAttempt - 1), 30000);

                // AUTONOMOUS REPAIR
                if (this.director && loop.status === 'running') {
                    console.log(`[AutoDev] 🔧 Requesting Director fix...`);
                    const goal = `Fix the following ${config.type} error in ${config.target || 'the project'}. 
Output:
${loop.lastOutput.substring(0, 2000)}

Please analyze the file, fix the code, and ensure it passes.`;

                    try {
                        // Give the director a few steps to fix it
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

    private getCommand(config: LoopConfig): string {
        if (config.command) return config.command;

        switch (config.type) {
            case 'test':
                return config.target
                    ? `npx vitest run ${config.target}`
                    : 'npm test';
            case 'lint':
                return config.target
                    ? `npx eslint --fix ${config.target}`
                    : 'npm run lint -- --fix';
            case 'build':
                return 'npm run build';
            default:
                return 'npm test';
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
