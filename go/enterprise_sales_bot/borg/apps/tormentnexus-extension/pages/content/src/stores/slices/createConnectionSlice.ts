import { StateCreator } from 'zustand';
import { eventBus } from '../../events';
import type { ConnectionStatus, ServerConfig, MCPTelemetry, ConnectionType } from '../../types/stores';
import { createLogger } from '@extension/shared/lib/logger';
import type { RootState } from '../root.store';

const logger = createLogger('ConnectionSlice');

export interface ConnectionProfile {
  id: string;
  name: string;
  uri: string;
  connectionType: ConnectionType;
}

export interface ProfileConnectionState {
  status: ConnectionStatus;
  serverConfig: ServerConfig;
  lastConnectedAt: number | null;
  connectionAttempts: number;
  error: string | null;
  isReconnecting: boolean;
  telemetry?: MCPTelemetry;
}

export interface ConnectionSlice {
  connection: {
    // Profile State
    profiles: ConnectionProfile[];
    activeProfileIds: string[];
    
    // Connection State
    connections: Record<string, ProfileConnectionState>;
  };

  // Profile Actions
  addProfile: (profile: Omit<ConnectionProfile, 'id'>) => void;
  removeProfile: (id: string) => void;
  updateProfile: (id: string, updates: Partial<ConnectionProfile>) => void;
  toggleProfileActive: (id: string) => void;
  setProfilesActive: (ids: string[]) => void;

  // Connection Actions
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
}

const defaultServerConfig: ServerConfig = {
  uri: 'http://localhost:3006/sse',
  connectionType: 'sse',
  timeout: 5000,
  retryAttempts: 3,
  retryDelay: 2000,
};

const defaultProfileState: ProfileConnectionState = {
  status: 'disconnected',
  serverConfig: defaultServerConfig,
  lastConnectedAt: null,
  connectionAttempts: 0,
  error: null,
  isReconnecting: false,
};

const initialConnectionState = {
  profiles: [
    {
      id: 'default-sse',
      name: 'Default (SSE)',
      uri: 'http://localhost:3006/sse',
      connectionType: 'sse' as ConnectionType,
    },
    {
      id: 'default-ws',
      name: 'Default (WebSocket)',
      uri: 'ws://localhost:3006/message',
      connectionType: 'websocket' as ConnectionType,
    },
  ],
  activeProfileIds: ['default-sse'],
  connections: {},
};

