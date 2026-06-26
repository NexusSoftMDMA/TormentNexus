// Core exports
import { McpClient } from './core/McpClient.js';
import { PluginRegistry } from './core/PluginRegistry.js';
import { EventEmitter } from './core/EventEmitter.js';
import { McpManager, mcpManager } from './core/McpManager.js';

// Plugin implementations
import { SSEPlugin } from './plugins/sse/SSEPlugin.js';
import { WebSocketPlugin } from './plugins/websocket/WebSocketPlugin.js';
import { WebSocketTransport } from './plugins/websocket/WebSocketTransport.js';

// Configuration
import { DEFAULT_CLIENT_CONFIG } from './types/config.js';
import { createLogger } from '@extension/shared/lib/logger';
import { validateToolSchema } from './utils/schemaValidator.js';

// Export core classes

const logger = createLogger('mcp_client');

export { McpClient, PluginRegistry, EventEmitter, McpManager, mcpManager };

// Export plugins
export { SSEPlugin, WebSocketPlugin, WebSocketTransport };

// Export configuration
export { DEFAULT_CLIENT_CONFIG };

// Re-export types
export type { ITransportPlugin, PluginMetadata, PluginConfig, TransportType } from './types/plugin.js';

export type {
  ClientConfig,
  ConnectionRequest,
  SSEPluginConfig,
  WebSocketPluginConfig,
  GlobalConfig,
} from './types/config.js';

export type {
  Primitive,
  NormalizedTool,
  PrimitivesResponse,
  ToolCallRequest,
  ToolCallResult,
} from './types/primitives.js';

export type { AllEvents } from './types/events.js';

// Singleton client instance for backward compatibility
let globalClient: McpClient | null = null;

/**
 * Get or create the global MCP client instance
 */
async function getGlobalClient(): Promise<McpClient> {
  if (!globalClient) {
    try {
      globalClient = new McpClient();
      await globalClient.initialize();

      // Set up global event listeners for connection status changes
      setupGlobalClientEventListeners(globalClient);
    } catch (error) {
      logger.error('[getGlobalClient] Failed to initialize client:', error);
      // Create a fallback client without plugin loading
      globalClient = new McpClient();
      // Don't initialize to avoid plugin loading issues
      setupGlobalClientEventListeners(globalClient);
    }
  }
  return globalClient;
}

/**
 * Set up event listeners on the global client to handle connection events
 */
function setupGlobalClientEventListeners(client: McpClient): void {
  // Listen for connection status changes and forward them to any registered listeners
  client.on('connection:status-changed', event => {
    logger.debug('[Global Client] Connection status changed:', event);

    // Emit a global event that can be caught by the background script
    if (typeof window !== 'undefined' && window.dispatchEvent) {
      window.dispatchEvent(
        new CustomEvent('mcp:connection-status-changed', {
          detail: event,
        }),
      );
    }

    // Also try to broadcast via chrome runtime if available
    if (typeof chrome !== 'undefined' && chrome.runtime && chrome.runtime.sendMessage) {
      chrome.runtime
        .sendMessage({
          type: 'mcp:connection-status-changed',
          payload: event,
          origin: 'mcpclient',
        })
        .catch(() => {
          // Ignore errors if background script isn't listening
        });
    }
  });

  client.on('client:connected', event => {
    logger.debug('[Global Client] Client connected:', event);
  });

  client.on('client:disconnected', event => {
    logger.debug('[Global Client] Client disconnected:', event);
  });

  client.on('client:error', event => {
    logger.error('[Global Client] Client error:', event);
  });
}

/**
 * Create a new MCP client instance
 */
export async function createMcpClient(config?: Partial<import('./types/config.js').ClientConfig>): Promise<McpClient> {
  const client = new McpClient(config);
  await client.initialize();
  return client;
}

/**
 * Auto-detect transport type from URI
 */
function detectTransportType(uri: string): import('./types/plugin.js').TransportType {
  try {
    const url = new URL(uri);
    if (url.protocol === 'ws:' || url.protocol === 'wss:') {
      return 'websocket';
    }
    // For HTTP/HTTPS, default to SSE (traditional behavior)
    // Users can manually select streamable-http if desired
    return 'sse';
  } catch {
    return 'sse';
  }
}

// BACKWARD COMPATIBILITY API

export function isMcpServerConnected(): boolean {
  if (!globalClient) return false;
  return globalClient.isConnected();
}

export async function checkMcpServerConnection(): Promise<boolean> {
  try {
    const client = await getGlobalClient();
    return await client.isHealthy();
  } catch (error) {
    logger.error('[Backward Compatibility] checkMcpServerConnection failed:', error);
    return false;
  }
}

export async function callToolWithBackwardsCompatibility(
  uri: string,
  toolName: string,
  args: { [key: string]: unknown },
  adapterName?: string,
  transportType?: import('./types/plugin.js').TransportType,
): Promise<any> {
  const client = await getGlobalClient();
  const type = transportType || detectTransportType(uri);

  if (!client.isConnected()) {
    await client.connect({ uri, type });
  }

  return await client.callTool(toolName, args, adapterName);
}

