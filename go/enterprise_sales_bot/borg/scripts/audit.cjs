const http = require('http');

const PORT = process.argv[2] || '4000';
const procs = [
  'health', 'mcp.listServers', 'mcp.getStatus', 'mcp.getWorkingSet', 'mcp.listTools',
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

async function check(proc) {
  return new Promise((resolve) => {
    const req = http.get(`http://127.0.0.1:${PORT}/trpc/${proc}`, { timeout: 5000 }, (res) => {
      let body = '';
      res.on('data', (c) => body += c);
      res.on('end', () => {
        if (res.statusCode === 200) {
          lines.push(`  OK  ${proc}`);
          ok++;
        } else if (res.statusCode === 500) {
          const m = body.match(/"message":"([^"]+)"/);
          lines.push(`  500 ${proc}: ${m ? m[1].substring(0, 50) : 'unknown'}`);
          err500++;
        } else {
          lines.push(`  ${res.statusCode} ${proc}`);
          other++;
        }
        resolve();
      });
    });
    req.on('error', (e) => { lines.push(`  ERR ${proc}: ${e.message.substring(0, 30)}`); other++; resolve(); });
    req.on('timeout', () => { lines.push(`  TMO ${proc}`); other++; req.destroy(); resolve(); });
  });
}

(async () => {
  for (const p of procs) await check(p);
  lines.push('');
  lines.push(`Summary: ${ok} ok, ${err500} runtime errors, ${other} other (${procs.length} total)`);
  console.log(lines.join('\n'));
})();
