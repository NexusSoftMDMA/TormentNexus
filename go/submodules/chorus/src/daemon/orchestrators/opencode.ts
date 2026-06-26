import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {
  DEFAULT_DAEMON_URL,
  type ConnectOpts,
  type ConnectResult,
  type OrchestratorDefinition,
  type OrchestratorStatus,
} from './shared.js';

const OPENCODE_USER_CONFIG_PATH = path.join(
  os.homedir(),
  '.config',
  'opencode',
  'opencode.json',
);

interface OpencodeConfig {
  mcp?: Record<string, unknown>;
  permission?: {
    edit?: 'ask' | 'allow' | 'deny';
    bash?: Record<string, 'ask' | 'allow' | 'deny'>;
    [key: string]: unknown;
  };
  [key: string]: unknown;
}

/**
 * Safe-bash allowlist for OpenCode reviewers. Read-only ops that
 * reviewers legitimately need to inspect a diff (git diff, ls, cat,
 * grep, find) — these shouldn't prompt every spawn. Destructive ops
 * (rm, git push, git reset) fall through to `*: ask` which preserves
 * the safety gate.
 *
 * Layer 1 of the three-layer permission model: prevent the dialog from
 * appearing in the first place. Layer 2 (shim.recoverKeys) catches
 * anything this misses.
 */
const OPENCODE_SAFE_BASH: Record<string, 'allow' | 'ask' | 'deny'> = {
  'git diff': 'allow',
  'git diff *': 'allow',
  'git status': 'allow',
  'git log': 'allow',
  'git log *': 'allow',
  'git show *': 'allow',
  'git blame *': 'allow',
  cat: 'allow',
  'cat *': 'allow',
  ls: 'allow',
  'ls *': 'allow',
  'find *': 'allow',
  'rg *': 'allow',
  'grep *': 'allow',
  'head *': 'allow',
  'tail *': 'allow',
  'wc *': 'allow',
  pwd: 'allow',
  'echo *': 'allow',
  '*': 'ask',
};

function readOpencodeConfig(): { config: OpencodeConfig; existed: boolean } {
  if (!fs.existsSync(OPENCODE_USER_CONFIG_PATH)) {
    return { config: {}, existed: false };
  }
  try {
    return {
      config: JSON.parse(
        fs.readFileSync(OPENCODE_USER_CONFIG_PATH, 'utf-8'),
      ) as OpencodeConfig,
      existed: true,
    };
  } catch {
    return { config: {}, existed: true };
  }
}

function getOpencodeStatus(): OrchestratorStatus {
  const detected = fs.existsSync(path.dirname(OPENCODE_USER_CONFIG_PATH));
  const { config } = readOpencodeConfig();
  const connected =
    detected &&
    Boolean(
      config.mcp && (config.mcp as Record<string, unknown>).chorus,
    );
  return {
    name: 'opencode',
    label: 'OpenCode',
    connected,
    approvedTools: connected ? 1 : 0,
    totalTools: 1,
    note: 'Registers Chorus as a local MCP server in ~/.config/opencode/opencode.json with `enabled: true` so the agent can call its tools.',
    supported: detected,
    firstCallBehavior: 'auto',
  };
}

/**
 * Register Chorus with OpenCode. `opencode mcp add` is interactive (no
 * flags for non-tty use), so we patch the user-scope config directly.
 * Idempotent.
 */
async function connectOpencode(
  opts: { binPath: string; daemonUrl?: string },
): Promise<ConnectResult> {
  const { config } = readOpencodeConfig();
  const mcp = (config.mcp ?? {}) as Record<string, unknown>;

  const existing = mcp.chorus as { command?: string[] } | undefined;
  const existingBinMatches =
    existing &&
    Array.isArray(existing.command) &&
    existing.command.includes(opts.binPath);

  // Bash pre-approval is independent of the MCP block — pre-approve
  // even if the MCP block is already correct, in case the user's
  // existing config doesn't have it. Idempotent merge: never overwrite
  // a user's explicit setting, only fill gaps.
  const existingPermission = (config.permission ?? {}) as NonNullable<
    OpencodeConfig['permission']
  >;
  const existingBash = (existingPermission.bash ?? {}) as Record<
    string,
    'allow' | 'ask' | 'deny'
  >;
  const mergedBash: Record<string, 'allow' | 'ask' | 'deny'> = {
    ...OPENCODE_SAFE_BASH,
  };
  // User's settings win over our defaults — preserves explicit deny rules.
  for (const [pattern, mode] of Object.entries(existingBash)) {
    mergedBash[pattern] = mode;
  }
  const bashChanged =
    JSON.stringify(existingBash) !== JSON.stringify(mergedBash);

  if (existingBinMatches && !bashChanged) {
    return {
      added: [],
      alreadyPresent: ['mcp.chorus', 'permission.bash'],
      configPath: OPENCODE_USER_CONFIG_PATH,
      slashCommand: 'skipped',
      slashCommandPath: '',
    };
  }

  const daemonUrl = opts.daemonUrl ?? DEFAULT_DAEMON_URL;
  mcp.chorus = {
    type: 'local',
    command: ['node', opts.binPath, 'mcp'],
    environment: { CHORUS_DAEMON_URL: daemonUrl },
    enabled: true,
  };

  const next: OpencodeConfig = {
    $schema: 'https://opencode.ai/config.json',
    ...config,
    mcp,
    permission: {
      ...existingPermission,
      bash: mergedBash,
    },
  };

  fs.mkdirSync(path.dirname(OPENCODE_USER_CONFIG_PATH), { recursive: true });
  fs.writeFileSync(
    OPENCODE_USER_CONFIG_PATH,
    JSON.stringify(next, null, 2) + '\n',
    'utf-8',
  );

  const added: string[] = [];
  if (!existingBinMatches) added.push('mcp.chorus');
  if (bashChanged) added.push('permission.bash');

  return {
    added,
    alreadyPresent: [],
    configPath: OPENCODE_USER_CONFIG_PATH,
    slashCommand: 'skipped',
    slashCommandPath: '',
  };
}

export const opencodeOrchestrator: OrchestratorDefinition = {
  name: 'opencode',
  label: 'OpenCode',
  getStatus: getOpencodeStatus,
  detect: () =>
    fs.existsSync(path.dirname(OPENCODE_USER_CONFIG_PATH)) ||
    fs.existsSync(path.join(os.homedir(), '.opencode')),
  connect: async (opts: ConnectOpts) => {
    const full = await connectOpencode(opts);
    return { registered: full.added.length > 0, toolsAdded: 0, full };
  },
};
