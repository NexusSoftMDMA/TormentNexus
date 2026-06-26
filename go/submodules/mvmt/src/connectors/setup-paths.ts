import fs from 'fs/promises';
import os from 'os';
import path from 'path';

export async function pathExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

// Setup writes resolved paths into config so later runs do not depend on the
// caller's current working directory. `~` expansion still follows the user's
// home directory.
export function resolveSetupPath(inputPath: string): string {
  if (inputPath === '~') return os.homedir();
  if (inputPath.startsWith(`~${path.sep}`)) {
    return path.join(os.homedir(), inputPath.slice(2));
  }
  return path.resolve(inputPath);
}

// Executable prompts preserve relative inputs like `./python` because callers
// may intentionally rely on PATH lookup or a repo-local executable.
export function normalizeExecutableInput(value: string): string {
  const trimmed = value.trim();
  if (trimmed === '~' || trimmed.startsWith(`~${path.sep}`)) {
    return resolveSetupPath(trimmed);
  }
  return trimmed;
}
