import { spawn } from "child_process";

const TIMEOUT_MS = 15000;
const MCP_SERVER_PATH =
	"/app/packages/core/dist/stdioLoader.js";
const ENV = {
	...process.env,
	TORMENTNEXUS_WORKSPACE_ROOT: "/app",
	TORMENTNEXUS_TRPC_PORT: "4100",
	TORMENTNEXUS_SIDECAR_PORT: "4300",
};

let requestId = 0;
let mcpProcess = null;
let buffer = "";
const pendingRequests = new Map();
const stderrLogs = [];
const testResults = [];

function sendRequest(method, params = {}) {
	const id = ++requestId;
	const msg = JSON.stringify({ jsonrpc: "2.0", id, method, params });
	return new Promise((resolve, reject) => {
		const timer = setTimeout(() => {
			pendingRequests.delete(id);
			reject(new Error("Timeout after " + TIMEOUT_MS + "ms"));
		}, TIMEOUT_MS);
		pendingRequests.set(id, { resolve, reject, timer, method });
		mcpProcess.stdin.write(msg + "\n");
	});
}

function sendNotification(method, params = {}) {
	const msg = JSON.stringify({ jsonrpc: "2.0", method, params });
	mcpProcess.stdin.write(msg + "\n");
}

function handleStdout(data) {
	buffer += data.toString();
	while (true) {
		const lineEnd = buffer.indexOf("\n");
		if (lineEnd === -1) break;
		const line = buffer.slice(0, lineEnd).trim();
		buffer = buffer.slice(lineEnd + 1);
		if (!line) continue;
		try {
			const msg = JSON.parse(line);
			if (msg.id && pendingRequests.has(msg.id)) {
				const { resolve, reject, timer } = pendingRequests.get(msg.id);
				clearTimeout(timer);
				pendingRequests.delete(msg.id);
				if (msg.error)
					reject(new Error(msg.error.message || JSON.stringify(msg.error)));
				else resolve(msg.result);
			}
		} catch (e) {
			/* not JSON */
		}
	}
}

async function testTool(name, args = {}) {
	const start = Date.now();
	try {
		const result = await sendRequest("tools/call", { name, arguments: args });
		const ms = Date.now() - start;
		const content = result?.content?.[0]?.text || "";
		const isError = result?.isError === true;
		const icon = isError ? "❌" : ms > 8000 ? "🐢" : "⚡";
		testResults.push({
			name,
			pass: !isError,
			ms,
			isError,
			content: content.slice(0, 200),
		});
		console.log(
			"  " +
				icon +
				" " +
				name +
				" (" +
				ms +
				"ms): " +
				content.slice(0, 80) +
				(content.length > 80 ? "..." : ""),
		);
		return { name, pass: !isError, ms, content };
	} catch (e) {
		const ms = Date.now() - start;
		testResults.push({
			name,
			pass: false,
			ms,
			isError: true,
			content: e.message,
		});
		console.log("  💀 " + name + " (" + ms + "ms): " + e.message);
		return { name, pass: false, ms, content: e.message };
	}
}

async function startMcpServer(label) {
	const proc = spawn("node", [MCP_SERVER_PATH], {
		env: ENV,
		stdio: ["pipe", "pipe", "pipe"],
		cwd: "/app",
	});
	proc.stdout.on("data", handleStdout);
	proc.stderr.on("data", (d) => {
		const lines = d
			.toString()
			.split("\n")
			.filter((l) => l.trim());
		for (const l of lines) {
			stderrLogs.push({ ts: Date.now(), msg: l });
			if (
				/error|fail|crash|reject|ECONNREFUSED|ENOENT|timeout/i.test(l) &&
				!/heartbeat/i.test(l)
			) {
				console.log("  \u{1FAB5} [" + label + "] " + l.slice(0, 150));
			}
		}
	});
	proc.on("error", (e) => console.log("  \u274C Process error: " + e.message));
	proc.on("exit", (code) => {
		if (code) console.log("  \u26A0\uFE0F Process exited with code " + code);
	});
	return proc;
}

async function initMcp() {
	await new Promise((r) => setTimeout(r, 2000));
	const initResult = await sendRequest("initialize", {
		protocolVersion: "2024-11-05",
		capabilities: { tools: {} },
		clientInfo: { name: "test-harness", version: "1.0.0" },
	});
	sendNotification("notifications/initialized");
	await new Promise((r) => setTimeout(r, 1000));
	return initResult;
}