export const createConnectionSlice: StateCreator<RootState, [], [], ConnectionSlice> = (set, get) => ({
  connection: initialConnectionState,

  // --- Profile Actions ---
  addProfile: profile =>
    set(state => ({
      connection: {
        ...state.connection,
        profiles: [...state.connection.profiles, { ...profile, id: crypto.randomUUID() }],
      }
    })),
    
  removeProfile: id =>
    set(state => {
      const { [id]: _, ...restConnections } = state.connection.connections;
      return {
        connection: {
          ...state.connection,
          profiles: state.connection.profiles.filter(p => p.id !== id),
          activeProfileIds: state.connection.activeProfileIds.filter(activeId => activeId !== id),
          connections: restConnections,
        }
      };
    }),
    
  updateProfile: (id, updates) =>
    set(state => ({
      connection: {
        ...state.connection,
        profiles: state.connection.profiles.map(p => (p.id === id ? { ...p, ...updates } : p)),
      }
    })),
    
  toggleProfileActive: id => 
    set(state => ({
      connection: {
        ...state.connection,
        activeProfileIds: state.connection.activeProfileIds.includes(id) 
          ? state.connection.activeProfileIds.filter(activeId => activeId !== id)
          : [...state.connection.activeProfileIds, id]
      }
    })),
    
  setProfilesActive: ids => 
    set(state => ({ connection: { ...state.connection, activeProfileIds: ids } })),


  // --- Connection Actions ---
  ensureProfile: (profileId: string) => {
    if (!get().connection.connections[profileId]) {
      set(state => ({
        connection: {
          ...state.connection,
          connections: {
            ...state.connection.connections,
            [profileId]: { ...defaultProfileState },
          },
        }
      }));
    }
  },

  getProfileState: (profileId: string) => {
    return get().connection.connections[profileId] || defaultProfileState;
  },

  setStatus: (profileId: string, status: ConnectionStatus) => {
    get().ensureProfile(profileId);
    const oldStatus = get().connection.connections[profileId]?.status;
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            status,
          },
        },
      }
    }));
    logger.debug(`Profile ${profileId} status changed from ${oldStatus} to: ${status}`);
    eventBus.emit('connection:status-changed', { status, error: get().connection.connections[profileId]?.error || undefined, profileId });
  },

  setServerConfig: (profileId: string, config: Partial<ServerConfig>) => {
    get().ensureProfile(profileId);
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            serverConfig: { ...state.connection.connections[profileId].serverConfig, ...config },
          },
        },
      }
    }));
    logger.debug(`[ConnectionSlice] Profile ${profileId} server config updated:`, get().connection.connections[profileId].serverConfig);
  },

  setLastError: (profileId: string, error: string | null) => {
    get().ensureProfile(profileId);
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            error,
          },
        },
      }
    }));
    if (error) {
      logger.error(`[ConnectionSlice] Profile ${profileId} error set:`, error);
      eventBus.emit('connection:error', { error: error, profileId });
    }
  },

  setTelemetry: (profileId: string, telemetry: MCPTelemetry) => {
    get().ensureProfile(profileId);
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            telemetry,
          },
        },
      }
    }));
    eventBus.emit('connection:telemetry', { telemetry, profileId });
  },

  incrementAttempts: (profileId: string) => {
    get().ensureProfile(profileId);
    const newAttempts = (get().connection.connections[profileId]?.connectionAttempts || 0) + 1;
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            connectionAttempts: newAttempts,
          },
        },
      }
    }));
    logger.debug(`Profile ${profileId} connection attempts: ${newAttempts}`);
    eventBus.emit('connection:attempt', { attempt: newAttempts, maxAttempts: get().connection.connections[profileId].serverConfig.retryAttempts, profileId });
  },

  resetAttempts: (profileId: string) => {
    get().ensureProfile(profileId);
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            connectionAttempts: 0,
          },
        },
      }
    }));
    logger.debug(`[ConnectionSlice] Profile ${profileId} connection attempts reset.`);
  },

  setConnected: (profileId: string, timestamp: number) => {
    get().ensureProfile(profileId);
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            status: 'connected',
            lastConnectedAt: timestamp,
            connectionAttempts: 0,
            error: null,
            isReconnecting: false,
          },
        },
      }
    }));
    logger.debug(`Profile ${profileId} connected at: ${new Date(timestamp).toISOString()}`);
    eventBus.emit('connection:status-changed', { status: 'connected', profileId });
  },

  setDisconnected: (profileId: string, error?: string) => {
    get().ensureProfile(profileId);
    set(state => {
      const profileConn = state.connection.connections[profileId];
      return {
        connection: {
          ...state.connection,
          connections: {
            ...state.connection.connections,
            [profileId]: {
              ...profileConn,
              status: error ? 'error' : 'disconnected',
              error: error || profileConn.error,
              isReconnecting: false,
            },
          },
        }
      };
    });
    logger.debug(`Profile ${profileId} disconnected. ${error ? 'Error: ' + error : ''}`);
    eventBus.emit('connection:status-changed', { status: get().connection.connections[profileId].status, error: error || get().connection.connections[profileId].error || undefined, profileId });
  },

  startReconnecting: (profileId: string) => {
    get().ensureProfile(profileId);
    if (get().connection.connections[profileId]?.status === 'connected') return;
    set(state => ({
      connection: {
        ...state.connection,
        connections: {
          ...state.connection.connections,
          [profileId]: {
            ...state.connection.connections[profileId],
            isReconnecting: true,
            status: 'reconnecting',
          },
        },
      }
    }));
    logger.debug(`[ConnectionSlice] Profile ${profileId} reconnecting started...`);
    eventBus.emit('connection:status-changed', { status: 'reconnecting', profileId });
  },

  stopReconnecting: (profileId: string) => {
    get().ensureProfile(profileId);
    if (get().connection.connections[profileId]?.isReconnecting) {
      const previousStatus = get().connection.connections[profileId].error ? 'error' : 'disconnected';
      set(state => ({
        connection: {
          ...state.connection,
          connections: {
            ...state.connection.connections,
            [profileId]: {
              ...state.connection.connections[profileId],
              isReconnecting: false,
              status: previousStatus,
            },
          },
        }
      }));
      logger.debug(`[ConnectionSlice] Profile ${profileId} reconnecting stopped.`);
    }
  },
});
