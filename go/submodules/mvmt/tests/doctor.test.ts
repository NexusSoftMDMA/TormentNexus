import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { collectDoctorReport, doctor, printDoctorReport } from '../src/cli/doctor.js';

afterEach(() => {
  process.exitCode = undefined;
  vi.restoreAllMocks();
});

describe('collectDoctorReport', () => {
  it('validates config and checks enabled mounts', async () => {
    const root = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-doctor-mount-'));
    const configPath = await writeConfig([
      'version: 1',
      'server:',
      '  port: 4141',
      'proxy: []',
      'mounts:',
      '  - name: workspace',
      '    type: local_folder',
      '    path: /workspace',
      `    root: ${root}`,
    ]);

    const report = await collectDoctorReport({ config: configPath, updateCheck: false });

    expect(report.ok).toBe(true);
    expect(report.config).toMatchObject({
      exists: true,
      valid: true,
      server: { port: 4141, allowedOrigins: [] },
    });
    expect(report.connectors).toEqual([
      expect.objectContaining({
        name: 'workspace',
        id: '/workspace',
        kind: 'mount',
        enabled: true,
        status: 'ok',
      }),
    ]);
  });

  it('reports invalid config without trying connector health checks', async () => {
    const configPath = await writeConfig([
      'version: 1',
      'proxy:',
      '  - name: bad',
      '    transport: stdio',
    ]);

    const report = await collectDoctorReport({ config: configPath, updateCheck: false });

    expect(report.ok).toBe(false);
    expect(report.config.valid).toBe(false);
    expect(report.config.errors.join('\n')).toContain('stdio transport requires "command"');
    expect(report.connectors).toEqual([]);
  });

  it('reports missing config files', async () => {
    const report = await collectDoctorReport({ config: '/tmp/mvmt-missing-config.yaml', updateCheck: false });

    expect(report.ok).toBe(false);
    expect(report.config.exists).toBe(false);
    expect(report.config.errors.join('\n')).toContain('Config file does not exist');
  });

  it('reports unavailable enabled mounts', async () => {
    const root = path.join(await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-doctor-missing-')), 'missing');
    const configPath = await writeConfig([
      'version: 1',
      'proxy: []',
      'mounts:',
      '  - name: missing',
      '    type: local_folder',
      '    path: /missing',
      `    root: ${root}`,
    ]);

    const report = await collectDoctorReport({ config: configPath, updateCheck: false });

    expect(report.ok).toBe(false);
    expect(report.connectors).toEqual([
      expect.objectContaining({
        name: 'missing',
        kind: 'mount',
        status: 'fail',
        message: expect.stringContaining('missing'),
      }),
    ]);
  });
});

describe('doctor output', () => {
  it('prints JSON and sets process exit status', async () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const root = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-doctor-mount-'));
    const configPath = await writeConfig([
      'version: 1',
      'proxy: []',
      'mounts:',
      '  - name: workspace',
      '    type: local_folder',
      '    path: /workspace',
      `    root: ${root}`,
    ]);

    await doctor({ config: configPath, json: true, updateCheck: false });

    expect(output).toHaveBeenCalledWith(expect.stringContaining('"ok": true'));
    expect(process.exitCode).toBe(0);
  });

  it('prints human diagnostics', async () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const root = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-doctor-mount-'));
    const configPath = await writeConfig([
      'version: 1',
      'proxy: []',
      'mounts:',
      '  - name: workspace',
      '    type: local_folder',
      '    path: /workspace',
      `    root: ${root}`,
    ]);
    const report = await collectDoctorReport({ config: configPath, updateCheck: false });

    printDoctorReport(report);

    const text = output.mock.calls.map((call) => call.join(' ')).join('\n');
    expect(text).toContain('mvmt doctor');
    expect(text).toContain('Config');
    expect(text).toContain('workspace');
  });
});

async function writeConfig(lines: string[]): Promise<string> {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-doctor-'));
  const configPath = path.join(dir, 'config.yaml');
  await fs.writeFile(configPath, `${lines.join('\n')}\n`, 'utf-8');
  if (process.platform !== 'win32') {
    await fs.chmod(configPath, 0o600);
  }
  return configPath;
}
