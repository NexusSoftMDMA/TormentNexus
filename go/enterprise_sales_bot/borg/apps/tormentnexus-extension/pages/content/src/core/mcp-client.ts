import { sendMessage, onMessage } from 'webext-bridge/content-script';
import { useConnectionStore, useToolStore, useDebuggerStore } from '../stores';
import { eventBus } from '../events/event-bus';
import type { ServerConfig, ConnectionStatus } from '../types/stores';
import { logMessage } from '../utils/helpers';
import { pluginRegistry } from '../plugins';

class McpClient {
  private static instance: McpClient | null = null;
  private isInitialized = false;
  private heartbeatInterval: number | null = null;
  private readonly HEARTBEAT_INTERVAL = 30000;

  private constructor() {
    this.initialize();
  }

  private initialize(): void {
    if (this.isInitialized) {
      logMessage('[McpClient] Already initialized');
      return;
    }

    try {
      this.setupMessageListeners();
      this.startHeartbeat();
      this.isInitialized = true;

      // Request initial status for default profile
      this.requestInitialState('default-sse').catch(error => {
        logMessage(`[McpClient] Initial state request failed: ${error}`);
      });
      // Optionally request for other profiles if needed, but the UI component should really trigger it.
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      this.isInitialized = false;
      eventBus.emit('error:unhandled', { error: error instanceof Error ? error : new Error(errorMessage), context: 'mcp-client-initialization' });
      throw error;
    }
  }

  public async requestInitialState(profileId: string = 'default-sse'): Promise<void> {
    try {
      const statusResponse = await this.getCurrentConnectionStatus(profileId);
      if (statusResponse) {
        this.handleConnectionStatusChange(profileId, statusResponse.status as ConnectionStatus, undefined);
      }

      const config = await this.getServerConfig(profileId);
      if (config) {
        useConnectionStore.getState().setServerConfig(profileId, config);
      }

      await this.getAvailableTools(true, profileId);
    } catch (error) {
      logMessage(`[McpClient] requestInitialState failed for profile ${profileId}`);
    }
  }

  private setupMessageListeners(): void {
    onMessage('connection:status-changed', ({ data }) => {
      try {
        const { status, error, isConnected, profileId } = data ?? {};
        if (status) {
          this.handleConnectionStatusChange(profileId || 'default-sse', status, error);
        }
      } catch (error) {
        logMessage(`[McpClient] Error processing connection status message: ${error}`);
      }
    });

    onMessage('mcp:tool-update', ({ data }) => {
      try {
        const { tools, profileId } = data ?? { tools: [] };
        const safeTools = Array.isArray(tools) ? tools : [];
        this.handleToolUpdate(profileId || 'default-sse', safeTools);
      } catch (error) {
        logMessage(`[McpClient] Error processing tool update: ${error}`);
      }
    });

    onMessage('mcp:server-config-updated', ({ data }) => {
      try {
        const { config, profileId } = data ?? {};
        if (config) {
          this.handleServerConfigUpdate(profileId || 'default-sse', config);
        }
      } catch (error) {
        logMessage(`[McpClient] Error processing server config config update: ${error}`);
      }
    });

    onMessage('mcp:heartbeat-response', ({ data }) => {
      try {
        const { timestamp, isConnected, profileId } = data ?? {};
        if (timestamp) {
          const targetProfileId = profileId || 'default-sse';
          if (typeof isConnected === 'boolean') {
            const currentStatus = useConnectionStore.getState().connections[targetProfileId]?.status;
            const expectedStatus = isConnected ? 'connected' : 'disconnected';
            if (currentStatus !== expectedStatus) {
              this.handleConnectionStatusChange(targetProfileId, expectedStatus);
            }
          }
          this.handleHeartbeatResponse(targetProfileId, timestamp);
        }
      } catch (error) {
        logMessage(`[McpClient] Error processing heartbeat response: ${error}`);
      }
    });
  }

