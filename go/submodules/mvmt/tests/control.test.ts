import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { describe, expect, it, vi } from 'vitest';
import { ControlAuthError, startJsonControlServer, streamJsonControl } from '../src/utils/control.js';

describe('control streaming', () => {
  it('restricts the Unix control socket to the current user', async () => {
    if (process.platform === 'win32') return;

    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-control-test-'));
    const socketPath = path.join(tmp, 'control.sock');
    const server = await startJsonControlServer(
      socketPath,
      () => undefined,
      { verifyToken: () => true },
    );

    try {
      expect(fs.statSync(socketPath).mode & 0o777).toBe(0o600);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('fails startup when Unix socket permissions cannot be hardened', async () => {
    if (process.platform === 'win32') return;

    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-control-test-'));
    const socketPath = path.join(tmp, 'control.sock');
    const chmod = vi.spyOn(fs, 'chmodSync').mockImplementationOnce(() => {
      throw new Error('chmod failed');
    });

    try {
      await expect(
        startJsonControlServer(socketPath, () => undefined, { verifyToken: () => true }),
      ).rejects.toThrow('chmod failed');
    } finally {
      chmod.mockRestore();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects stream startup on control-token auth failures instead of hanging', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-control-test-'));
    const socketPath = path.join(tmp, 'control.sock');
    const server = await startJsonControlServer(
      socketPath,
      () => undefined,
      { verifyToken: () => false },
    );

    try {
      await expect(
        streamJsonControl(socketPath, { type: 'tunnel.logs.stream' }, () => undefined, 'bad-token'),
      ).rejects.toBeInstanceOf(ControlAuthError);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});
