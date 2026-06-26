import { describe, expect, it } from 'vitest';

import { normalizeShellHistory } from './tools-page-normalizers';

describe('tools page normalizers', () => {
  it('normalizes malformed shell history rows into safe renderable values', () => {
    const rows = normalizeShellHistory([
      null,
      123,
      {
        id: ' cmd-1 ',
        cwd: ' /repo ',
        command: ' pnpm test ',
        duration: 120,
        exitCode: 0,
        outputSnippet: 'ok',
      },
      {
        id: '',
        cwd: 42,
        command: null,
        duration: 'fast',
        exitCode: '0',
        outputSnippet: 999,
      },
    ] as any);

    expect(rows).toEqual([
      {
        id: 'cmd-1',
        cwd: '/repo',
        command: 'pnpm test',
        duration: 120,
        exitCode: 0,
        outputSnippet: 'ok',
      },
      {
        id: 'history-3',
        cwd: '~',
        command: '(no command)',
        duration: null,
        exitCode: null,
        outputSnippet: '',
      },
    ]);
  });

  it('returns empty list when shell history payload is not an array', () => {
    expect(normalizeShellHistory({ bad: true } as any)).toEqual([]);
    expect(normalizeShellHistory(undefined as any)).toEqual([]);
  });
});
