import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { describe, expect, it } from 'vitest';
import { createTemporaryFilesystemConfig } from '../src/cli/config.js';

describe('createTemporaryFilesystemConfig', () => {
  it('creates a temporary read-only filesystem config without touching saved config', async () => {
    const result = await createTemporaryFilesystemConfig({
      paths: ['~/Documents', '/tmp/demo'],
      port: 4142,
    });

    try {
      expect(result.config.server).toMatchObject({
        port: 4142,
        access: 'local',
      });
      expect(result.config.proxy).toHaveLength(0);
      expect(result.config.mounts).toEqual([
        expect.objectContaining({
          name: 'documents',
          path: '/documents',
          root: path.join(os.homedir(), 'Documents'),
          writeAccess: false,
        }),
        expect.objectContaining({
          name: 'demo',
          path: '/demo',
          root: '/tmp/demo',
          writeAccess: false,
        }),
      ]);

      const raw = await fs.readFile(result.configPath, 'utf-8');
      expect(raw).toContain('mounts:');
      expect(raw).toContain('/tmp/demo');
    } finally {
      await result.cleanup();
    }
  });

  it('deduplicates paths after home expansion and resolution', async () => {
    const result = await createTemporaryFilesystemConfig({
      paths: ['~/Documents', path.join(os.homedir(), 'Documents')],
    });

    try {
      expect(result.config.mounts).toEqual([
        expect.objectContaining({
          name: 'documents',
          root: path.join(os.homedir(), 'Documents'),
        }),
      ]);
    } finally {
      await result.cleanup();
    }
  });

  it('keeps temporary mounts read-only', async () => {
    const result = await createTemporaryFilesystemConfig({
      paths: ['/tmp/demo'],
    });

    try {
      expect(result.config.mounts[0]).toMatchObject({
        name: 'demo',
        writeAccess: false,
      });
    } finally {
      await result.cleanup();
    }
  });

  it('removes the temporary config directory on cleanup', async () => {
    const result = await createTemporaryFilesystemConfig({
      paths: ['/tmp/demo'],
    });

    const tempDir = path.dirname(result.configPath);
    await result.cleanup();
    await expect(fs.access(tempDir)).rejects.toThrow();
  });
});
