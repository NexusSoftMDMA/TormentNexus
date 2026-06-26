/**
 * `tormentnexus mcp` - MCP Router management commands
 *
 * Comprehensive MCP server lifecycle management, tool browsing,
 * traffic inspection, config sync, and directory access.
 *
 * @example
 *   tormentnexus mcp list                    # List all MCP servers
 *   tormentnexus mcp start filesystem        # Start a specific server
 *   tormentnexus mcp traffic                 # Watch live MCP traffic
 *   tormentnexus mcp install @anthropic/mcp  # Install from directory
 *   tormentnexus mcp sync                    # Sync configs to all AI tools
 */

import type { Command } from "commander";

export function registerMcpCommand(program: Command): void {
	const mcp = program
		.command("mcp")
		.description(
			"MCP Router — manage servers, tools, traffic, config, and directory",
		);

	mcp
		.command("list")
		.description(
			"List all configured MCP servers with status, transport, latency, and tool count",
		)
		.option("--json", "Output as JSON")
		.option("--running", "Show only running servers")
		.option("--namespace <ns>", "Filter by namespace")
		.action(async (opts, cmd) => {
			// Merge global opts (program-level --json) with subcommand opts
			const allOpts = cmd ? cmd.optsWithGlobals() : opts;
			const isJson = allOpts.json === true;

			// Query the running server for real data
			let servers: any[] = [];
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.listServers", {
					signal: AbortSignal.timeout(3000),
				});
				if (res.ok) {
					const json = await res.json();
					servers = json?.result?.data ?? [];
				}
			} catch (e) {
				void e;
			}

			// Load fleet PIDs for status enrichment and filtering
			const fleetAlive = new Map<string, number>();
			try {
				const fs = await import("fs");
				const path = await import("path");
				const pidDir = path.join(
					process.env.HOME ?? "",
					".tormentnexus",
					"mcp-pids",
				);
				const pidFiles = fs
					.readdirSync(pidDir)
					.filter((f) => f.endsWith(".pid"));
				for (const pf of pidFiles) {
					try {
						const name = pf.replace(".pid", "");
						const pid = parseInt(
							fs.readFileSync(path.join(pidDir, pf), "utf8").trim(),
						);
						process.kill(pid, 0);
						fleetAlive.set(name, pid);
					} catch (e) {
						void e;
					}
				}
			} catch (e) {
				void e;
			}

			// Filter
			if (opts.running)
				servers = servers.filter(
					(s: any) => s.runtimeConnected || fleetAlive.has(s.name),
				);
			if (opts.namespace)
				servers = servers.filter((s: any) =>
					(s.tags ?? []).includes(opts.namespace),
				);

			if (isJson) {
				console.log(JSON.stringify({ servers }, null, 2));
				return;
			}

			const chalk = (await import("chalk")).default;
			const Table = (await import("cli-table3")).default;

			if (servers.length === 0) {
				console.log(chalk.bold.cyan("\n  MCP Servers\n"));
				console.log(
					chalk.dim(
						"  No servers found." +
							(servers.length === 0
								? " Is the server running? Use `tormentnexus start`."
								: "") +
							"\n",
					),
				);
				return;
			}

			// Fleet PIDs already loaded above

			const table = new Table({
				head: ["Name", "Status", "Tools", "PID", "Tags"],
				style: { head: ["cyan"] },
			});

			for (const s of servers) {
				let status: string;
				if (s.runtimeConnected) {
					status = chalk.green("● Connected");
				} else if (fleetAlive.has(s.name)) {
					status = chalk.green("● Spawned");
				} else if (s.status === "cached") {
					status = chalk.yellow("◐ Cached");
				} else {
					status = chalk.dim("○ Stopped");
				}
				const tools = String(s.toolCount ?? s.advertisedToolCount ?? 0);
				const pid = fleetAlive.has(s.name)
					? String(fleetAlive.get(s.name))
					: "—";
				const tags = (s.tags ?? []).slice(0, 3).join(", ");
				table.push([s.displayName ?? s.name, status, tools, pid, tags]);
			}

			console.log(chalk.bold.cyan(`\n  MCP Servers (${servers.length})\n`));
			console.log(table.toString());
			console.log(
				chalk.dim(
					`\n  ${servers.filter((s: any) => s.runtimeConnected).length} connected · ${fleetAlive.size} spawned · ${servers.reduce((a: number, s: any) => a + (s.toolCount ?? 0), 0)} tools\n`,
				),
			);
		});

	mcp
		.command("start <name>")
		.description("Start an MCP server by name")
		.action(async (name) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.yellow(`  Starting MCP server: ${name}...`));
			let started = false;

			// Try tRPC first
			try {
				const res = await fetch(
					"http://127.0.0.1:4100/trpc/mcp.connectServer",
					{
						method: "POST",
						headers: { "Content-Type": "application/json" },
						body: JSON.stringify({ json: { name } }),
						signal: AbortSignal.timeout(15000),
					},
				);
				if (res.ok) {
					started = true;
					console.log(chalk.green(`  ✓ Server '${name}' started`));
				}
			} catch (e) {
				void e;
			}

			// Fallback: spawn directly from config
			if (!started) {
				try {
					const listRes = await fetch(
						"http://127.0.0.1:4100/trpc/mcp.listServers",
						{ signal: AbortSignal.timeout(5000) },
					);
					if (listRes.ok) {
						const servers = (await listRes.json())?.result?.data ?? [];
						const server = servers.find((s: any) => s.name === name);
						if (server?.config?.command) {
							const { spawn } = await import("child_process");
							const cmd = server.config.command;
							const args = server.config.args ?? [];
							const env = { ...process.env, ...(server.config.env ?? {}) };

							console.log(chalk.dim(`    Spawning: ${cmd} ${args.join(" ")}`));
							const proc = spawn(cmd, args, {
								stdio: "ignore",
								detached: true,
								env,
								shell: true,
							});
							proc.unref();

							// Write PID file for fleet tracking
							try {
								const fs = await import("fs");
								const path = await import("path");
								const pidDir = path.join(
									process.env.HOME ?? "",
									".tormentnexus",
									"mcp-pids",
								);
								fs.mkdirSync(pidDir, { recursive: true });
								fs.writeFileSync(
									path.join(pidDir, `${name}.pid`),
									String(proc.pid),
								);
							} catch (e) {
								void e;
							}
							await new Promise((r) => setTimeout(r, 1000));
							try {
								process.kill(proc.pid!, 0);
								started = true;
							} catch {
								started = false;
							}

							if (started) {
								console.log(
									chalk.green(`  ✓ Server '${name}' spawned (PID ${proc.pid})`),
								);
								console.log(chalk.dim(`    ${cmd} ${args.join(" ")}`));
							} else {
								console.log(
									chalk.red(`  ✗ Server '${name}' exited immediately`),
								);
							}
						} else {
							console.log(
								chalk.red(`  ✗ Server '${name}' not found or has no command`),
							);
							console.log(
								chalk.dim(
									`    Use tormentnexus mcp list to see available servers`,
								),
							);
						}
					}
				} catch (e: any) {
					console.log(chalk.red(`  ✗ Error: ${e.message}`));
				}
			}
		});

	mcp
		.command("stop <name>")
		.description("Stop a running MCP server")
		.action(async (name) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.yellow(`  Stopping MCP server: ${name}...`));
			let stopped = false;

			// Try tRPC first
			try {
				const res = await fetch(
					"http://127.0.0.1:4100/trpc/mcp.disconnectServer",
					{
						method: "POST",
						headers: { "Content-Type": "application/json" },
						body: JSON.stringify({ json: { name } }),
						signal: AbortSignal.timeout(10000),
					},
				);
				if (res.ok) {
					stopped = true;
				}
			} catch (e) {
				void e;
			}

			// Fallback: kill via PID file
			if (!stopped) {
				try {
					const fs = await import("fs");
					const path = await import("path");
					const pidFile = path.join(
						process.env.HOME ?? "",
						".tormentnexus",
						"mcp-pids",
						`${name}.pid`,
					);
					const pid = parseInt(fs.readFileSync(pidFile, "utf8").trim());
					process.kill(pid, "SIGTERM");
					stopped = true;
					// Clean up PID file
					try {
						fs.unlinkSync(pidFile);
					} catch (e) {
						void e;
					}
				} catch (e) {
					void e;
				}
			}

			if (stopped) {
				console.log(chalk.green(`  ✓ Server '${name}' stopped`));
			} else {
				console.log(
					chalk.red(`  ✗ Could not stop '${name}' (not found or not running)`),
				);
			}
		});

	mcp
		.command("restart <name>")
		.description("Restart an MCP server (stop + start)")
		.action(async (name) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.yellow(`  Restarting MCP server: ${name}...`));

			// Stop via PID file kill
			try {
				const fs = await import("fs");
				const path = await import("path");
				const pidFile = path.join(
					process.env.HOME ?? "",
					".tormentnexus",
					"mcp-pids",
					`${name}.pid`,
				);
				const pid = parseInt(fs.readFileSync(pidFile, "utf8").trim());
				process.kill(pid, "SIGTERM");
				try {
					fs.unlinkSync(pidFile);
				} catch (e) {
					void e;
				}
				await new Promise((r) => setTimeout(r, 500));
			} catch (e) {
				void e;
			}

			// Get server config and spawn
			try {
				const listRes = await fetch(
					"http://127.0.0.1:4100/trpc/mcp.listServers",
					{ signal: AbortSignal.timeout(5000) },
				);
				if (listRes.ok) {
					const servers = (await listRes.json())?.result?.data ?? [];
					const server = servers.find((s: any) => s.name === name);
					if (server?.config?.command) {
						const { spawn } = await import("child_process");
						const proc = spawn(
							server.config.command,
							server.config.args ?? [],
							{
								stdio: "ignore",
								detached: true,
								env: { ...process.env, ...(server.config.env ?? {}) },
								shell: true,
							},
						);
						proc.unref();
						// Write PID file
						try {
							const fs = await import("fs");
							const path = await import("path");
							const pidDir = path.join(
								process.env.HOME ?? "",
								".tormentnexus",
								"mcp-pids",
							);
							fs.mkdirSync(pidDir, { recursive: true });
							fs.writeFileSync(
								path.join(pidDir, `${name}.pid`),
								String(proc.pid),
							);
						} catch (e) {
							void e;
						}
						console.log(
							chalk.green(`  ✓ Server '${name}' restarted (PID ${proc.pid})`),
						);
					} else {
						console.log(
							chalk.red(`  ✗ Server '${name}' not found or no command`),
						);
					}
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	mcp
		.command("add <name> <command>")
		.description("Add a new MCP server to the router configuration")
		.option(
			"-t, --transport <type>",
			"Transport type: stdio, sse, streamable-http",
			"stdio",
		)
		.option("-n, --namespace <ns>", "Server namespace", "default")
		.option("--args <args...>", "Command arguments")
		.option("--env <vars...>", "Environment variables (KEY=VALUE)")
		.option("--auto-start", "Auto-start on TormentNexus launch", true)
		.addHelpText(
			"after",
			`
Examples:
  $ tormentnexus mcp add filesystem npx -- -y @modelcontextprotocol/server-filesystem /home
  $ tormentnexus mcp add github npx -- -y @modelcontextprotocol/server-github --env GITHUB_TOKEN=xxx
  $ tormentnexus mcp add remote-api http://localhost:8080/mcp -t streamable-http
    `,
		)
		.action(async (name, command, opts) => {
			const chalk = (await import("chalk")).default;
			try {
				const res = await fetch(
					"http://127.0.0.1:4100/trpc/mcpServers.create",
					{
						method: "POST",
						headers: { "Content-Type": "application/json" },
						body: JSON.stringify({
							json: {
								name,
								command,
								transport: opts.transport,
								namespace: opts.namespace,
								args: opts.args,
								env: opts.env?.reduce((acc: any, v: string) => {
									const [k, ...r] = v.split("=");
									acc[k] = r.join("=");
									return acc;
								}, {}),
								autoStart: opts.autoStart,
							},
						}),
						signal: AbortSignal.timeout(5000),
					},
				);
				if (res.ok) {
					console.log(
						chalk.green(`  ✓ Added MCP server '${name}' (${opts.transport})`),
					);
					console.log(chalk.dim(`    Command: ${command}`));
					console.log(chalk.dim(`    Namespace: ${opts.namespace}`));
				} else {
					const json = await res.json().catch(() => ({}));
					console.log(
						chalk.red(`  ✗ Failed: ${json.error?.message ?? res.statusText}`),
					);
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
				console.log(chalk.dim(`    Is the server running?`));
			}
		});

	mcp
		.command("remove <name>")
		.description("Remove an MCP server from configuration")
		.option("-f, --force", "Skip confirmation")
		.action(async (name, opts) => {
			const chalk = (await import("chalk")).default;
			try {
				const res = await fetch(
					"http://127.0.0.1:4100/trpc/mcpServers.delete",
					{
						method: "POST",
						headers: { "Content-Type": "application/json" },
						body: JSON.stringify({ json: { name } }),
						signal: AbortSignal.timeout(5000),
					},
				);
				if (res.ok) {
					console.log(chalk.green(`  ✓ Removed MCP server '${name}'`));
				} else {
					const json = await res.json().catch(() => ({}));
					console.log(
						chalk.red(`  ✗ Failed: ${json.error?.message ?? res.statusText}`),
					);
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
				console.log(chalk.dim(`    Is the server running?`));
			}
		});

	mcp
		.command("inspect <name>")
		.description(
			"Show detailed info about an MCP server (tools, traffic stats, latency histogram)",
		)
		.action(async (name) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.bold.cyan(`\n  MCP Server: ${name}\n`));

			try {
				// Get server list and find the one matching name
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.listServers", {
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					const json = await res.json();
					const servers = json?.result?.data ?? [];
					const server = servers.find((s: any) => s.name === name);
					if (!server) {
						console.log(chalk.red(`  ✗ Server '${name}' not found`));
						console.log(
							chalk.dim(
								`    Use tormentnexus mcp list to see available servers`,
							),
						);
						return;
					}
					console.log(
						chalk.dim("  Status:     ") +
							(server.runtimeConnected
								? chalk.green("● Connected")
								: chalk.dim("○ Disconnected")),
					);

					// Check PID file for spawn status
					try {
						const fs = await import("fs");
						const path = await import("path");
						const pidFile = path.join(
							process.env.HOME ?? "",
							".tormentnexus",
							"mcp-pids",
							`${name}.pid`,
						);
						const pid = parseInt(fs.readFileSync(pidFile, "utf8").trim());
						process.kill(pid, 0);
						console.log(
							chalk.dim("  Fleet:      ") +
								chalk.green(`● Spawned (PID ${pid})`),
						);
					} catch {
						try {
							const fs = await import("fs");
							const path = await import("path");
							const pidFile = path.join(
								process.env.HOME ?? "",
								".tormentnexus",
								"mcp-pids",
								`${name}.pid`,
							);
							if (fs.existsSync(pidFile)) {
								console.log(
									chalk.dim("  Fleet:      ") + chalk.red("✗ Process dead"),
								);
							}
						} catch (e) {
							void e;
						}
					}
					console.log(
						chalk.dim("  Transport:  ") + (server.transport ?? "stdio"),
					);
					console.log(chalk.dim("  Tools:      ") + (server.toolCount ?? 0));
					console.log(
						chalk.dim("  Source:     ") + (server.source ?? "unknown"),
					);
					if (server.description)
						console.log(chalk.dim("  Description: ") + server.description);
					if (server.tags?.length)
						console.log(chalk.dim("  Tags:       ") + server.tags.join(", "));
					if (server.config?.command)
						console.log(
							chalk.dim("  Command:    ") +
								`${server.config.command} ${(server.config.args ?? []).join(" ")}`,
						);

					// Fetch tool details with descriptions
					try {
						const toolsRes = await fetch(
							"http://127.0.0.1:4100/trpc/mcp.listTools",
							{ signal: AbortSignal.timeout(5000) },
						);
						if (toolsRes.ok) {
							const allTools = (await toolsRes.json())?.result?.data ?? [];
							const serverTools = allTools.filter(
								(t: any) => (t.server ?? t.serverName) === name,
							);
							if (serverTools.length > 0) {
								console.log(
									chalk.dim(`\n  Tool Details (${serverTools.length}):`),
								);
								for (const t of serverTools.slice(0, 20)) {
									const desc = (t.description ?? "").substring(0, 80);
									console.log(`    ${chalk.cyan(t.name)}`);
									if (desc) console.log(chalk.dim(`      ${desc}`));
								}
								if (serverTools.length > 20)
									console.log(
										chalk.dim(`    ... and ${serverTools.length - 20} more`),
									);
							}
						}
					} catch (e) {
						void e;
					}
				}
			} catch {
				console.log(
					chalk.dim("  Status:    ") + "unknown (server not reachable)",
				);
				console.log(chalk.dim("  Transport: ") + "stdio");
				console.log(chalk.dim("  Tools:     ") + "0");
			}
			console.log("");
		});

	mcp
		.command("traffic")
		.description(
			"Show live MCP traffic log (JSON-RPC messages with latency and direction)",
		)
		.option("--server <name>", "Filter by server name")
		.option("--method <method>", "Filter by JSON-RPC method")
		.option("-n, --limit <count>", "Max messages to show", "50")
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.bold.cyan("  MCP Traffic Inspector"));
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.traffic", {
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					const json = await res.json();
					const events = json?.result?.data ?? [];
					if (events.length === 0) {
						console.log(chalk.dim("  No traffic events recorded."));
					} else {
						let filtered = events;
						if (opts.server)
							filtered = filtered.filter((e: any) => e.server === opts.server);
						if (opts.method)
							filtered = filtered.filter((e: any) => e.method === opts.method);
						const limit = parseInt(opts.limit) || 50;
						filtered = filtered.slice(-limit);
						for (const e of filtered) {
							const dir =
								e.direction === "in" ? chalk.green("→") : chalk.blue("←");
							const lat = e.latencyMs ? chalk.dim(`${e.latencyMs}ms`) : "";
							console.log(
								`  ${dir} [${e.server ?? "unknown"}] ${e.method ?? "unknown"} ${lat}`,
							);
						}
						console.log(chalk.dim(`\n  ${filtered.length} events shown`));
					}
				} else {
					console.log(chalk.dim("  Traffic endpoint not available."));
				}
			} catch {
				console.log(
					chalk.dim(
						"  Server not reachable. Use tormentnexus start to launch.",
					),
				);
			}
		});

	mcp
		.command("tools")
		.description(
			"List all tools across all MCP servers with namespace, priority, and usage stats",
		)
		.option("--json", "Output as JSON")
		.option("--server <name>", "Filter by server")
		.option("--namespace <ns>", "Filter by namespace")
		.option("-s, --search <query>", "Semantic search for tools")
		.action(async (opts, cmd) => {
			const allOpts = cmd ? cmd.optsWithGlobals() : opts;
			const isJson = allOpts.json === true;

			// Query the running server for tools
			let tools: any[] = [];
			try {
				if (opts.search) {
					const input = encodeURIComponent(
						JSON.stringify({ query: opts.search, limit: 100 }),
					);
					const res = await fetch(
						`http://127.0.0.1:4100/trpc/mcp.searchTools?input=${input}`,
						{ signal: AbortSignal.timeout(5000) },
					);
					if (res.ok) {
						const json = await res.json();
						tools = json?.result?.data?.tools ?? json?.result?.data ?? [];
					}
				} else {
					// Full listing — use mcp.listTools which has all 1302 tools
					const res = await fetch("http://127.0.0.1:4100/trpc/mcp.listTools", {
						signal: AbortSignal.timeout(5000),
					});
					if (res.ok) {
						const json = await res.json();
						tools = json?.result?.data ?? [];
					}
				}
			} catch (e) {
				void e;
			}

			if (opts.server)
				tools = tools.filter(
					(t: any) => (t.server ?? t.serverName) === opts.server,
				);
			if (opts.namespace)
				tools = tools.filter((t: any) =>
					(t.tags ?? []).includes(opts.namespace),
				);

			if (isJson) {
				console.log(JSON.stringify({ tools }, null, 2));
				return;
			}

			const chalk = (await import("chalk")).default;

			if (tools.length === 0) {
				console.log(chalk.bold.cyan("\n  MCP Tools\n"));
				console.log(
					chalk.dim(
						"  No tools found." +
							(opts.search
								? ` for "${opts.search}"`
								: " Is the server running?") +
							"\n",
					),
				);
				return;
			}

			const Table = (await import("cli-table3")).default;
			const table = new Table({
				head: ["Tool", "Server", "Status", "Priority"],
				style: { head: ["cyan"] },
			});

			for (const t of tools.slice(0, 50)) {
				const status = t.loaded
					? chalk.green("● Loaded")
					: t.serverStatus === "connected"
						? chalk.green("● Connected")
						: chalk.dim("○ Available");
				const priority = t.priority ?? t.rank ?? "—";
				const server = t.server ?? t.serverName ?? t.serverDisplayName ?? "—";
				table.push([t.name, server, status, String(priority)]);
			}

			const label = opts.search ? `search: "${opts.search}"` : "all";
			console.log(
				chalk.bold.cyan(`\n  MCP Tools (${tools.length} for ${label})\n`),
			);
			console.log(table.toString());
			if (tools.length > 50)
				console.log(chalk.dim(`\n  ... and ${tools.length - 50} more\n`));
		});

	mcp
		.command("config")
		.description("Show or edit MCP router configuration")
		.option("--json", "Output raw JSON config")
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;

			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/settings.get", {
					signal: AbortSignal.timeout(3000),
				});
				if (res.ok) {
					const json = await res.json();
					const settings = json?.result?.data ?? {};
					if (opts.json) {
						console.log(JSON.stringify(settings, null, 2));
					} else {
						console.log(chalk.bold.cyan("\n  TormentNexus Configuration\n"));
						const mcp = settings.mcp ?? {};
						console.log(chalk.bold("  MCP Router:"));
						console.log(
							chalk.dim("    progressiveDisclosure: ") +
								(mcp.progressiveDisclosure ?? true),
						);
						console.log(
							chalk.dim("    semanticSearch:        ") +
								(mcp.semanticSearch ?? true),
						);
						console.log(
							chalk.dim("    codeMode:              ") +
								(mcp.codeMode ?? false),
						);
						console.log(
							chalk.dim("    toolRenaming:          ") +
								(mcp.toolRenaming ?? true),
						);
						console.log(
							chalk.dim("    keepAlive:             ") +
								(mcp.keepAlive ?? true),
						);
						console.log("");
						console.log(chalk.bold("  Server:"));
						console.log(
							chalk.dim("    port:    ") + (settings.server?.port ?? 4100),
						);
						console.log(
							chalk.dim("    host:    ") + (settings.server?.host ?? "0.0.0.0"),
						);
						console.log(
							chalk.dim("    logLevel: ") +
								(settings.server?.logLevel ?? "info"),
						);
						console.log("");
						const providers = settings.providers ?? [];
						console.log(
							chalk.bold("  Providers:"),
							providers.length > 0
								? providers.length + " configured"
								: "none (use env vars)",
						);
						console.log("");
					}
				} else {
					console.log(chalk.yellow("  Server returned " + res.status));
				}
			} catch {
				// Fallback to static config
				const config = {
					namespaces: { default: { enabled: true } },
					progressiveDisclosure: true,
					semanticSearch: true,
					codeMode: false,
					toolRenaming: true,
					keepAlive: true,
				};
				if (opts.json) {
					console.log(JSON.stringify(config, null, 2));
				} else {
					console.log(chalk.bold.cyan("\n  MCP Router Config (default)\n"));
					for (const [key, val] of Object.entries(config)) {
						console.log(chalk.dim(`  ${key}: `) + JSON.stringify(val));
					}
					console.log(
						chalk.dim("\n  Note: Server not reachable — showing defaults"),
					);
					console.log("");
				}
			}
		});

	mcp
		.command("install <package>")
		.description(
			"Install an MCP server from the directory (npm, pip, or GitHub)",
		)
		.option("--npm", "Install from npm")
		.option("--pip", "Install from pip")
		.option("--github <repo>", "Install from GitHub repo")
		.addHelpText(
			"after",
			`
Examples:
  $ tormentnexus mcp install @modelcontextprotocol/server-filesystem
  $ tormentnexus mcp install --pip mcp-server-sqlite
  $ tormentnexus mcp install --github anthropics/mcp-servers
    `,
		)
		.action(async (pkg, opts) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.yellow(`  Installing MCP server: ${pkg}...`));

			// Derive server name from package
			const serverName = pkg
				.replace(/^@.*?\//, "")
				.replace(/^mcp-server-/, "")
				.replace(/\/.*$/, "");
			const command = opts.npm ? "npm" : opts.pip ? "pip" : "npx";
			const args = opts.pip
				? ["install", pkg]
				: opts.npm
					? ["install", "-g", pkg]
					: ["-y", pkg];

			try {
				// Add to TormentNexus MCP config via tRPC
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.addServer", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						json: {
							name: serverName,
							command: "npx",
							args: ["-y", pkg],
							env: {},
						},
					}),
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					console.log(chalk.green(`  ✓ Installed '${pkg}' as '${serverName}'`));
					console.log(
						chalk.dim(
							`    Use 'tormentnexus mcp start ${serverName}' to connect`,
						),
					);
				} else {
					const json = await res.json().catch(() => ({}));
					console.log(
						chalk.red(`  ✗ Failed: ${json.error?.message ?? res.statusText}`),
					);
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
				console.log(
					chalk.dim("    Is the server running? Use tormentnexus start"),
				);
			}
		});

	mcp
		.command("search <query>")
		.description("Search the MCP directory for available servers")
		.option("-c, --category <cat>", "Filter by category")
		.option("-n, --limit <count>", "Max results", "20")
		.action(async (query, opts, cmd) => {
			const allOpts = cmd ? cmd.optsWithGlobals() : opts;
			const isJson = allOpts.json === true;

			let servers: any[] = [];
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.listServers", {
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					const json = await res.json();
					servers = json?.result?.data ?? [];
				}
			} catch (e) {
				void e;
			}

			// Fuzzy search across name, tags, description
			const q = query.toLowerCase();
			let results = servers.filter((s: any) => {
				const name = (s.name ?? "").toLowerCase();
				const tags = (s.tags ?? []).join(" ").toLowerCase();
				const display = (s.displayName ?? "").toLowerCase();
				return name.includes(q) || tags.includes(q) || display.includes(q);
			});

			// Deduplicate by name
			const seen = new Set<string>();
			results = results.filter((s: any) => {
				const key = s.name ?? s.displayName ?? "";
				if (seen.has(key)) return false;
				seen.add(key);
				return true;
			});

			const limit = parseInt(opts.limit) || 20;
			const limited = results.slice(0, limit);

			if (isJson) {
				console.log(JSON.stringify({ query, results: limited }, null, 2));
				return;
			}

			const chalk = (await import("chalk")).default;
			console.log(
				chalk.bold.cyan(
					`\n  MCP Directory Search: "${query}" (${limited.length}/${results.length} results)\n`,
				),
			);

			if (limited.length === 0) {
				console.log(chalk.dim("  No matching servers found.\n"));
				return;
			}

			const Table = (await import("cli-table3")).default;
			const table = new Table({
				head: ["Name", "Tools", "Tags"],
				style: { head: ["cyan"] },
			});
			for (const s of limited) {
				table.push([
					s.name,
					String(s.toolCount ?? 0),
					(s.tags ?? []).slice(0, 3).join(", "),
				]);
			}
			console.log(table.toString());
			console.log("");
		});

	mcp
		.command("export")
		.description("Export MCP configuration to JSON file")
		.option(
			"-o, --output <file>",
			"Output file path",
			"tormentnexus-mcp-export.json",
		)
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			const { writeFileSync } = await import("fs");

			let servers: any[] = [];
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.listServers", {
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) {
					const json = await res.json();
					servers = json?.result?.data ?? [];
				}
			} catch (e) {
				void e;
			}

			const exported = {
				exportedAt: new Date().toISOString(),
				version: await (async () => {
					try {
						const { readFileSync } = await import("fs");
						const { resolve } = await import("path");
						let d = process.cwd();
						for (let i = 0; i < 20; i++) {
							try {
								return readFileSync(resolve(d, "VERSION"), "utf8").trim();
							} catch (e) {
								void e;
							}
							const p = resolve(d, "..");
							if (p === d) break;
							d = p;
						}
					} catch (e) {
						void e;
					}
					return "unknown";
				})(),
				serverCount: servers.length,
				totalTools: servers.reduce(
					(a: number, s: any) => a + (s.toolCount ?? 0),
					0,
				),
				servers: servers.map((s: any) => ({
					name: s.name,
					tags: s.tags,
					toolCount: s.toolCount,
					status: s.status,
					alwaysOn: s.alwaysOn,
					config: s.config,
				})),
			};

			writeFileSync(opts.output, JSON.stringify(exported, null, 2));
			console.log(
				chalk.green(
					`  ✓ Exported ${servers.length} servers (${exported.totalTools} tools) to ${opts.output}`,
				),
			);
		});

	mcp
		.command("import <file>")
		.description("Import MCP configuration from JSON file")
		.option("--merge", "Merge with existing config instead of replacing")
		.action(async (file, opts) => {
			const chalk = (await import("chalk")).default;
			const { readFileSync, existsSync, writeFileSync } = await import("fs");
			const { resolve } = await import("path");

			const filePath = resolve(process.cwd(), file);
			if (!existsSync(filePath)) {
				console.log(chalk.red(`  ✗ File not found: ${filePath}`));
				return;
			}

			try {
				const raw = readFileSync(filePath, "utf8");
				const imported = JSON.parse(raw);
				const servers = imported?.servers ?? imported?.mcpServers ?? [];
				const serverList = Array.isArray(servers)
					? servers
					: Object.entries(servers).map(([name, config]: [string, any]) => ({
							name,
							...config,
						}));

				console.log(
					chalk.yellow(
						`  Importing ${serverList.length} servers from ${file}...`,
					),
				);

				let imported_count = 0,
					failed = 0;
				for (const server of serverList.slice(0, 50)) {
					try {
						const res = await fetch(
							"http://127.0.0.1:4100/trpc/mcp.addServer",
							{
								method: "POST",
								headers: { "Content-Type": "application/json" },
								body: JSON.stringify({
									json: {
										name: server.name,
										command: server.command ?? "npx",
										args: server.args ?? [],
										env: server.env ?? {},
									},
								}),
								signal: AbortSignal.timeout(3000),
							},
						);
						if (res.ok) {
							imported_count++;
						} else {
							failed++;
						}
					} catch {
						failed++;
					}
				}

				// Fallback: write directly to mcp.jsonc if tRPC failed
				if (imported_count === 0 && failed > 0) {
					console.log(
						chalk.yellow(
							"  tRPC unavailable, writing directly to mcp.jsonc...",
						),
					);
					try {
						const mcpPath = resolve(process.cwd(), "mcp.jsonc");
						if (existsSync(mcpPath)) {
							// Parse JSONC (strip comments carefully)
							let mcpContent = readFileSync(mcpPath, "utf8");
							// Strip single-line comments (not inside strings)
							mcpContent = mcpContent.replace(
								/("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')|[//].*$/gm,
								"$1",
							);
							mcpContent = mcpContent.replace(
								/("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')|[/][*][\s\S]*?[*][/]/g,
								"$1",
							);
							// Strip trailing commas before } or ]
							mcpContent = mcpContent.replace(/,\s*([}\]])/g, "$1");
							const mcpConfig = JSON.parse(mcpContent);
							const mcpServers = mcpConfig.mcpServers ?? {};
							let added = 0;
							for (const server of serverList) {
								if (server.config?.command || server.command) {
									mcpServers[server.name] = {
										command: server.config?.command ?? server.command ?? "npx",
										args: server.config?.args ?? server.args ?? [],
									};
									if (server.config?.env)
										mcpServers[server.name].env = server.config.env;
									else if (server.env) mcpServers[server.name].env = server.env;
									added++;
								}
							}
							mcpConfig.mcpServers = mcpServers;
							writeFileSync(mcpPath, JSON.stringify(mcpConfig, null, 2));
							console.log(
								chalk.green(`  ✓ Wrote ${added} servers to mcp.jsonc`),
							);
							imported_count = added;
							failed = 0;
						} else {
							console.log(chalk.yellow("  ⚠ mcp.jsonc not found in cwd"));
						}
					} catch (e: any) {
						console.log(
							chalk.red(`  ✗ mcp.jsonc fallback failed: ${e.message}`),
						);
					}
				}

				console.log(
					chalk.green(`  ✓ Imported ${imported_count} servers from ${file}`),
				);
				if (failed > 0)
					console.log(chalk.yellow(`  ⚠ ${failed} servers failed to import`));
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	mcp
		.command("sync")
		.description(
			"Auto-detect and sync MCP configs to all AI tools (Claude, Cursor, VS Code, etc.)",
		)
		.option("--dry-run", "Show what would be synced without writing")
		.option("--client <name>", "Sync to specific client only")
		.addHelpText(
			"after",
			`
Supported clients:
  claude-desktop (claude)
  cursor
  vscode
    `,
		)
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.bold.cyan("\n  MCP Config Sync (Go Native)\n"));

			const SIDECAR_URL =
				process.env.TORMENTNEXUS_SIDECAR_URL || "http://127.0.0.1:4300";

			if (opts.dryRun) {
				console.log(chalk.yellow("  [DRY RUN] No changes will be written\n"));
			}

			try {
				// 1. Get sync targets from Go sidecar
				const targetsRes = await fetch(
					`${SIDECAR_URL}/api/mcp/servers/sync-targets`,
					{ signal: AbortSignal.timeout(5000) },
				);
				if (!targetsRes.ok)
					throw new Error(
						`Go sidecar returned ${targetsRes.status} on sync-targets`,
					);

				const targetsJson = await targetsRes.json();
				const targets = targetsJson.data || [];

				const clientMap: Record<string, string> = {
					claude: "claude-desktop",
					"claude-desktop": "claude-desktop",
					cursor: "cursor",
					vscode: "vscode",
				};

				const targetClient = opts.client
					? clientMap[opts.client.toLowerCase()] || opts.client
					: null;

				for (const target of targets) {
					if (targetClient && target.client !== targetClient) continue;

					if (!target.exists) {
						console.log(chalk.dim(`  ○ ${target.client}: not detected`));
						continue;
					}

					if (opts.dryRun) {
						console.log(
							chalk.green(`  ✓ ${target.client}`) +
								chalk.dim(` detected at ${target.path}`),
						);
						continue;
					}

					// 2. Execute sync via Go sidecar
					const syncRes = await fetch(
						`${SIDECAR_URL}/api/mcp/servers/sync-client-config`,
						{
							method: "POST",
							headers: { "Content-Type": "application/json" },
							body: JSON.stringify({ client: target.client }),
							signal: AbortSignal.timeout(10000),
						},
					);

					if (syncRes.ok) {
						const result = (await syncRes.json()).data;
						console.log(
							chalk.green(`  ✓ ${result.client}`) +
								chalk.dim(
									` synced ${result.serverCount} servers to ${result.targetPath}`,
								),
						);
					} else {
						const errorJson = await syncRes.json().catch(() => ({}));
						console.log(
							chalk.red(`  ✗ ${target.client}: Sync failed`) +
								chalk.dim(` (${errorJson.error || syncRes.statusText})`),
						);
					}
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
				console.log(
					chalk.dim(
						"    Ensure the TormentNexus Go sidecar is running (typically via `tormentnexus start`).",
					),
				);
			}
			console.log("");
		});

	mcp
		.command("connect-all")
		.description("Attempt to connect all configured MCP servers")
		.option("--timeout <ms>", "Connection timeout per server", "10000")
		.action(async (opts) => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.bold.cyan("\n  MCP Connect All\n"));

			try {
				// Get list of servers
				const listRes = await fetch(
					"http://127.0.0.1:4100/trpc/mcp.listServers",
					{
						signal: AbortSignal.timeout(5000),
					},
				);
				if (!listRes.ok) {
					console.log(chalk.red("  ✗ Failed to fetch server list"));
					return;
				}
				const listJson = await listRes.json();
				const servers = listJson?.result?.data ?? [];
				const disconnected = servers.filter((s: any) => !s.runtimeConnected);
				const alreadyConnected = servers.length - disconnected.length;

				console.log(
					chalk.dim(
						`  Found ${servers.length} servers (${alreadyConnected} already connected)`,
					),
				);
				console.log(
					chalk.dim(
						`  Attempting to connect ${disconnected.length} servers...\n`,
					),
				);

				let connected = 0,
					failed = 0,
					spawned = 0;
				const { spawn } = await import("child_process");

				for (const server of disconnected.slice(0, 808)) {
					const name = server.name;

					// Try tRPC first
					let ok = false;
					try {
						const res = await fetch(
							"http://127.0.0.1:4100/trpc/mcp.connectServer",
							{
								method: "POST",
								headers: { "Content-Type": "application/json" },
								body: JSON.stringify({ json: { name } }),
								signal: AbortSignal.timeout(parseInt(opts.timeout)),
							},
						);
						if (res.ok) {
							ok = true;
							connected++;
							console.log(chalk.green(`  ✓ ${name}`));
						}
					} catch (e) {
						void e;
					}

					// Fallback: spawn directly
					if (!ok && server.config?.command) {
						try {
							const proc = spawn(
								server.config.command,
								server.config.args ?? [],
								{
									stdio: "ignore",
									detached: true,
									env: { ...process.env, ...(server.config.env ?? {}) },
									shell: true,
								},
							);
							proc.unref();

							// Write PID file
							try {
								const fs = await import("fs");
								const path = await import("path");
								const pidDir = path.join(
									process.env.HOME ?? "",
									".tormentnexus",
									"mcp-pids",
								);
								fs.mkdirSync(pidDir, { recursive: true });
								fs.writeFileSync(
									path.join(pidDir, `${name}.pid`),
									String(proc.pid),
								);
							} catch (e) {
								void e;
							}
							console.log(
								chalk.green(`  ✓ ${name}`) +
									chalk.dim(` (spawned PID ${proc.pid})`),
							);
						} catch (e: any) {
							failed++;
							console.log(chalk.red(`  ✗ ${name}: ${e.message}`));
						}
					} else if (!ok) {
						failed++;
						console.log(chalk.red(`  ✗ ${name}: no command`));
					}
				}
				console.log(
					chalk.dim(
						`\n  ${connected} connected, ${spawned} spawned, ${failed} failed`,
					),
				);
				if (disconnected.length > 20) {
					console.log(
						chalk.dim(
							`  (${disconnected.length - 20} servers skipped — use individual connect for more)`,
						),
					);
				}
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	mcp
		.command("stop-all")
		.description("Stop all spawned MCP servers")
		.action(async () => {
			const chalk = (await import("chalk")).default;
			console.log(
				chalk.bold.cyan("\n  MCP Stop All — Killing spawned servers\n"),
			);

			try {
				const fs = await import("fs");
				const path = await import("path");
				const pidDir = path.join(
					process.env.HOME ?? "",
					".tormentnexus",
					"mcp-pids",
				);
				const pidFiles = fs
					.readdirSync(pidDir)
					.filter((f) => f.endsWith(".pid"));

				let killed = 0,
					alreadyDead = 0;
				for (const pf of pidFiles) {
					const name = pf.replace(".pid", "");
					try {
						const pid = parseInt(
							fs.readFileSync(path.join(pidDir, pf), "utf8").trim(),
						);
						process.kill(pid, "SIGTERM");
						killed++;
						console.log(chalk.green(`  ✓ ${name} (PID ${pid}) killed`));
					} catch {
						alreadyDead++;
						console.log(chalk.dim(`  ○ ${name} already dead`));
					}
					try {
						fs.unlinkSync(path.join(pidDir, pf));
					} catch (e) {
						void e;
					}
				}
				console.log(
					chalk.dim(
						`\n  ${killed} killed, ${alreadyDead} already dead, ${pidFiles.length} total`,
					),
				);
			} catch (e: any) {
				console.log(chalk.red(`  ✗ Error: ${e.message}`));
			}
		});

	mcp
		.command("fleet")
		.description("Show spawned MCP server fleet status")
		.action(async () => {
			const chalk = (await import("chalk")).default;
			console.log(chalk.bold.cyan("\n  MCP Fleet — Spawned Servers\n"));

			// Get servers with commands from tRPC
			let servers: any[] = [];
			try {
				const res = await fetch("http://127.0.0.1:4100/trpc/mcp.listServers", {
					signal: AbortSignal.timeout(5000),
				});
				if (res.ok) servers = (await res.json())?.result?.data ?? [];
			} catch (e) {
				void e;
			}

			const startable = servers.filter((s: any) => s.config?.command);
			const noCommand = servers.length - startable.length;

			let running = 0,
				stopped = 0;
			for (const s of startable) {
				const pidFile = `${process.env.HOME}/.tormentnexus/mcp-pids/${s.name}.pid`;
				let alive = false;
				try {
					const fs = await import("fs");
					const pid = parseInt(fs.readFileSync(pidFile, "utf8").trim());
					process.kill(pid, 0);
					alive = true;
					running++;
				} catch (e) {
					void e;
				}

				const icon = alive ? chalk.green("●") : chalk.dim("○");
				const cmd = (
					s.config.command +
					" " +
					(s.config.args ?? []).join(" ")
				).substring(0, 60);
				console.log(`  ${icon} ${s.name} ${chalk.dim(cmd)}`);
			}

			console.log(
				chalk.dim(
					`\n  ${running} running, ${startable.length - running} stopped, ${noCommand} no-command, ${servers.length} total`,
				),
			);
			console.log(
				chalk.dim("  Use tormentnexus mcp connect-all to spawn servers\n"),
			);
		});
}