  private handleConnectionStatusChange(profileId: string, status: ConnectionStatus, error?: string): void {
    const store = useConnectionStore.getState();

    switch (status) {
      case 'connected':
        store.setConnected(profileId, Date.now());
        eventBus.emit('connection:status-changed', { status, error: undefined, profileId });
        break;
      case 'reconnecting':
        store.startReconnecting(profileId);
        eventBus.emit('connection:status-changed', { status, error: undefined, profileId });
        break;
      case 'error':
        store.setDisconnected(profileId, error ?? 'Unknown connection error');
        eventBus.emit('connection:status-changed', { status, error: error ?? 'Unknown connection error', profileId });
        break;
      case 'disconnected':
      default:
        store.setDisconnected(profileId, error);
        eventBus.emit('connection:status-changed', { status: 'disconnected', error, profileId });
    }
  }

  private handleToolUpdate(profileId: string, tools: any[]): void {
    const normalizedTools = tools.map(tool => ({
      name: tool.name,
      description: tool.description || '',
      input_schema: tool.input_schema || tool.schema || {},
      schema: typeof tool.schema === 'string' ? tool.schema : JSON.stringify(tool.input_schema || {}),
      profileId,
    }));
    useToolStore.getState().setAvailableTools(profileId, normalizedTools);
    eventBus.emit('tool:list-updated', { tools: normalizedTools });
  }

  private handleServerConfigUpdate(profileId: string, config: Partial<ServerConfig>): void {
    useConnectionStore.getState().setServerConfig(profileId, config);
  }

  private handleHeartbeatResponse(profileId: string, timestamp: number): void {
    eventBus.emit('connection:heartbeat', { timestamp, profileId });
  }

