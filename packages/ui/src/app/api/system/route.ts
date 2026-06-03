import { NextResponse } from 'next/server';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import fs from 'fs';

export const dynamic = 'force-dynamic';

const execAsync = promisify(exec);

type SyncStatus = 'synced' | 'behind' | 'ahead' | 'diverged' | 'unknown';

async function getSyncStatus(submodulePath: string, rootDir: string): Promise<SyncStatus> {
  try {
    const fullPath = path.join(rootDir, submodulePath);
    if (!fs.existsSync(fullPath)) return 'unknown';

    await execAsync('git fetch origin --quiet', { cwd: fullPath, timeout: 10000 });
    
    const { stdout: local } = await execAsync('git rev-parse HEAD', { cwd: fullPath });
    const { stdout: remote } = await execAsync('git rev-parse @{u}', { cwd: fullPath });
    
    if (local.trim() === remote.trim()) return 'synced';
    
    const { stdout: base } = await execAsync('git merge-base HEAD @{u}', { cwd: fullPath });
    
    if (base.trim() === local.trim()) return 'behind';
    if (base.trim() === remote.trim()) return 'ahead';
    return 'diverged';
  } catch {
    return 'unknown';
  }
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const includeSyncStatus = searchParams.get('syncStatus') === 'true';

  try {
    // Assuming the API is running from packages/ui, the root is two levels up
    // But in production/build, it might be different. 
    // We'll try to find the root by looking for pnpm-workspace.yaml
    
    let rootDir = process.cwd();
    while (rootDir !== path.parse(rootDir).root) {
      if (fs.existsSync(path.join(rootDir, 'pnpm-workspace.yaml'))) {
        break;
      }
      rootDir = path.dirname(rootDir);
    }

    // Get submodules status
    const { stdout: submoduleStatus } = await execAsync('git submodule status', { cwd: rootDir });
    
    const submodules = await Promise.all(submoduleStatus.trim().split('\n').map(async (line) => {
      const match = line.match(/^([+\-U ]?)([0-9a-f]+)\s+(.*?)(\s+\(.*\))?$/);
      if (!match) return null;
      
      const [, statusChar, commit, pathStr, describe] = match;
      const fullPath = path.join(rootDir, pathStr);
      
      let date = 'Unknown';
      let branch = 'Unknown';
      
      try {
        const { stdout: dateOut } = await execAsync('git log -1 --format=%cd', { cwd: fullPath });
        date = dateOut.trim();
        
        const { stdout: branchOut } = await execAsync('git rev-parse --abbrev-ref HEAD', { cwd: fullPath });
        branch = branchOut.trim();
      } catch (e) {
        console.error(`Failed to get info for ${pathStr}`, e);
      }

      return {
        name: path.basename(pathStr),
        path: pathStr,
        commit,
        branch,
        date,
        status: statusChar === '+' ? 'Modified' : statusChar === '-' ? 'Uninitialized' : statusChar === 'U' ? 'Conflict' : 'Clean',
        syncStatus: includeSyncStatus ? await getSyncStatus(pathStr, rootDir) : undefined
      };
    }));

    const structure = [
      { path: 'packages/core', description: 'Main Node.js Hub service (Backend), handles MCP connections, agents, and memory.' },
      { path: 'packages/ui', description: 'Next.js Web Dashboard (Frontend), provides the user interface for Jules and system management.' },
      { path: 'submodules/jules-app', description: 'Jules Application Logic, contains the core components and business logic for the Jules assistant.' },
      { path: 'mcp-hubs/tormentnexus', description: 'Meta-Orchestrator, handles complex routing and tool aggregation.' },
      { path: 'submodules/mcpenetes', description: 'Configuration Injector, manages client configurations for Claude Desktop and VSCode.' },
      { path: 'agents/', description: 'Autonomous Agent Definitions (JSON), defines the behavior and tools for different agents.' },
      { path: 'mcp-servers/', description: 'Managed Local MCP Servers, contains source code for local tools.' },
      { path: 'docs/', description: 'Project Documentation, includes architecture, guides, and strategy documents.' },
      { path: 'prompts/', description: 'System Prompts and Jailbreaks, stores the prompt templates used by agents.' },
      { path: 'skills/', description: 'Skill Definitions, markdown files defining specific capabilities for the LLM.' },
    ];

    const packageJsonPath = path.join(rootDir, 'package.json');
    let rootVersion = 'Unknown';
    try {
      const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));
      rootVersion = packageJson.version;
    } catch (e) {
      console.error('Failed to read package.json', e);
    }

    return NextResponse.json({
      submodules: submodules.filter(Boolean),
      structure,
      rootVersion
    });

  } catch (error) {
    console.error('System info error:', error);
    return NextResponse.json({ error: 'Failed to fetch system info' }, { status: 500 });
  }
}
