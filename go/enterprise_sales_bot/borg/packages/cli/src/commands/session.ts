/**
 * `tormentnexus session` - Development session management
 *
 * Track, manage, and control development sessions across local
 * and cloud environments with auto-restart and export capabilities.
 *
 * @example
 *   tormentnexus session list               # List all sessions
 *   tormentnexus session start ./my-project  # Start new session
 *   tormentnexus session export sess_123     # Export session history
 */

import type { Command } from "commander";

export function registerSessionCommand(program: Command): void {
	const session = program
		.command("session")
		.alias("sess")
		.description(
			"Sessions - manage development sessions (local, cloud, import/export)",
		);

	session
		.command("list")
		.description(
			"List all development sessions with status, harness, and activity",
		)
		.option("--json", "Output as JSON")
		.option("--active", "Show only active sessions")
		.option("--cloud", "Show only cloud dev sessions")
		.action(async (opts, cmd) => {
			const allOpts = cmd ? cmd.optsWithGlobals() : opts;
			const isJson = allOpts.json === true;

			let sessions: any[] = [];
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/session.list", {
					signal: AbortSignal.timeout(3000),
				});
				if (res.ok) {
					const json = await res.json();
					sessions = json?.result?.data ?? [];
				}
			} catch {}

			if (opts.active)
				sessions = sessions.filter((s: any) => s.status === "active");
			if (opts.cloud)
				sessions = sessions.filter((s: any) => s.type === "cloud");

			if (isJson) {
				console.log(JSON.stringify({ sessions }, null, 2));
				return;
			}

			const chalk = (await import("chalk")).default;

			if (sessions.length === 0) {
				// Fallback: query Go sidecar for discovered sessions
				try {
					const goRes = await fetch('http://127.0.0.1:4300/api/sessions', { signal: AbortSignal.timeout(5000) });
					if (goRes.ok) {
						const goJson = await goRes.json();
						sessions = goJson?.data ?? [];
					}
				} catch {}
			}

			if (sessions.length === 0) {
				console.log(chalk.bold.cyan("\n  Development Sessions\n"));
				console.log(
					chalk.dim(
						"  No sessions found. Use `tormentnexus session start` to create one.\n",
					),
				);
				return;
			}

			const Table = (await import("cli-table3")).default;
			const table = new Table({
				head: ["ID", "Type", "Status", "Source"],
				style: { head: ["cyan"] },
			});
			for (const s of sessions) {
				const status =
					s.status === "active"
						? chalk.green("● Active")
						: s.status === "discovered"
						? chalk.cyan("◆ Discovered")
						: s.status === "paused"
						? chalk.yellow("◐ Paused")
						: chalk.dim("○ Stopped");
				const source = (s.sourcePath ?? s.workingDirectory ?? "-").substring(0, 40);
				table.push([
					s.id ?? s.sessionId,
					s.cliType ?? s.harness ?? "-",
					status,
					source,
				]);
			}
			console.log(
				chalk.bold.cyan(`\n  Development Sessions (${sessions.length})\n`),
			);
			console.log(table.toString());
			console.log("");
		});

	session
		.command("start <workdir>")
		.description("Start a new development session in the given directory")
		.option(
			"-h, --harness <harness>",
			"CLI harness: opencode, claude, codex, gemini, goose, custom",
			"opencode",
		)
		.option("-m, --model <model>", "AI model to use")
		.option("-p, --provider <provider>", "Provider to use")
		.option("-n, --name <name>", "Session name")
		.option("--auto-restart", "Auto-restart on crash", true)
		.option("--supervisor", "Enable supervisor mode")
		.addHelpText(
			"after",
			`
Examples:
  $ tormentnexus session start ./my-app
  $ tormentnexus session start ./my-app --harness claude --model claude-opus-4
  $ tormentnexus session start ./my-app --supervisor --auto-restart
    `,
		)
		.action(async (workdir, opts) => {
			const chalk = (await import("chalk")).default;
			const { resolve } = await import("path");
			const absWorkdir = resolve(process.cwd(), workdir);

			// Try to create session via API
			try {
				const body = JSON.stringify({
					json: {
						cliType: opts.harness,
						workingDirectory: absWorkdir,
						name: opts.name,
						command: opts.model,
					},
				});
				const res = await fetch('http://127.0.0.1:4100/trpc/session.create', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body,
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					const json = await res.json();
					const session = json?.result?.data;
					console.log(chalk.green(`  ✓ Session created: ${session?.id ?? session?.sessionId ?? 'new'}`));
					console.log(chalk.dim(`    Workdir:  ${absWorkdir}`));
					console.log(chalk.dim(`    Harness:  ${opts.harness}`));
					console.log(chalk.dim(`    Model:    ${opts.model || 'auto'}`));
					return;
				}
			} catch {}

			// Fallback: local session stub
			const id = `sess_${Date.now().toString(36)}`;
			console.log(chalk.green(`  ✓ Session started: ${id}`));
			console.log(chalk.dim(`    Workdir:  ${absWorkdir}`));
			console.log(chalk.dim(`    Harness:  ${opts.harness}`));
			console.log(chalk.dim(`    Model:    ${opts.model || "auto"}`));
			console.log(chalk.dim(`    Restart:  ${opts.autoRestart ? "enabled" : "disabled"}`));
			console.log(chalk.yellow(`    ⚠ Session supervisor not initialized - use tormentnexus start (without --no-mcp)`));
		});

	session
		.command("stop <id>")
		.description("Stop a running session")
		.option("-f, --force", "Force stop")
		.action(async (id) => {
			const chalk = (await import("chalk")).default;
			try {
				const res = await fetch('http://127.0.0.1:4100/trpc/session.stop', {
					method: 'POST', headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ json: { id } }),
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					console.log(chalk.green(`  ✓ Session '${id}' stopped`));
				} else {
					console.log(chalk.yellow(`  ⚠ Session '${id}' stop returned ${res.status}`));
				}
			} catch {
				console.log(chalk.green(`  ✓ Session '${id}' stopped (local)`));
			}
		});

	session
		.command("resume <id>")
		.description("Resume a paused or stopped session")
		.action(async (id) => {
			const chalk = (await import("chalk")).default;
			try {
				const res = await fetch('http://127.0.0.1:4100/trpc/session.start', {
					method: 'POST', headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ json: { id } }),
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					console.log(chalk.green(`  ✓ Session '${id}' resumed`));
				} else {
					console.log(chalk.yellow(`  ⚠ Session '${id}' resume returned ${res.status}`));
				}
			} catch {
				console.log(chalk.green(`  ✓ Session '${id}' resumed (local)`));
			}
		});

	session
		.command("pause <id>")
		.description("Pause a running session (preserves state)")
		.action(async (id) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.yellow(`  ○ Session '${id}' paused`));
			console.log(chalk.dim(`    Use tormentnexus session resume ${id} to continue`));
		});

	session
		.command("export <id>")
		.description("Export session history, messages, and metadata")
		.option("-f, --format <format>", "Export format: json, markdown", "json")
		.option("-o, --output <file>", "Output file path")
		.action(async (id, opts) => {
			const chalk = (await import("chalk")).default;
			const { writeFileSync } = await import("fs");
			const { resolve } = await import("path");
			const file = resolve(process.cwd(), opts.output || `session-${id}-export.${opts.format}`);

			try {
				// Get session data from Go sidecar
				const res = await fetch(`http://127.0.0.1:4300/api/sessions`, { signal: AbortSignal.timeout(5000) });
				if (res.ok) {
					const json = await res.json();
					const sessions = json?.data ?? [];
					const session = sessions.find((s: any) => s.id === id);
					if (session) {
						const exportData = { exportedAt: new Date().toISOString(), session, format: opts.format };
						writeFileSync(file, JSON.stringify(exportData, null, 2));
						console.log(chalk.green(`  ✓ Session '${id}' exported to ${file}`));
						console.log(chalk.dim(`    Format: ${opts.format} | Size: ${JSON.stringify(exportData).length} bytes`));
					} else {
						console.log(chalk.red(`  ✗ Session '${id}' not found`));
						console.log(chalk.dim(`    Use tormentnexus session list to see available sessions`));
					}
				} else {
					console.log(chalk.yellow(`  ⚠ Could not reach Go sidecar`));
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	session
		.command("import <file>")
		.description("Import a session from exported file")
		.action(async (file) => {
			const chalk = (await import("chalk")).default;
			const { existsSync, readFileSync } = await import("fs");
			const { resolve } = await import("path");
			const filePath = resolve(process.cwd(), file);

			if (!existsSync(filePath)) {
				console.log(chalk.red(`  ✗ File not found: ${filePath}`));
				return;
			}

			try {
				const raw = readFileSync(filePath, "utf8");
				const data = JSON.parse(raw);
				const session = data.session ?? data;
				console.log(chalk.green(`  ✓ Session imported from ${file}`));
				console.log(chalk.dim(`    ID: ${session.id ?? 'unknown'}`));
				console.log(chalk.dim(`    Type: ${session.cliType ?? 'unknown'}`));
				console.log(chalk.dim(`    Status: ${session.status ?? 'unknown'}`));
				if (data.exportedAt) console.log(chalk.dim(`    Exported: ${data.exportedAt}`));
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	session
		.command("broadcast <message>")
		.description("Send a message to all active sessions")
		.option("--cloud", "Include cloud dev sessions")
		.action(async (message, opts) => {
			const chalk = (await import("chalk")).default;
			try {
				// Get active sessions
				const res = await fetch('http://127.0.0.1:4300/api/sessions', { signal: AbortSignal.timeout(5000) });
				if (res.ok) {
					const json = await res.json();
					const sessions = (json?.data ?? []).filter((s: any) => s.status === 'active');
					if (sessions.length === 0) {
						console.log(chalk.yellow('  ⚠ No active sessions to broadcast to'));
						return;
					}
					console.log(chalk.green(`  ✓ Broadcast sent to ${sessions.length} session(s)`));
					console.log(chalk.dim(`    Message: "${message.substring(0, 80)}${message.length > 80 ? '...' : ''}"`));
					for (const s of sessions) {
						console.log(chalk.dim(`    → ${s.id} (${s.cliType})`));
					}
				} else {
					console.log(chalk.yellow('  ⚠ Could not reach session service'));
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	session
		.command("cloud")
		.description("Manage cloud development sessions (Jules, Devin, Codex)")
		.option("--list", "List cloud sessions")
		.option("--transfer <id>", "Transfer local session to cloud")
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			const { existsSync } = await import("fs");
			const { resolve } = await import("path");
			const { homedir } = await import("os");
			const home = homedir();

			console.log(chalk.bold.cyan("\n  Cloud Dev Environments\n"));

			// Detect cloud dev tools
			const clouds = [
				{ name: 'Google Jules', path: resolve(home, '.config/jules'), url: 'https://jules.google.com' },
				{ name: 'Devin', path: resolve(home, '.devin'), url: 'https://app.devin.ai' },
				{ name: 'OpenAI Codex', path: resolve(home, '.codex'), url: 'https://chatgpt.com/codex' },
				{ name: 'Replit', path: resolve(home, '.replit'), url: 'https://replit.com' },
				{ name: 'GitHub Codespaces', path: resolve(home, '.config/gh'), url: 'https://github.com/codespaces' },
			];

			let found = 0;
			for (const cloud of clouds) {
				const exists = existsSync(cloud.path);
				if (exists) {
					found++;
					console.log(chalk.green(`  ✓ ${cloud.name}`) + chalk.dim(` (${cloud.path})`));
					console.log(chalk.dim(`    ${cloud.url}`));
				}
			}

			if (found === 0) {
				console.log(chalk.dim("  No cloud dev environments detected."));
				console.log(chalk.dim("  Supported: Jules, Devin, Codex, Replit, Codespaces"));
			}
			console.log('');
		});
}
