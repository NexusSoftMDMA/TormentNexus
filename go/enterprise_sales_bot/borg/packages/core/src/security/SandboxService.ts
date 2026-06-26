
import vm from 'vm';
import { spawn } from 'child_process';
import { v4 as uuidv4 } from 'uuid';

export class SandboxService {

    constructor() {
        console.log("[SandboxService] Initialized (Mode: Process Isolation)");
    }

    /**
     * Reason: VM execution may return sync values or thenables from async wrappers.
     * What: Detects promise-like results via runtime narrowing on a `then` function.
     * Why: Allows async handling without unsafe structural casts.
     */
    private isPromiseLike(value: unknown): value is Promise<unknown> {
        if (!value || (typeof value !== 'object' && typeof value !== 'function')) {
            return false;
        }
        const thenCandidate = Reflect.get(value as object, 'then');
        return typeof thenCandidate === 'function';
    }

    /**
     * Execute code in a sandboxed environment.
     * @param language 'node' | 'python'
     * @param code The code to execute
     * @param timeoutMs Max execution time (default 5000ms)
     */
    public async execute(language: 'node' | 'python', code: string, timeoutMs: number = 5000, context: Record<string, any> = {}): Promise<{ output: string, result?: any, error?: string }> {
        if (language === 'node') {
            return this.executeNode(code, timeoutMs, context);
        } else if (language === 'python') {
            return this.executePython(code, timeoutMs);
        } else {
            return { output: '', error: `Unsupported language: ${language}` };
        }
    }

    private async executeNode(code: string, timeoutMs: number, contextOverrides: Record<string, any> = {}): Promise<{ output: string, result?: any, error?: string }> {
        // Capture console.log
        let outputBuffer = "";
        const sandbox = {
            console: {
                log: (...args: any[]) => { outputBuffer += args.map(a => String(a)).join(' ') + "\n"; },
                error: (...args: any[]) => { outputBuffer += "[ERR] " + args.map(a => String(a)).join(' ') + "\n"; },
                warn: (...args: any[]) => { outputBuffer += "[WARN] " + args.map(a => String(a)).join(' ') + "\n"; },
            },
            setTimeout,
            clearTimeout,
            setInterval,
            clearInterval,
            Buffer,
            ...contextOverrides, // Inject tools/variables here
            module: {},
            exports: {},
            require: (name: string) => {
                // Whitelist simple packages if needed? For now, strict no-require for safety.
                throw new Error(`require('${name}') is blocked in sandbox.`);
            }
        };

        const context = vm.createContext(sandbox);

        try {
            const script = new vm.Script(code);

            const result = script.runInContext(context, {
                timeout: timeoutMs,
                displayErrors: true
            });

            // If the result is a Promise (async code), wait for it with timeout
            if (this.isPromiseLike(result)) {
                const resolved = await Promise.race([
                    result,
                    new Promise((_, reject) => setTimeout(() => reject(new Error("Execution timed out (Async)")), timeoutMs))
                ]);
                return { output: outputBuffer.trim(), result: resolved };
            }

            return { output: outputBuffer.trim(), result };

        } catch (e: any) {
            return {
                output: outputBuffer.trim(),
                error: e.message || "Unknown Sandbox Error"
            };
        }
    }

    private executePython(code: string, timeoutMs: number): Promise<{ output: string, error?: string }> {
        return new Promise((resolve) => {
            const child = spawn('python', ['-c', code], {
                timeout: timeoutMs
            });

            let stdout = "";
            let stderr = "";
            let killed = false;

            const timer = setTimeout(() => {
                killed = true;
                child.kill();
                resolve({ output: stdout, error: "Execution timed out" });
            }, timeoutMs);

            child.stdout.on('data', (data) => {
                stdout += data.toString();
            });

            child.stderr.on('data', (data) => {
                stderr += data.toString();
            });

            child.on('error', (err) => {
                clearTimeout(timer);
                if (!killed) resolve({ output: stdout, error: err.message });
            });

            child.on('close', (code) => {
                clearTimeout(timer);
                if (!killed) {
                    if (code !== 0) {
                        resolve({ output: stdout, error: stderr || `Process exited with code ${code}` });
                    } else {
                        resolve({ output: stdout.trim() });
                    }
                }
            });
        });
    }
}
