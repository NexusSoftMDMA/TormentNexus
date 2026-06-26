import { createLogger } from '@extension/shared/lib/logger';
import { useRootStore } from './root.store';
import type { RootState } from './root.store';
import type { UISlice } from './slices/createUISlice';
import type { ConfigSlice } from './slices/createConfigSlice';
import type { ToolSlice } from './slices/createToolSlice';
import type { ResourceSlice } from './slices/createResourceSlice';
import type { ConnectionSlice } from './slices/createConnectionSlice';
import type { AdapterSlice } from './slices/createAdapterSlice';

const logger = createLogger('Stores');

// ── Re-exports ────────────────────────────────────────────────────────
export { useRootStore } from './root.store';
export type { RootState } from './root.store';

// ── Type-safe Store API helper ────────────────────────────────────────
/**
 * Utility type that mirrors the Zustand store API surface.
 * Used for backward-compatible wrappers so consumers keep full type inference.
 */
interface StoreApi<S> {
  (): S;
  <T>(selector: (state: S) => T): T;
  getState: () => S;
  setState: (partial: Partial<S> | ((state: S) => Partial<S>), replace?: false) => void;
  subscribe: (listener: (state: S, prevState: S) => void) => () => void;
}

/**
 * Create a type-safe backward-compatible store wrapper that maps to a slice of the Root Store.
 */
function createSliceWrapper<K extends keyof RootState>(sliceKey: K): StoreApi<RootState[K]> {
  const hook = <T>(selector?: (state: RootState[K]) => T): T | RootState[K] => {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    return useRootStore(state => (selector ? selector(state[sliceKey]) : state[sliceKey]));
  };

  return Object.assign(hook, {
    getState: (): RootState[K] => useRootStore.getState()[sliceKey],
    setState: (partial: Partial<RootState[K]> | ((state: RootState[K]) => Partial<RootState[K]>), replace?: false): void => {
      useRootStore.setState(
        (state: RootState) => ({
          [sliceKey]: {
            ...state[sliceKey],
            ...(typeof partial === 'function' ? partial(state[sliceKey]) : partial),
          },
        } as Partial<RootState>),
        replace
      );
    },
    subscribe: (listener: (state: RootState[K], prevState: RootState[K]) => void): (() => void) => {
      let prevState = useRootStore.getState()[sliceKey];
      return useRootStore.subscribe((state) => {
        if (state[sliceKey] !== prevState) {
          const newSlice = state[sliceKey];
          listener(newSlice, prevState);
          prevState = newSlice;
        }
      });
    },
  }) as StoreApi<RootState[K]>;
}

// ── Legacy Store Mappings (Backwards Compatibility) ───────────────────
// These intercept calls to legacy stores and map them to unified Root Store slices.
// Full type safety — no `as any` casts.

export const useUIStore = createSliceWrapper('ui');
export const useAppStore = createSliceWrapper('ui'); // AppStore merged into UI
export const useConfigStore = createSliceWrapper('config');
export const useToolStore = createSliceWrapper('tool');
export const useResourceStore = createSliceWrapper('resource');
export const useConnectionStore = createSliceWrapper('connection');
export const useProfileStore = createSliceWrapper('connection'); // Profile merged into Connection
export const useAdapterStore = createSliceWrapper('adapter');

// ── Type Exports (canonical source: slice files) ──────────────────────
// Backward-compatible type aliases mapping legacy names to slice types.
export type UIState = UISlice;
export type AppState = UISlice; // AppStore merged into UI
export type ConfigState = ConfigSlice;
export type ConnectionState = ConnectionSlice;
export type ToolState = ToolSlice;
export type ResourceState = ResourceSlice;
export type AdapterState = AdapterSlice;

// Re-export slice-native types for consumers
export type { UISlice } from './slices/createUISlice';
export type { ConfigSlice, FeatureFlag, UserProperties, NotificationConfig, RemoteNotification } from './slices/createConfigSlice';
export type { ConnectionSlice, ConnectionProfile, ProfileConnectionState } from './slices/createConnectionSlice';
export type { ToolSlice } from './slices/createToolSlice';
export type { ResourceSlice, Resource, ResourceTemplate } from './slices/createResourceSlice';
export type { AdapterSlice } from './slices/createAdapterSlice';

// ── Satellite Stores (not sliced — small, domain-specific, standalone) ─
export { useDebuggerStore, type DebugPacket } from './debugger.store';
export { useMacroStore, type Macro } from './macro.store';
export { useContextStore, type ContextItem } from './context.store';
export { useToastStore, type Toast } from './toast.store';
export { useActivityStore, type LogType, type LogStatus } from './activity.store';
export { usePromptStore, type PromptTemplate } from './prompt.store';

// ── Initialization ────────────────────────────────────────────────────
export async function initializeAllStores(): Promise<void> {
  logger.debug('Initializing all stores...');

  useRootStore.getState();

  const rootState = useRootStore.getState();

  rootState.setUserProperties({
    extensionVersion: chrome?.runtime?.getManifest?.()?.version || '0.0.0',
    lastActiveDate: new Date().toISOString(),
    browserVersion: navigator.userAgent,
    platform: navigator.platform,
    language: navigator.language,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  });

  if (!rootState.ui.isInitialized) {
    await rootState.initializeAppInfo();
  }

  logger.debug('All stores accessed/initialized.');
}
