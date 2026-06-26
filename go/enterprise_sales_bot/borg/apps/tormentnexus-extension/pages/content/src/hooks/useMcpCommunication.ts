import { useEffect, useCallback, useRef } from 'react';
import { useConnectionStatus, useAvailableTools, useToolExecution, useServerConfig } from './useStores';
import { mcpClient } from '../core/mcp-client';
import { createLogger } from '@extension/shared/lib/logger';
import { useToastStore } from '@src/stores';

const logger = createLogger('useMcpCommunication');

export const useMcpCommunication = (profileId: string = 'default-sse') => {
  const { isConnected, isReconnecting, status, error: connectionError } = useConnectionStatus(profileId);
  // Connection store actions directly since we don't return them from useConnectionStatus anymore
  const { setAvailableTools } = useAvailableTools();
  const { startExecution, updateExecution, completeExecution } = useToolExecution();
  const { config } = useServerConfig(profileId);
  const { addToast } = useToastStore.getState();

  const retryCountRef = useRef(0);
  const maxRetries = 5;
  const baseDelay = 2000;

  // Function to connect to the MCP server
  const connect = useCallback(async () => {
    if (!config.uri) {
      return;
    }

    try {
      // If we are already connected and config matches, maybe we don't need to force reconnect?
      // But for now, let's trust the logic.

      // Update client configuration if needed
      await mcpClient.updateServerConfig(config, profileId);

      const currentStatus = await mcpClient.getCurrentConnectionStatus(profileId);

      if (currentStatus.status === 'connected') {
        retryCountRef.current = 0;
        const tools = await mcpClient.getAvailableTools(false, profileId);
      } else {
        const success = await mcpClient.forceReconnect(profileId);
        if (success) {
          retryCountRef.current = 0;
        } else {
          throw new Error('Connection failed');
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      logger.error(`[useMcpCommunication] Connection error for ${profileId}:`, errorMessage);

      // Auto-Retry Logic
      if (retryCountRef.current < maxRetries) {
        const delay = baseDelay * Math.pow(1.5, retryCountRef.current);
        logger.info(
          `[useMcpCommunication] Auto-retrying in ${delay}ms (Attempt ${retryCountRef.current + 1}/${maxRetries})`,
        );

        setTimeout(() => {
          retryCountRef.current++;
          connect();
        }, delay);
      } else {
        addToast({
          title: 'Connection Failed',
          message: 'Max retries reached. Please check server status.',
          type: 'error',
          duration: 5000,
        });
      }
    }
  }, [config, profileId]); // Removed setStatus, setIsConnected, setError, setTools which no longer exist in this scope

  // Handle server config updates
  const updateServerConfig = useCallback(
    async (newConfig: { uri: string; connectionType: any }) => {
      logger.debug(`[useMcpCommunication] Updating server config for ${profileId}:`, newConfig);
      await mcpClient.updateServerConfig(newConfig, profileId);
      retryCountRef.current = 0;
      return connect();
    },
    [connect, profileId],
  );

  const forceReconnect = useCallback(async () => {
    logger.debug(`[useMcpCommunication] Forcing reconnection for ${profileId}`);
    retryCountRef.current = 0;
    const success = await mcpClient.forceReconnect(profileId);
    return success;
  }, [profileId]);

  // Check connection status
  const forceConnectionStatusCheck = useCallback(async () => {
    const status = await mcpClient.getCurrentConnectionStatus(profileId);
    return status.isConnected;
  }, [profileId]);

  // Send message / Execute tool
  const sendMessage = useCallback(
    async (toolOrName: string | any, args?: any) => {
      let toolName: string;
      let toolArgs: any;

      if (typeof toolOrName === 'string') {
        toolName = toolOrName;
        toolArgs = args || {};
      } else {
        // Assume it's a tool object { name, arguments } or { name }
        toolName = toolOrName.name;
        toolArgs = toolOrName.arguments || args || {};
      }

      try {
        logger.debug('[useMcpCommunication] Executing tool:', toolName);

        // startExecution handles ID generation and store initialization
        const executionId = startExecution(toolName, toolArgs);

        const result = await mcpClient.callTool(toolName, toolArgs, profileId);

        completeExecution(executionId, result, 'success');

        return result;
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        logger.error('[useMcpCommunication] Tool execution error:', errorMessage);
        throw error;
      }
    },
    [startExecution, completeExecution, profileId],
  );

  // Refresh tools
  const refreshTools = useCallback(
    async (force = false) => {
      try {
        const tools = await mcpClient.getAvailableTools(force, profileId);
        return tools;
      } catch (error) {
        logger.error(`[useMcpCommunication] Error refreshing tools for ${profileId}:`, error);
        return [];
      }
    },
    [profileId],
  );

  const getServerConfig = useCallback(async () => {
    return mcpClient.getServerConfig(profileId);
  }, [profileId]);

  // Initial connection on mount
  useEffect(() => {
    // We only trigger connect if config is available.
    // mcpClient might already be initialized by global singleton, but we sync state here.
    if (config.uri) {
      connect();
    }

    return () => {
      // Optional: mcpClient.cleanup() ?
      // Usually we don't want to kill the global client on unmount of a hook,
      // unless this hook controls the lifecycle.
      // Since Sidebar uses this, and Sidebar is persistent...
      // But if Sidebar unmounts, maybe we should?
      // For now, let's NOT cleanup global client to be safe.
    };
  }, [connect, config.uri]);

  return {
    availableTools: useAvailableTools().tools,
    sendMessage,
    refreshTools,
    forceReconnect,
    serverStatus: status,
    updateServerConfig,
    getServerConfig,
    forceConnectionStatusCheck,
    lastConnectionError: connectionError,
  };
};
