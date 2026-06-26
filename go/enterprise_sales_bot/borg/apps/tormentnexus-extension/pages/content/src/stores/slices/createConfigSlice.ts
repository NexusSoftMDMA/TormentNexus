import { StateCreator } from 'zustand';
import type { RootState } from '../root.store';

export interface FeatureFlag {
  enabled: boolean;
  rollout: number;
  config?: Record<string, any>;
  dependencies?: string[];
  targeting?: {
    versions?: string[];
    regions?: string[];
    userSegments?: string[];
  };
}

export interface UserProperties {
  extensionVersion: string;
  installDate: string;
  usageDays: number;
  featuresUsed: string[];
  userSegment: string;
  sessionCount: number;
  lastActiveDate: string;
  browserVersion: string;
  platform: string;
  language: string;
  timezone: string;
}

export interface NotificationConfig {
  enabled: boolean;
  maxPerDay: number;
  cooldownHours: number;
  respectDoNotDisturb: boolean;
  channels: string[];
}

export interface RemoteNotification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  duration?: number;
  actions?: NotificationAction[];
  targeting?: NotificationTargeting;
  campaignId?: string;
  priority?: number;
  expiresAt?: string;
}

export interface NotificationAction {
  text: string;
  action: string;
  style?: 'primary' | 'secondary' | 'danger';
}

export interface NotificationTargeting {
  versions?: string[];
  userSegments?: string[];
  featureFlags?: string[];
  regions?: string[];
  installDateRange?: {
    start?: string;
    end?: string;
  };
}

export interface ConfigSlice {
  config: {
    featureFlags: Record<string, FeatureFlag>;
    lastFetchTime: number | null;
    lastUpdateTime: number | null;
    isLoading: boolean;
    error: string | null;
    userProperties: UserProperties;
    userSegment: string;
    notificationConfig: NotificationConfig;
    shownNotifications: string[];
    notificationHistory: Array<{ id: string; shownAt: number; action?: string }>;
  };

  updateFeatureFlags: (flags: Record<string, FeatureFlag>) => void;
  setUserProperties: (properties: Partial<UserProperties>) => void;
  setUserSegment: (segment: string) => void;
  updateNotificationConfig: (config: Partial<NotificationConfig>) => void;
  markNotificationShown: (notificationId: string) => void;
  addNotificationToHistory: (notificationId: string, action?: string) => void;
  isFeatureEnabled: (featureName: string) => boolean;
  getFeatureConfig: (featureName: string) => FeatureFlag | undefined;
  setConfigLoading: (loading: boolean) => void;
  setConfigError: (error: string | null) => void;
  updateLastFetchTime: (timestamp: number) => void;
  canShowNotification: (notification: RemoteNotification) => boolean;
  resetConfigState: () => void;
}

const initialUserProperties: UserProperties = {
  extensionVersion: typeof chrome !== 'undefined' ? chrome?.runtime?.getManifest?.()?.version || '0.0.0' : '0.0.0',
  installDate: new Date().toISOString(),
  usageDays: 0,
  featuresUsed: [],
  userSegment: 'new',
  sessionCount: 0,
  lastActiveDate: new Date().toISOString(),
  browserVersion: navigator.userAgent,
  platform: navigator.platform,
  language: navigator.language,
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
};

const initialNotificationConfig: NotificationConfig = {
  enabled: true,
  maxPerDay: 3,
  cooldownHours: 4,
  respectDoNotDisturb: true,
  channels: ['in-app'],
};

const initialConfigState = {
  featureFlags: {},
  lastFetchTime: null,
  lastUpdateTime: null,
  isLoading: false,
  error: null,
  userProperties: initialUserProperties,
  userSegment: 'new',
  notificationConfig: initialNotificationConfig,
  shownNotifications: [],
  notificationHistory: [],
};

function hashUserProperties(properties: UserProperties): number {
  const userString = JSON.stringify({
    version: properties.extensionVersion,
    install: properties.installDate.split('T')[0],
    segment: properties.userSegment,
    platform: properties.platform,
    language: properties.language,
  });

  let hash = 0;
  for (let i = 0; i < userString.length; i++) {
    const char = userString.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }
  return Math.abs(hash);
}

