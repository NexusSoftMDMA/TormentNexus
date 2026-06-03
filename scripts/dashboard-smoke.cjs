#!/usr/bin/env node
const http = require('http');

const BASE_URL = process.env.DASHBOARD_URL || 'http://127.0.0.1:3000';
const TIMEOUT_MS = Number.parseInt(process.env.DASHBOARD_TIMEOUT_MS || '12000', 10);
const ROUTES = [
  '/dashboard',
  '/dashboard/health',
  '/dashboard/mcp',
  '/dashboard/mcp/inspector',
  '/dashboard/mcp/search',
  '/dashboard/mcp/settings',
  '/dashboard/billing',
  '/dashboard/memory',
  '/dashboard/swarm',
  '/dashboard/session',
  '/dashboard/integrations',
  '/dashboard/knowledge',
  '/dashboard/council',
  '/dashboard/director',
  '/dashboard/skills',
  '/dashboard/library',
  '/dashboard/submodules',
  '/dashboard/workflows',
  '/dashboard/metrics',
  '/dashboard/squads',
];

let passed = 0;
let failed = 0;
let skipped = 0;

function logResult(symbol, message) {
  console.log(`  ${symbol} ${message}`);
}

function get(url) {
  return new Promise((resolve) => {
    const req = http.get(url, { timeout: TIMEOUT_MS }, (res) => {
      let body = '';
      res.on('data', (chunk) => {
        body += chunk;
      });
      res.on('end', () => {
        resolve({ status: res.statusCode ?? 0, body, headers: res.headers });
      });
    });
    req.on('error', (error) => resolve({ status: 0, body: '', error }));
    req.on('timeout', () => {
      req.destroy();
      resolve({ status: 0, body: '', error: new Error('timeout') });
    });
  });
}

async function test(name, fn) {
  try {
    const result = await fn();
    if (result === 'skip') {
      skipped += 1;
      logResult('⊘', name);
      return;
    }

    if (result) {
      passed += 1;
      logResult('✓', name);
      return;
    }

    failed += 1;
    logResult('✗', name);
  } catch (error) {
    failed += 1;
    logResult('✗', `${name} — ${error instanceof Error ? error.message : String(error)}`);
  }
}

async function runHttpChecks() {
  console.log('  --- HTTP route checks ---');
  for (const route of ROUTES) {
    await test(`GET ${route}`, async () => {
      const response = await get(`${BASE_URL}${route}`);
      return response.status === 200 && /text\/html/i.test(String(response.headers?.['content-type'] || ''));
    });
  }
}

async function runBrowserChecks() {
  let chromium;
  try {
    ({ chromium } = require('playwright'));
  } catch {
    return 'skip';
  }

  console.log('\n  --- Browser hydration checks ---');
  const browser = await chromium.launch({ headless: true });
  try {
    for (const route of ROUTES) {
      await test(`Hydrate ${route}`, async () => {
        const page = await browser.newPage();
        const failures = [];
        page.on('response', (response) => {
          if (response.status() >= 500) {
            failures.push({ status: response.status(), url: response.url() });
          }
        });

        try {
          await page.goto(`${BASE_URL}${route}`, { waitUntil: 'domcontentloaded', timeout: TIMEOUT_MS });
          await page.waitForTimeout(1500);
          const bodyText = await page.locator('body').innerText();
          return failures.length === 0 && bodyText.trim().length > 120;
        } finally {
          await page.close();
        }
      });
    }
  } finally {
    await browser.close();
  }

  return true;
}

(async () => {
  console.log('\n  TormentNexus Dashboard Smoke Test\n');
  console.log(`  Base URL: ${BASE_URL}\n`);

  await runHttpChecks();
  const browserResult = await runBrowserChecks();
  if (browserResult === 'skip') {
    skipped += 1;
    logResult('⊘', 'Browser hydration checks (playwright not available)');
  }

  console.log(`\n  --- Results ---`);
  console.log(`  ${passed} passed, ${failed} failed, ${skipped} skipped\n`);
  process.exit(failed > 0 ? 1 : 0);
})();
