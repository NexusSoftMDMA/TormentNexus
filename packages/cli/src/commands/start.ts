/**
 * `tormentnexus start` - Start the TORMENTNEXUS backend server
 *
 * Launches the TormentNexus core server with Express/tRPC/WebSocket/MCP endpoints.
 * The server provides the API backend for the WebUI dashboard, CLI commands,
 * and external MCP clients.
 *
 * @example
 *   tormentnexus start                    # Start on default port 3000
 *   tormentnexus start --port 8080        # Start on custom port
 *   tormentnexus start --no-mcp           # Start without MCP server
 *   tormentnexus start --config ./my.json # Use custom config file
 */

import {
	closeSync,
	mkdirSync,
	openSync,
	readFileSync,
	rmSync,
	writeFileSync,
} from "node:fs";
import net from "node:net";
import { homedir } from "node:os";
import { isAbsolute, join, resolve, sep } from "node:path";

import type { Command } from "commander";

/** Read version from nearest ancestor VERSION file */
async function getVersion(): Promise<string> {
	try {
		let dir = process.cwd();
		for (let i = 0; i < 20; i++) {
			try {
				return readFileSync(resolve(dir, "VERSION"), "utf8").trim();
			} catch {}
			const parent = resolve(dir, "..");
			if (parent === dir) break;
			dir = parent;
		}
	} catch {}
	return "dev";
}

export interface TormentNexusStartLockRecord {
	instanceId: string;
	pid: number;
	port: number;
	host: string;
	createdAt: string;
	/** Go-sidecar compatible fields */
	version?: string;
	startedAt?: string;
}

export interface TormentNexusStartLockHandle {
	port: number;
	lockPath: string;
	clearedStaleLock: boolean;
	reusedStalePort: boolean;
	release: () => Promise<void>;
	releaseSync: () => void;
}

export interface TormentNexusStartLifecycleHandlers {
	cleanup: () => void;
	handleSigint: () => void;
	handleSigterm: () => void;
	handleUncaughtException: (error: unknown) => void;
	handleUnhandledRejection: (reason: unknown) => void;
}

interface CreateLockLifecycleHandlersDeps {
	exit?: (code: number) => void;
	logError?: (message?: unknown, ...optionalParams: unknown[]) => void;
}

interface AcquireSingleInstanceLockOptions {
	dataDir: string;
	requestedPort: number;
	explicitPort: boolean;
	host: string;
}

interface AcquireSingleInstanceLockDeps {
	now?: () => Date;
	getPid?: () => number;
	isProcessRunning?: (pid: number) => boolean;
	isPortFree?: (port: number) => Promise<boolean>;
}

type CoreRuntimeModule = {
	startOrchestrator?: (options?: {
		host?: string;
		trpcPort?: number;
		startSupervisor?: boolean;
		startMcp?: boolean;
		autoDrive?: boolean;
	}) => Promise<{
		host: string;
		trpcPort: number;
		bridgePort: number | null;
	}>;
};

export function resolveDataDir(
	dataDir: string,
	homeDirectory: string = homedir(),
): string {
	if (dataDir === "~") {
		return homeDirectory;
	}

	if (
		dataDir.startsWith("~/") ||
		dataDir.startsWith("~\\") ||
		dataDir.startsWith(`~${sep}`)
	) {
		return resolve(homeDirectory, dataDir.slice(2));
	}

	return isAbsolute(dataDir) ? dataDir : resolve(dataDir);
}

export function isProcessRunning(pid: number): boolean {
	try {
		process.kill(pid, 0);
		return true;
	} catch (error) {
		return (error as NodeJS.ErrnoException).code === "EPERM";
	}
}

export async function isPortFree(port: number): Promise<boolean> {
	return await new Promise((resolvePort) => {
		const server = net.createServer();

		server.once("error", () => {
			resolvePort(false);
		});

		server.once("listening", () => {
			server.close(() => resolvePort(true));
		});

		server.listen(port);
	});
}

async function isGoAvailable(): Promise<boolean> {
	try {
		const { existsSync } = await import("fs");
		const { resolve } = await import("path");
		const goBin = resolve(process.cwd(), "tormentnexus.exe");
		if (existsSync(goBin)) return true;
		const altLocations = [
			resolve(process.cwd(), "go", "tormentnexus.exe"),
			resolve(process.cwd(), "bin", "tormentnexus.exe"),
		];
		for (const alt of altLocations) {
			if (existsSync(alt)) return true;
		}
		return false;
	} catch {
		return false;
	}
}

