import { describe, expect, it } from 'vitest';

import {
  normalizeSubmodules,
  normalizeUserLinks,
  summarizeSubmoduleCounts,
} from './submodules/submodules-page-normalizers';

describe('submodules page normalizers', () => {
  it('normalizes malformed submodule and link payloads safely', () => {
    const submodules = normalizeSubmodules([
      {
        path: '  packages/core  ',
        url: '  https://example.com/core.git  ',
        status: 'clean',
        lastCommit: '  abcdef123  ',
      },
      {
        path: '',
        url: null,
        status: 'weird',
        version: 100,
      },
      null,
    ] as any);

    expect(submodules).toEqual([
      {
        path: 'packages/core',
        url: 'https://example.com/core.git',
        status: 'clean',
        lastCommit: 'abcdef123',
        lastCommitDate: undefined,
        lastCommitMessage: undefined,
        version: undefined,
        pkgName: undefined,
      },
      {
        path: 'unknown/submodule-2',
        url: 'unknown-url',
        status: 'error',
        lastCommit: undefined,
        lastCommitDate: undefined,
        lastCommitMessage: undefined,
        version: undefined,
        pkgName: undefined,
      },
      {
        path: 'unknown/submodule-3',
        url: 'unknown-url',
        status: 'error',
        lastCommit: undefined,
        lastCommitDate: undefined,
        lastCommitMessage: undefined,
        version: undefined,
        pkgName: undefined,
      },
    ]);

    const links = normalizeUserLinks([
      { name: '  Docs  ', links: [' https://a.dev ', '', 5] },
      { name: '', links: null },
      null,
    ] as any);

    expect(links).toEqual([
      { name: 'Docs', links: ['https://a.dev'] },
      { name: 'Category 2', links: [] },
      { name: 'Category 3', links: [] },
    ]);
  });

  it('computes safe summary counts from normalized inputs', () => {
    const summary = summarizeSubmoduleCounts(
      normalizeSubmodules([
        { path: 'a', url: 'u', status: 'clean' },
        { path: 'b', url: 'u', status: 'dirty' },
        { path: 'c', url: 'u', status: 'missing' },
        { path: 'd', url: 'u', status: 'error' },
      ] as any),
      normalizeUserLinks([
        { name: 'one', links: ['x', 'y'] },
        { name: 'two', links: ['z'] },
      ] as any),
    );

    expect(summary).toEqual({
      clean: 1,
      dirty: 1,
      missing: 1,
      error: 1,
      resources: 3,
    });
  });
});
