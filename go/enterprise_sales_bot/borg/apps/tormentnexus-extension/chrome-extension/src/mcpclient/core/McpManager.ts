import { McpClient } from './McpClient.js';
import type { TransportType } from '../types/plugin.js';
import type { ConnectionRequest } from '../types/config.js';
import { createLogger } from '@extension/shared/lib/logger';
import { EventEmitter } from './EventEmitter.js';

const logger = createLogger('McpManager');

export interface McpManagerEvents {
  'manager:client-added': { profileId: string; uri: string; type: TransportType };
  'manager:client-removed': { profileId: string };
  'connection:status-changed': { profileId: string; isConnected: boolean; type: TransportType | null; error?: string };
  'tools:list-updated': { profileId: string; tools: any[]; type: TransportType };
}

export class McpManager extends EventEmitter<McpManagerEvents> {
  private static instance: McpManager | null = null;
  private clients: Map<string, McpClient> = new Map();

  private constructor() {
    super();
  }

  public static getInstance(): McpManager {
    if (!McpManager.instance) {
      McpManager.instance = new McpManager();
    }
    return McpManager.instance;
  }

  /**
   * Initialize a new connection for a specific profile
   */
  public async connectProfile(profileId: string, request: ConnectionRequest): Promise<void> {
    logger.debug(`[McpManager] Connecting profile ${profileId} to ${request.uri} via ${request.type}`);

    // Disconnect existing if we have one for this profile
    if (this.clients.has(profileId)) {
      await this.disconnectProfile(profileId);
    }

    const client = new McpClient();
    await client.initialize();

    // Attach event listeners specifically tagged with profileId
    this.setupClientListeners(profileId, client);

    this.clients.set(profileId, client);
    this.emit('manager:client-added', { profileId, uri: request.uri, type: request.type });

    try {
      await client.connect(request);
    } catch (error) {
      logger.error(`[McpManager] Failed to connect profile ${profileId}:`, error);
      throw error;
    }
  }

  /**
   * Disconnect and remove a profile's client
   */
  public async disconnectProfile(profileId: string): Promise<void> {
    const client = this.clients.get(profileId);
    if (client) {
      logger.debug(`[McpManager] Disconnecting profile ${profileId}`);
      try {
        await client.disconnect();
      } catch (error) {
        logger.error(`[McpManager] Error disconnecting profile ${profileId}:`, error);
      }
      this.clients.delete(profileId);
      this.emit('manager:client-removed', { profileId });
    }
  }

  /**
   * Forward connection events from individual clients
   */
  private setupClientListeners(profileId: string, client: McpClient): void {
    client.on('connection:status-changed', event => {
      this.emit('connection:status-changed', {
        profileId,
        ...event,
      });
    });

    client.on('tools:list-updated', event => {
      // Tag tools with profileId before emitting
      const taggedTools = event.tools.map(tool => ({ ...tool, _profileId: profileId }));
      this.emit('tools:list-updated', {
        profileId,
        tools: taggedTools,
        type: event.type,
      });
    });
  }

  public getClient(profileId: string): McpClient | undefined {
    return this.clients.get(profileId);
  }

  public getAllClients(): Map<string, McpClient> {
    return this.clients;
  }
}

export const mcpManager = McpManager.getInstance();
