import { StateCreator } from 'zustand';
import { eventBus } from '../../events';
import type { UserPreferences, SidebarState, Notification, GlobalSettings } from '../../types/stores';
import type { RemoteNotification } from '../config.store';
import { createLogger } from '@extension/shared/lib/logger';
import type { RootState } from '../root.store';

const logger = createLogger('UISlice');

const initialSidebarState: SidebarState = {
  isVisible: true,
  isMinimized: false,
  position: 'left',
  width: 320,
};

const initialUserPreferences: UserPreferences = {
  autoSubmit: false,
  autoInsert: false,
  autoExecute: false,
  notifications: true,
  theme: 'system',
  language: navigator.language || 'en-US',
  isPushMode: false,
  sidebarWidth: 320,
  isMinimized: false,
  customInstructions: '',
  customInstructionsEnabled: false,
  autoInsertDelay: 2,
  autoExecuteDelay: 2,
  autoSubmitDelay: 2,
  accentColor: 'indigo',
  autoExecuteWhitelist: [],
};

export interface UISlice {
  ui: {
    sidebar: SidebarState;
    preferences: UserPreferences;
    notifications: Notification[];
    activeModal: string | null;
    isLoading: boolean;
    theme: GlobalSettings['theme'];
    mcpEnabled: boolean;
    
    // Formerly AppStore state
    isInitialized: boolean;
    initializationError: string | null;
    currentSite: string;
    currentHost: string;
    globalSettings: GlobalSettings;
  };

  // UI Actions
  toggleSidebar: (reason?: string) => void;
  toggleMinimize: (reason?: string) => void;
  resizeSidebar: (width: number) => void;
  setSidebarVisibility: (visible: boolean, reason?: string) => void;
  updatePreferences: (prefs: Partial<UserPreferences>) => void;
  addNotification: (notificationData: Omit<Notification, 'id' | 'timestamp'>) => string;
  addRemoteNotification: (notification: RemoteNotification) => string;
  removeNotification: (id: string) => void;
  dismissNotification: (id: string, reason?: string) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  clearNotifications: () => void;
  openModal: (modalName: string) => void;
  closeModal: () => void;
  setGlobalLoading: (loading: boolean) => void;
  setTheme: (theme: GlobalSettings['theme']) => void;
  setMCPEnabled: (enabled: boolean, reason?: string) => void;

  // App Actions
  initializeAppInfo: () => Promise<void>;
  setCurrentSite: (siteInfo: { site: string; host: string }) => void;
  updateGlobalSettings: (settings: Partial<GlobalSettings>) => void;
  resetUIState: () => void;
}

const initialUIState = {
  sidebar: initialSidebarState,
  preferences: initialUserPreferences,
  notifications: [],
  activeModal: null,
  isLoading: false,
  theme: initialUserPreferences.theme,
  mcpEnabled: true,
  
  isInitialized: false,
  initializationError: null,
  currentSite: window.location.href,
  currentHost: window.location.hostname,
  globalSettings: {
    theme: 'system' as GlobalSettings['theme'],
    autoSubmit: false,
    debugMode: false,
    sidebarWidth: 320,
    isPushMode: false,
    language: navigator.language || 'en-US',
    notifications: true,
  },
};

