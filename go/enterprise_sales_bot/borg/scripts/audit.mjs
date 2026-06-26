import { writeFileSync } from 'fs';

const PORT = process.argv[2] || '4000';
const procs = [
  'health',
  'mcp.listServers', 'mcp.getStatus', 'mcp.getWorkingSet', 'mcp.listTools',
  'mcpServers.list', 'savedScripts.list',
  'memory.getAgentStats', 'settings.get', 'secrets.list',
  'session.list', 'session.catalog',
  'billing.getStatus', 'squad.list',
  'director.status', 'supervisor.status',
  'tools.list', 'toolSets.list', 'catalog.list',
  'skills.list', 'browser.status',
  'apiKeys.list', 'policies.list', 'unifiedDirectory.list',
  'workspace.list', 'marketplace.list', 'linksBacklog.list',
  'healer.getHistory', 'knowledge.getStats',
  'suggestions.list', 'commands.list',
];

let ok = 0, err500 = 0, other = 0;
const lines = [];

for (const proc of procs) {
  try {
    const res = await fetch(`http://127.0.0.1:${PORT}/trpc/${proc}`, { signal: AbortSignal.timeout(5000) });
    if (res.ok) { lines.push(`  ✓ ${proc}`); ok++; }
    else if (res.status === 500) {
      const body = await res.text();
      const match = body.match(/"message":"([^"]+)"/);
      lines.push(`  ⚠ ${proc} → 500: ${match ? match[1].substring(0,60) : 'unknown'}`);
      err500++;
    } else {
      lines.push(`  ? ${proc} → HTTP ${res.status}`);
      other++;
    }
  } catch (e) {
    lines.push(`  ✗ ${proc} → ${e.message?.substring(0, 40)}`);
    other++;
  }
}

lines.push('');
lines.push(`Summary: ${ok} ok, ${err500} runtime errors, ${other} other (${procs.length} total)`);
console.log(lines.join('\n'));