export const createConfigSlice: StateCreator<RootState, [], [], ConfigSlice> = (set, get) => ({
  config: initialConfigState,

  updateFeatureFlags: (flags: Record<string, FeatureFlag>) => {
    set(state => ({
      config: {
        ...state.config,
        featureFlags: { ...state.config.featureFlags, ...flags },
        lastUpdateTime: Date.now(),
      }
    }));
  },

  setUserProperties: (properties: Partial<UserProperties>) => {
    set(state => {
      const newProperties = { ...state.config.userProperties, ...properties };
      let segment = 'new';
      if (newProperties.usageDays > 30) segment = 'power';
      else if (newProperties.usageDays > 7) segment = 'regular';
      else if (newProperties.sessionCount > 5) segment = 'engaged';

      return {
        config: {
          ...state.config,
          userProperties: newProperties,
          userSegment: segment,
        }
      };
    });
  },

  setUserSegment: (segment: string) => {
    set(state => ({ config: { ...state.config, userSegment: segment } }));
  },

  updateNotificationConfig: (config: Partial<NotificationConfig>) => {
    set(state => ({
      config: {
        ...state.config,
        notificationConfig: { ...state.config.notificationConfig, ...config },
      }
    }));
  },

  markNotificationShown: (notificationId: string) => {
    set(state => ({
      config: {
        ...state.config,
        shownNotifications: [...state.config.shownNotifications, notificationId],
      }
    }));
  },

  addNotificationToHistory: (notificationId: string, action?: string) => {
    set(state => ({
      config: {
        ...state.config,
        notificationHistory: [
          ...state.config.notificationHistory,
          { id: notificationId, shownAt: Date.now(), action },
        ].slice(-100),
      }
    }));
  },

  isFeatureEnabled: (featureName: string): boolean => {
    const state = get().config;
    const feature = state.featureFlags[featureName];

    if (!feature) return false;
    if (!feature.enabled) return false;

    if (feature.rollout < 100) {
      const userHash = hashUserProperties(state.userProperties);
      const rolloutThreshold = (feature.rollout / 100) * 100;
      if (userHash % 100 >= rolloutThreshold) return false;
    }

    if (feature.targeting) {
      const { userProperties, userSegment } = state;

      if (feature.targeting.versions && feature.targeting.versions.length > 0) {
        const currentVersion = userProperties.extensionVersion;
        const versionMatch = feature.targeting.versions.some(targetVersion =>
          currentVersion.startsWith(targetVersion),
        );
        if (!versionMatch) return false;
      }

      if (feature.targeting.userSegments && feature.targeting.userSegments.length > 0) {
        if (!feature.targeting.userSegments.includes(userSegment)) return false;
      }
    }

    return true;
  },

  getFeatureConfig: (featureName: string): FeatureFlag | undefined => {
    return get().config.featureFlags[featureName];
  },

  setConfigLoading: (loading: boolean) => {
    set(state => ({ config: { ...state.config, isLoading: loading } }));
  },

  setConfigError: (error: string | null) => {
    set(state => ({ config: { ...state.config, error } }));
  },

  updateLastFetchTime: (timestamp: number) => {
    set(state => ({ config: { ...state.config, lastFetchTime: timestamp } }));
  },

  canShowNotification: (notification: RemoteNotification): boolean => {
    const state = get().config;
    const { notificationConfig, shownNotifications, notificationHistory } = state;

    if (!notificationConfig.enabled) return false;
    if (shownNotifications.includes(notification.id)) return false;
    if (notification.expiresAt && new Date(notification.expiresAt) < new Date()) return false;

    const today = new Date().toDateString();
    const todayNotifications = notificationHistory.filter(
      n => new Date(n.shownAt).toDateString() === today,
    ).length;

    if (todayNotifications >= notificationConfig.maxPerDay) return false;

    const lastNotificationTime = Math.max(...notificationHistory.map(n => n.shownAt), 0);
    const cooldownMs = notificationConfig.cooldownHours * 60 * 60 * 1000;
    if (Date.now() - lastNotificationTime < cooldownMs) return false;

    if (notification.targeting) {
      const { userProperties, userSegment } = state;

      if (notification.targeting.versions && notification.targeting.versions.length > 0) {
        const versionMatch = notification.targeting.versions.some(targetVersion =>
          userProperties.extensionVersion.startsWith(targetVersion),
        );
        if (!versionMatch) return false;
      }

      if (notification.targeting.userSegments && notification.targeting.userSegments.length > 0) {
        if (!notification.targeting.userSegments.includes(userSegment)) return false;
      }

      if (notification.targeting.installDateRange) {
        const installDate = new Date(userProperties.installDate);
        const { start, end } = notification.targeting.installDateRange;

        if (start && installDate < new Date(start)) return false;
        if (end && installDate > new Date(end)) return false;
      }
    }

    return true;
  },

  resetConfigState: () => {
    set(state => ({ config: initialConfigState }));
  },
});
