import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, describe, expect, it } from 'vitest';
import {
  applyPathCompletion,
  completeDirectoryPath,
  completeFilePath,
  shouldFinishFolderPrompt,
  validateExistingFilePath,
  validateExistingFolderPath,
} from '../src/cli/folder-prompt.js';

const tempDirs: string[] = [];

afterEach(async () => {
  while (tempDirs.length > 0) {
    await fs.rm(tempDirs.pop()!, { recursive: true, force: true });
  }
});

async function makeTempDir(): Promise<string> {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-folder-prompt-'));
  tempDirs.push(dir);
  return dir;
}

describe('folder prompt helpers', () => {
  it('completes existing directories and ignores files', async () => {
    const root = await makeTempDir();
    await fs.mkdir(path.join(root, 'Documents'));
    await fs.mkdir(path.join(root, 'Downloads'));
    await fs.writeFile(path.join(root, 'Dockerfile'), 'FROM scratch\n');

    const [matches] = completeDirectoryPath(`${root}${path.sep}Do`);

    expect(matches).toContain(`${root}${path.sep}Documents${path.sep}`);
    expect(matches).toContain(`${root}${path.sep}Downloads${path.sep}`);
    expect(matches).not.toContain(`${root}${path.sep}Dockerfile${path.sep}`);
  });

  it('applies unambiguous tab completions and common prefixes', () => {
    expect(applyPathCompletion('~/co', () => [['~/code/'], '~/co'])).toBe('~/code/');
    expect(applyPathCompletion('/tmp/pro', () => [['/tmp/project-a/', '/tmp/project-b/'], '/tmp/pro'])).toBe('/tmp/project-');
    expect(applyPathCompletion('/tmp/project-', () => [['/tmp/project-a/', '/tmp/project-b/'], '/tmp/project-'])).toBeUndefined();
    expect(applyPathCompletion('/tmp/none', () => [[], '/tmp/none'])).toBeUndefined();
  });

  it('completes files when the caller asks for a file path', async () => {
    const root = await makeTempDir();
    await fs.mkdir(path.join(root, 'cloudflared'));
    await fs.writeFile(path.join(root, 'config.yml'), 'tunnel: test\n');

    const [matches] = completeFilePath(`${root}${path.sep}c`);

    expect(matches).toContain(`${root}${path.sep}config.yml`);
    expect(matches).toContain(`${root}${path.sep}cloudflared${path.sep}`);
  });

  it('validates that a mount root exists and is a directory', async () => {
    const root = await makeTempDir();
    const file = path.join(root, 'note.txt');
    await fs.writeFile(file, 'hello\n');

    await expect(validateExistingFolderPath(root)).resolves.toBe(true);
    await expect(validateExistingFolderPath(file)).resolves.toBe('Mount root must be a folder, not a file');
    await expect(validateExistingFolderPath(path.join(root, 'missing'))).resolves.toContain('Folder not found');
  });

  it('validates that a path exists and is a file', async () => {
    const root = await makeTempDir();
    const file = path.join(root, 'config.yml');
    await fs.writeFile(file, 'tunnel: test\n');

    await expect(validateExistingFilePath(file)).resolves.toBe(true);
    await expect(validateExistingFilePath(root)).resolves.toBe('Path must be a file, not a folder');
    await expect(validateExistingFilePath(path.join(root, 'missing.yml'))).resolves.toContain('File not found');
  });

  it('allows Enter to finish only when the caller opts in', () => {
    expect(shouldFinishFolderPrompt('', { allowEmpty: true })).toBe(true);
    expect(shouldFinishFolderPrompt('   ', { allowEmpty: true })).toBe(true);
    expect(shouldFinishFolderPrompt('', { allowEmpty: false })).toBe(false);
    expect(shouldFinishFolderPrompt('~/Documents', { allowEmpty: true })).toBe(false);
  });
});