async function runTests() {
	console.log(
		"\u2554\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2557",
	);
	console.log(
		"\u2551  TORMENTNEXUS AIOS \u2014 COMPREHENSIVE MCP TEST                 \u2551",
	);
	console.log(
		"\u255A\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u255D\n",
	);

	// Phase 0: Pre-flight
	console.log(
		"\u2550\u2550\u2550 Phase 0: Pre-flight Service Checks \u2550\u2550\u2550",
	);
	for (const s of [
		{ name: "tRPC Server", url: "http://localhost:4100/health" },
		{ name: "Go Sidecar", url: "http://localhost:4300/health" },
		{ name: "Dashboard", url: "http://localhost:3000/" },
	]) {
		try {
			const resp = await fetch(s.url, { signal: AbortSignal.timeout(5000) });
			console.log("  \u2705 " + s.name + ": HTTP " + resp.status);
		} catch (e) {
			console.log("  \u274C " + s.name + ": " + e.message);
		}
	}
	console.log("");

	// Phase 1: MCP server startup
	console.log(
		"\u2550\u2550\u2550 Phase 1: MCP Server Startup (stdio) \u2550\u2550\u2550",
	);
	const startTime = Date.now();
	mcpProcess = await startMcpServer("client1");
	const initResult = await initMcp();
	const initMs = Date.now() - startTime;
	console.log("  \u2705 MCP initialized in " + initMs + "ms");
	console.log(
		"     Server: " +
			(initResult?.serverInfo?.name || "?") +
			" v" +
			(initResult?.serverInfo?.version || "?"),
	);
	console.log("");

	// Phase 2: Tool Discovery
	console.log("\u2550\u2550\u2550 Phase 2: Tool Discovery \u2550\u2550\u2550");
	const toolsResult = await sendRequest("tools/list", {});
	const tools = toolsResult?.tools || [];
	console.log("  Discovered " + tools.length + " tools");
	const toolNames = tools.map((t) => t.name).sort();
	console.log("  Names: " + toolNames.join(", "));
	let schemaOk = 0,
		schemaBad = 0;
	for (const t of tools) {
		if (t.inputSchema && t.inputSchema.type === "object") schemaOk++;
		else schemaBad++;
	}
	console.log(
		"  Schema validation: " + schemaOk + " OK, " + schemaBad + " bad",
	);
	console.log("");

	// Phase 3: Execute ALL tools
	console.log(
		"\u2550\u2550\u2550 Phase 3: Execute All Internal Tools \u2550\u2550\u2550",
	);

	// System/Status
	await testTool("router_status");
	await testTool("system_status");
	await testTool("system_diagnostics");
	await testTool("health_check");
	await testTool("code_mode_status");

	// Progressive discovery
	await testTool("search_tools", { query: "memory" });
	await testTool("list_loaded_tools");
	await testTool("set_capacity", { maxLoadedTools: 30 });
	await testTool("get_eviction_history");
	await testTool("clear_eviction_history");

	// Search/Index
	await testTool("symbol_search", { query: "MCPServer" });
	await testTool("lsp_symbol_search", { query: "index" });
	await testTool("search_codebase", { query: "memory manager" });
	await testTool("index_codebase", {
		path: "/app/packages/core/src",
	});

	// Agent/Squad/Skill
	await testTool("list_squads");
	await testTool("list_agents");
	await testTool("list_skills");
	await testTool("search_skills", { query: "search" });
	await testTool("list_workflows");
	await testTool("plan_mode_status");

	// Memory
	await testTool("memory_store", {
		content: "test memory from harness",
		type: "working",
	});
	await testTool("memory_recall", { query: "test" });
	await testTool("add_memory", {
		content: "harness test memory",
		type: "working",
		namespace: "test",
	});
	await testTool("search_memory", { query: "harness test" });
	await testTool("get_recent_memories", { limit: 5 });
	await testTool("memory_stats");
	await testTool("save_memory", { content: "saved test", type: "long_term" });

	// File/Project
	await testTool("read_file", {
		path: "/app/package.json",
	});
	await testTool("list_directory", { path: "/app" });
	await testTool("get_logs", { lines: 10 });
	await testTool("config_get", { key: "version" });
	await testTool("git_worktree_list");
	await testTool("get_project_context");

	// Execution
	await testTool("run_code", { code: "return 1+1" });
	await testTool("run_python", { code: 'print("hello from python")' });
	await testTool("execute_sandbox", {
		language: "javascript",
		code: "return 2+2",
	});
	await testTool("toolset_list");

	// Session/Context
	await testTool("handoff_session");
	await testTool("compact_context");
	await testTool("export_chat");

	// Healer
	await testTool("auto_heal", { error: "test error" });
	await testTool("healer_diagnose", { error: "TypeError: test" });
	await testTool("healer_heal", { error: "test" });

	// Note
	await testTool("process_note", {
		content: "Test note from harness",
		type: "observation",
	});
	console.log("");

	// Phase 4: Progressive tool loading
	console.log(
		"\u2550\u2550\u2550 Phase 4: Progressive Tool Discovery \u2550\u2550\u2550",
	);
	await testTool("load_tool", { name: "memory_stats" });
	await testTool("get_tool_schema", { name: "memory_stats" });
	await testTool("get_tool_context", { name: "memory_stats" });
	await testTool("list_loaded_tools");
	await testTool("unload_tool", { name: "memory_stats" });
	await testTool("list_loaded_tools");
	console.log("");

	// Phase 5: Go Sidecar
	console.log(
		"\u2550\u2550\u2550 Phase 5: Go Sidecar Native Endpoints \u2550\u2550\u2550",
	);
	for (const ep of [
		{ name: "Health", url: "http://localhost:4300/health" },
		{ name: "Code Index", url: "http://localhost:4300/api/index" },
		{ name: "Runtime Status", url: "http://localhost:4300/api/runtime/status" },
		{ name: "CLI Summary", url: "http://localhost:4300/api/cli/summary" },
		{
			name: "Provider Routing",
			url: "http://localhost:4300/api/providers/routing-summary",
		},
		{ name: "Import Summary", url: "http://localhost:4300/api/import/summary" },
	]) {
		try {
			const t0 = Date.now();
			const resp = await fetch(ep.url, { signal: AbortSignal.timeout(10000) });
			const body = await resp.text();
			console.log(
				"  \u2705 " +
					ep.name +
					": HTTP " +
					resp.status +
					", " +
					body.length +
					" bytes, " +
					(Date.now() - t0) +
					"ms",
			);
		} catch (e) {
			console.log("  \u274C " + ep.name + ": " + e.message);
		}
	}
	console.log("");

	// Phase 6: Dashboard resilience
	console.log(
		"\u2550\u2550\u2550 Phase 6: Dashboard Resilience Check \u2550\u2550\u2550",
	);
	for (const page of [
		"",
		"dashboard",
		"dashboard/mcp",
		"dashboard/council",
		"dashboard/config",
	]) {
		try {
			const resp = await fetch("http://localhost:3000/" + page, {
				signal: AbortSignal.timeout(5000),
			});
			console.log("  \u2705 /" + (page || "(root)") + ": HTTP " + resp.status);
		} catch (e) {
			console.log("  \u274C /" + (page || "(root)") + ": " + e.message);
		}
	}
	console.log("");

	// Phase 7: tRPC API
	console.log(
		"\u2550\u2550\u2550 Phase 7: tRPC API Endpoints \u2550\u2550\u2550",
	);
	for (const ep of [
		{ name: "Health", path: "health" },
		{ name: "Tools List", path: "tools.list?input=%7B%7D" },
		{ name: "Memory Stats", path: "agentMemory.stats?input=%7B%7D" },
		{ name: "Process List", path: "process.list?input=%7B%7D" },
	]) {
		try {
			const resp = await fetch("http://localhost:4100/trpc/" + ep.path, {
				signal: AbortSignal.timeout(10000),
			});
			const body = await resp.text();
			console.log(
				"  \u2705 " +
					ep.name +
					": HTTP " +
					resp.status +
					", " +
					body.length +
					" bytes",
			);
		} catch (e) {
			console.log("  \u274C " + ep.name + ": " + e.message);
		}
	}
	console.log("");

	// Phase 8: Second MCP connection
	console.log(
		"\u2550\u2550\u2550 Phase 8: Second MCP Client (fresh spawn) \u2550\u2550\u2550",
	);
	mcpProcess.kill();
	await new Promise((r) => setTimeout(r, 2000));
	requestId = 0;
	buffer = "";
	pendingRequests.clear();
	mcpProcess = await startMcpServer("client2");
	const init2 = await initMcp();
	console.log(
		"  \u2705 Second client: " +
			(init2?.serverInfo?.name || "?") +
			" v" +
			(init2?.serverInfo?.version || "?"),
	);
	await testTool("router_status");
	await testTool("health_check");
	await testTool("memory_stats");
	await testTool("list_agents");
	await testTool("search_skills", { query: "test" });
	await testTool("run_code", { code: "return 42" });
	await testTool("read_file", {
		path: "/app/package.json",
	});
	mcpProcess.kill();
	console.log("");

	// Phase 9: Log analysis
	console.log("\u2550\u2550\u2550 Phase 9: Log Analysis \u2550\u2550\u2550");
	const patterns = {};
	for (const log of stderrLogs) {
		let cat = "other";
		if (/error|fail|crash|ENOENT|ECONNREFUSED/i.test(log.msg)) cat = "errors";
		else if (/warn|deprecat/i.test(log.msg)) cat = "warnings";
		else if (/timeout|timed out/i.test(log.msg)) cat = "timeouts";
		else if (/init|connect|ready/i.test(log.msg)) cat = "startup";
		else if (/heartbeat/i.test(log.msg)) cat = "heartbeat";
		else continue;
		if (cat !== "heartbeat") {
			const key = log.msg.slice(0, 100).replace(/\d{4,}/g, "N");
			patterns[key] = (patterns[key] || 0) + 1;
		}
	}
	if (Object.keys(patterns).length > 0) {
		console.log("  Unique log patterns (non-heartbeat):");
		const sorted = Object.entries(patterns).sort((a, b) => b[1] - a[1]);
		for (const [pattern, count] of sorted.slice(0, 25)) {
			const icon = /error|fail|crash|ENOENT|ECONNREFUSED|timeout/i.test(pattern)
				? "\uD83D\uDD34"
				: "\uD83D\uDFE1";
			console.log("  " + icon + " [x" + count + "] " + pattern);
		}
	} else {
		console.log("  \u2705 No error/warning patterns found");
	}
	console.log("");

	// Final summary
	console.log(
		"\u2554\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2557",
	);
	console.log(
		"\u2551  FINAL COMPREHENSIVE RESULTS                                 \u2551",
	);
	console.log(
		"\u255A\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u255D",
	);

	const pass = testResults.filter((r) => r.pass).length;
	const fail = testResults.filter((r) => !r.pass).length;
	const slow = testResults.filter((r) => r.ms > 8000).length;
	const total = testResults.length;

	console.log("  \u2705 Pass:   " + pass);
	console.log("  \u274C Error:  " + fail);
	console.log("  \uD83D\uDC7E Slow:   " + slow + " (>8s)");
	console.log(
		"  \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500",
	);
	console.log("  Total:    " + total);
	console.log("  Rate:     " + ((pass / total) * 100).toFixed(1) + "%");

	if (fail > 0) {
		console.log("\n  \u274C ALL ERRORS:");
		testResults
			.filter((r) => !r.pass)
			.forEach((r) =>
				console.log("    " + r.name + ": " + r.content.slice(0, 120)),
			);
	}

	if (slow > 0) {
		console.log("\n  \uD83D\uDC7E SLOW TOOLS (>8s):");
		testResults
			.filter((r) => r.ms > 8000)
			.sort((a, b) => b.ms - a.ms)
			.forEach((r) =>
				console.log(
					"    " +
						r.name +
						": " +
						(r.ms / 1000).toFixed(1) +
						"s \u2014 " +
						r.content.slice(0, 80),
				),
			);
	}

	// Recommendations
	console.log("\n  \uD83D\uDCCB RECOMMENDATIONS:");
	const vecTimeouts = testResults.filter((r) =>
		r.content?.includes("Vector store timed out"),
	);
	if (vecTimeouts.length > 0)
		console.log(
			"  \u26A1 " +
				vecTimeouts.length +
				" vector store timeouts \u2014 LanceDB init issue",
		);
	const notFound = testResults.filter(
		(r) => r.content?.includes("not found") || r.content?.includes("NotFound"),
	);
	if (notFound.length > 0) {
		console.log(
			"  \uD83D\uDC80 " +
				notFound.length +
				' "not found" errors \u2014 missing tool handlers',
		);
		notFound.forEach((r) => console.log("     " + r.name));
	}
	const slowTools = testResults.filter((r) => r.ms > 5000 && r.pass);
	if (slowTools.length > 0) {
		console.log(
			"  \uD83D\uDC7E " +
				slowTools.length +
				" tools >5s \u2014 investigate performance",
		);
		slowTools
			.sort((a, b) => b.ms - a.ms)
			.forEach((r) =>
				console.log("     " + r.name + ": " + (r.ms / 1000).toFixed(1) + "s"),
			);
	}

	// Final dashboard check
	console.log(
		"\n\u2550\u2550\u2550 Post-test Dashboard Verification \u2550\u2550\u2550",
	);
	try {
		const resp = await fetch("http://localhost:3000/dashboard", {
			signal: AbortSignal.timeout(5000),
		});
		console.log(
			"  Dashboard: HTTP " +
				resp.status +
				(resp.status === 200
					? " \u2705 STILL RUNNING"
					: " \u26A0\uFE0F UNEXPECTED STATUS"),
		);
	} catch (e) {
		console.log("  Dashboard: \u274C " + e.message);
	}
}

runTests().catch((e) => {
	console.error("Test harness error:", e);
	if (mcpProcess) mcpProcess.kill();
	process.exit(1);
});
