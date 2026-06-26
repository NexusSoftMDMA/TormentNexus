import chalk from 'chalk';
import { clearActiveTerminalProgress } from '../cli/progress.js';

let verbose = false;

export interface Logger {
  info(message: string): void;
  warn(message: string): void;
  error(message: string): void;
  debug(message: string): void;
}

export function createLogger(isVerbose: boolean): Logger {
  return {
    info(message: string): void {
      clearActiveTerminalProgress();
      console.log(message);
    },
    warn(message: string): void {
      clearActiveTerminalProgress();
      console.warn(chalk.yellow(message));
    },
    error(message: string): void {
      clearActiveTerminalProgress();
      console.error(chalk.red(message));
    },
    debug(message: string): void {
      if (isVerbose) {
        clearActiveTerminalProgress();
        console.error(chalk.dim(message));
      }
    },
  };
}

export function setVerbose(value: boolean): void {
  verbose = value;
}

export const log = {
  info(message: string): void {
    clearActiveTerminalProgress();
    console.log(message);
  },
  warn(message: string): void {
    clearActiveTerminalProgress();
    console.warn(chalk.yellow(message));
  },
  error(message: string): void {
    clearActiveTerminalProgress();
    console.error(chalk.red(message));
  },
  debug(message: string): void {
    if (verbose) {
      clearActiveTerminalProgress();
      console.error(chalk.dim(message));
    }
  },
};
