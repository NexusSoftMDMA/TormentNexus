/**
 * Cursor + Windsurf are IDE forks of VSCode that share the
 * `{ mcpServers: { ... } }` JSON shape (same as Gemini/Kimi). Neither
 * ships an `mcp add` CLI — direct JSON edit. Detection looks for the
 * config dir (Cursor/Windsurf is installed); the mcp.json itself
 * usually doesn't exist until first registration.
 */

import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {
  DEFAULT_DAEMON_URL,
  hasMcpEntry,
  writeMcpEntry,
  type ConnectOpts,
  type ConnectResult,
  type OrchestratorDefinition,
  type OrchestratorStatus,
} from './shared.js';

const CURSOR_CONFIG_DIR = path.join(os.homedir(), '.cursor');
const CURSOR_MCP_PATH = path.join(CURSOR_CONFIG_DIR, 'mcp.json');

const WINDSURF_CONFIG_DIR = path.join(os.homedir(), '.codeium', 'windsurf');
const WINDSURF_MCP_PATH = path.join(
  WINDSURF_CONFIG_DIR,
  'mcp_config.json',
);

function getCursorStatus(): OrchestratorStatus {
  const detected = fs.existsSync(CURSOR_CONFIG_DIR);
  const connected = detected && hasMcpEntry(CURSOR_MCP_PATH);
  return {
    name: 'cursor',
    label: 'Cursor',
    connected,
    approvedTools: connected ? 1 : 0,
    totalTools: 1,
    note: 'Registers Chorus in ~/.cursor/mcp.json. Cursor will prompt on first tool call — choose "Always allow" to make it stick.',
    supported: detected,
    firstCallBehavior: 'prompts_once',
  };
}

async function connectCursor(
  opts: { binPath: string; daemonUrl?: string },
): Promise<ConnectResult> {
  if (hasMcpEntry(CURSOR_MCP_PATH, opts.binPath)) {
    return {
      added: [],
      alreadyPresent: ['mcpServers.chorus'],
      configPath: CURSOR_MCP_PATH,
      slashCommand: 'skipped',
      slashCommandPath: '',
    };
  }

  writeMcpEntry({
    filePath: CURSOR_MCP_PATH,
    binPath: opts.binPath,
    daemonUrl: opts.daemonUrl ?? DEFAULT_DAEMON_URL,
  });

  return {
    added: ['mcpServers.chorus'],
    alreadyPresent: [],
    configPath: CURSOR_MCP_PATH,
    slashCommand: 'skipped',
    slashCommandPath: '',
  };
}

function getWindsurfStatus(): OrchestratorStatus {
  const detected = fs.existsSync(WINDSURF_CONFIG_DIR);
  const connected = detected && hasMcpEntry(WINDSURF_MCP_PATH);
  return {
    name: 'windsurf',
    label: 'Windsurf',
    connected,
    approvedTools: connected ? 1 : 0,
    totalTools: 1,
    note: 'Registers Chorus in ~/.codeium/windsurf/mcp_config.json. Windsurf will prompt on first tool call — choose "Always allow".',
    supported: detected,
    firstCallBehavior: 'prompts_once',
  };
}

async function connectWindsurf(
  opts: { binPath: string; daemonUrl?: string },
): Promise<ConnectResult> {
  if (hasMcpEntry(WINDSURF_MCP_PATH, opts.binPath)) {
    return {
      added: [],
      alreadyPresent: ['mcpServers.chorus'],
      configPath: WINDSURF_MCP_PATH,
      slashCommand: 'skipped',
      slashCommandPath: '',
    };
  }

  writeMcpEntry({
    filePath: WINDSURF_MCP_PATH,
    binPath: opts.binPath,
    daemonUrl: opts.daemonUrl ?? DEFAULT_DAEMON_URL,
  });

  return {
    added: ['mcpServers.chorus'],
    alreadyPresent: [],
    configPath: WINDSURF_MCP_PATH,
    slashCommand: 'skipped',
    slashCommandPath: '',
  };
}

export const cursorOrchestrator: OrchestratorDefinition = {
  name: 'cursor',
  label: 'Cursor',
  getStatus: getCursorStatus,
  detect: () => fs.existsSync(CURSOR_CONFIG_DIR),
  connect: async (opts: ConnectOpts) => {
    const before = hasMcpEntry(CURSOR_MCP_PATH, opts.binPath);
    const full = await connectCursor(opts);
    return { registered: !before, toolsAdded: 0, full };
  },
};

export const windsurfOrchestrator: OrchestratorDefinition = {
  name: 'windsurf',
  label: 'Windsurf',
  getStatus: getWindsurfStatus,
  detect: () => fs.existsSync(WINDSURF_CONFIG_DIR),
  connect: async (opts: ConnectOpts) => {
    const before = hasMcpEntry(WINDSURF_MCP_PATH, opts.binPath);
    const full = await connectWindsurf(opts);
    return { registered: !before, toolsAdded: 0, full };
  },
};