function readStartLock(lockPath: string): TormentNexusStartLockRecord | null {
	try {
		const parsed = JSON.parse(
			readFileSync(lockPath, "utf8"),
		) as Partial<TormentNexusStartLockRecord>;
		if (
			typeof parsed.instanceId !== "string" ||
			typeof parsed.pid !== "number" ||
			typeof parsed.port !== "number" ||
			typeof parsed.host !== "string" ||
			typeof parsed.createdAt !== "string"
		) {
			return null;
		}

		return parsed as TormentNexusStartLockRecord;
	} catch {
		return null;
	}
}

function writeStartLock(
	lockPath: string,
	record: TormentNexusStartLockRecord,
): void {
	const fd = openSync(lockPath, "wx");
	try {
		writeFileSync(fd, `${JSON.stringify(record, null, 2)}\n`, "utf8");
	} finally {
		closeSync(fd);
	}
}

export async function acquireSingleInstanceLock(
	options: AcquireSingleInstanceLockOptions,
	deps: AcquireSingleInstanceLockDeps = {},
): Promise<TormentNexusStartLockHandle> {
	const now = deps.now ?? (() => new Date());
	const getPid = deps.getPid ?? (() => process.pid);
	const checkProcessRunning = deps.isProcessRunning ?? isProcessRunning;
	const checkPortFree = deps.isPortFree ?? isPortFree;

	const resolvedDataDir = resolveDataDir(options.dataDir);
	const lockPath = join(resolvedDataDir, "lock");
	mkdirSync(resolvedDataDir, { recursive: true });

	const pid = getPid();
	const instanceId = `tormentnexus-${pid}-${now().getTime()}`;
	let selectedPort = options.requestedPort;
	let clearedStaleLock = false;
	let reusedStalePort = false;

	for (let attempt = 0; attempt < 2; attempt += 1) {
		try {
			writeStartLock(lockPath, {
				instanceId,
				pid,
				port: selectedPort,
				host: options.host,
				createdAt: now().toISOString(),
				version: process.env.TORMENTNEXUS_VERSION || "1.0.0-alpha.60",
				startedAt: now().toISOString(),
			});

			const releaseSync = () => {
				const current = readStartLock(lockPath);
				if (current?.instanceId === instanceId) {
					rmSync(lockPath, { force: true });
				}
			};

			return {
				port: selectedPort,
				lockPath,
				clearedStaleLock,
				reusedStalePort,
				release: async () => {
					releaseSync();
				},
				releaseSync,
			};
		} catch (error) {
			const fsError = error as NodeJS.ErrnoException;
			if (fsError.code !== "EEXIST") {
				throw error;
			}

			const existingLock = readStartLock(lockPath);
			if (existingLock && checkProcessRunning(existingLock.pid)) {
				const lockedPortIsFree =
					existingLock.port > 0
						? await checkPortFree(existingLock.port)
						: false;

				if (!lockedPortIsFree) {
					throw new Error(
						`TormentNexus is already running (PID ${existingLock.pid}) on port ${existingLock.port}. ` +
							`Stop that process before starting another instance, or remove ${lockPath} if it is incorrect.`,
					);
				}
			}

			clearedStaleLock = true;
			if (
				!options.explicitPort &&
				existingLock?.port &&
				existingLock.port > 0
			) {
				const stalePortIsFree = await checkPortFree(existingLock.port);
				if (stalePortIsFree) {
					selectedPort = existingLock.port;
					reusedStalePort = true;
				}
			}

			rmSync(lockPath, { force: true });
		}
	}

	throw new Error(`Unable to acquire TormentNexus startup lock at ${lockPath}`);
}

export async function startCoreRuntime(
	options: {
		host: string;
		port: number;
		mcp: boolean;
		supervisor?: boolean;
		autoDrive?: boolean;
	},
	loadCore: () => Promise<CoreRuntimeModule> = async () =>
		await import("@tormentnexus/core"),
) {
	const core = await loadCore();

	if (typeof core.startOrchestrator !== "function") {
		throw new Error("Core orchestrator entrypoint is unavailable");
	}

	return await core.startOrchestrator({
		host: options.host,
		trpcPort: options.port,
		startMcp: options.mcp,
		startSupervisor: options.supervisor ?? false,
		autoDrive: options.autoDrive ?? false,
	});
}

