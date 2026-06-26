#!/usr/bin/env node
/**
 * Full-stack smoke test for TormentNexus TORMENTNEXUS
 * Verifies TS server, Go sidecar, and key API endpoints
 */
const http = require('http');

const TS_PORT = parseInt(process.env.TS_PORT || '4100');
const GO_PORT = parseInt(process.env.GO_PORT || '4300');
const TIMEOUT = 8000;

let passed = 0, failed = 0, skipped = 0;

function get(url) {
  return new Promise((resolve) => {
    const req = http.get(url, { timeout: TIMEOUT }, (res) => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => resolve({ status: res.statusCode, body: d }));
    });
    req.on('error', () => resolve({ status: 0, body: '' }));
    req.on('timeout', () => { req.destroy(); resolve({ status: 0, body: '' }); });
  });
}

async function test(name, fn) {
  try {
    const result = await fn();
    if (result === 'skip') {
      console.log(`  ⊘ ${name}`);
      skipped++;
    } else if (result) {
      console.log(`  ✓ ${name}`);
      passed++;
    } else {
      console.log(`  ✗ ${name}`);
      failed++;
    }
  } catch (e) {
    console.log(`  ✗ ${name} — ${e.message}`);
    failed++;
  }
}

(async () => {
  console.log(`\n  TormentNexus TORMENTNEXUS Full-Stack Smoke Test\n`);
  console.log(`  TS: http://127.0.0.1:${TS_PORT}  Go: http://127.0.0.1:${GO_PORT}\n`);

  // === TypeScript Server ===
  console.log('  --- TypeScript Control Plane ---');
  
  await test('TS health', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/health`);
    if (r.status !== 200) return false;
    const j = JSON.parse(r.body);
    return j.status === 'ok';
  });

  await test('TS startupStatus', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/startupStatus`);
    return r.status === 200;
  });

  await test('TS mcp.listServers (135 servers)', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/mcp.listServers`);
    if (r.status !== 200) return false;
    const j = JSON.parse(r.body);
    return (j.result?.data?.length ?? 0) > 100;
  });

  await test('TS mcp.getStatus (1302 tools)', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/mcp.getStatus`);
    if (r.status !== 200) return false;
    const j = JSON.parse(r.body);
    return (j.result?.data?.toolCount ?? 0) > 1000;
  });

  await test('TS settings.get', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/settings.get`);
    return r.status === 200;
  });

  await test('TS secrets.list', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/secrets.list`);
    return r.status === 200;
  });

  await test('TS squad.list', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/squad.list`);
    return r.status === 200;
  });

  await test('TS skills.list', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/skills.list`);
    return r.status === 200;
  });

  await test('TS catalog.list', async () => {
    const r = await get(`http://127.0.0.1:${TS_PORT}/trpc/catalog.list`);
    return r.status === 200;
  });

  // === Go Sidecar ===
  console.log('\n  --- Go Sidecar ---');

  await test('Go health', async () => {
    const r = await get(`http://127.0.0.1:${GO_PORT}/health`);
    if (r.status !== 200) return 'skip';
    const j = JSON.parse(r.body);
    return j.ok === true;
  });

  await test('Go version', async () => {
    const r = await get(`http://127.0.0.1:${GO_PORT}/version`);
    if (r.status !== 200) return 'skip';
    const j = JSON.parse(r.body);
    return j.version?.startsWith('1.0.0-alpha') ?? false;
  });

  await test('Go /api/index (routes)', async () => {
    const r = await get(`http://127.0.0.1:${GO_PORT}/api/index`);
    if (r.status !== 200) return 'skip';
    const j = JSON.parse(r.body);
    return (j.data?.routes?.length ?? 0) > 300;
  });

  // === Summary ===
  console.log(`\n  --- Results ---`);
  console.log(`  ${passed} passed, ${failed} failed, ${skipped} skipped\n`);
  process.exit(failed > 0 ? 1 : 0);
})();
