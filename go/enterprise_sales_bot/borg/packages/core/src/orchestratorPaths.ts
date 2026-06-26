import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const MODULE_DIR = path.dirname(fileURLToPath(import.meta.url));

export function resolveMonorepoRoot(startDir: string): string | null {
  let current = path.resolve(startDir);
  const root = path.parse(current).root;

  while (true) {
    if (fs.existsSync(path.join(current, 'turbo.json'))) {
      return current;
    }

    if (current === root) {
      return null;
    }

    current = path.dirname(current);
  }
}

export function resolveSupervisorEntryPath(startDir: string = process.cwd()): string | null {
  const candidateRoots = [
    resolveMonorepoRoot(startDir),
    resolveMonorepoRoot(MODULE_DIR),
    path.resolve(MODULE_DIR, '../..'),
  ].filter((value, index, array): value is string => Boolean(value) && array.indexOf(value) === index);

  for (const root of candidateRoots) {
    const candidate = path.join(root, 'packages', 'tormentnexus-supervisor', 'dist', 'index.js');
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  return null;
}

export function resolveCliEntryPath(startDir: string = process.cwd()): string | null {
  const candidateRoots = [
    resolveMonorepoRoot(startDir),
    resolveMonorepoRoot(MODULE_DIR),
    path.resolve(MODULE_DIR, '../..'),
  ].filter((value, index, array): value is string => Boolean(value) && array.indexOf(value) === index);

  for (const root of candidateRoots) {
    const candidates = [
      path.join(root, 'packages', 'cli', 'dist', 'cli', 'src', 'index.js'),
      path.join(root, 'packages', 'cli', 'dist', 'index.js'),
    ];

    for (const candidate of candidates) {
      if (fs.existsSync(candidate)) {
        return candidate;
      }
    }
  }

  return null;
}
