import fs from 'fs';
import os from 'os';
import path from 'path';
import readline from 'readline';
import { resolveSetupPath } from '../connectors/setup-paths.js';

export interface FolderPromptOptions {
  allowEmpty?: boolean;
  defaultValue?: string;
}

export async function promptForExistingFolder(
  message = 'Folder on this computer:',
  options: FolderPromptOptions = {},
): Promise<string> {
  while (true) {
    const answer = (await askWithPathCompletion(`${message} `, completeDirectoryPath, options.defaultValue)).trim();
    if (shouldFinishFolderPrompt(answer, options)) return '';
    const valid = await validateExistingFolderPath(answer);
    if (valid === true) return answer;
    console.error(valid);
  }
}

export async function promptForExistingFile(
  message = 'File on this computer:',
  options: FolderPromptOptions = {},
): Promise<string> {
  while (true) {
    const answer = (await askWithPathCompletion(`${message} `, completeFilePath, options.defaultValue)).trim();
    if (shouldFinishFolderPrompt(answer, options)) return '';
    const valid = await validateExistingFilePath(answer);
    if (valid === true) return answer;
    console.error(valid);
  }
}

export function shouldFinishFolderPrompt(answer: string, options: FolderPromptOptions = {}): boolean {
  return answer.trim() === '' && options.allowEmpty === true;
}

export async function validateExistingFolderPath(value: string): Promise<true | string> {
  const trimmed = value.trim();
  if (!trimmed) return 'Enter a folder path such as ~/Documents';

  try {
    const stats = await fs.promises.stat(resolveSetupPath(trimmed));
    return stats.isDirectory() ? true : 'Mount root must be a folder, not a file';
  } catch {
    return `Folder not found: ${trimmed}`;
  }
}

export async function validateExistingFilePath(value: string): Promise<true | string> {
  const trimmed = value.trim();
  if (!trimmed) return 'Enter a file path such as ~/.cloudflared/config.yml';

  try {
    const stats = await fs.promises.stat(resolveSetupPath(trimmed));
    return stats.isFile() ? true : 'Path must be a file, not a folder';
  } catch {
    return `File not found: ${trimmed}`;
  }
}

export function completeDirectoryPath(line: string): [string[], string] {
  return completeLocalPath(line, { includeFiles: false });
}

export function completeFilePath(line: string): [string[], string] {
  return completeLocalPath(line, { includeFiles: true });
}

export function applyPathCompletion(
  line: string,
  completer: (line: string) => [string[], string],
): string | undefined {
  const current = stripTabCharacters(line);
  const [matches] = completer(current);
  if (matches.length === 0) return undefined;
  if (matches.length === 1) return matches[0];
  const shared = commonPrefix(matches);
  return shared.length > current.length ? shared : undefined;
}

function completeLocalPath(line: string, options: { includeFiles: boolean }): [string[], string] {
  const { baseDir, prefix } = directorySearch(line);
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(baseDir, { withFileTypes: true });
  } catch {
    return [[], line];
  }

  const matches = entries
    .filter((entry) => entry.isDirectory() || options.includeFiles)
    .filter((entry) => prefix.startsWith('.') || !entry.name.startsWith('.'))
    .filter((entry) => entry.name.toLowerCase().startsWith(prefix.toLowerCase()))
    .sort((a, b) => a.name.localeCompare(b.name))
    .map((entry) => formatCompletion(path.join(baseDir, entry.name), line, entry.isDirectory()));

  return [matches, line];
}

function askWithPathCompletion(
  message: string,
  completer: (line: string) => [string[], string],
  defaultValue?: string,
): Promise<string> {
  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    return questionWithoutCompletion(message, defaultValue);
  }
  return questionWithRawPathCompletion(message.trimEnd(), completer, defaultValue);
}

function completeCurrentReadlineValue(
  line: string,
  completer: (line: string) => [string[], string],
): string {
  const current = stripTabCharacters(line);
  return applyPathCompletion(current, completer) ?? current;
}