export function createLockLifecycleHandlers(
	lockHandle: TormentNexusStartLockHandle,
	deps: CreateLockLifecycleHandlersDeps = {},
): TormentNexusStartLifecycleHandlers {
	const exit = deps.exit ?? ((code: number) => process.exit(code));
	void deps.logError;

	const cleanup = () => {
		lockHandle.releaseSync();
	};

	return {
		cleanup,
		handleSigint: () => {
			cleanup();
			exit(130);
		},
		handleSigterm: () => {
			cleanup();
			exit(143);
		},
		async handleUncaughtException(error: unknown) {
			this.cleanup();
			const chalk = (await import("chalk")).default;
			const msg =
				error instanceof Error ? (error.stack ?? error.message) : String(error);
			console.error(chalk.red(`\n  ✗ Uncaught Exception: ${msg}`));
			process.exit(1);
		},
		async handleUnhandledRejection(reason: unknown) {
			this.cleanup();
			const chalk = (await import("chalk")).default;
			const msg =
				reason instanceof Error
					? (reason.stack ?? reason.message)
					: String(reason);
			console.error(chalk.red(`\n  ✗ Unhandled Rejection: ${msg}`));
			process.exit(1);
		},
	};
}

export function registerStartCommand(program: Command): void {
	program
		.command("start")
		.description(
			"Start the TormentNexus TORMENTNEXUS backend server (Express/tRPC/WebSocket/MCP)",
		)
		.option("-p, --port <number>", "tRPC control-plane port", "4100")
		.option("-H, --host <address>", "Server host address", "0.0.0.0")
		.option("--no-mcp", "Disable the MCP server endpoint")
		.option("--supervisor", "Enable TormentNexus supervisor startup")
		.option("--auto-drive", "Enable Director auto-drive after startup")
		.option("--no-dashboard", "Disable serving the WebUI dashboard")
		.option("-c, --config <path>", "Path to config file")
		.option(
			"-d, --data-dir <path>",
			"Data directory for TormentNexus state",
			"~/.tormentnexus",
		)
		.option("--daemon", "Run as background daemon")
		.option(
			"--runtime <mode>",
			"Runtime mode: auto (prefer Go), go (Go-only), node (TS-primary)",
			"auto",
		)
		.addHelpText(
			"after",
			`
Examples:
  $ tormentnexus start                     Start with defaults (tRPC on port 4100)
  $ tormentnexus start -p 8080             Start on port 8080
  $ tormentnexus start --no-mcp            Start without MCP server
  $ tormentnexus start --auto-drive        Start the Director after boot completes
  $ tormentnexus start --daemon            Run as background service
  $ tormentnexus start --host 127.0.0.1    Bind to localhost only
  $ tormentnexus start --runtime go        Go-primary startup (Go control plane + optional TS)
  $ tormentnexus start --runtime node      TS-primary startup (legacy behavior)
    `,
		)
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			const requestedPort = parseInt(opts.port, 10);
			const host = opts.host;
			const explicitPort =
				process.argv.includes("--port") || process.argv.includes("-p");
			let lockHandle: TormentNexusStartLockHandle | null = null;

			console.log(
				chalk.bold.cyan(
					`\n  ⬡ TormentNexus TORMENTNEXUS v${await getVersion()}`,
				),
			);
			console.log(chalk.dim("  The Neural Operating System\n"));

			try {
				lockHandle = await acquireSingleInstanceLock({
					dataDir: opts.dataDir,
					requestedPort,
					explicitPort,
					host,
				});

				const port = lockHandle.port;
				const lifecycle = createLockLifecycleHandlers(lockHandle);

				console.log(chalk.yellow("  Starting server..."));
				console.log(chalk.dim(`  Host: ${host}:${port}`));
				console.log(chalk.dim(`  MCP:  ${opts.mcp ? "enabled" : "disabled"}`));
				console.log(
					chalk.dim(
						`  Supervisor: ${opts.supervisor ? "enabled" : "disabled"}`,
					),
				);
				console.log(
					chalk.dim(`  Auto-Drive: ${opts.autoDrive ? "enabled" : "disabled"}`),
				);
				console.log(
					chalk.dim(`  Dashboard: ${opts.dashboard ? "enabled" : "disabled"}`),
				);
				console.log(chalk.dim(`  Lock: ${lockHandle.lockPath}`));
				if (lockHandle.clearedStaleLock) {
					console.log(
						chalk.yellow(
							`  ↺ Cleared stale TormentNexus lock${lockHandle.reusedStalePort ? ` and reused port ${port}` : ""}`,
						),
					);
				}
				console.log("");

				const runtime = await startCoreRuntime({
					host,
					port,
					mcp: Boolean(opts.mcp),
					supervisor: Boolean(opts.supervisor),
					autoDrive: Boolean(opts.autoDrive),
				});

				console.log(chalk.dim("  Core loaded: orchestrator started"));

				// Detect available providers from environment
				const providerEnvMap: Record<string, string> = {
					OPENAI_API_KEY: "OpenAI",
					ANTHROPIC_API_KEY: "Anthropic",
					GOOGLE_API_KEY: "Google",
					GEMINI_API_KEY: "Gemini",
					XAI_API_KEY: "xAI",
					DEEPSEEK_API_KEY: "DeepSeek",
					MISTRAL_API_KEY: "Mistral",
					OPENROUTER_API_KEY: "OpenRouter",
				};
				const detectedProviders = Object.entries(providerEnvMap)
					.filter(([env]) => process.env[env])
					.map(([, name]) => name);
				if (detectedProviders.length > 0) {
					console.log(
						chalk.dim(`  Providers: ${detectedProviders.join(", ")}`),
					);
				}

				console.log(
					chalk.green(
						`  ✓ tRPC control plane running at http://${runtime.host}:${runtime.trpcPort}/trpc`,
					),
				);
				if (opts.mcp) {
					console.log(
						chalk.green(
							`  ✓ MCP bridge ready at ws://127.0.0.1:${runtime.bridgePort ?? 3001} (+ HTTP health on /health)`,
						),
					);
				}
				if (opts.supervisor) {
					console.log(
						chalk.green("  ✓ Supervisor startup enabled for this run"),
					);
				}
				if (opts.dashboard) {
					console.log(
						chalk.green(
							"  ✓ Dashboard proxy can now reach Core via the web app runtime",
						),
					);
				}
				console.log(chalk.dim("\n  Press Ctrl+C to stop\n"));

				// Runtime selection: Go-primary vs TS-primary
				const runtimeMode = (opts.runtime || "auto").toLowerCase();
				const wantGo =
					runtimeMode === "go" ||
					(runtimeMode === "auto" && (await isGoAvailable()));

				if (wantGo) {
					// Launch Go primary control plane
					try {
						const { spawn } = await import("child_process");
						const { resolve } = await import("path");
						const goBin = resolve(process.cwd(), "tormentnexus.exe");
						const goPort = 7778;
						const goProc = spawn(
							goBin,
							["serve", "--port", String(goPort), "--host", host],
							{
								stdio: "ignore",
								detached: true,
								env: { ...process.env, TORMENTNEXUS_WORKSPACE: process.cwd() },
							},
						);
						goProc.unref();
						console.log(
							chalk.green(`  ✓ Go control plane launched on port ${goPort}`),
						);

						// Start TS as compatibility supplement (optional, always starts for now)
						if (runtimeMode !== "go") {
							console.log(
								chalk.dim("  TS compatibility layer: started alongside Go"),
							);
						}
					} catch (e: any) {
						console.log(
							chalk.yellow(
								"  Go control plane: not available, falling back to TS-primary",
							),
						);
					}
				} else {
					// TS-primary mode (legacy): Go sidecar is optional
					console.log(
						chalk.dim(
							`  Runtime: TS-primary${runtimeMode === "auto" ? " (Go binary not found)" : ""}`,
						),
					);
					try {
						const { existsSync } = await import("fs");
						const { resolve } = await import("path");
						const goBin = resolve(process.cwd(), "tormentnexus.exe");
						if (existsSync(goBin)) {
							const { spawn } = await import("child_process");
							const goProc = spawn(goBin, ["serve", "--port", "4300"], {
								stdio: "ignore",
								detached: true,
								env: { ...process.env, TORMENTNEXUS_WORKSPACE: process.cwd() },
							});
							goProc.unref();
							console.log(chalk.green("  ✓ Go sidecar launched on port 4300"));
						}
					} catch (e: any) {
						console.log(chalk.dim("  Go sidecar: not available (optional)"));
					}
				}

				process.once("exit", lifecycle.cleanup);
				process.once("SIGINT", lifecycle.handleSigint);
				process.once("SIGTERM", lifecycle.handleSigterm);
				process.once("uncaughtException", lifecycle.handleUncaughtException);
				process.once("unhandledRejection", lifecycle.handleUnhandledRejection);
			} catch (err: unknown) {
				lockHandle?.releaseSync();
				const msg =
					err instanceof Error ? `${err.message}\n${err.stack}` : String(err);
				console.error(chalk.red(`  ✗ Failed to start: ${msg}`));
				process.exit(1);
			}
		});
}
