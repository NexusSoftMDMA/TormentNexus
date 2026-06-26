import { describe, expect, it } from 'vitest';

import { normalizeSavedScripts } from './scripts-page-normalizers';

describe('scripts page normalizers', () => {
  it('normalizes malformed script payload rows into safe renderable values', () => {
    const rows = normalizeSavedScripts([
      null,
      123,
      {
        uuid: ' script-1 ',
        name: ' Daily Cleanup ',
        description: 42,
        code: 'console.log("ok")',
      },
      {
        name: '',
        description: ' desc ',
        code: null,
      },
    ] as any);

    expect(rows).toEqual([
      {
        uuid: 'script-1',
        name: 'Daily Cleanup',
        description: '',
        code: 'console.log("ok")',
      },
      {
        uuid: 'script-3',
        name: 'Unnamed script',
        description: 'desc',
        code: '',
      },
    ]);
  });

  it('returns empty array when scripts payload is not an array', () => {
    expect(normalizeSavedScripts({ bad: true } as any)).toEqual([]);
    expect(normalizeSavedScripts(undefined as any)).toEqual([]);
  });
});
