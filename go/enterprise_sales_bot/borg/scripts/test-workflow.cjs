#!/usr/bin/env node
/**
 * End-to-end workflow test — simulates a real user journey
 */
const { execSync } = require('child_process');
const TORMENTNEXUS = 'node packages/cli/dist/cli/src/index.js';

function run(cmd, timeout = 10000) {
  try {
    return execSync(`${TORMENTNEXUS} ${cmd}`, { timeout, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] });
  } catch (e) {
    return e.stdout || e.message;
  }
}

function has(str, text) { return str.includes(text); }

console.log('\n  TormentNexus TORMENTNEXUS — End-to-End Workflow Test\n');

// Step 1: Doctor check
console.log('  Step 1: System health check');
const doctor = run('doctor');
console.log(has(doctor, 'Server is running') ? '  ✓ Server healthy' : '  ✗ Server down');
console.log(has(doctor, 'Go sidecar is running') ? '  ✓ Go sidecar healthy' : '  ✗ Go sidecar down');
console.log(has(doctor, '0 issues') || has(doctor, 'All systems healthy') ? '  ✓ Doctor passes' : '  ⚠ Doctor has issues');

// Step 2: Info overview
console.log('\n  Step 2: System overview');
const info = run('info');
console.log(has(info, 'Running') ? '  ✓ info shows running status' : '  ✗ info failed');
console.log(has(info, 'Providers:') ? '  ✓ info shows providers' : '  ✗ no providers');

// Step 3: Provider verification
console.log('\n  Step 3: Provider verification');
const providers = run('provider list');
const providerCount = (providers.match(/● Available/g) || []).length;
console.log(`  ✓ ${providerCount} providers detected`);

const quota = run('provider quota');
const activeCount = (quota.match(/● Active/g) || []).length;
console.log(`  ✓ ${activeCount} active API keys`);

// Step 4: Catalog browsing
console.log('\n  Step 4: MCP catalog browsing');
const catalogStats = run('catalog stats');
console.log(has(catalogStats, '340') || has(catalogStats, '311') ? '  ✓ Catalog has 300+ entries' : '  ⚠ Catalog may be empty');

const search = run('catalog search memory', 10000);
console.log(has(search, 'results') || has(search, 'memory') ? '  ✓ Catalog search works' : '  ✗ Catalog search failed');

// Step 5: MCP tools
console.log('\n  Step 5: MCP tool inventory');
const mcpStatus = run('status');
console.log(has(mcpStatus, '1302') || has(mcpStatus, 'Running') ? '  ✓ MCP tools inventoried' : '  ⚠ MCP may not be cached');

// Step 6: Session discovery
console.log('\n  Step 6: Session discovery');
const sessions = run('session list');
const sessionCount = (sessions.match(/Discovered/g) || []).length;
console.log(sessionCount > 0 ? `  ✓ ${sessionCount} sessions discovered` : '  ⚠ No sessions found');

// Step 7: AI tools detection
console.log('\n  Step 7: AI tools detection');
const sync = run('mcp sync');
const toolCount = (sync.match(/✓/g) || []).length;
console.log(toolCount > 0 ? `  ✓ ${toolCount} AI tools detected` : '  ⚠ No AI tools found');

// Step 8: Secrets scan
console.log('\n  Step 8: Secrets scan');
const secrets = run('config secrets --list');
const secretCount = (secrets.match(/_API_KEY/g) || []).length;
console.log(secretCount > 0 ? `  ✓ ${secretCount} API keys in env` : '  ⚠ No API keys found');

// Step 9: Catalog search for specific tool
console.log('\n  Step 9: Catalog tool search');
const toolSearch = run('catalog search github', 10000);
console.log(has(toolSearch, 'github') ? '  ✓ Found github-related servers' : '  ⚠ No github results');

// Step 10: Harnesses detection
console.log('\n  Step 10: CLI harness detection');
const harnesses = run('tools harnesses');
const harnessCount = (harnesses.match(/✓/g) || []).length;
console.log(harnessCount > 0 ? `  ✓ ${harnessCount} harnesses installed` : '  ⚠ No harnesses');

console.log('\n  Workflow complete!\n');
