import { NextRequest, NextResponse } from 'next/server';
import { spawn } from 'child_process';
import path from 'path';

export const dynamic = 'force-dynamic';

const CLI_PATH = path.join(process.cwd(), '../../cli/mcp-router-cli/dist/mcp-router-cli.js');

interface CLIResult {
  success: boolean;
  result?: string;
  error?: string;
}

async function spawnCLI(args: string[]): Promise<CLIResult> {
  return new Promise((resolve) => {
    const proc = spawn('node', [CLI_PATH, ...args], {
      cwd: path.join(process.cwd(), '../../'),
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let stdout = '';
    let stderr = '';

    if (proc.stdout) {
      proc.stdout.on('data', (data: Buffer) => {
        stdout += data.toString();
      });
    }

    if (proc.stderr) {
      proc.stderr.on('data', (data: Buffer) => {
        stderr += data.toString();
      });
    }

    proc.on('close', (code) => {
      if (code === 0) {
        resolve({ success: true, result: stdout });
      } else {
        resolve({ success: false, error: stderr || `Process exited with code ${code}` });
      }
    });

    proc.on('error', (err) => {
      resolve({ success: false, error: err.message });
    });
  });
}

const argsMap: Record<string, string[]> = {
  'discover': ['discover', '--data-dir', './data', '--format', 'json'],
  'search': ['search', '$1', '--format', 'json'],
  'stats': ['stats', '--data-dir', './data'],
  'install': ['install', '$1', '--type', 'github', '--auto-start', 'true'],
  'uninstall': ['uninstall', '$1'],
  'check-updates': ['check-updates', '--format', 'json'],
  'update': ['update', '$1'],
  'health': ['health', '$1'],
  'detect-configs': ['detect-configs', '--recursive', 'false'],
  'import-configs': ['import-configs', ...('$2')],
  'export-configs': ['export-configs', '$1', '--format', '$2'],
  'init-sessions': ['init-sessions'],
  'session-stats': ['session-stats'],
  'list-sessions': ['list-sessions'],
  'start-session': ['start-session', '$1'],
  'stop-session': ['stop-session', '$1'],
  'restart-session': ['restart-session', '$1'],
  'shutdown-sessions': ['shutdown-sessions'],
  'session-metrics': ['session-metrics', '$1']
};

export async function POST(request: NextRequest, { params }: { params?: Promise<{ path?: string[] }> }) {
  const pathParams = await params;
  const path = pathParams?.path || [];
  const command = path[0];

  if (!command) {
    return NextResponse.json({ error: 'Command required' }, { status: 400 });
  }

  try {
    const body = await request.json();
    const resolvedArgs = resolveArgs(command, body);
    const result = await spawnCLI(resolvedArgs);
    return NextResponse.json(result);
  } catch (error) {
    return NextResponse.json({ success: false, error: 'Invalid JSON body' }, { status: 400 });
  }
}

function resolveArgs(command: string, body: any): string[] {
  switch (command) {
    case 'import-configs':
      return argsMap['import-configs'].map(arg => arg === '$1' ? body.query : arg === '$2' ? body.files : arg);
    case 'session-metrics':
      return argsMap['session-metrics'].map(arg => arg === '$1' ? body.serverName : arg);
    case 'search':
      return argsMap['search'].map(arg => arg === '$1' ? body.query : arg);
    case 'install':
    case 'uninstall':
    case 'check-updates':
    case 'update':
    case 'health':
    case 'start-session':
    case 'stop-session':
    case 'restart-session':
      return argsMap[command].map(arg => arg === '$1' ? body.name || body.serverId : arg);
    default:
      return argsMap[command];
  }
}

export async function GET(request: NextRequest) {
  return NextResponse.json({
    status: 'healthy',
    service: 'MCP Router API',
    endpoints: [
      { path: 'POST /api/mcp-router/discover', description: 'Discover servers' },
      { path: 'POST /api/mcp-router/search', description: 'Search servers' },
      { path: 'POST /api/mcp-router/install', description: 'Install server' },
      { path: 'POST /api/mcp-router/uninstall', description: 'Uninstall server' }
    ]
  });
}