export async function getPrimitivesWithBackwardsCompatibility(
  uri: string,
  forceRefresh: boolean = false,
  transportType?: import('./types/plugin.js').TransportType,
): Promise<any[]> {
  const client = await getGlobalClient();
  const type = transportType || detectTransportType(uri);

  if (!client.isConnected()) {
    await client.connect({ uri, type });
  }

  const response = await client.getPrimitives(forceRefresh);

  // Convert back to old format
  const primitives: any[] = [];

  response.tools.forEach(tool => {
    primitives.push({ type: 'tool', value: tool });
  });

  response.resources.forEach(resource => {
    primitives.push({ type: 'resource', value: resource });
  });

  response.prompts.forEach(prompt => {
    primitives.push({ type: 'prompt', value: prompt });
  });

  return primitives;
}

export async function forceReconnectToMcpServer(
  uri: string,
  transportType?: import('./types/plugin.js').TransportType,
): Promise<void> {
  const client = await getGlobalClient();
  const type = transportType || detectTransportType(uri);

  if (client.isConnected()) {
    await client.disconnect();
  }

  await client.connect({ uri, type });
}

export async function runWithBackwardsCompatibility(
  uri: string,
  transportType?: import('./types/plugin.js').TransportType,
): Promise<void> {
  const client = await getGlobalClient();
  const type = transportType || detectTransportType(uri);

  await client.connect({ uri, type });

  // Setup reverse command listener for browser control (MCP-to-web)
  if (type === 'websocket') {
    const transport = (client as any).transport;
    if (transport && typeof transport.on === 'function') {
      transport.on('message', async (message: any) => {
        // Detect non-standard JSON-RPC calls from TormentNexus Core (e.g. read_page, browser_*)
        // Detect non-standard JSON-RPC calls from tormentnexus Core (e.g. read_page, browser_*)
        if (message.jsonrpc === '2.0' && message.method && message.id && !message.result && !message.error) {
          const { method, params, id } = message;
          
          if (method === 'read_page' || method === 'browser_scrape' || method.startsWith('browser_')) {
            logger.debug(`[Global Client] Received browser command: ${method}`, { id, params });
            
            // Forward to active tab via webext-bridge
            try {
              const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
              if (tab?.id) {
                const result = await sendMessage('mcp:execute-browser-action', {
                  method,
                  params,
                  requestId: id
                }, { context: 'content-script', tabId: tab.id });
                
                // Send result back over WebSocket
                await transport.send({
                  jsonrpc: '2.0',
                  id,
                  result: result
                });
              } else {
                await transport.send({
                  jsonrpc: '2.0',
                  id,
                  error: { code: -32000, message: 'No active tab found' }
                });
              }
            } catch (err) {
              logger.error(`[Global Client] Failed to execute browser command: ${method}`, err);
              await transport.send({
                jsonrpc: '2.0',
                id,
                error: { code: -32603, message: err instanceof Error ? err.message : String(err) }
              });
            }
          }
        }
      });
    }
  }

  const response = await client.getPrimitives();
  logger.debug(
    `Connected, found ${response.tools.length} tools, ${response.resources.length} resources, ${response.prompts.length} prompts`,
  );
}

export function resetMcpConnectionState(): void {
  if (globalClient && globalClient.isConnected()) {
    globalClient.disconnect().catch(error => {
      logger.error('[Backward Compatibility] resetMcpConnectionState failed:', error);
    });
  }
}

export function resetMcpConnectionStateForRecovery(): void {
  logger.debug('[Backward Compatibility] resetMcpConnectionStateForRecovery - handled by plugin health monitoring');
}

export function abortMcpConnection(): void {
  if (globalClient) {
    globalClient.disconnect().catch(error => {
      logger.error('[Backward Compatibility] abortMcpConnection failed:', error);
    });
  }
}

// Legacy aliases
export const callToolWithSSE = callToolWithBackwardsCompatibility;
export const getPrimitivesWithSSE = getPrimitivesWithBackwardsCompatibility;
export const runWithSSE = runWithBackwardsCompatibility;

// WebSocket-specific functions
export async function connectWithWebSocket(
  uri: string,
  config?: Partial<import('./types/config.js').ClientConfig>,
): Promise<McpClient> {
  const client = new McpClient(config);
  await client.initialize();
  await client.connect({ uri, type: 'websocket' });
  return client;
}

export async function callToolWithWebSocket(
  uri: string,
  toolName: string,
  args: { [key: string]: unknown },
): Promise<any> {
  const client = await getGlobalClient();
  await client.connect({ uri, type: 'websocket' });
  return await client.callTool(toolName, args);
}

export async function getPrimitivesWithWebSocket(uri: string, forceRefresh: boolean = false): Promise<any[]> {
  const client = await getGlobalClient();
  await client.connect({ uri, type: 'websocket' });

  const response = await client.getPrimitives(forceRefresh);

  const primitives: any[] = [];
  response.tools.forEach(tool => primitives.push({ type: 'tool', value: tool }));
  response.resources.forEach(resource => primitives.push({ type: 'resource', value: resource }));
  response.prompts.forEach(prompt => primitives.push({ type: 'prompt', value: prompt }));

  return primitives;
}

// Utility function for normalizing tools
export function normalizeToolsFromPrimitives(primitives: any[]): any[] {
  return primitives
    .filter(p => p.type === 'tool')
    .map(p => {
      const tool = p.value;
      const parsedSchema = tool.inputSchema || tool.input_schema || {};
      return {
        name: tool.name,
        description: tool.description || '',
        input_schema: parsedSchema,
        schema: Object.keys(parsedSchema).length ? JSON.stringify(parsedSchema) : '{}',
        ...(tool.uri && { uri: tool.uri }),
        ...(tool.arguments && { arguments: tool.arguments }),
      };
    })
    .filter(tool => validateToolSchema(tool.name, tool.input_schema));
}
