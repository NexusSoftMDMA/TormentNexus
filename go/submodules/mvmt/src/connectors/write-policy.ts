import { CallToolResult } from './types.js';

export interface ProxyWritePolicyConfig {
  name: string;
  command?: string;
  args?: readonly string[];
  url?: string;
  writeAccess?: boolean;
}

export function createProxyToolPolicy(config: ProxyWritePolicyConfig): (name: string) => boolean {
  const isFs = isFilesystemProxy(config);
  const readOnlyFilesystem = isFs && config.writeAccess !== true;
  const readOnlyGeneric = !isFs && config.writeAccess === false;

  return (name) => {
    if (readOnlyFilesystem && !isReadOnlyFilesystemTool(name)) return false;
    if (readOnlyGeneric && isLikelyWriteTool(name)) return false;
    return true;
  };
}

export function writeAccessDisabledResult(name: string, displayName: string): CallToolResult {
  return {
    content: [
      {
        type: 'text',
        text: `Error: write access is disabled for tool "${name}". Set writeAccess: true for proxy "${displayName}" in ~/.mvmt/config.yaml or re-run mvmt config setup to enable writes.`,
      },
    ],
    isError: true,
  };
}

function isFilesystemProxy(config: ProxyWritePolicyConfig): boolean {
  const fingerprint = [config.name, config.command, config.url, ...(config.args ?? [])]
    .filter((value): value is string => typeof value === 'string')
    .join(' ')
    .toLowerCase();
  return fingerprint.includes('filesystem');
}

function isReadOnlyFilesystemTool(name: string): boolean {
  return READ_ONLY_FILESYSTEM_TOOLS.has(name);
}

const READ_ONLY_FILESYSTEM_TOOLS = new Set([
  'directory_tree',
  'get_file_info',
  'list_allowed_directories',
  'list_directory',
  'read_file',
  'read_media_file',
  'read_multiple_files',
  'read_text_file',
  'search_files',
]);

const WRITE_TOOL_PREFIXES = [
  'write_',
  'edit_',
  'create_',
  'delete_',
  'remove_',
  'move_',
  'rename_',
  'append_',
  'update_',
  'patch_',
  'put_',
  'upsert_',
  'insert_',
  'drop_',
  'set_',
  'mkdir',
];

export function isLikelyWriteTool(name: string): boolean {
  const lower = name.toLowerCase();
  return WRITE_TOOL_PREFIXES.some((prefix) => lower.startsWith(prefix));
}
