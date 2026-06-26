import { confirm } from '@inquirer/prompts';
import path from 'path';
import {
  DEFAULT_MOUNT_EXCLUDE_PATTERNS,
  DEFAULT_MOUNT_PROTECT_PATTERNS,
  LocalFolderMountConfig,
  MvmtConfig,
} from '../config/schema.js';
import type { ConnectorSetupDefinition } from './setup-registry.js';
import { resolveSetupPath } from './setup-paths.js';
import { promptForExistingFolder } from '../cli/folder-prompt.js';

export interface FilesystemConfigInput {
  paths: string[];
  writeAccess: boolean;
}

export function createFilesystemMountConfigs(filesystem: FilesystemConfigInput): LocalFolderMountConfig[] {
  const seen = new Map<string, number>();
  return filesystem.paths.map((root) => {
    const baseName = sanitizeMountName(path.basename(root) || 'folder');
    const count = seen.get(baseName) ?? 0;
    seen.set(baseName, count + 1);
    const name = count === 0 ? baseName : `${baseName}-${count + 1}`;
    return {
      name,
      type: 'local_folder',
      path: `/${name}`,
      root,
      description: '',
      guidance: '',
      exclude: [...DEFAULT_MOUNT_EXCLUDE_PATTERNS],
      protect: [...DEFAULT_MOUNT_PROTECT_PATTERNS],
      writeAccess: filesystem.writeAccess,
      enabled: true,
    };
  });
}

export async function promptForFilesystemFolders(): Promise<FilesystemConfigInput> {
  const wantFilesystem = await confirm({
    message: 'Expose specific local folders through mvmt?',
    default: false,
  });
  if (!wantFilesystem) return { paths: [], writeAccess: false };

  const folders = new Set<string>();

  while (true) {
    const folder = await promptForExistingFolder(
      folders.size === 0
        ? 'Folder on this computer:'
        : 'Another folder on this computer (Enter to finish):',
      { allowEmpty: folders.size > 0 },
    );
    if (!folder) break;

    folders.add(resolveSetupPath(folder.trim()));

    const addAnother = await confirm({
      message: 'Allow another folder?',
      default: false,
    });
    if (!addAnother) break;
  }

  const paths = [...folders].sort((a, b) => a.localeCompare(b));
  const writeAccess = await confirm({
    message: 'Allow filesystem write tools for these folders? Read-only is recommended.',
    default: false,
  });

  return { paths, writeAccess };
}

export const filesystemSetupDefinition = {
  id: 'filesystem',
  displayName: 'Filesystem',
  isAddable: false,
  async detect(): Promise<null> {
    return null;
  },
  async prompt(): Promise<FilesystemConfigInput | undefined> {
    const filesystem = await promptForFilesystemFolders();
    return filesystem.paths.length > 0 ? filesystem : undefined;
  },
  isConfigured(config: MvmtConfig): boolean {
    return config.mounts.some((mount) => mount.enabled !== false);
  },
  apply(config: MvmtConfig, input: FilesystemConfigInput): MvmtConfig {
    const mounts = createFilesystemMountConfigs(input);
    const replacedNames = new Set(mounts.map((mount) => mount.name));
    const replacedPaths = new Set(mounts.map((mount) => mount.path));
    return {
      ...config,
      mounts: [
        ...config.mounts.filter((mount) => !replacedNames.has(mount.name) && !replacedPaths.has(mount.path)),
        ...mounts,
      ],
    };
  },
} satisfies ConnectorSetupDefinition<null, FilesystemConfigInput, 'filesystem'>;

function sanitizeMountName(value: string): string {
  const chars: string[] = [];
  let pendingSeparator = false;

  for (const char of value.toLowerCase()) {
    if (isMountNameChar(char)) {
      if (pendingSeparator && chars.length > 0) chars.push('-');
      chars.push(char);
      pendingSeparator = false;
    } else {
      pendingSeparator = true;
    }
  }

  let start = 0;
  let end = chars.length;
  while (start < end && chars[start] === '-') start += 1;
  while (end > start && chars[end - 1] === '-') end -= 1;

  return chars.slice(start, end).join('') || 'folder';
}

function isMountNameChar(char: string): boolean {
  const code = char.charCodeAt(0);
  return (
    (code >= 97 && code <= 122)
    || (code >= 48 && code <= 57)
    || char === '_'
    || char === '-'
  );
}
