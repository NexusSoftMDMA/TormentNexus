import { writeFileSync } from "fs";

// Correct procedure names based on router source
const procs = [
	// Core
	"health",
	"mcp.listServers",
	"mcp.getStatus",
	"mcp.getWorkingSet",
	"mcp.listTools",
	"mcpServers.list",
	"savedScripts.list",
	"serverHealth.reset",

	// Memory
	"memory.getAgentStats",

	// Settings
	"settings.get",
	"secrets.list",

	// Sessions
	"session.list",
	"session.catalog",

	// Billing
	"billing.getStatus",

	// Agents
	"squad.list",
	"director.status",
	"supervisor.status",
	"agent.list",

	// Tools
	"tools.list",
	"toolSets.list",
	"catalog.list",
	"skills.list",

	// Browser
	"browser.status",

	// Infrastructure
	"logs.list",
	"apiKeys.list",
	"policies.list",
	"unifiedDirectory.list",
	"workspace.list",
	"marketplace.list",
	"linksBacklog.list",

	// Healer
	"healer.getHistory",

	// Knowledge
	"knowledge.getStats",

	// Suggestions
	"suggestions.list",

	// Commands
	"commands.list",
];

let ok = 0,
	fail = 0,
	err = 0;
const lines = [];

for (const proc of procs) {
	try {
		const res = await fetch(`http://127.0.0.1:4000/trpc/${proc}`, {
			signal: AbortSignal.timeout(5000),
		});
		if (res.ok) {
			lines.push(`  ✓ ${proc}`);
			ok++;
		} else {
			const body = await res.text();
			const msg = body.substring(0, 80).replace(/\n/g, " ");
			lines.push(`  ✗ ${proc} → HTTP ${res.status}: ${msg}`);
			err++;
		}
	} catch (e) {
		lines.push(`  ✗ ${proc} → ${e.message?.substring(0, 40)}`);
		fail++;
	}
}

lines.push("");
lines.push(`Summary: ${ok} ok, ${err} error, ${fail} timeout`);
lines.push(`Total procedures tested: ${procs.length}`);
const out = lines.join("\n");
console.log(out);
writeFileSync("/tmp/trpc-audit.txt", out);
