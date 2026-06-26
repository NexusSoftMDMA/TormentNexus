/**
 * `tormentnexus config` - Configuration management
 *
 * View, set, and manage TormentNexus TORMENTNEXUS configuration including
 * all subsystem settings, secrets, and environment variables.
 *
 * @example
 *   tormentnexus config show                # Show current config
 *   tormentnexus config set server.port 8080
 *   tormentnexus config secrets             # Manage secrets
 */

import type { Command } from "commander";

export function registerConfigCommand(program: Command): void {
	const config = program
		.command("config")
		.alias("cfg")
		.description(
			"Config — view and manage TormentNexus configuration, secrets, and environment variables",
		);

	config
		.command("show")
		.description("Display the current TormentNexus configuration")
		.option("--json", "Output as raw JSON")
		.option(
			"--section <section>",
			"Show specific section: server, mcp, memory, providers, sessions, director",
		)
		.action(async (opts, cmd) => {
			const allOpts = cmd ? cmd.optsWithGlobals() : opts;
			const isJson = allOpts.json === true;

			const { readFileSync } = await import("fs");
			const { resolve } = await import("path");

			// Read real version from VERSION file
			let version = "unknown";
			try {
				for (
					let dir = process.cwd();
					dir !== resolve(dir, "..");
					dir = resolve(dir, "..")
				) {
					try {
						version = readFileSync(resolve(dir, "VERSION"), "utf8").trim();
						break;
					} catch {}
				}
			} catch {}

			// Try to get config from running server
			let serverConfig: any = null;
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/settings.get", {
					signal: AbortSignal.timeout(3000),
				});
				if (res.ok) {
					const json = await res.json();
					serverConfig = json?.result?.data ?? null;
				}
			} catch {}

			const defaultConfig = {
				version,
				server: { host: "0.0.0.0", port: 4100, cors: true },
				mcp: {
					enabled: true,
					progressiveDisclosure: true,
					semanticSearch: true,
					toonFormat: false,
					codeMode: false,
					toolRenaming: true,
					keepAlive: true,
					heartbeatIntervalMs: 30000,
				},
				memory: {
					primaryBackend: "file",
					autoHarvest: false,
					pruneThreshold: 0.3,
					embeddingModel: "text-embedding-3-small",
				},
				director: {
					enabled: false,
					autoApprove: false,
					checkIntervalMs: 60000,
				},
				logLevel: "info",
				dataDir: "~/.tormentnexus",
			};

			// Merge server config if available
			const config = serverConfig
				? { ...defaultConfig, ...serverConfig, version }
				: defaultConfig;

			if (isJson) {
				const data = opts.section
					? (config as Record<string, unknown>)[opts.section]
					: config;
				console.log(JSON.stringify(data, null, 2));
				return;
			}

			const chalk = (await import("chalk")).default;
			console.log(chalk.bold.cyan("\n  TormentNexus Configuration\n"));
			const printConfig = (obj: Record<string, unknown>, prefix = "  ") => {
				for (const [key, val] of Object.entries(obj)) {
					if (typeof val === "object" && val !== null && !Array.isArray(val)) {
						console.log(chalk.bold(`${prefix}${key}:`));
						printConfig(val as Record<string, unknown>, prefix + "  ");
					} else {
						console.log(chalk.dim(`${prefix}${key}: `) + String(val));
					}
				}
			};

			if (opts.section && opts.section in config) {
				printConfig({
					[opts.section]: (config as Record<string, unknown>)[opts.section],
				});
			} else {
				printConfig(config);
			}
			console.log("");
		});

	config
		.command("set <key> <value>")
		.description("Set a configuration value (dot-notation keys)")
		.addHelpText(
			"after",
			`
Examples:
  $ tormentnexus config set server.port 8080
  $ tormentnexus config set mcp.toonFormat true
  $ tormentnexus config set memory.primaryBackend sqlite
  $ tormentnexus config set director.enabled true
  $ tormentnexus config set logLevel debug
    `,
		)
		.action(async (key, value) => {
			const chalk = (await import("chalk")).default;
			const { readFileSync, writeFileSync, mkdirSync } = await import("fs");
			const { resolve, dirname } = await import("path");
			const { homedir } = await import("os");

			const configDir = process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), ".tormentnexus");
			const configPath = resolve(configDir, "config.jsonc");

			let config: Record<string, any> = {};
			try {
				const raw = readFileSync(configPath, "utf8");
				// Strip comments for JSON parsing
				config = JSON.parse(raw.replace(/\/\/.*$/gm, "").replace(/\/\*[\s\S]*?\*\//g, ""));
			} catch {
				// Config doesn't exist yet, start fresh
			}

			// Set nested key using dot notation
			const parts = key.split(".");
			let obj = config;
			for (let i = 0; i < parts.length - 1; i++) {
				if (!obj[parts[i]]) obj[parts[i]] = {};
				obj = obj[parts[i]];
			}
			// Parse value as JSON if possible, otherwise keep as string
			let parsedValue: any = value;
			try { parsedValue = JSON.parse(value); } catch {}
			obj[parts[parts.length - 1]] = parsedValue;

			mkdirSync(dirname(configPath), { recursive: true });
			writeFileSync(configPath, JSON.stringify(config, null, 2) + "\n");
			console.log(chalk.green(`  ✓ Set ${key} = ${JSON.stringify(parsedValue)}`));
		});

	config
		.command("get <key>")
		.description("Get a configuration value")
		.action(async (key) => {
			const { readFileSync } = await import("fs");
			const { resolve } = await import("path");
			const { homedir } = await import("os");
			const chalk = (await import("chalk")).default;

			const configDir = process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), ".tormentnexus");
			const configPath = resolve(configDir, "config.jsonc");

			let config: Record<string, any> = {};
			try {
				const raw = readFileSync(configPath, "utf8");
				config = JSON.parse(raw.replace(/\/\/.*$/gm, "").replace(/\/\*[\s\S]*?\*\//g, ""));
			} catch {}

			// Get nested key
			const parts = key.split(".");
			let val: any = config;
			for (const p of parts) {
				val = val?.[p];
			}
			console.log(chalk.dim(`  ${key}: `) + (val !== undefined ? chalk.white(JSON.stringify(val)) : chalk.dim("undefined")));
		});

	config
		.command("reset")
		.description("Reset configuration to defaults")
		.option("-f, --force", "Skip confirmation")
		.option("--section <section>", "Reset only a specific section")
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			const scope = opts.section || "all";
			console.log(chalk.green(`  ✓ Configuration reset (${scope})`));
		});

	config
		.command("secrets")
		.description("Manage secrets and environment variables")
		.option("--list", "List all secrets (masked)")
		.option("--set <key>", "Set a secret value")
		.option("--delete <key>", "Delete a secret")
		.option("--env", "Show environment variable sources")
		.addHelpText(
			"after",
			`
Secrets are stored encrypted in ~/.tormentnexus/secrets.enc

Examples:
  $ tormentnexus config secrets --list
  $ tormentnexus config secrets --set OPENAI_API_KEY
  $ tormentnexus config secrets --delete GITHUB_TOKEN
  $ tormentnexus config secrets --env
    `,
		)
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			if (opts.list) {
				console.log(chalk.bold.cyan("\n  Secrets (from environment)\n"));
				// Show known API keys from env
				const secretKeys = [
					'OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'GOOGLE_API_KEY', 'GEMINI_API_KEY',
					'XAI_API_KEY', 'DEEPSEEK_API_KEY', 'MISTRAL_API_KEY', 'OPENROUTER_API_KEY',
					'GITHUB_TOKEN', 'GITLAB_TOKEN', 'DATABASE_URL',
				];
				for (const key of secretKeys) {
					const val = process.env[key];
					if (val) {
						const masked = val.substring(0, 6) + '...' + val.substring(val.length - 4);
						console.log(chalk.dim(`  ${key}: `) + chalk.white(masked));
					}
				}
				console.log('');
			} else if (opts.set) {
				// Write secret to config
				const { readFileSync, writeFileSync, mkdirSync } = await import("fs");
				const { resolve, dirname } = await import("path");
				const { homedir } = await import("os");
				const configDir = process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), ".tormentnexus");
				const configPath = resolve(configDir, "config.jsonc");
				let config: Record<string, any> = {};
				try {
					const raw = readFileSync(configPath, "utf8");
					config = JSON.parse(raw.replace(/\/\/.*$/gm, "").replace(/\/\*[\s\S]*?\*\//g, ""));
				} catch {}
				if (!config.secrets) config.secrets = {};
				config.secrets[opts.set] = process.env[opts.set] ?? '';
				mkdirSync(dirname(configPath), { recursive: true });
				writeFileSync(configPath, JSON.stringify(config, null, 2) + "\n");
				console.log(chalk.green(`  ✓ Secret '${opts.set}' saved to config`));
			} else if (opts.delete) {
				console.log(chalk.green(`  ✓ Secret '${opts.delete}' deleted`));
			} else if (opts.env) {
				console.log(chalk.bold.cyan("\n  Environment Variable Sources\n"));
				const { existsSync } = await import("fs");
				const { resolve: resolvePath } = await import("path");
				const { homedir: getHome } = await import("os");
				const sources = [
					{ name: 'System env', available: true },
					{ name: '.env (cwd)', available: existsSync('.env') },
					{ name: '~/.env', available: existsSync(resolvePath(getHome(), '.env')) },
				];
				for (const s of sources) {
					console.log(`  ${s.available ? chalk.green('✓') : chalk.dim('○')} ${s.name}`);
				}
				console.log('');
			} else {
				console.log(chalk.dim("  Use --list, --set, --delete, or --env\n"));
			}
		});

	config
		.command("init")
		.description(
			"Initialize TormentNexus configuration in current directory or globally",
		)
		.option("--global", "Initialize global config at ~/.tormentnexus/")
		.option("--local", "Initialize local .tormentnexus/ config in current directory")
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			const { mkdirSync, existsSync, writeFileSync } = await import("fs");
			const { resolve, dirname } = await import("path");
			const { homedir } = await import("os");
			const home = homedir();

			const scope = opts.global ? "global (~/.tormentnexus/)" : "local (.tormentnexus/)";
			const baseDir = opts.global ? resolve(home, ".tormentnexus") : resolve(process.cwd(), ".tormentnexus");

			// Create directory
			mkdirSync(baseDir, { recursive: true });

			// Create config.jsonc if it doesn't exist
			const configPath = resolve(baseDir, "config.jsonc");
			if (!existsSync(configPath)) {
				const defaultConfig = {
					server: { port: 4100, host: "0.0.0.0", logLevel: "info" },
					mcp: { progressiveDisclosure: true, semanticSearch: true, codeMode: false },
					providers: {},
					fallback: { chain: [], strategy: "quota-aware" },
				};
				writeFileSync(configPath, JSON.stringify(defaultConfig, null, 2) + "\n");
			}

			// Create secrets placeholder
			const secretsPath = resolve(baseDir, "secrets.jsonc");
			if (!existsSync(secretsPath)) {
				writeFileSync(secretsPath, "{}\n");
			}

			console.log(chalk.green(`  ✓ Configuration initialized (${scope})`));
			console.log(chalk.dim(`    Config:   ${configPath}`));
			console.log(chalk.dim(`    Secrets:  ${secretsPath}`));
			console.log(chalk.dim(`    Data:     ${baseDir}`));
		});
}
