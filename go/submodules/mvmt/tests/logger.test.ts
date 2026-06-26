import { afterEach, describe, expect, it, vi } from 'vitest';
import { createLogger, log, setVerbose } from '../src/utils/logger.js';

describe('logger', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    setVerbose(false);
  });

  it('writes info, warnings, and errors', () => {
    const info = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    const error = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    log.info('hello');
    log.warn('careful');
    log.error('broken');

    expect(info).toHaveBeenCalledWith('hello');
    expect(warn).toHaveBeenCalledWith(expect.stringContaining('careful'));
    expect(error).toHaveBeenCalledWith(expect.stringContaining('broken'));
  });

  it('only writes debug output when verbose is enabled', () => {
    const error = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    log.debug('hidden');
    setVerbose(true);
    log.debug('shown');

    expect(error).toHaveBeenCalledTimes(1);
    expect(error).toHaveBeenCalledWith(expect.stringContaining('shown'));
  });

  it('creates independent verbose loggers', () => {
    const error = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const quiet = createLogger(false);
    const loud = createLogger(true);

    quiet.debug('hidden');
    loud.debug('shown');

    expect(error).toHaveBeenCalledTimes(1);
    expect(error).toHaveBeenCalledWith(expect.stringContaining('shown'));
  });
});
