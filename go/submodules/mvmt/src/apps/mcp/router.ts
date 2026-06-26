import { CallToolResult } from '../../connectors/types.js';
import { TextContextIndex } from '../../context/text-index.js';
import { ClientIdentity } from '../../core/auth/client-identity.js';
import { actionAvailable, pathAllowed, pathMayExposeEntry, type PermissionAction } from '../../core/auth/permissions.js';
import { AuditLogger, summarizeArgs } from '../../utils/audit.js';
import { accessDeniedResult } from './tools/helpers.js';
import {
  CONTEXT_TOOL_BY_NAME,
  CONTEXT_TOOLS,
  isContextToolName,
  type ContextToolName,
  type NamespacedTool,
} from './tools/index.js';

export type { NamespacedTool } from './tools/index.js';

export interface ToolRouterOptions {
  contextIndex?: TextContextIndex;
}

export class ToolRouter {
  private readonly contextToolDefinitions: NamespacedTool[] = [];
  private readonly contextIndex?: TextContextIndex;
  private initialized = false;

  constructor(
    private readonly audit?: AuditLogger,
    options: ToolRouterOptions = {},
  ) {
    this.contextIndex = options.contextIndex;
  }

  async initialize(): Promise<void> {
    if (this.initialized) return;
    this.initialized = true;
    if (this.contextIndex) {
      this.contextToolDefinitions.push(...CONTEXT_TOOLS.map((tool) => tool.definition));
    }
  }

  getAllTools(identity?: ClientIdentity): NamespacedTool[] {
    return this.contextToolDefinitions
      .filter((tool) => this.isContextToolVisible(tool.namespacedName, identity))
      .map((tool) => ({ ...tool }));
  }

  async callTool(
    name: string,
    args: Record<string, unknown>,
    identity?: ClientIdentity,
  ): Promise<CallToolResult> {
    if (!isContextToolName(name)) {
      throw new Error(`Unknown tool: ${name}`);
    }
    return this.callContextTool(name, args, identity);
  }

  private async callContextTool(
    name: ContextToolName,
    args: Record<string, unknown>,
    identity?: ClientIdentity,
  ): Promise<CallToolResult> {
    const start = Date.now();
    let result: CallToolResult | undefined;
    let threw = false;
    let deniedReason: string | undefined;

    try {
      if (!this.contextIndex) {
        deniedReason = 'text_index_disabled';
        result = accessDeniedResult(deniedReason);
        return result;
      }

      result = await this.dispatchContextTool(name, args, identity);
      return result;
    } catch (err) {
      threw = true;
      result = {
        content: [{ type: 'text', text: `Error: ${err instanceof Error ? err.message : 'Unknown error'}` }],
        isError: true,
      };
      return result;
    } finally {
      this.recordAudit(
        name,
        args,
        start,
        result,
        threw,
        identity,
        deniedReason ?? (result?.isError ? deniedReasonFromResult(result) : undefined),
      );
    }
  }

  private async dispatchContextTool(
    name: ContextToolName,
    args: Record<string, unknown>,
    identity?: ClientIdentity,
  ): Promise<CallToolResult> {
    if (!this.contextIndex) return accessDeniedResult('text_index_disabled');
    return CONTEXT_TOOL_BY_NAME.get(name)!.handle(args, {
      index: this.contextIndex,
      identity,
      access: {
        allowedMounts: (action, requestedMountNames) => this.allowedContextMounts(action, identity, requestedMountNames),
        pathAllowed: (inputPath, action) => pathAllowed(inputPath, action, identity),
        pathMayExposeEntry: (inputPath, action) => pathMayExposeEntry(inputPath, action, identity),
      },
    });
  }

  private isContextToolVisible(name: string, identity?: ClientIdentity): boolean {
    const tool = CONTEXT_TOOL_BY_NAME.get(name as ContextToolName);
    if (!tool || !this.contextIndex) return false;
    return actionAvailable(tool.definition.requiredAction, identity);
  }

  private allowedContextMounts(
    action: PermissionAction,
    identity?: ClientIdentity,
    requestedMountNames?: string[],
  ): string[] {
    if (!this.contextIndex) return [];
    const requested = requestedMountNames ? new Set(requestedMountNames) : undefined;
    return this.contextIndex.mountNames().filter((mountName) => {
      const mountPath = this.contextIndex?.mountPathForName(mountName);
      return Boolean(mountPath)
        && (!requested || requested.has(mountName))
        && pathMayExposeEntry(mountPath!, action, identity);
    });
  }

  private recordAudit(
    name: string,
    args: Record<string, unknown>,
    start: number,
    result: CallToolResult | undefined,
    threw: boolean,
    identity?: ClientIdentity,
    deniedReason?: string,
  ): void {
    if (!this.audit) return;
    const { argKeys, argPreview } = summarizeArgs(args);
    this.audit.record({
      ts: new Date().toISOString(),
      ...(identity && (identity.source === 'token' || identity.source === 'oauth')
        ? {
            event: 'token.use' as const,
            name: identity.id,
            result: threw || Boolean(result?.isError) ? 'error' as const : 'success' as const,
          }
        : {}),
      connectorId: 'mvmt',
      tool: name,
      ...(identity ? { clientId: identity.id } : {}),
      argKeys,
      argPreview,
      isError: threw || Boolean(result?.isError),
      ...(deniedReason ? { deniedReason } : {}),
      durationMs: Date.now() - start,
    });
  }
}

function extractToolText(raw: CallToolResult): string {
  return raw.content
    .filter((item): item is { type: 'text'; text: string } => item.type === 'text')
    .map((item) => item.text)
    .join('\n');
}

function errorTextFromResult(result: CallToolResult): string | undefined {
  const text = extractToolText(result);
  return text.startsWith('Error: ') ? text.slice('Error: '.length, 180) : undefined;
}

function deniedReasonFromResult(result: CallToolResult): string | undefined {
  const text = extractToolText(result);
  const match = text.match(/^Error: access denied \((.+)\)\.?$/);
  return match?.[1] ?? errorTextFromResult(result);
}