function questionWithRawPathCompletion(
  message: string,
  completer: (line: string) => [string[], string],
  defaultValue = '',
): Promise<string> {
  return new Promise((resolve, reject) => {
    const input = process.stdin;
    const output = process.stdout;
    const previousRawMode = input.isRaw;
    let value = '';
    let settled = false;

    const render = () => {
      readline.clearLine(output, 0);
      readline.cursorTo(output, 0);
      const defaultHint = defaultValue && !value ? ` (${defaultValue})` : '';
      output.write(`? ${message}${defaultHint} ${value}`);
    };

    const cleanup = () => {
      input.off('data', onData);
      if (input.isTTY) input.setRawMode(previousRawMode);
    };

    const finish = (callback: () => void) => {
      if (settled) return;
      settled = true;
      cleanup();
      output.write('\n');
      callback();
    };

    const onData = (chunk: Buffer) => {
      for (const char of chunk.toString('utf-8')) {
        if (char === '\u0003') {
          finish(() => reject(new Error('User force closed the prompt with SIGINT')));
          return;
        }
        if (char === '\r' || char === '\n') {
          finish(() => resolve(value || defaultValue || ''));
          return;
        }
        if (char === '\t') {
          value = completeCurrentReadlineValue(value, completer);
          render();
          continue;
        }
        if (char === '\u007f' || char === '\b') {
          value = value.slice(0, -1);
          render();
          continue;
        }
        if (char === '\u0015') {
          value = '';
          render();
          continue;
        }
        if (char >= ' ') {
          value += char;
          render();
        }
      }
    };

    input.setRawMode(true);
    input.resume();
    input.on('data', onData);
    render();
  });
}

function questionWithoutCompletion(message: string, defaultValue?: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout, terminal: false });
    rl.once('SIGINT', () => {
      rl.close();
      reject(new Error('User force closed the prompt with SIGINT'));
    });
    rl.question(message, (answer) => {
      rl.close();
      resolve(answer || defaultValue || '');
    });
  });
}

function directorySearch(line: string): { baseDir: string; prefix: string } {
  if (!line) return { baseDir: os.homedir(), prefix: '' };

  const expanded = expandHomePath(line);
  if (endsWithPathSeparator(line) || isDirectorySync(expanded)) {
    return { baseDir: expanded, prefix: '' };
  }

  return {
    baseDir: path.dirname(expanded),
    prefix: path.basename(expanded),
  };
}

function formatCompletion(absolutePath: string, originalInput: string, isDirectory: boolean): string {
  const completedPath = isDirectory ? `${absolutePath}${path.sep}` : absolutePath;
  if (originalInput.startsWith('~/') || originalInput === '~') {
    return collapseHomePath(completedPath);
  }
  if (path.isAbsolute(originalInput)) {
    return completedPath;
  }

  const relative = path.relative(process.cwd(), completedPath);
  const display = relative.startsWith('..') ? relative : `.${path.sep}${relative}`;
  return isDirectory && !display.endsWith(path.sep) ? `${display}${path.sep}` : display;
}

function expandHomePath(value: string): string {
  const trimmed = value.trim();
  if (trimmed === '~') return os.homedir();
  if (trimmed.startsWith('~/')) return path.join(os.homedir(), trimmed.slice(2));
  return path.resolve(trimmed || os.homedir());
}

function collapseHomePath(absolutePath: string): string {
  const home = os.homedir();
  if (absolutePath === home) return '~';
  if (absolutePath.startsWith(`${home}${path.sep}`)) {
    return `~/${path.relative(home, absolutePath)}`;
  }
  return absolutePath;
}

function endsWithPathSeparator(value: string): boolean {
  return value.endsWith('/') || value.endsWith(path.sep);
}

function isDirectorySync(candidate: string): boolean {
  try {
    return fs.statSync(candidate).isDirectory();
  } catch {
    return false;
  }
}

function commonPrefix(values: string[]): string {
  if (values.length === 0) return '';
  let prefix = values[0];
  for (const value of values.slice(1)) {
    while (!value.startsWith(prefix)) {
      prefix = prefix.slice(0, -1);
      if (!prefix) return '';
    }
  }
  return prefix;
}

function stripTabCharacters(value: string): string {
  return value.replaceAll('\t', '');
}
