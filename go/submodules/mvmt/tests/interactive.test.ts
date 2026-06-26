import { describe, expect, it, vi, beforeEach } from 'vitest';
import { 
  InteractiveAuditLogger, 
  formatAuditEntry, 
  formatHttpRequestEntry,
  isPromptCancelError,
  printAdvancedHelp,
  printInteractiveHelp,
  shouldShutdownOnSigint,
} from '../src/cli/interactive.js';
import { AuditLogger } from '../src/utils/audit.js';

describe('InteractiveAuditLogger', () => {
  let inner: AuditLogger;
  let logger: InteractiveAuditLogger;
  let output: string[] = [];

  beforeEach(() => {
    output = [];
    inner = { record: vi.fn(), recordHttp: vi.fn() };
    logger = new InteractiveAuditLogger(inner);
    logger.setWriter((msg) => output.push(msg));
  });

  it('forwards records to inner logger', () => {
    const entry = { ts: new Date().toISOString(), connectorId: 'test', tool: 'test', argKeys: [], durationMs: 10, isError: false };
    logger.record(entry);
    expect(inner.record).toHaveBeenCalledWith(entry);
  });

  it('writes to output when live logs are enabled', () => {
    const entry = { ts: new Date().toISOString(), connectorId: 'test', tool: 'test', argKeys: [], durationMs: 10, isError: false };
    logger.record(entry);
    expect(output.length).toBe(1);
    expect(output[0]).toContain('test test');
  });

  it('suppresses output when live logs are disabled', () => {
    logger.setLiveLogs(false);
    const entry = { ts: new Date().toISOString(), connectorId: 'test', tool: 'test', argKeys: [], durationMs: 10, isError: false };
    logger.record(entry);
    expect(output.length).toBe(0);
  });

  it('records HTTP entries to inner logger and writer', () => {
    const entry = { ts: new Date().toISOString(), kind: 'test', method: 'GET', path: '/test', status: 200 };
    logger.recordHttp(entry);
    expect(inner.recordHttp).toHaveBeenCalledWith(entry);
    expect(output.length).toBe(1);
    expect(output[0]).toContain('GET /test');
  });

});

describe('Formatting helpers', () => {
  it('formatAuditEntry returns expected shape', () => {
    const entry = { 
      ts: '2025-01-01T12:00:00Z', 
      connectorId: 'notes',
      tool: 'read_note', 
      argKeys: ['path'], 
      durationMs: 123, 
      isError: false 
    };
    const formatted = formatAuditEntry(entry);
    expect(formatted).toContain('OK');
    expect(formatted).toContain('notes read_note');
    expect(formatted).toContain('args=path');
    expect(formatted).toContain('123ms');
  });

  it('formatHttpRequestEntry returns expected shape', () => {
    const entry = {
      ts: '2025-01-01T12:00:00Z',
      kind: 'mcp.request',
      method: 'POST',
      path: '/mcp',
      status: 200,
      clientId: 'test-client',
      ip: '127.0.0.1',
      detail: 'test-detail'
    };
    const formatted = formatHttpRequestEntry(entry);
    expect(formatted).toContain('200');
    expect(formatted).toContain('mcp.request POST /mcp');
    expect(formatted).toContain('client=test-client');
    expect(formatted).toContain('from=127.0.0.1');
    expect(formatted).toContain('test-detail');
  });
});

describe('Interactive prompt control helpers', () => {
  it('requires a second Ctrl-C within the exit window', () => {
    expect(shouldShutdownOnSigint(0, 1000)).toBe(false);
    expect(shouldShutdownOnSigint(1000, 2500)).toBe(true);
    expect(shouldShutdownOnSigint(1000, 4001)).toBe(false);
  });

  it('detects Inquirer prompt cancellation errors', () => {
    const err = new Error('User force closed the prompt with SIGINT');
    err.name = 'ExitPromptError';

    expect(isPromptCancelError(err)).toBe(true);
    expect(isPromptCancelError(new Error('regular failure'))).toBe(false);
  });
});

describe('Interactive help', () => {
  it('presents leases as the primary surface', () => {
    const lines = captureConsole(() => printInteractiveHelp());

    expect(lines).toContain('  dashboard           show local/public dashboard URLs');
    expect(lines).toContain('  lease               list shared links');
    expect(lines).toContain('  lease create        create a shared link for a file or folder');
    expect(lines).toContain('  lease add-path      add files/folders to a shared link');
    expect(lines).toContain('  lease revoke        turn off a shared link');
    expect(lines).toContain('  users               list dashboard users');
    expect(lines).toContain('  users add           create dashboard user');
    expect(lines).toContain('  users grant/revoke  allow local source management');
    expect(lines).toContain('  url                 show dashboard and MCP URLs');
    expect(lines).toContain('  advanced            show source, token, and MCP commands');
    expect(lines).not.toContain('  token               list scoped API tokens');
    expect(lines).not.toContain('  mounts      list configured mounts');
    expect(lines).not.toContain('  share               list browser download links');
  });

  it('keeps implementation-level commands under advanced help', () => {
    const lines = captureConsole(() => printAdvancedHelp());

    expect(lines).toContain('  advanced mounts              list local sources');
    expect(lines).toContain('  advanced token               list scoped MCP/API tokens');
    expect(lines).toContain('  advanced connectors          list compatibility MCP connectors');
  });
});

function captureConsole(fn: () => void): string {
  const lines: string[] = [];
  const spy = vi.spyOn(console, 'log').mockImplementation((line = '') => {
    lines.push(String(line));
  });
  try {
    fn();
    return lines.join('\n');
  } finally {
    spy.mockRestore();
  }
}
