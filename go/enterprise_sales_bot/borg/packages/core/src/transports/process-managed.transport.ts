import { ChildProcess, IOType } from "node:child_process";
import process from "node:process";
import { PassThrough, Stream } from "node:stream";

import { Transport } from "@modelcontextprotocol/sdk/shared/transport.js";
import { JSONRPCMessage } from "@modelcontextprotocol/sdk/types.js";
import spawn from "cross-spawn";

import { ReadBuffer, serializeMessage } from "@modelcontextprotocol/sdk/shared/stdio.js";
import { getDefaultEnvironment } from "../services/utils.service.js";

export type StdioServerParameters = {
    /**
     * The executable to run to start the server.
     */
    command: string;

    /**
     * Command line arguments to pass to the executable.
     */
    args?: string[];

    /**
     * The environment to use when spawning the process.
     *
     * If not specified, the result of getDefaultEnvironment() will be used.
     */
    env?: Record<string, string>;

    /**
     * How to handle stderr of the child process. This matches the semantics of Node's `child_process.spawn`.
     *
     * The default is "inherit", meaning messages to stderr will be printed to the parent process's stderr.
     */
    stderr?: IOType | Stream | number;

    /**
     * The working directory to use when spawning the process.
     *
     * If not specified, the current working directory will be inherited.
     */
    cwd?: string;
};

/**
 * Client transport for stdio: this will connect to a server by spawning a process and communicating with it over stdin/stdout.
 * 
 * This transport is only available in Node.js environments.
 */
export class ProcessManagedStdioTransport implements Transport {
    private _process?: ChildProcess;
    private _abortController: AbortController = new AbortController();
    private _readBuffer: ReadBuffer = new ReadBuffer();
    private _serverParams: StdioServerParameters;
    private _stderrStream: PassThrough | null = null;
    private _stdoutStream: PassThrough = new PassThrough();
    private _isCleanup: boolean = false;

    private isIgnorableKillError(error: unknown): boolean {
        if (!(error instanceof Error)) {
            return false;
        }

        const maybeCode = (error as NodeJS.ErrnoException).code;
        return maybeCode === "ESRCH" || maybeCode === "EPERM";
    }

    onclose?: () => void;
    onerror?: (error: Error) => void;
    onmessage?: (message: JSONRPCMessage) => void;
    onprocesscrash?: (exitCode: number | null, signal: string | null) => void;

    constructor(server: StdioServerParameters) {
        this._serverParams = server;
        if (server.stderr === "pipe" || server.stderr === "overlapped") {
            this._stderrStream = new PassThrough();
        }
    }

    /**
     * Starts the server process and prepares to communicate with it.
     */
    async start(): Promise<void> {
        if (this._process) {
            throw new Error(
                "StdioClientTransport already started! If using Client class, note that connect() calls start() automatically.",
            );
        }

        return new Promise((resolve, reject) => {
            this._process = spawn(
                this._serverParams.command,
                this._serverParams.args ?? [],
                {
                    // merge default env with server env because mcp server needs some env vars
                    env: {
                        ...getDefaultEnvironment(),
                        ...this._serverParams.env,
                    },
                    stdio: ["pipe", "pipe", this._serverParams.stderr ?? "inherit"],
                    shell: false,
                    signal: this._abortController.signal,
                    // Always hide child process console windows on Windows.
                    windowsHide: process.platform === "win32",
                    cwd: this._serverParams.cwd,
                    // Keep attached so lifecycle stays deterministic and no extra terminal window/process group appears.
                    detached: false,
                },
            );

            this._process.on("error", (error) => {
                if (error.name === "AbortError") {
                    // Expected when close() is called.
                    this.onclose?.();
                    return;
                }

                reject(error);
                this.onerror?.(error);
            });

            this._process.on("spawn", () => {
                resolve();
            });

            this._process.on("close", (code, signal) => {
                // Only emit crash event if this wasn't a clean shutdown
                if (!this._isCleanup && (code !== 0 || signal)) {
                    console.warn(`Process crashed with code: ${code}, signal: ${signal}`);
                    console.info(
                        `Calling onprocesscrash handler: ${this.onprocesscrash ? "handler exists" : "no handler"}`,
                    );
                    this.onprocesscrash?.(code, signal);
                }

                this._process = undefined;
                this.onclose?.();
            });

            this._process.stdin?.on("error", (error) => {
                this.onerror?.(error);
            });

            this._process.stdout?.on("data", (chunk) => {
                this._stdoutStream.write(chunk);
                this._readBuffer.append(chunk);
                this.processReadBuffer();
            });

            this._process.stdout?.on("error", (error) => {
                this.onerror?.(error);
            });

            if (this._stderrStream && this._process.stderr) {
                this._process.stderr.pipe(this._stderrStream);
            }
        });
    }

    /**
     * The stderr stream of the child process, if `StdioServerParameters.stderr` was set to "pipe" or "overlapped".
     *
     * If stderr piping was requested, a PassThrough stream is returned _immediately_, allowing callers to
     * attach listeners before the start method is invoked. This prevents loss of any early
     * error output emitted by the child process.
     */
    get stderr(): Stream | null {
        if (this._stderrStream) {
            return this._stderrStream;
        }

        return this._process?.stderr ?? null;
    }

    /**
     * Raw stdout stream from the child process. This still carries JSON-RPC frames,
     * but exposing it allows external observers to log process output for diagnostics.
     */
    get stdout(): Stream | null {
        return this._stdoutStream ?? this._process?.stdout ?? null;
    }

    /**
     * The child process pid spawned by this transport.
     *
     * This is only available after the transport has been started.
     */
    get pid(): number | null {
        return this._process?.pid ?? null;
    }

    private processReadBuffer() {
        while (true) {
            try {
                const message = this._readBuffer.readMessage();
                if (message === null) {
                    break;
                }

                this.onmessage?.(message);
            } catch (error) {
                this.onerror?.(error as Error);
            }
        }
    }

    async close(): Promise<void> {
        this._isCleanup = true;
        this._abortController.abort();

        const child = this._process;
        this._process = undefined;
        this._readBuffer.clear();

        // AbortController above usually terminates the child. This explicit kill is a fallback only.
        if (!child?.pid || child.killed || child.exitCode !== null || child.signalCode !== null) {
            return;
        }

        try {
            if (process.platform === "win32") {
                child.kill("SIGTERM");
            } else {
                // On POSIX we can target the process group to avoid lingering grandchildren.
                process.kill(-child.pid, "SIGTERM");
            }
        } catch (error) {
            if (!this.isIgnorableKillError(error)) {
                this.onerror?.(error as Error);
            }
        }
    }

    send(message: JSONRPCMessage): Promise<void> {
        return new Promise((resolve) => {
            if (!this._process?.stdin) {
                throw new Error("Not connected");
            }

            const json = serializeMessage(message);
            if (this._process.stdin.write(json)) {
                resolve();
            } else {
                this._process.stdin.once("drain", resolve);
            }
        });
    }
}
