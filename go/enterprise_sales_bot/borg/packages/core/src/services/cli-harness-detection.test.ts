import { describe, expect, test } from 'vitest';

import { detectCliHarnesses } from './cli-harness-detection.js';

describe('detectCliHarnesses', () => {
  test('returns truthful manual-install metadata for non-PATH harnesses', async () => {
    const results = await detectCliHarnesses([
      {
        id: 'antigravity',
        name: 'Antigravity',
        command: 'antigravity',
        args: [],
        homepage: 'https://antigravity.google/',
        docsUrl: 'https://antigravity.google/docs/home',
        installHint: 'Download and launch the desktop app.',
        category: 'editor',
        sessionCapable: false,
        versionArgs: [],
        detectionMode: 'manual',
      },
    ]);

    expect(results).toEqual([
      expect.objectContaining({
        id: 'antigravity',
        installed: false,
        resolvedPath: null,
        version: null,
        detectionMode: 'manual',
        detectionError: 'Manual install surface; no PATH-detectable command is currently modeled for this harness.',
      }),
    ]);
  });
});
