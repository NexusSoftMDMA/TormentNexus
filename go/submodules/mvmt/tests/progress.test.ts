import { afterEach, describe, expect, it, vi } from 'vitest';
import { clearActiveTerminalProgress, createTerminalProgress, withTerminalProgress } from '../src/cli/progress.js';

describe('terminal progress', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders and clears progress on TTY streams', () => {
    vi.useFakeTimers();
    const writes: string[] = [];
    const stream = {
      isTTY: true,
      columns: 80,
      write: vi.fn((chunk: string) => {
        writes.push(chunk);
        return true;
      }),
    };

    const progress = createTerminalProgress('Waiting for tunnel URL', {
      stream: stream as any,
      frames: ['|', '/'],
      intervalMs: 100,
    });

    progress.start();
    vi.advanceTimersByTime(100);
    progress.stop();

    expect(writes).toEqual([
      '\rWaiting for tunnel URL |',
      '\rWaiting for tunnel URL /',
      '\r' + ' '.repeat('Waiting for tunnel URL /'.length) + '\r',
    ]);
  });

  it('lets logger output clear an active progress line before printing', () => {
    const writes: string[] = [];
    const stream = {
      isTTY: true,
      columns: 80,
      write: vi.fn((chunk: string) => {
        writes.push(chunk);
        return true;
      }),
    };

    const progress = createTerminalProgress('Waiting', {
      stream: stream as any,
      frames: ['|'],
    });

    progress.start();
    clearActiveTerminalProgress();
    progress.stop();

    expect(writes).toEqual([
      '\rWaiting |',
      '\r' + ' '.repeat('Waiting |'.length) + '\r',
    ]);
  });

  it('does nothing on non-TTY streams', () => {
    const stream = {
      isTTY: false,
      write: vi.fn(),
    };

    const progress = createTerminalProgress('Waiting', { stream: stream as any });

    progress.start();
    progress.stop();

    expect(stream.write).not.toHaveBeenCalled();
  });

  it('truncates progress text to fit narrow terminals', () => {
    const stream = {
      isTTY: true,
      columns: 8,
      write: vi.fn(),
    };

    const progress = createTerminalProgress('Waiting for tunnel URL', {
      stream: stream as any,
      frames: ['|'],
    });

    progress.start();
    progress.stop();

    expect(stream.write).toHaveBeenNthCalledWith(1, '\rWaiting');
    expect(stream.write).toHaveBeenNthCalledWith(2, '\r' + ' '.repeat('Waiting'.length) + '\r');
  });

  it('stops progress after async success', async () => {
    const progress = {
      start: vi.fn(),
      stop: vi.fn(),
    };

    await expect(withTerminalProgress('Waiting', async () => 'ok', { progress })).resolves.toBe('ok');

    expect(progress.start).toHaveBeenCalledOnce();
    expect(progress.stop).toHaveBeenCalledOnce();
  });

  it('stops progress after async failure', async () => {
    const progress = {
      start: vi.fn(),
      stop: vi.fn(),
    };

    await expect(
      withTerminalProgress('Waiting', async () => {
        throw new Error('boom');
      }, { progress }),
    ).rejects.toThrow('boom');

    expect(progress.start).toHaveBeenCalledOnce();
    expect(progress.stop).toHaveBeenCalledOnce();
  });
});