  private startHeartbeat(): void {
    if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
    this.heartbeatInterval = window.setInterval(() => {
      this.sendHeartbeat().catch(() => {});
    }, this.HEARTBEAT_INTERVAL);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private async sendHeartbeat(): Promise<void> {
    try {
      const response = await sendMessage('mcp:heartbeat', { timestamp: Date.now() }, 'background');
      if (response && response.telemetry) {
        useConnectionStore.getState().setTelemetry(response.profileId || 'default-sse', response.telemetry);
      }
    } catch (e) {}
  }

  async callTool(toolName: string, args: Record<string, unknown>, profileId: string = 'default-sse'): Promise<any> {
    if (!this.isInitialized) throw new Error('McpClient not initialized');
    
    const connectionStore = useConnectionStore.getState();
    const connState = connectionStore.connections[profileId];
    if (connState?.status !== 'connected') {
      throw new Error(`Not connected to MCP server for profile ${profileId}.`);
    }

    const executionId = useToolStore.getState().startToolExecution(toolName, args);
    const activePlugin = pluginRegistry.getActivePlugin();
    const adapterName = activePlugin?.name || window.location.hostname || 'unknown';

    const startTime = Date.now();
    useDebuggerStore.getState().addPacket({
      type: 'request',
      direction: 'outbound',
      method: 'mcp:call-tool',
      toolName,
      payload: args,
    });

    try {
      const response = await sendMessage(
        'mcp:call-tool',
        { toolName, args: args as any, profileId },
        'background'
      );

      const result = response?.result;
      useToolStore.getState().completeToolExecution(executionId, result, 'success');
      
      useDebuggerStore.getState().addPacket({
        type: 'response',
        direction: 'inbound',
        method: 'mcp:call-tool',
        toolName,
        payload: result,
        durationMs: Date.now() - startTime,
      });
      
      return result;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      useToolStore.getState().completeToolExecution(executionId, null, 'error', errorMessage);
      if (this.isConnectionError(errorMessage)) {
        connectionStore.setDisconnected(profileId, `Tool call failed: ${errorMessage}`);
      }
      
      useDebuggerStore.getState().addPacket({
        type: 'error',
        direction: 'inbound',
        method: 'mcp:call-tool',
        toolName,
        payload: { error: errorMessage },
        durationMs: Date.now() - startTime,
      });
      
      throw error;
    }
  }

  private isConnectionError(errorMessage: string): boolean {
    const patterns = [/connection refused/i, /econnrefused/i, /timeout/i, /etimedout/i, /network error/i, /fetch failed/i];
    return patterns.some(pattern => pattern.test(errorMessage));
  }

  async getAvailableTools(forceRefresh = false, profileId: string = 'default-sse'): Promise<any[]> {
    if (!this.isInitialized) throw new Error('McpClient not initialized');

    try {
      const startTime = Date.now();
      useDebuggerStore.getState().addPacket({
        type: 'request',
        direction: 'outbound',
        method: 'mcp:get-tools',
        payload: { forceRefresh, profileId },
      });

      const response = await sendMessage(
        'mcp:get-tools',
        { forceRefresh, profileId },
        'background'
      );

      const tools = response?.tools || [];
      
      useDebuggerStore.getState().addPacket({
        type: 'response',
        direction: 'inbound',
        method: 'mcp:get-tools',
        payload: { count: tools.length },
        durationMs: Date.now() - startTime,
      });

      const validatedTools = Array.isArray(tools) ? tools : [];
      const normalizedTools = validatedTools.map(tool => ({
        name: tool.name,
        description: tool.description || '',
        input_schema: tool.input_schema || tool.schema || {},
        schema: typeof tool.schema === 'string' ? tool.schema : JSON.stringify(tool.input_schema || {}),
        profileId,
      }));
      // In a real Multi-Proxy setup, we merge tools from all connected profiles into the store.
      useToolStore.getState().setAvailableTools(profileId, normalizedTools);

      return normalizedTools;
    } catch (error) {
      useDebuggerStore.getState().addPacket({
        type: 'error',
        direction: 'inbound',
        method: 'mcp:get-tools',
        payload: { error: error instanceof Error ? error.message : String(error) },
      });
      throw error;
    }
  }

  async forceReconnect(profileId: string = 'default-sse'): Promise<boolean> {
    if (!this.isInitialized) throw new Error('McpClient not initialized');
    const connectionStore = useConnectionStore.getState();

    try {
      connectionStore.startReconnecting(profileId);
      eventBus.emit('connection:status-changed', { status: 'reconnecting', error: undefined, profileId });

      const uriObj = connectionStore.connections[profileId]?.serverConfig;
      const response = await sendMessage(
        'mcp:force-reconnect',
        { profileId, uri: uriObj?.uri, transportType: uriObj?.connectionType },
        'background'
      );

      const isConnected = response?.isConnected ?? false;
      if (isConnected) {
        connectionStore.setConnected(profileId, Date.now());
        try { await this.getAvailableTools(true, profileId); } catch(e) {}
      } else {
        const errorMsg = response?.error || 'Reconnect attempt failed';
        connectionStore.setDisconnected(profileId, errorMsg);
      }
      return isConnected;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      connectionStore.setDisconnected(profileId, `Reconnect failed: ${errorMessage}`);
      throw error;
    }
  }

  async forceConnectionStatusCheck(profileId: string = 'default-sse'): Promise<void> {
    try {
      const statusResponse = await this.getCurrentConnectionStatus(profileId);
      if (statusResponse) {
        this.handleConnectionStatusChange(profileId, statusResponse.status as ConnectionStatus, undefined);
      }
    } catch (e) {}
  }

  async getServerConfig(profileId: string = 'default-sse'): Promise<ServerConfig> {
    const response = await sendMessage('mcp:get-server-config', { profileId }, 'background');
    return response?.config as ServerConfig;
  }

  async getCurrentConnectionStatus(profileId: string = 'default-sse'): Promise<{ status: string; isConnected: boolean; timestamp: number }> {
    return await sendMessage('mcp:get-connection-status', { profileId }, 'background');
  }

  async updateServerConfig(config: Partial<ServerConfig>, profileId: string = 'default-sse'): Promise<boolean> {
    try {
      const response = await sendMessage(
        'mcp:update-server-config',
        { config, profileId },
        'background'
      );
      if (response?.success) {
        useConnectionStore.getState().setServerConfig(profileId, config);
        return true;
      }
      return false;
    } catch (error) {
      throw error;
    }
  }

  getConnectionStatus(profileId: string = 'default-sse'): ConnectionStatus {
    return useConnectionStore.getState().connections[profileId]?.status || 'disconnected';
  }

  isReady(): boolean { return this.isInitialized; }
  cleanup(): void {
    this.stopHeartbeat();
    this.isInitialized = false;
  }

  public static getInstance(): McpClient {
    if (!McpClient.instance) {
      McpClient.instance = new McpClient();
    }
    return McpClient.instance;
  }
}

export const mcpClient = McpClient.getInstance();
export type { McpClient };
