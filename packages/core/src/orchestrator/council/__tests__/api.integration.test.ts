import { describe, test, expect, beforeAll, afterAll } from 'vitest';

const API_BASE = process.env.TEST_API_BASE || 'http://localhost:3847';
const describeApiIntegration = process.env.RUN_API_INTEGRATION_TESTS === '1' ? describe : describe.skip;

async function waitForServer(timeout = 10000): Promise<boolean> {
  const start = Date.now();
  while (Date.now() - start < timeout) {
    try {
      const res = await fetch(`${API_BASE}/health`);
      if (res.ok) return true;
    } catch {
      await new Promise(r => setTimeout(r, 500));
    }
  }
  return false;
}

describeApiIntegration('API Integration Tests', () => {
  beforeAll(async () => {
    const ready = await waitForServer();
    if (!ready) {
      throw new Error('Server not available. Set RUN_API_INTEGRATION_TESTS=1 and point TEST_API_BASE at a live TormentNexus API.');
    }
  });

  describe('Health & Root', () => {
    test('GET /health returns healthy status', async () => {
      const res = await fetch(`${API_BASE}/health`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(['healthy', 'degraded']).toContain(data.status);
      expect(data.uptime).toBeDefined();
      expect(data.supervisors).toBeDefined();
      expect(data.sessions).toBeDefined();
    });

    test('GET / returns server info', async () => {
      const res = await fetch(`${API_BASE}/`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.name).toBe('cli-orchestrator');
      expect(data.endpoints).toBeDefined();
      expect(data.endpoints.sessions).toBe('/api/sessions');
      expect(data.endpoints.council).toBe('/api/council');
      expect(data.endpoints.smartPilot).toBe('/api/smart-pilot');
    });
  });

  describe('Council API', () => {
    test('GET /api/council/status returns council status', async () => {
      const res = await fetch(`${API_BASE}/api/council/status`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(data.data).toBeDefined();
      expect(typeof data.data.enabled).toBe('boolean');
      expect(typeof data.data.supervisorCount).toBe('number');
    });

    test('POST /api/council/add-mock adds mock supervisor', async () => {
      const res = await fetch(`${API_BASE}/api/council/add-mock`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'TestMock', behavior: 'approve' }),
      });
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
    });

    test('POST /api/council/debate runs debate with mock', async () => {
      const res = await fetch(`${API_BASE}/api/council/debate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: 'test-task-1',
          description: 'Test task for integration testing',
          context: 'Integration test context',
          files: ['test.ts'],
        }),
      });
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(data.data).toBeDefined();
      expect(typeof data.data.approved).toBe('boolean');
      expect(typeof data.data.consensus).toBe('number');
      expect(Array.isArray(data.data.votes)).toBe(true);
    }, 30000);
  });

  describe('Sessions API', () => {
    test('GET /api/sessions returns sessions list', async () => {
      const res = await fetch(`${API_BASE}/api/sessions`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(Array.isArray(data.data)).toBe(true);
    });

    test('GET /api/sessions/active returns active sessions only', async () => {
      const res = await fetch(`${API_BASE}/api/sessions/active`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(Array.isArray(data.data)).toBe(true);
    });
  });

  describe('Smart Pilot API', () => {
    test('GET /api/smart-pilot/status returns status', async () => {
      const res = await fetch(`${API_BASE}/api/smart-pilot/status`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(typeof data.enabled).toBe('boolean');
      expect(data.config).toBeDefined();
      expect(typeof data.config.autoApproveThreshold).toBe('number');
      expect(typeof data.config.maxAutoApprovals).toBe('number');
    });

    test('POST /api/smart-pilot/config updates config', async () => {
      const res = await fetch(`${API_BASE}/api/smart-pilot/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ autoApproveThreshold: 0.9 }),
      });
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(data.config.autoApproveThreshold).toBe(0.9);
    });

    test('POST /api/smart-pilot/toggle toggles state', async () => {
      const statusBefore = await fetch(`${API_BASE}/api/smart-pilot/status`).then(r => r.json());
      
      const res = await fetch(`${API_BASE}/api/smart-pilot/toggle`, { method: 'POST' });
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(data.enabled).toBe(!statusBefore.enabled);
      
      await fetch(`${API_BASE}/api/smart-pilot/toggle`, { method: 'POST' });
    });
  });

  describe('Hooks API', () => {
    test('GET /api/hooks returns registered hooks', async () => {
      const res = await fetch(`${API_BASE}/api/hooks`);
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(Array.isArray(data.hooks)).toBe(true);
    });

    test('POST /api/hooks/register registers webhook', async () => {
      const res = await fetch(`${API_BASE}/api/hooks/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          phase: 'pre-debate',
          webhookUrl: 'http://localhost:9999/test-hook',
          priority: 10,
        }),
      });
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
      expect(data.hookId).toBeDefined();
      
      await fetch(`${API_BASE}/api/hooks/${data.hookId}`, { method: 'DELETE' });
    });

    test('DELETE /api/hooks/:id unregisters hook', async () => {
      const regRes = await fetch(`${API_BASE}/api/hooks/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          phase: 'post-debate',
          webhookUrl: 'http://localhost:9999/cleanup-hook',
        }),
      });
      const regData = await regRes.json();
      
      const res = await fetch(`${API_BASE}/api/hooks/${regData.hookId}`, { method: 'DELETE' });
      expect(res.ok).toBe(true);
      const data = await res.json();
      expect(data.success).toBe(true);
    });
  });
});