export const createUISlice: StateCreator<RootState, [], [], UISlice> = (set, get) => ({
  ui: initialUIState,

  toggleSidebar: (reason?: string) => {
    const newVisibility = !get().ui.sidebar.isVisible;
    set(state => ({ ui: { ...state.ui, sidebar: { ...state.ui.sidebar, isVisible: newVisibility } } }));
    logger.debug(`Sidebar toggled to ${newVisibility ? 'visible' : 'hidden'}. Reason: ${reason || 'user action'}`);
    eventBus.emit('ui:sidebar-toggle', { visible: newVisibility, reason: reason || 'user action' });
  },

  toggleMinimize: (reason?: string) => {
    const newMinimized = !get().ui.sidebar.isMinimized;
    set(state => ({
      ui: { 
        ...state.ui, 
        sidebar: { ...state.ui.sidebar, isMinimized: newMinimized },
        preferences: { ...state.ui.preferences, isMinimized: newMinimized }
      }
    }));
    logger.debug(`Sidebar ${newMinimized ? 'minimized' : 'expanded'}. Reason: ${reason || 'user action'}`);
    eventBus.emit('ui:sidebar-minimize', { minimized: newMinimized, reason: reason || 'user action' });
  },

  resizeSidebar: (width: number) => {
    set(state => ({ ui: { ...state.ui, sidebar: { ...state.ui.sidebar, width } } }));
    logger.debug(`Sidebar resized to: ${width}px`);
    eventBus.emit('ui:sidebar-resize', { width });
  },

  setSidebarVisibility: (visible: boolean, reason?: string) => {
    set(state => ({ ui: { ...state.ui, sidebar: { ...state.ui.sidebar, isVisible: visible } } }));
    logger.debug(`Sidebar visibility set to ${visible}. Reason: ${reason || 'programmatic'}`);
    eventBus.emit('ui:sidebar-toggle', { visible, reason: reason || 'programmatic' });
  },

  updatePreferences: (prefs: Partial<UserPreferences>) => {
    const oldPrefs = get().ui.preferences;
    const newPrefs = { ...oldPrefs, ...prefs };
    set(state => ({ ui: { ...state.ui, preferences: newPrefs } }));
    logger.debug('[UISlice] Preferences updated:', newPrefs);
    eventBus.emit('ui:preferences-updated', { preferences: newPrefs });
    
    // Deduplicate theme update
    if (prefs.theme && prefs.theme !== oldPrefs.theme) {
      get().setTheme(prefs.theme);
    }
  },

  addNotification: (notificationData: Omit<Notification, 'id' | 'timestamp'>): string => {
    const newNotification: Notification = {
      ...notificationData,
      id: `notif_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
      timestamp: Date.now(),
    };
    set(state => ({ ui: { ...state.ui, notifications: [...state.ui.notifications, newNotification] } }));
    logger.debug('[UISlice] Notification added:', newNotification);
    eventBus.emit('ui:notification-added', { notification: newNotification });
    return newNotification.id;
  },

  addRemoteNotification: (notification: RemoteNotification): string => {
    const configStore = get().config;

    // Check if notifications are enabled
    if (!configStore.notificationConfig.enabled) {
      logger.debug('[UISlice] Remote notifications disabled, ignoring:', notification.id);
      return '';
    }

    // Check frequency limits
    const today = new Date().toDateString();
    const todayNotifications = get().ui.notifications.filter(
      n => new Date(n.timestamp).toDateString() === today && 'source' in n && n.source === 'remote',
    ).length;

    if (todayNotifications >= configStore.notificationConfig.maxPerDay) {
      logger.debug('[UISlice] Daily notification limit reached, ignoring:', notification.id);
      eventBus.emit('notification:frequency-limited', {
        notificationId: notification.id,
        reason: 'Daily limit exceeded',
      });
      return '';
    }

    // Create enhanced notification
    const newNotification: Notification & {
      source: 'remote';
      campaignId?: string;
      actions?: any[];
      priority?: number;
    } = {
      id: notification.id || `remote_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
      type: notification.type,
      title: notification.title,
      message: notification.message,
      duration: notification.duration,
      timestamp: Date.now(),
      source: 'remote',
      campaignId: notification.campaignId,
      actions: notification.actions,
      priority: notification.priority || 1,
    };

    set(state => ({
      ui: {
        ...state.ui,
        notifications: [...state.ui.notifications, newNotification].sort(
          (a, b) => ('priority' in b ? (b as any).priority : 1) - ('priority' in a ? (a as any).priority : 1),
        ),
      }
    }));

    get().markNotificationShown(newNotification.id);
    get().addNotificationToHistory(newNotification.id);

    eventBus.emit('ui:notification-added', { notification: newNotification });
    eventBus.emit('notification:shown', {
      notificationId: newNotification.id,
      source: 'remote',
      timestamp: Date.now(),
    });

    eventBus.emit('analytics:track', {
      event: 'notification_shown',
      parameters: {
        notification_id: newNotification.id,
        campaign_id: notification.campaignId,
        type: notification.type,
        source: 'remote',
      },
    });

    logger.debug('[UISlice] Remote notification added:', newNotification);
    return newNotification.id;
  },

  removeNotification: (id: string) => {
    set(state => ({ ui: { ...state.ui, notifications: state.ui.notifications.filter(n => n.id !== id) } }));
    logger.debug(`Notification removed: ${id}`);
    eventBus.emit('ui:notification-removed', { id });
  },

  dismissNotification: (id: string, reason?: string) => {
    const notification = get().ui.notifications.find(n => n.id === id);
    if (notification) {
      if ('source' in notification && notification.source === 'remote') {
        eventBus.emit('notification:dismissed', {
          notificationId: id,
          reason: reason || 'user_dismissed',
          timestamp: Date.now(),
        });

        eventBus.emit('analytics:track', {
          event: 'notification_dismissed',
          parameters: {
            notification_id: id,
            campaign_id: 'campaignId' in notification ? (notification as any).campaignId : undefined,
            reason: reason || 'user_dismissed',
            source: 'remote',
          },
        });
      }
    }
    get().removeNotification(id);
  },

  markAsRead: (id: string) => {
    set(state => ({
      ui: {
        ...state.ui,
        notifications: state.ui.notifications.map(n =>
          n.id === id ? { ...n, read: true } : n
        )
      }
    }));
    logger.debug(`Notification marked as read: ${id}`);
  },

  markAllAsRead: () => {
    set(state => ({
      ui: {
        ...state.ui,
        notifications: state.ui.notifications.map(n => ({ ...n, read: true }))
      }
    }));
    logger.debug('[UISlice] All notifications marked as read.');
  },

  clearNotifications: () => {
    get().ui.notifications.forEach(n => eventBus.emit('ui:notification-removed', { id: n.id }));
    set(state => ({ ui: { ...state.ui, notifications: [] } }));
    logger.debug('[UISlice] All notifications cleared.');
  },

  openModal: (modalName: string) => {
    set(state => ({ ui: { ...state.ui, activeModal: modalName } }));
    logger.debug(`Modal opened: ${modalName}`);
  },

  closeModal: () => {
    logger.debug(`Modal closed: ${get().ui.activeModal}`);
    set(state => ({ ui: { ...state.ui, activeModal: null } }));
  },

  setGlobalLoading: (loading: boolean) => {
    set(state => ({ ui: { ...state.ui, isLoading: loading } }));
    logger.debug(`Global loading state: ${loading}`);
  },

  setTheme: (theme: GlobalSettings['theme']) => {
    set(state => ({ ui: { ...state.ui, theme } }));
    logger.debug(`Theme changed to: ${theme}`);
    eventBus.emit('ui:theme-changed', { theme });
    
    if (get().ui.preferences.theme !== theme) {
      set(state => ({ ui: { ...state.ui, preferences: { ...state.ui.preferences, theme } } }));
      eventBus.emit('ui:preferences-updated', { preferences: get().ui.preferences });
    }
    
    // Sync the old redundant app-store global settings theme too
    if (get().ui.globalSettings.theme !== theme) {
      set(state => ({ ui: { ...state.ui, globalSettings: { ...state.ui.globalSettings, theme } } }));
    }
  },

  setMCPEnabled: (enabled: boolean, reason?: string) => {
    const previousState = get().ui.mcpEnabled;
    set(state => ({ ui: { ...state.ui, mcpEnabled: enabled } }));

    logger.debug(`MCP toggle set to ${enabled}. Reason: ${reason || 'user action'}`);

    if (enabled !== previousState && !enabled) {
      get().setSidebarVisibility(false, reason || 'mcp-toggle');
    }

    eventBus.emit('ui:mcp-toggle', { enabled, reason: reason || 'user action', previousState });
  },

  // App Actions
  initializeAppInfo: async () => {
    if (get().ui.isInitialized) {
      logger.debug('[UISlice] Already initialized.');
      return;
    }

    logger.debug('[UISlice] Initializing app info...');
    try {
      set(state => ({ ui: { ...state.ui, initializationError: null } }));

      set(state => ({ ui: { ...state.ui, isInitialized: true, initializationError: null } }));
      logger.debug('[UISlice] Initialization complete.');
      const version = typeof chrome !== 'undefined' && chrome.runtime?.getManifest?.()?.version || '0.0.0';
      eventBus.emit('app:initialized', { version, timestamp: Date.now() });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown initialization error';
      logger.error('[UISlice] Initialization failed:', errorMessage, error);
      set(state => ({ ui: { ...state.ui, isInitialized: false, initializationError: errorMessage } }));
    }
  },

  setCurrentSite: (siteInfo: { site: string; host: string }) => {
    set(state => ({
      ui: { ...state.ui, currentSite: siteInfo.site, currentHost: siteInfo.host },
    }));
    logger.debug(`Site changed to: ${siteInfo.site}`);
    eventBus.emit('app:site-changed', { site: siteInfo.site, hostname: siteInfo.host });
  },

  updateGlobalSettings: (settings: Partial<GlobalSettings>) => {
    set(state => ({
      ui: { ...state.ui, globalSettings: { ...state.ui.globalSettings, ...settings } },
    }));
    logger.debug('[UISlice] Settings updated:', settings);
    eventBus.emit('app:settings-updated', { settings });
    
    // Deduplicate theme update if changed from globalSettings sync
    if (settings.theme && settings.theme !== get().ui.theme) {
      get().setTheme(settings.theme);
    }
  },

  resetUIState: () => {
    logger.debug('[UISlice] Resetting state.');
    set(state => ({ ui: initialUIState }));
  },
});
