import { logMessage } from './helpers';

// Types
export interface SidebarPreferences {
  isPushMode: boolean;
  sidebarWidth: number;
  isMinimized: boolean;
  autoSubmit: boolean;
  theme: 'light' | 'dark' | 'system';
  customInstructions: string;
  customInstructionsEnabled: boolean;
}

const STORAGE_KEY = 'mcp_sidebar_preferences';
const TOOL_ENABLEMENT_KEY = 'mcp_tool_enablement';

// Default preferences
const DEFAULT_PREFERENCES: SidebarPreferences = {
  isPushMode: false,
  sidebarWidth: 320,
  isMinimized: false,
  autoSubmit: false,
  theme: 'system',
  customInstructions: '',
  customInstructionsEnabled: false,
};

/**
 * Get sidebar preferences from chrome.storage.local
 */
export const getSidebarPreferences = async (): Promise<SidebarPreferences> => {
  try {
    if (!chrome.storage || !chrome.storage.local) {
      logMessage('[Storage] Chrome storage API not available');
      return DEFAULT_PREFERENCES;
    }

    const result = await chrome.storage.local.get(STORAGE_KEY);
    const preferences = result && typeof result === 'object' ? (result[STORAGE_KEY] as SidebarPreferences) : undefined;

    if (!preferences) {
      logMessage('[Storage] No stored sidebar preferences found, using defaults');
      return DEFAULT_PREFERENCES;
    }

    logMessage('[Storage] Retrieved sidebar preferences from storage');
    return {
      ...DEFAULT_PREFERENCES,
      ...(preferences || {}),
    };
  } catch (error) {
    logMessage(
      `[Storage] Error retrieving sidebar preferences: ${error instanceof Error ? error.message : String(error)}`,
    );
    return DEFAULT_PREFERENCES;
  }
};

/**
 * Save sidebar preferences to chrome.storage.local
 */
export const saveSidebarPreferences = async (preferences: Partial<SidebarPreferences>): Promise<void> => {
  try {
    if (!chrome.storage || !chrome.storage.local) {
      logMessage('[Storage] Chrome storage API not available');
      return;
    }

    // Get current preferences first to merge with new ones
    const currentPrefs = await getSidebarPreferences();
    const updatedPrefs = {
      ...currentPrefs,
      ...preferences,
    };

    await chrome.storage.local.set({ [STORAGE_KEY]: updatedPrefs });
    logMessage(`[Storage] Saved sidebar preferences: ${JSON.stringify(updatedPrefs)}`);
  } catch (error) {
    logMessage(`[Storage] Error saving sidebar preferences: ${error instanceof Error ? error.message : String(error)}`);
  }
};

/**
 * Get tool enablement state from chrome.storage.local
 * Returns a Set of enabled tool names
 */
export const getToolEnablementState = async (): Promise<Set<string>> => {
  try {
    if (!chrome.storage || !chrome.storage.local) {
      logMessage('[Storage] Chrome storage API not available');
      return new Set();
    }

    const result = await chrome.storage.local.get(TOOL_ENABLEMENT_KEY);
    const enabledToolsArray =
      result && typeof result === 'object' ? (result[TOOL_ENABLEMENT_KEY] as string[]) : undefined;

    if (!enabledToolsArray || !Array.isArray(enabledToolsArray)) {
      logMessage('[Storage] No stored tool enablement state found, returning empty set');
      return new Set();
    }

    logMessage(`[Storage] Retrieved tool enablement state: ${enabledToolsArray.length} enabled tools`);
    return new Set(enabledToolsArray);
  } catch (error) {
    logMessage(
      `[Storage] Error retrieving tool enablement state: ${error instanceof Error ? error.message : String(error)}`,
    );
    return new Set();
  }
};

/**
 * Save tool enablement state to chrome.storage.local
 * Takes a Set of enabled tool names and persists it
 */
export const saveToolEnablementState = async (enabledTools: Set<string>): Promise<void> => {
  try {
    if (!chrome.storage || !chrome.storage.local) {
      logMessage('[Storage] Chrome storage API not available');
      return;
    }

    const enabledToolsArray = Array.from(enabledTools);
    await chrome.storage.local.set({ [TOOL_ENABLEMENT_KEY]: enabledToolsArray });
    logMessage(`[Storage] Saved tool enablement state: ${enabledToolsArray.length} enabled tools`);
  } catch (error) {
    logMessage(
      `[Storage] Error saving tool enablement state: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
};

/**
 * Clear all tool enablement state from storage
 */
export const clearToolEnablementState = async (): Promise<void> => {
  try {
    if (!chrome.storage || !chrome.storage.local) {
      logMessage('[Storage] Chrome storage API not available');
      return;
    }

    await chrome.storage.local.remove(TOOL_ENABLEMENT_KEY);
    logMessage('[Storage] Cleared tool enablement state from storage');
  } catch (error) {
    logMessage(
      `[Storage] Error clearing tool enablement state: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
};
