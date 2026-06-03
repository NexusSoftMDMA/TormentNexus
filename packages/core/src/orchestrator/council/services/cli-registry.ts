import type { CLITool, CLIType } from './types.js';
import { spawn } from 'child_process';

interface CLIDefinition {
  type: CLIType;
  name: string;
  command: string;
  serveArgs: string[];
  versionArgs: string[];
  healthEndpoint?: string;
  detectCommands: string[];
  capabilities: string[];
  interactive?: boolean;
  promptRegex?: string;
}

export const CLI_DEFINITIONS: CLIDefinition[] = [
  {
    type: 'tormentnexus',
    name: 'TormentNexus CLI',
    command: 'tormentnexus',
    serveArgs: [],
    versionArgs: ['version'],
    detectCommands: ['tormentnexus'],
    capabilities: ['chat', 'edit', 'repl', 'tormentnexus-adapter'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'antigravity',
    name: 'Antigravity CLI',
    command: 'antigravity',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['antigravity'],
    capabilities: ['chat', 'desktop', 'editor', 'automation'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'opencode',
    name: 'OpenCode CLI',
    command: 'opencode',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['opencode', 'npx opencode'],
    capabilities: ['chat', 'edit', 'multi-file', 'autonomous'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'claude',
    name: 'Claude CLI',
    command: 'claude',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['claude'],
    capabilities: ['chat', 'code', 'analyze'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'aider',
    name: 'Aider CLI',
    command: 'aider',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['aider', 'python -m aider'],
    capabilities: ['chat', 'edit', 'git-aware', 'multi-file'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'amazon-q',
    name: 'Amazon Q CLI',
    command: 'q',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['q', 'amazon-q'],
    capabilities: ['chat', 'aws', 'deploy'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'amazon-q-developer',
    name: 'Amazon Q Developer CLI',
    command: 'q-developer',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['q-developer'],
    capabilities: ['chat', 'code', 'aws'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'amp-code',
    name: 'Amp Code CLI',
    command: 'amp',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['amp'],
    capabilities: ['chat', 'edit', 'terminal'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'auggie',
    name: 'Auggie CLI',
    command: 'auggie',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['auggie'],
    capabilities: ['chat', 'review', 'git'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'azure-openai',
    name: 'Azure OpenAI CLI',
    command: 'az-openai',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['az-openai', 'az openai'],
    capabilities: ['chat', 'code', 'azure'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'claude-code',
    name: 'Claude Code CLI',
    command: 'claude-code',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['claude-code'],
    capabilities: ['chat', 'code', 'analyze'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'code-codex',
    name: 'Code CLI (Codex fork)',
    command: 'code-codex',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['code-codex'],
    capabilities: ['chat', 'code', 'edit'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'codebuff',
    name: 'Codebuff CLI',
    command: 'codebuff',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['codebuff'],
    capabilities: ['chat', 'code', 'refactor'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'codemachine',
    name: 'Codemachine CLI',
    command: 'codemachine',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['codemachine'],
    capabilities: ['chat', 'code', 'generate'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'codex',
    name: 'Codex CLI',
    command: 'codex',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['codex'],
    capabilities: ['chat', 'code'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'copilot',
    name: 'GitHub Copilot CLI',
    command: 'github-copilot-cli',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['github-copilot-cli', 'gh copilot'],
    capabilities: ['explain', 'suggest', 'chat', 'terminal', 'shell'],
    interactive: true,
    promptRegex: '(?i)(?:\\?|>)\\s*$',
  },
  {
    type: 'cursor',
    name: 'Cursor CLI',
    command: 'cursor',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['cursor'],
    capabilities: ['chat', 'code', 'editor'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'continue',
    name: 'Continue CLI',
    command: 'continue',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['continue'],
    capabilities: ['chat', 'autocomplete', 'editor'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'cody',
    name: 'Cody CLI',
    command: 'cody',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['cody'],
    capabilities: ['chat', 'search', 'code'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'adrenaline',
    name: 'Adrenaline CLI',
    command: 'adrenaline',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['adrenaline'],
    capabilities: ['chat', 'shell', 'automation'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'crush',
    name: 'Crush CLI',
    command: 'crush',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['crush'],
    capabilities: ['chat', 'code', 'data'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'bito',
    name: 'Bito CLI',
    command: 'bito',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['bito'],
    capabilities: ['chat', 'code', 'review'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'byterover',
    name: 'Byterover CLI',
    command: 'byterover',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['byterover'],
    capabilities: ['chat', 'code', 'automation'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'factory',
    name: 'Factory CLI',
    command: 'factory',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['factory'],
    capabilities: ['chat', 'code', 'build'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'dolt',
    name: 'Dolt CLI',
    command: 'dolt',
    serveArgs: [],
    versionArgs: ['version'],
    detectCommands: ['dolt'],
    capabilities: ['database', 'sql', 'git'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'factory-droid',
    name: 'Factory Droid CLI',
    command: 'factory-droid',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['factory-droid'],
    capabilities: ['chat', 'code', 'build'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'gemini',
    name: 'Gemini CLI',
    command: 'gemini',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['gemini'],
    capabilities: ['chat', 'code', 'multimodal'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'goose',
    name: 'Goose CLI',
    command: 'goose',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['goose'],
    capabilities: ['chat', 'code', 'agent'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'grok',
    name: 'Grok CLI',
    command: 'grok',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['grok'],
    capabilities: ['chat', 'code', 'realtime'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'jules',
    name: 'Jules CLI',
    command: 'jules',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['jules'],
    capabilities: ['chat', 'code', 'agent'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'kilo-code',
    name: 'Kilo Code CLI',
    command: 'kilo',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['kilo'],
    capabilities: ['chat', 'code', 'editor'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'kimi',
    name: 'Kimi CLI',
    command: 'kimi',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['kimi'],
    capabilities: ['chat', 'code', 'long-context'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'llm',
    name: 'LLM CLI',
    command: 'llm',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['llm'],
    capabilities: ['chat', 'models', 'prompt'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'litellm',
    name: 'LiteLLM CLI',
    command: 'litellm',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['litellm'],
    capabilities: ['models', 'proxy', 'routing'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'llamafile',
    name: 'Llamafile CLI',
    command: 'llamafile',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['llamafile'],
    capabilities: ['chat', 'local', 'models'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'manus',
    name: 'Manus CLI',
    command: 'manus',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['manus'],
    capabilities: ['chat', 'agent', 'automation'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'mistral-vibe',
    name: 'Mistral Vibe CLI',
    command: 'mistral',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['mistral'],
    capabilities: ['chat', 'code', 'local'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'ollama',
    name: 'Ollama CLI',
    command: 'ollama',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['ollama'],
    capabilities: ['chat', 'local', 'models'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'open-interpreter',
    name: 'Open Interpreter CLI',
    command: 'interpreter',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['interpreter', 'open-interpreter'],
    capabilities: ['chat', 'python', 'shell'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'pi',
    name: 'Pi CLI',
    command: 'pi',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['pi'],
    capabilities: ['chat', 'personal', 'assistant'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'qwen-code',
    name: 'Qwen Code CLI',
    command: 'qwen',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['qwen'],
    capabilities: ['chat', 'code', 'local'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'rowboatx',
    name: 'RowboatX CLI',
    command: 'rowboatx',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['rowboatx'],
    capabilities: ['chat', 'code', 'data'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'rovo',
    name: 'Rovo CLI',
    command: 'rovo',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['rovo'],
    capabilities: ['chat', 'code', 'search'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'shell-pilot',
    name: 'Shell Pilot CLI',
    command: 'shell-pilot',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['shell-pilot'],
    capabilities: ['shell', 'chat', 'automation'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'smithery',
    name: 'Smithery CLI',
    command: 'smithery',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['smithery'],
    capabilities: ['mcp', 'registry', 'tools'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'superai-cli',
    name: 'SuperAI CLI',
    command: 'superai',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['superai', 'superai-cli'],
    capabilities: ['chat', 'code', 'automation'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'trae',
    name: 'Trae CLI',
    command: 'trae',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['trae'],
    capabilities: ['chat', 'code', 'review'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
  {
    type: 'warp',
    name: 'Warp CLI',
    command: 'warp',
    serveArgs: [],
    versionArgs: ['--version'],
    detectCommands: ['warp'],
    capabilities: ['chat', 'terminal', 'collaborative'],
    interactive: true,
    promptRegex: '(?i)(?:>)\\s*$',
  },
];

function stripWrappingQuotes(token: string): string {
  if (
    (token.startsWith('"') && token.endsWith('"')) ||
    (token.startsWith("'") && token.endsWith("'"))
  ) {
    return token.slice(1, -1);
  }
  return token;
}

export function splitCommandSpec(commandSpec: string): { command: string; args: string[] } {
  const tokens = commandSpec.match(/"[^"]*"|'[^']*'|\S+/g) ?? [];
  if (tokens.length === 0) {
    return { command: '', args: [] };
  }

  const command = tokens[0]!;
  const args = tokens.slice(1);

  return {
    command: stripWrappingQuotes(command),
    args: args.map(stripWrappingQuotes),
  };
}

class CLIRegistry {
  private tools: Map<CLIType, CLITool> = new Map();
  private customTools: Map<string, CLITool> = new Map();
  private detectionPromise: Promise<void> | null = null;

  async detectAll(): Promise<CLITool[]> {
    if (this.detectionPromise) {
      await this.detectionPromise;
      return this.getAllTools();
    }

    this.detectionPromise = this.runDetection();
    await this.detectionPromise;
    return this.getAllTools();
  }

  private async runDetection(): Promise<void> {
    const detectionPromises = CLI_DEFINITIONS.map(async (def) => {
      const tool = await this.detectTool(def);
      if (tool) {
        this.tools.set(def.type, tool);
      }
    });

    await Promise.all(detectionPromises);
  }

  private async detectTool(def: CLIDefinition): Promise<CLITool | null> {
    for (const cmd of def.detectCommands) {
      try {
        const result = await this.runCommand(cmd, def.versionArgs);
        const parsedCommand = splitCommandSpec(cmd);
        if (result.success) {
          return {
            type: def.type,
            name: def.name,
            command: parsedCommand.command || def.command,
            args: def.serveArgs,
            healthEndpoint: def.healthEndpoint,
            detectCommand: cmd,
            available: true,
            version: this.parseVersion(result.output),
            capabilities: def.capabilities,
            interactive: def.interactive,
            promptRegex: def.promptRegex,
          };
        }
      } catch {
        continue;
      }
    }

    return {
      type: def.type,
      name: def.name,
      command: def.command,
      args: def.serveArgs,
      healthEndpoint: def.healthEndpoint,
      available: false,
      capabilities: def.capabilities,
      interactive: def.interactive,
      promptRegex: def.promptRegex,
    };
  }

  private runCommand(command: string, args: string[]): Promise<{ success: boolean; output: string }> {
    return new Promise((resolve) => {
      const parsedCommand = splitCommandSpec(command);
      if (!parsedCommand.command) {
        resolve({ success: false, output: '' });
        return;
      }

      const cmdArgs = [...parsedCommand.args, ...args];
      let settled = false;
      const finish = (result: { success: boolean; output: string }) => {
        if (settled) return;
        settled = true;
        clearTimeout(timeoutHandle);
        resolve(result);
      };

      const proc = spawn(parsedCommand.command, cmdArgs, {
        shell: false,
        windowsHide: true,
      });

      let output = '';
      proc.stdout?.on('data', (data) => { output += data.toString(); });
      proc.stderr?.on('data', (data) => { output += data.toString(); });

      proc.on('close', (code) => {
        finish({ success: code === 0, output: output.trim() });
      });

      proc.on('error', () => {
        finish({ success: false, output: '' });
      });

      const timeoutHandle = setTimeout(() => {
        proc.kill();
        finish({ success: false, output: '' });
      }, 5000);
    });
  }

  private parseVersion(output: string): string {
    const match = output.match(/v?(\d+\.\d+\.\d+)/);
    return match ? match[1] : 'unknown';
  }

  getTool(type: CLIType): CLITool | undefined {
    return this.tools.get(type);
  }

  getAvailableTools(): CLITool[] {
    return [...this.tools.values(), ...this.customTools.values()].filter(t => t.available);
  }

  getAllTools(): CLITool[] {
    return [...this.tools.values(), ...this.customTools.values()];
  }

  registerCustomTool(tool: CLITool): void {
    this.customTools.set(tool.name, tool);
  }

  unregisterCustomTool(name: string): boolean {
    return this.customTools.delete(name);
  }

  getServeCommand(type: CLIType, port: number): { command: string; args: string[] } | null {
    const tool = this.tools.get(type);
    if (!tool || !tool.available) return null;

    const args = [...tool.args];
    const portArgIndex = args.findIndex(a => a.includes('port'));
    if (portArgIndex >= 0) {
      args.splice(portArgIndex + 1, 0, String(port));
    } else {
      args.push(String(port));
    }

    return { command: tool.command, args };
  }

  getHealthEndpoint(type: CLIType): string {
    const tool = this.tools.get(type);
    return tool?.healthEndpoint || '/health';
  }

  async refreshDetection(): Promise<CLITool[]> {
    this.tools.clear();
    this.detectionPromise = null;
    return this.detectAll();
  }
}

export const cliRegistry = new CLIRegistry();
