export interface TerminalProgress {
  start(): void;
  stop(): void;
}

interface ProgressStream {
  isTTY?: boolean;
  columns?: number;
  write(chunk: string): unknown;
}

export interface TerminalProgressOptions {
  stream?: ProgressStream;
  frames?: string[];
  intervalMs?: number;
  enabled?: boolean;
}

export interface WithTerminalProgressOptions extends TerminalProgressOptions {
  progress?: TerminalProgress;
}

const DEFAULT_FRAMES = ['|', '/', '-', '\\'];
const DEFAULT_INTERVAL_MS = 120;
let activeClearLine: (() => void) | undefined;

export function createTerminalProgress(message: string, options: TerminalProgressOptions = {}): TerminalProgress {
  const stream = options.stream ?? process.stdout;
  const enabled = options.enabled ?? Boolean(stream.isTTY);
  const frames = options.frames?.length ? options.frames : DEFAULT_FRAMES;
  const intervalMs = options.intervalMs ?? DEFAULT_INTERVAL_MS;
  let timer: NodeJS.Timeout | undefined;
  let frameIndex = 0;
  let lastLine = '';
  const clearLine = () => {
    if (lastLine) {
      stream.write(`\r${' '.repeat(lastLine.length)}\r`);
      lastLine = '';
    }
  };

  const render = () => {
    lastLine = fitTerminalLine(`${message} ${frames[frameIndex % frames.length]}`, stream.columns);
    stream.write(`\r${lastLine}`);
    frameIndex += 1;
  };

  return {
    start() {
      if (!enabled || timer) return;
      render();
      activeClearLine = clearLine;
      timer = setInterval(render, intervalMs);
      timer.unref?.();
    },
    stop() {
      if (!enabled) return;
      if (timer) {
        clearInterval(timer);
        timer = undefined;
      }
      clearLine();
      if (activeClearLine === clearLine) activeClearLine = undefined;
    },
  };
}

export function clearActiveTerminalProgress(): void {
  activeClearLine?.();
}

export async function withTerminalProgress<T>(
  message: string,
  task: () => Promise<T>,
  options: WithTerminalProgressOptions = {},
): Promise<T> {
  const progress = options.progress ?? createTerminalProgress(message, options);
  progress.start();
  try {
    return await task();
  } finally {
    progress.stop();
  }
}

function fitTerminalLine(line: string, columns: number | undefined): string {
  if (!columns || columns <= 1 || line.length < columns) return line;
  return line.slice(0, columns - 1);
}
