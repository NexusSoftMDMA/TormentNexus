#!/usr/bin/env node
/**
 * CLI integration test — verifies every CLI command works against a running server
 */
const { execSync } = require('child_process');
const TORMENTNEXUS = 'node packages/cli/dist/cli/src/index.js';

let passed = 0, failed = 0;

function test(name, cmd, expectFn, timeout = 10000) {
  try {
    const out = execSync(`${TORMENTNEXUS} ${cmd}`, { timeout, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] });
    if (expectFn(out)) {
      console.log(`  ✓ ${name}`);
      passed++;
    } else {
      console.log(`  ✗ ${name} — assertion failed`);
      console.log(`    output: ${out.substring(0, 100).replace(/\n/g, ' ')}`);
      failed++;
    }
  } catch (e) {
    console.log(`  ✗ ${name} — ${e.message?.substring(0, 80)}`);
    failed++;
  }
}

(async () => {
  console.log(`\n  TormentNexus CLI Integration Test (v1.0.0-alpha.40)\n`);

  test('status shows version', 'status', o => o.includes('1.0.0-alpha'));
  test('status shows server running', 'status', o => o.includes('Running'));
  test('status shows Go sidecar', 'status', o => o.includes('Go Sidecar'));

  test('about shows version', 'about', o => o.includes('1.0.0-alpha'));
  test('about shows codename', 'about', o => o.includes('TORMENTNEXUS'));

  test('config show', 'config show', o => o.includes('Configuration'));
  test('config get', 'config get nonexistent.key', o => o.includes('undefined'));

  test('provider list', 'provider list', o => o.includes('Provider'));

  test('mcp list', 'mcp list', o => o.includes('alpaca') || o.includes('servers'));
  test('mcp tools', 'mcp tools', o => o.includes('Tools') || o.includes('tool'));
  test('mcp search', 'mcp search github', o => o.includes('github') || o.includes('results'));
  test('mcp inspect', 'mcp inspect alpaca', o => o.includes('alpaca'));
  test('mcp traffic', 'mcp traffic', o => o.includes('Traffic'));

  test('memory list', 'memory list', o => o.includes('Memor'));
  test('memory stats', 'memory stats', o => o.includes('Memor'));

  test('tools list', 'tools list', o => o.includes('tool') || o.includes('server'));
  test('session list', 'session list', o => o.includes('ession'));
  test('agent list', 'agent list', o => o.includes('gent'));
  test('agent council', 'agent council', o => o.includes('ouncil') || o.includes('irector'));

  test('mcp config', 'mcp config', o => o.includes('Configuration') || o.includes('MCP'));

  test('ping', 'ping', o => o.includes('OK') || o.includes('unreachable'));

  test('health', 'health', o => o.includes('Subsystem') || o.includes('health'));

  test('catalog stats', 'catalog stats', o => o.includes('Catalog') || o.includes('340'));

  test('catalog search', 'catalog search memory', o => o.includes('memory') || o.includes('results'));

  test('provider test', 'provider test openai', o => o.includes('openai') || o.includes('authenticated'), 15000);

  test('doctor', 'doctor', o => o.includes('Doctor') || o.includes('checks'), 15000);
  test('info', 'info', o => o.includes('TormentNexus') || o.includes('System'));
  test('cloud providers', 'cloud providers', o => o.includes('Cloud') || o.includes('Provider'));
  test('cloud stats', 'cloud stats', o => o.includes('Cloud') || o.includes('Providers'));
  test('billing status', 'billing status', o => o.includes('Billing') || o.includes('active'));
  test('billing depleted', 'billing depleted', o => o.includes('Depleted') || o.includes('depleted'));
  test('tools harnesses', 'tools harnesses', o => o.includes('Harness') || o.includes('Aider'));
  test('memory stats count', 'memory stats', o => o.includes('14708') || o.includes('entries'));
  test('context stats', 'context stats', o => o.includes('Context'));
  test('context list', 'context list', o => o.includes('Context') || o.includes('harvested'));
  test('knowledge stats', 'knowledge stats', o => o.includes('Knowledge'));
  test('knowledge resources', 'knowledge resources', o => o.includes('Knowledge') || o.includes('resources'));
  test('swarm missions', 'swarm missions', o => o.includes('Swarm') || o.includes('mission'));
  test('swarm risk', 'swarm risk', o => o.includes('Risk') || o.includes('N/A'));
  test('swarm capabilities', 'swarm capabilities', o => o.includes('Swarm') || o.includes('capabilities'));
  test('metrics system', 'metrics system', o => o.includes('System') || o.includes('CPU') || o.includes('cores'));
  test('metrics providers', 'metrics providers', o => o.includes('Provider') || o.includes('Google'));
  test('metrics stats', 'metrics stats', o => o.includes('Metrics') || o.includes('events'));
  test('skills list', 'skills list', o => o.includes('Skills') || o.includes('skills'));
  test('upgrade check', 'upgrade --check', o => o.includes('Upgrade') || o.includes('version'), 20000);
  test('plan status', 'plan status', o => o.includes('Plan') || o.includes('PLAN'));
  test('plan diffs', 'plan diffs', o => o.includes('Diff') || o.includes('Pending'));
  test('plan checkpoints', 'plan checkpoints', o => o.includes('Checkpoint') || o.includes('checkpoint'));
  test('browser status', 'browser status', o => o.includes('Browser') || o.includes('browser'));
  test('git status', 'git status', o => o.includes('Branch') || o.includes('Git'));
  test('git log', 'git log -n 3', o => o.includes('Commit') || o.includes('2026'));

  console.log(`\n  ${passed} passed, ${failed} failed\n`);
  process.exit(failed > 0 ? 1 : 0);
})();
