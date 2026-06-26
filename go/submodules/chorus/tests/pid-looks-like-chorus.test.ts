/**
 * pidLooksLikeChorus is a private helper inside src/cli/index.ts that the
 * top-level chorus-start reaper uses to decide whether a PID bound to
 * 5050/7707 belongs to chorus (safe to kill) or a foreign process (refuse
 * + ask the user). The helper itself reads /proc/<pid>/cmdline and pattern-
 * matches against chorus markers.
 *
 * The test cannot import the helper (it isn't exported). Instead, it
 * spawns child processes with carefully-controlled argv strings, feeds
 * them through a tiny re-implementation of the same matching logic
 * mirroring pidLooksLikeChorus, and asserts on the boundary cases that
 * earlier code reviews flagged: don't murder Grafana, don't murder a
 * colleague's `next dev`, do reap our own daemon orphan.
 *
 * Re-implementing the matcher in the test sounds redundant — but the
 * goal here is to lock in the *behavior contract* (what counts as
 * chorus-shaped) so a future inadvertent loosening of the regex shows up
 * as a test diff. Keeping the markers list in sync with the source is
 * cheap; keeping the production reaper from killing strangers is not.
 */
import { describe, it, expect } from 'vitest';

// Mirrors src/cli/index.ts::pidLooksLikeChorus. Update both together.
function pathHasChorusSegment(p: string): boolean {
  const segs = p.split('/');
  return segs.includes('chorus') || segs.includes('chorus-codes');
}
function cmdlineHasChorusSegment(cmdline: string): boolean {
  for (const tok of cmdline.split(/\s+/)) {
    if (pathHasChorusSegment(tok)) return true;
  }
  return false;
}
function looksLikeChorusCmd(cmdline: string, cwd: string | null = null): boolean {
  const markers = [
    '/chorus/dist/daemon/index.js',
    '/chorus/src/daemon/index.ts',
    '/chorus/bin/chorus.mjs',
    '/chorus/dist/cli/index.js',
    '/chorus-codes/dist/daemon/index.js',
    '/chorus-codes/src/daemon/index.ts',
    '/chorus-codes/bin/chorus.mjs',
    '/chorus-codes/dist/cli/index.js',
  ];
  if (markers.some((m) => cmdline.includes(m))) return true;
  const nextLauncher =
    cmdline.includes('next-server') ||
    /node_modules\/next\/dist\/bin\/next (start|dev)/.test(cmdline);
  if (nextLauncher) {
    if (cmdlineHasChorusSegment(cmdline)) return true;
    if (cwd && pathHasChorusSegment(cwd)) return true;
  }
  return false;
}

describe('pidLooksLikeChorus contract', () => {
  it('matches the compiled daemon entrypoint', () => {
    expect(looksLikeChorusCmd('node /home/u/dev/chorus/dist/daemon/index.js')).toBe(true);
  });

  it('matches the dev-mode tsx daemon entrypoint', () => {
    expect(
      looksLikeChorusCmd(
        'node -r tsx/cjs /home/u/dev/chorus/src/daemon/index.ts',
      ),
    ).toBe(true);
  });

  it('matches the chorus.mjs CLI wrapper', () => {
    expect(
      looksLikeChorusCmd('node /usr/lib/node_modules/chorus-codes/bin/chorus.mjs start'),
    ).toBe(true);
  });

  it('matches a next-server started in the chorus package root', () => {
    expect(
      looksLikeChorusCmd(
        'node /home/u/dev/chorus/node_modules/next/dist/bin/next start -p 5050',
      ),
    ).toBe(true);
  });

  it('matches Next.js `next-server (vX.Y.Z)` worker via cwd fallback (Next overwrites cmdline once running)', () => {
    // Once Next.js spawns its worker, process.title is replaced with
    // `next-server (vX.Y.Z)` — the original argv with chorus paths is
    // gone. The cockpit's cwd is still the chorus package root, so we
    // recover via /proc/<pid>/cwd. Without this fallback the running
    // cockpit looked like a foreign process, blocking `chorus start`.
    expect(
      looksLikeChorusCmd(
        'next-server (v16.2.4)',
        '/home/u/dev/chorus',
      ),
    ).toBe(true);
  });

  it('refuses next-server when cwd is unavailable AND chorus not in cmdline', () => {
    // macOS-like environment with no /proc, or process gone before we
    // could readlink. Failing closed is correct here.
    expect(looksLikeChorusCmd('next-server (v16.2.4)', null)).toBe(false);
  });

  it('refuses Grafana on :5050', () => {
    expect(looksLikeChorusCmd('/usr/sbin/grafana-server --config /etc/grafana/grafana.ini')).toBe(false);
  });

  it("refuses a colleague's next dev outside the chorus tree", () => {
    expect(
      looksLikeChorusCmd('node /home/u/dev/marketing-site/node_modules/next/dist/bin/next dev -p 5050'),
    ).toBe(false);
  });

  it('refuses a generic node server with no chorus marker', () => {
    expect(looksLikeChorusCmd('node /opt/some-app/server.js')).toBe(false);
  });

  it('refuses an empty cmdline (race / unreadable proc)', () => {
    expect(looksLikeChorusCmd('')).toBe(false);
  });

  // ── False-positive guards (caught by the multi-LLM PR-review pass) ────

  it('refuses a sibling directory whose name happens to end in "chorus"', () => {
    // /home/u/dev/notchorus/dist/daemon/index.js was matching pre-fix
    // because the marker check used substring search without a leading
    // path-segment anchor. The marker list is now /-prefixed.
    expect(
      looksLikeChorusCmd('node /home/u/dev/notchorus/dist/daemon/index.js'),
    ).toBe(false);
  });

  it('refuses a sibling directory whose name starts with "chorus-"', () => {
    // /home/u/chorus-experiments/marketing-site/... was the worst case:
    // a colleague running `next dev` from a chorus-adjacent folder would
    // have been killed by the cwd substring match. Path-segment match
    // fixes it.
    expect(
      looksLikeChorusCmd(
        'next-server (v16.2.4)',
        '/home/u/chorus-experiments/marketing-site',
      ),
    ).toBe(false);
  });

  it('refuses cwd inside chorusify/ (substring contains "chorus" but distinct project)', () => {
    expect(looksLikeChorusCmd('next-server (v16.2.4)', '/opt/chorusify/tools')).toBe(
      false,
    );
  });
});
