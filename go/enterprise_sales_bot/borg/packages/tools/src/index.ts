// @tormentnexus/tools — runtime-safe class stubs with full surface area
// Provides class stubs that @tormentnexus/core uses as both types and values.
// Real implementations live in @tormentnexus/core runtime.

// Tool shape for iterable tool groups
export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

// Helper: creates a tool definition
function defTool(name: string, description: string, handler: (args: any) => Promise<any>): ToolDefinition {
  return { name, description, inputSchema: { type: 'object', properties: {} }, handler };
}

// ── Classes (used with `new` in @tormentnexus/core) ────────────────────────────────

export class BrowserTool {
  async execute(_input: any): Promise<any> { return { result: '[BrowserTool stub]' }; }
  async executeTask(_task: string, _autonomous?: boolean): Promise<any> { return { result: '[BrowserTool executeTask stub]' }; }
  getName(): string { return 'browser'; }
}

export class ChainExecutor {
  private _server: any = null;
  constructor(_server?: any) { this._server = _server; }
  async execute(_request: any): Promise<any> { return { result: '[ChainExecutor stub]' }; }
  async runChain(_request: any): Promise<any> { return { result: '[ChainExecutor runChain stub]' }; }
  async executeChain(_request: any): Promise<any> { return { result: '[ChainExecutor executeChain stub]' }; }
}

export interface ChainRequest {
  prompt: string;
  model?: string;
  tools?: any[];
  maxSteps?: number;
}

export class InputTools {
  private _server: any = null;
  constructor(_server?: any) { this._server = _server; }
  async sendKeys(_keys: string, _forceFocus?: boolean, _targetWindow?: string): Promise<any> {
    return { status: 'ok' };
  }
  async execute(_input: any): Promise<any> { return { result: '[InputTools stub]' }; }
  async run(_input: any): Promise<any> { return { result: '[InputTools run stub]' }; }
}

export class ProcessRegistry {
  private _processes: Map<string, any> = new Map();
  register(_id: string, _proc: any): void {}
  unregister(_id: string): void {}
  get(_id: string): any { return undefined; }
  list(): any[] { return []; }
}

export class SystemStatusTool {
  private _server: any = null;
  constructor(_server?: any) { this._server = _server; }
  async execute(_input: any): Promise<any> { return { result: '[SystemStatusTool stub]' }; }
  getStatus(): any { return { status: 'ok' }; }
}

export class TerminalService {
  private _registry: any = null;
  constructor(_registry?: any) { this._registry = _registry; }
  async execute(_input: any): Promise<any> { return { result: '[TerminalService stub]' }; }
  getTools(): ToolDefinition[] { return []; }
  createSession(_opts?: any): any { return null; }
  killSession(_id: string): void {}
  listSessions(): any[] { return []; }
}

// ── Tool Groups (iterable arrays used with spread [...Tools]) ──────────────

export const ConfigTools: ToolDefinition[] = [
  defTool('config_get', 'Get a configuration value', async () => ({ content: [] })),
  defTool('config_set', 'Set a configuration value', async () => ({ content: [] })),
];

export const FileSystemTools: ToolDefinition[] = [
  defTool('read_file', 'Read a file', async () => ({ content: [] })),
  defTool('write_file', 'Write a file', async () => ({ content: [] })),
  defTool('list_directory', 'List directory contents', async () => ({ content: [] })),
];

export const LogTools: ToolDefinition[] = [
  defTool('get_logs', 'Get logs', async () => ({ content: [] })),
];

export const MemoryTools: ToolDefinition[] = [
  defTool('memory_store', 'Store a memory', async () => ({ content: [] })),
  defTool('memory_recall', 'Recall memories', async () => ({ content: [] })),
];

export const MetaTools: ToolDefinition[] = [
  defTool('meta_status', 'Get system status', async () => ({ content: [] })),
];

export const ReaderTools: ToolDefinition[] = [
  defTool('read_page', 'Read a web page', async () => ({ content: [] })),
  defTool('read_file', 'Read a file', async () => ({ content: [] })),
];

export const SearchTools: ToolDefinition[] = [
  defTool('search', 'Search the workspace', async () => ({ content: [] })),
];

export const TunnelTools: ToolDefinition[] = [
  defTool('tunnel_create', 'Create a tunnel', async () => ({ content: [] })),
];

export const WorktreeTools: ToolDefinition[] = [
  defTool('worktree_create', 'Create a worktree', async () => ({ content: [] })),
];

export function getAllParityTools(): ToolDefinition[] {
  return [];
}
