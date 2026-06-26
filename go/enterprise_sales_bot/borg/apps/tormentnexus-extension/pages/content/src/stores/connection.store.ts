/**
 * @deprecated This standalone store has been migrated to `createConnectionSlice` in the unified Root Store.
 * Import from `@src/stores` instead. This file is retained only for type reference.
 */
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { eventBus } from '../events';
import type { ConnectionStatus, ServerConfig, MCPTelemetry } from '../types/stores';
import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('useConnectionStore');

export interface ProfileConnectionState {
  status: ConnectionStatus;
  serverConfig: ServerConfig;
  lastConnectedAt: number | null;
  connectionAttempts: number;
  error: string | null;
  isReconnecting: boolean;
  telemetry?: MCPTelemetry;
}

export interface ConnectionState {
  connections: Record<string, ProfileConnectionState>;

  // Actions
  setStatus: (profileId: string, status: ConnectionStatus) => void;
  setServerConfig: (profileId: string, config: Partial<ServerConfig>) => void;
  setLastError: (profileId: string, error: string | null) => void;
  setTelemetry: (profileId: string, telemetry: MCPTelemetry) => void;
  incrementAttempts: (profileId: string) => void;
  resetAttempts: (profileId: string) => void;
  setConnected: (profileId: string, timestamp: number) => void;
  setDisconnected: (profileId: string, error?: string) => void;
  startReconnecting: (profileId: string) => void;
  stopReconnecting: (profileId: string) => void;
  getProfileState: (profileId: string) => ProfileConnectionState;
  ensureProfile: (profileId: string) => void;
  removeProfile: (profileId: string) => void;
}

const defaultServerConfig: ServerConfig = {
  uri: 'http://localhost:3006/sse', // Default from migration guide, should be configurable
  connectionType: 'sse',
  timeout: 5000, // ms
  retryAttempts: 3,
  retryDelay: 2000, // ms
};

const defaultProfileState: ProfileConnectionState = {
  status: 'disconnected',
  serverConfig: defaultServerConfig,
  lastConnectedAt: null,
  connectionAttempts: 0,
  error: null,
  isReconnecting: false,
};

export const useConnectionStore = create<ConnectionState>()(
  devtools(
    (set, get) => ({
      connections: {},

      ensureProfile: (profileId: string) => {
        if (!get().connections[profileId]) {
          set(state => ({
            connections: {
              ...state.connections,
              [profileId]: { ...defaultProfileState },
            },
          }));
        }
      },

      removeProfile: (profileId: string) => {
        set(state => {
          const { [profileId]: _, ...rest } = state.connections;
          return { connections: rest };
        });
      },

      getProfileState: (profileId: string) => {
        return get().connections[profileId] || defaultProfileState;
      },

      setStatus: (profileId: string, status: ConnectionStatus) => {
        get().ensureProfile(profileId);
        const oldStatus = get().connections[profileId]?.status;
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              status,
            },
          },
        }));
        logger.debug(`Profile ${profileId} status changed from ${oldStatus} to: ${status}`);
        eventBus.emit('connection:status-changed', { status, error: get().connections[profileId]?.error || undefined, profileId });
      },

      setServerConfig: (profileId: string, config: Partial<ServerConfig>) => {
        get().ensureProfile(profileId);
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              serverConfig: { ...state.connections[profileId].serverConfig, ...config },
            },
          },
        }));
        logger.debug(`[ConnectionStore] Profile ${profileId} server config updated:`, get().connections[profileId].serverConfig);
      },

      setLastError: (profileId: string, error: string | null) => {
        get().ensureProfile(profileId);
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              error,
            },
          },
        }));
        if (error) {
          logger.error(`[ConnectionStore] Profile ${profileId} error set:`, error);
          eventBus.emit('connection:error', { error: error, profileId });
        }
      },

      setTelemetry: (profileId: string, telemetry: MCPTelemetry) => {
        get().ensureProfile(profileId);
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              telemetry,
            },
          },
        }));
        eventBus.emit('connection:telemetry', { telemetry, profileId });
      },

      incrementAttempts: (profileId: string) => {
        get().ensureProfile(profileId);
        const newAttempts = (get().connections[profileId]?.connectionAttempts || 0) + 1;
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              connectionAttempts: newAttempts,
            },
          },
        }));
        logger.debug(`Profile ${profileId} connection attempts: ${newAttempts}`);
        eventBus.emit('connection:attempt', { attempt: newAttempts, maxAttempts: get().connections[profileId].serverConfig.retryAttempts, profileId });
      },

      resetAttempts: (profileId: string) => {
        get().ensureProfile(profileId);
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              connectionAttempts: 0,
            },
          },
        }));
        logger.debug(`[ConnectionStore] Profile ${profileId} connection attempts reset.`);
      },

      setConnected: (profileId: string, timestamp: number) => {
        get().ensureProfile(profileId);
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              status: 'connected',
              lastConnectedAt: timestamp,
              connectionAttempts: 0,
              error: null,
              isReconnecting: false,
            },
          },
        }));
        logger.debug(`Profile ${profileId} connected at: ${new Date(timestamp).toISOString()}`);
        eventBus.emit('connection:status-changed', { status: 'connected', profileId });
      },

      setDisconnected: (profileId: string, error?: string) => {
        get().ensureProfile(profileId);
        set(state => {
          const profileConn = state.connections[profileId];
          return {
            connections: {
              ...state.connections,
              [profileId]: {
                ...profileConn,
                status: error ? 'error' : 'disconnected',
                error: error || profileConn.error,
                isReconnecting: false,
              },
            },
          };
        });
        logger.debug(`Profile ${profileId} disconnected. ${error ? 'Error: ' + error : ''}`);
        eventBus.emit('connection:status-changed', { status: get().connections[profileId].status, error: error || get().connections[profileId].error || undefined, profileId });
      },

      startReconnecting: (profileId: string) => {
        get().ensureProfile(profileId);
        if (get().connections[profileId]?.status === 'connected') return;
        set(state => ({
          connections: {
            ...state.connections,
            [profileId]: {
              ...state.connections[profileId],
              isReconnecting: true,
              status: 'reconnecting',
            },
          },
        }));
        logger.debug(`[ConnectionStore] Profile ${profileId} reconnecting started...`);
        eventBus.emit('connection:status-changed', { status: 'reconnecting', profileId });
      },

      stopReconnecting: (profileId: string) => {
        get().ensureProfile(profileId);
        if (get().connections[profileId]?.isReconnecting) {
          const previousStatus = get().connections[profileId].error ? 'error' : 'disconnected';
          set(state => ({
            connections: {
              ...state.connections,
              [profileId]: {
                ...state.connections[profileId],
                isReconnecting: false,
                status: previousStatus,
              },
            },
          }));
          logger.debug(`[ConnectionStore] Profile ${profileId} reconnecting stopped.`);
        }
      },
    }),
    { name: 'ConnectionStore', store: 'connection' },
  ),
);
