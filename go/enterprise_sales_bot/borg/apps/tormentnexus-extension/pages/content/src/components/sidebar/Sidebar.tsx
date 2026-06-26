import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { useCurrentAdapter } from '@src/hooks/useAdapter';
import { useTheme, useSidebarState, useUserPreferences, useConnectionStatus } from '@src/hooks';
import { useKeyboardShortcuts } from '@src/hooks/useKeyboardShortcuts';
import { useUIStore } from '@src/stores';
import ServerStatus from './ServerStatus/ServerStatus';
import AvailableTools from './AvailableTools/AvailableTools';
import InstructionManager from './Instructions/InstructionManager';
import InputArea from './InputArea/InputArea';
import Settings from './Settings/Settings';
import Help from './Help/Help';
import ActivityLog from './Activity/ActivityLog';
import Dashboard from './Dashboard/Dashboard';
import MacroList from './Macros/MacroList';
import SystemInfo from './System/SystemInfo';
import CommandPalette from './CommandPalette/CommandPalette';
import Onboarding from './Onboarding/Onboarding';
import PromptTemplates from './PromptTemplates/PromptTemplates';
import { Debugger } from './Debugger';
import { ResourceBrowser } from './ResourceBrowser';
import { useMcpCommunication } from '@src/hooks/useMcpCommunication';
import { logMessage } from '@src/utils/helpers';
import { eventBus } from '@src/events/event-bus';
import { Typography, Toggle, ToggleWithoutLabel, ResizeHandle, Icon, Button } from './ui';
import { NotificationCenter } from './ui/NotificationCenter';
import { ToastContainer } from './ui/Toast';
import { useToastStore } from '@src/stores';
import { useActivityStore } from '@src/stores';
import { cn } from '@src/lib/utils';
import { Card, CardContent } from '@src/components/ui/card';
import type { UserPreferences } from '@src/types/stores';
import { createLogger } from '@extension/shared/lib/logger';
// Debug helper function to check if activeSidebarManager is available

const logger = createLogger('Sidebar');

const checkActiveSidebarManager = (): boolean => {
  const available = !!(window as any).activeSidebarManager;
  logMessage(`[Sidebar] checkActiveSidebarManager: ${available}`);
  return available;
};

// Define Theme type
type Theme = 'light' | 'dark' | 'system';
const THEME_CYCLE: Theme[] = ['light', 'dark', 'system']; // Define the cycle order

// Define a constant for minimized width (should match BaseSidebarManager and CSS logic)
const SIDEBAR_MINIMIZED_WIDTH = 56;
const SIDEBAR_DEFAULT_WIDTH = 320;

interface SidebarProps {
  initialPreferences?: UserPreferences | null;
}

const Sidebar: React.FC<SidebarProps> = ({ initialPreferences }) => {
  // Add unique ID to track component instances
  const componentId = useRef(`sidebar-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`);
  logMessage(
    `[Sidebar] Component initializing with preferences: ${initialPreferences ? 'loaded' : 'null'} (ID: ${componentId.current})`,
  );

  const currentAdapter = useCurrentAdapter();

  // Create a compatibility adapter for legacy components
  const adapter = useMemo(
    () => ({
      // Legacy methods for backward compatibility
      insertTextIntoInput: (text: string) => currentAdapter.insertText(text),
      triggerSubmission: () => currentAdapter.submitForm(),
      supportsFileUpload: () => currentAdapter.hasCapability('file-attachment'),
      attachFile: (file: File) => currentAdapter.attachFile(file),
      // Pass through other properties that might be needed
      name: currentAdapter.activeAdapterName || 'Unknown',
      isReady: currentAdapter.isReady,
      status: currentAdapter.status,
      capabilities: currentAdapter.capabilities,
    }),
    [currentAdapter],
  );

  // Use Zustand hooks for state management
  const { theme, setTheme } = useTheme();
  const {
    isVisible: sidebarVisible,
    isMinimized: storeSidebarMinimized,
    width: storeSidebarWidth,
    toggleSidebar,
    toggleMinimize,
    resizeSidebar,
    setSidebarVisibility,
  } = useSidebarState();
  const { preferences, updatePreferences } = useUserPreferences();
  const { status: connectionStatus } = useConnectionStatus();

  // Apply accent color
  useEffect(() => {
    const accentColor = preferences.accentColor || 'indigo';
    const colorMap: Record<string, Record<number, string>> = {
      indigo: {
        50: '#eef2ff', 100: '#e0e7ff', 200: '#c7d2fe', 300: '#a5b4fc',
        400: '#818cf8', 500: '#6366f1', 600: '#4f46e5', 700: '#4338ca',
        800: '#3730a3', 900: '#312e81'
      },
      blue: {
        50: '#eff6ff', 100: '#dbeafe', 200: '#bfdbfe', 300: '#93c5fd',
        400: '#60a5fa', 500: '#3b82f6', 600: '#2563eb', 700: '#1d4ed8',
        800: '#1e40af', 900: '#1e3a8a'
      },
      green: {
        50: '#f0fdf4', 100: '#dcfce7', 200: '#bbf7d0', 300: '#86efac',
        400: '#4ade80', 500: '#22c55e', 600: '#16a34a', 700: '#15803d',
        800: '#166534', 900: '#14532d'
      },
      purple: {
        50: '#faf5ff', 100: '#f3e8ff', 200: '#e9d5ff', 300: '#d8b4fe',
        400: '#c084fc', 500: '#a855f7', 600: '#9333ea', 700: '#7e22ce',
        800: '#6b21a8', 900: '#581c87'
      },
      red: {
        50: '#fef2f2', 100: '#fee2e2', 200: '#fecaca', 300: '#fca5a5',
        400: '#f87171', 500: '#ef4444', 600: '#dc2626', 700: '#b91c1c',
        800: '#991b1b', 900: '#7f1d1d'
      },
      orange: {
        50: '#fff7ed', 100: '#ffedd5', 200: '#fed7aa', 300: '#fdba74',
        400: '#fb923c', 500: '#f97316', 600: '#ea580c', 700: '#c2410c',
        800: '#9a3412', 900: '#7c2d12'
      },
    };

    const colors = colorMap[accentColor] || colorMap['indigo'];
    const root = sidebarRef.current;

    if (root) {
      Object.entries(colors).forEach(([shade, value]) => {
        root.style.setProperty(`--color-primary-${shade}`, value);
      });
      root.setAttribute('data-accent', accentColor);
    }
  }, [preferences.accentColor]);

  // Error states that could block rendering
  const [initializationError, setInitializationError] = useState<string | null>(null);
  const [extensionContextInvalid, setExtensionContextInvalid] = useState<boolean>(false);
  const [isComponentMounted, setIsComponentMounted] = useState<boolean>(false);
  const [renderKey, setRenderKey] = useState<number>(0); // Force re-render key
  const [isInitializing, setIsInitializing] = useState<boolean>(true); // Track initialization state

  // Get communication methods with guaranteed safe fallbacks and error boundaries
  let communicationMethods;
  try {
    communicationMethods = useMcpCommunication();
  } catch (error) {
    // Handle extension context invalidation gracefully
    if (error instanceof Error && error.message.includes('Extension context invalidated')) {
      logMessage('[Sidebar] Extension context invalidated during hook initialization');
      // Don't set state during render - use useEffect instead
      React.useEffect(() => {
        setExtensionContextInvalid(true);
        setInitializationError('Extension was reloaded. Please refresh the page to restore functionality.');
      }, []);

      // Provide fallback methods
      communicationMethods = {
        availableTools: [],
        sendMessage: async () => 'Extension context invalidated',
        refreshTools: async () => [],
        forceReconnect: async () => false,
        serverStatus: 'disconnected' as const,
        updateServerConfig: async () => false,
        getServerConfig: async () => ({ uri: '' }),
      };
    } else {
      logMessage(
        `[Sidebar] Unexpected error in useMcpCommunication: ${error instanceof Error ? error.message : String(error)}`,
      );
      // Provide safe fallback methods for any other error
      communicationMethods = {
        availableTools: [],
        sendMessage: async () => 'Communication error',
        refreshTools: async () => [],
        forceReconnect: async () => false,
        serverStatus: 'disconnected' as const,
        updateServerConfig: async () => false,
        getServerConfig: async () => ({ uri: '' }),
      };
    }
  }

  // Always render immediately - use safe defaults for all communication methods
  const serverStatus = connectionStatus || communicationMethods?.serverStatus || 'disconnected';
  const availableTools = communicationMethods?.availableTools || [];
  const sendMessage = communicationMethods?.sendMessage || (async () => 'Communication not available');
  const refreshTools = communicationMethods?.refreshTools || (async () => []);
  const forceReconnect = communicationMethods?.forceReconnect || (async () => false);

  // Component mounting and stability tracking
  useEffect(() => {
    setIsComponentMounted(true);
    logMessage(`[Sidebar] Component mounted (ID: ${componentId.current})`);

    // Mark initialization as complete after a brief delay
    const initTimer = setTimeout(() => {
      setIsInitializing(false);
      logMessage(`[Sidebar] Component initialization completed (ID: ${componentId.current})`);
    }, 100);

    return () => {
      clearTimeout(initTimer);
      setIsComponentMounted(false);
      logMessage(`[Sidebar] Component unmounting (ID: ${componentId.current})`);
    };
  }, []);

  // Prevent rendering if component is not properly mounted
  const isStable = isComponentMounted && !isInitializing;

  // Debug logging for serverStatus changes
  useEffect(() => {
    if (isStable) {
      logMessage(`[Sidebar] serverStatus changed to: "${serverStatus}", passing to ServerStatus component`);
    }
  }, [serverStatus, isStable]);

  // Monitor activeSidebarManager availability for debugging
  useEffect(() => {
    // Initial check
    checkActiveSidebarManager();

    // Periodic monitoring to detect if reference gets lost
    const monitorInterval = setInterval(() => {
      const available = checkActiveSidebarManager();
      if (!available) {
        logMessage('[Sidebar] WARNING: activeSidebarManager reference lost - this may cause push mode issues');
      }
    }, 2000); // Check every 2 seconds

    return () => {
      clearInterval(monitorInterval);
    };
  }, []);

  // Enhanced event bus integration for real-time updates
  useEffect(() => {
    const unsubscribeCallbacks: (() => void)[] = [];
    const { addToast } = useToastStore.getState();
    const { addLog } = useActivityStore.getState();

    // Listen for connection status changes
    const unsubscribeConnection = eventBus.on('connection:status-changed', data => {
      logMessage(`[Sidebar] Connection status event received: ${data.status}${data.error ? ` (${data.error})` : ''}`);

      if (data.status === 'connected') {
        addToast({
          title: 'Connected',
          message: 'Successfully connected to MCP server',
          type: 'success',
          duration: 3000,
        });
        addLog({
          type: 'connection',
          title: 'Connected',
          detail: 'Successfully connected to MCP server',
          status: 'success',
        });
        // Automatically refresh tools when connection is established
        logMessage('[Sidebar] Connection established, refreshing tools...');
        refreshTools(true).catch(error => {
          logMessage(`[Sidebar] Failed to refresh tools after connection: ${error}`);
        });
      }
    });
    unsubscribeCallbacks.push(unsubscribeConnection);

    // Listen for tool updates
    const unsubscribeTools = eventBus.on('tool:list-updated', data => {
      logMessage(`[Sidebar] Tool list updated event received: ${data.tools.length} tools`);
      if (data.tools.length > 0) {
        addToast({
          title: 'Tools Updated',
          message: `Loaded ${data.tools.length} available tools`,
          type: 'info',
          duration: 2000,
        });
      }
    });
    unsubscribeCallbacks.push(unsubscribeTools);

    // Listen for tool execution events for better user feedback
    const unsubscribeToolExecution = eventBus.on('tool:execution-completed', data => {
      logMessage(`[Sidebar] Tool execution completed: ${data.execution.toolName} (ID: ${data.execution.id})`);
      addToast({
        title: 'Tool Executed',
        message: `Successfully ran ${data.execution.toolName}`,
        type: 'success',
        duration: 3000,
      });
      addLog({
        type: 'tool_execution',
        title: `Executed: ${data.execution.toolName}`,
        detail: 'Tool execution completed successfully',
        status: 'success',
        metadata: {
          executionId: data.execution.id,
          result: data.execution.result,
        },
      });
    });
    unsubscribeCallbacks.push(unsubscribeToolExecution);

    const unsubscribeToolError = eventBus.on('tool:execution-failed', data => {
      logMessage(`[Sidebar] Tool execution failed: ${data.toolName} - ${data.error}`);
      addToast({
        title: 'Execution Failed',
        message: `${data.toolName}: ${data.error}`,
        type: 'error',
        duration: 5000,
      });
      addLog({
        type: 'error',
        title: `Failed: ${data.toolName}`,
        detail: data.error,
        status: 'error',
      });
    });
    unsubscribeCallbacks.push(unsubscribeToolError);

    // Listen for context bridge events to handle extension lifecycle
    const unsubscribeBridgeInvalidated = eventBus.on('context:bridge-invalidated', data => {
      logMessage(`[Sidebar] Context bridge invalidated: ${data.error}`);
      setExtensionContextInvalid(true);
      setInitializationError('Extension was reloaded. Please refresh the page to restore functionality.');
    });
    unsubscribeCallbacks.push(unsubscribeBridgeInvalidated);

    const unsubscribeBridgeRestored = eventBus.on('context:bridge-restored', () => {
      logMessage('[Sidebar] Context bridge restored');
      setExtensionContextInvalid(false);
      if (initializationError?.includes('Extension was reloaded')) {
        setInitializationError(null);
      }
      // Try to reconnect when context is restored
      forceReconnect().catch(error => {
        logMessage(`[Sidebar] Failed to reconnect after context restoration: ${error}`);
      });
    });
    unsubscribeCallbacks.push(unsubscribeBridgeRestored);

    // Listen for context menu save action
    const unsubscribeContextSave = eventBus.on('context:save', data => {
      logMessage(`[Sidebar] Received context save request: ${data.content.substring(0, 20)}...`);
      addToast({
        title: data.saved ? (data.duplicate ? 'Context Updated' : 'Context Saved') : 'Text Selected',
        message: data.saved
          ? `${data.name || 'Selection'} is now in the Context Library.`
          : 'Text copied from context menu. Open Context Manager to save.',
        type: data.saved ? 'success' : 'info',
        duration: 3000,
      });
    });
    unsubscribeCallbacks.push(unsubscribeContextSave);

    // Cleanup all event listeners
    return () => {
      unsubscribeCallbacks.forEach(unsubscribe => unsubscribe());
    };
  }, [refreshTools, forceReconnect, initializationError]);

  // Initial tool loading when component mounts and connection is available
  useEffect(() => {
    const loadInitialTools = async () => {
      if (serverStatus === 'connected' && availableTools.length === 0) {
        logMessage('[Sidebar] Component mounted with connection, loading initial tools...');
        try {
          await refreshTools(true);
        } catch (error) {
          logMessage(`[Sidebar] Failed to load initial tools: ${error}`);
        }
      }
    };

    // Small delay to ensure everything is initialized
    const timeoutId = setTimeout(loadInitialTools, 1000);
    return () => clearTimeout(timeoutId);
  }, [serverStatus, availableTools.length]);

  // Use store values with fallbacks to initial preferences
  const isMinimized = storeSidebarMinimized ?? initialPreferences?.isMinimized ?? false;
  const sidebarWidth = storeSidebarWidth || initialPreferences?.sidebarWidth || SIDEBAR_DEFAULT_WIDTH;
  const isPushMode = preferences.isPushMode ?? initialPreferences?.isPushMode ?? false;
  const autoSubmit = preferences.autoSubmit ?? initialPreferences?.autoSubmit ?? false;

  // Debug logging for state tracking
  useEffect(() => {
    logMessage(
      `[Sidebar] State update - visible: ${sidebarVisible}, minimized: ${isMinimized}, pushMode: ${isPushMode}, width: ${sidebarWidth}`,
    );
  }, [sidebarVisible, isMinimized, isPushMode, sidebarWidth]);

  // Local UI state that doesn't need to be  // Active Tab State
  const [activeTab, setActiveTab] = useState<
    | 'availableTools'
    | 'instructions'
    | 'activity'
    | 'resources'
    | 'dashboard'
    | 'macros'
    | 'prompts'
    | 'debugger'
    | 'settings'
    | 'system'
    | 'help'
  >('availableTools');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isTransitioning, setIsTransitioning] = useState(false);
  const [isInputMinimized, setIsInputMinimized] = useState(false);

  const sidebarRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null); // Ref for available tools search input
  const isResizingRef = useRef(false);
  const previousWidthRef = useRef(SIDEBAR_DEFAULT_WIDTH);
  const transitionTimerRef = useRef<number | null>(null);

  // Keyboard Shortcuts Integration
  useKeyboardShortcuts({
    toggleSidebar: () => {
      toggleSidebar();
      logMessage('[Sidebar] Toggled via shortcut');
    },
    closeSidebar: () => {
      if (!isMinimized) toggleMinimize('shortcut');
    },
    toggleCommandPalette: () => {
      // Handled globally by Spotlight
    },
    focusSearch: () => {
      // This requires passing a ref down to AvailableTools or managing focus globally
      // For now, let's just switch to the tab
      setActiveTab('availableTools');
      // Ideally we would focus the input inside AvailableTools
    },
    switchTab: direction => {
      const tabs: ('availableTools' | 'instructions' | 'activity' | 'dashboard' | 'macros' | 'prompts' | 'debugger' | 'resources' | 'settings' | 'help' | 'system')[] = [
        'availableTools',
        'instructions',
        'activity',
        'resources',
        'dashboard',
        'macros',
        'prompts',
        'debugger',
        'prompts',
        'resources',
        'settings',
        'help',
        'system',
      ];
      const currentIndex = tabs.indexOf(activeTab);
      let nextIndex = direction === 'next' ? currentIndex + 1 : currentIndex - 1;
      if (nextIndex >= tabs.length) nextIndex = 0;
      if (nextIndex < 0) nextIndex = tabs.length - 1;
      setActiveTab(tabs[nextIndex]);
    },
  });

  // Helper function to wait for SidebarManager to become available with retry mechanism
  const waitForSidebarManager = useCallback(async (maxRetries = 10, baseDelay = 50): Promise<any> => {
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      const sidebarManager = (window as any).activeSidebarManager;
      if (sidebarManager) {
        logMessage(`[Sidebar] activeSidebarManager found after ${attempt} attempts`);
        return sidebarManager;
      }

      // Exponential backoff: 50ms, 100ms, 200ms, 400ms, etc.
      const delay = baseDelay * Math.pow(2, attempt);
      logMessage(
        `[Sidebar] activeSidebarManager not available, retrying in ${delay}ms (attempt ${attempt + 1}/${maxRetries})`,
      );

      await new Promise(resolve => setTimeout(resolve, delay));
    }

    logMessage(`[Sidebar] activeSidebarManager not available after ${maxRetries} attempts`);
    return null;
  }, []);

  // --- Theme Application Logic ---
  const applyTheme = useCallback(
    async (selectedTheme: Theme) => {
      try {
        // Use retry mechanism to wait for SidebarManager
        const sidebarManager = await waitForSidebarManager(5, 50); // Shorter retry for theme application

        if (!sidebarManager) {
          logMessage('[Sidebar] Sidebar manager not available for theme application - will apply when ready.');
          return;
        }

        // OPTIMIZATION: Theme application is now CSS-only and doesn't trigger re-renders
        try {
          const success = sidebarManager.applyThemeClass(selectedTheme);
          if (!success) {
            logMessage('[Sidebar] Theme application failed but continuing...');
          }
        } catch (error) {
          logMessage(`[Sidebar] Theme application error: ${error instanceof Error ? error.message : String(error)}`);
        }
      } catch (error) {
        logMessage(
          `[Sidebar] Error waiting for SidebarManager during theme application: ${error instanceof Error ? error.message : String(error)}`,
        );
      }
    },
    [waitForSidebarManager],
  );

  // Effect to apply theme and listen for system changes
  // OPTIMIZATION: Throttle theme changes to avoid excessive calls
  const lastThemeChangeRef = useRef<number>(0);

  useEffect(() => {
    // Throttle theme applications to once every 100ms
    const now = Date.now();
    if (now - lastThemeChangeRef.current < 100) {
      return;
    }
    lastThemeChangeRef.current = now;

    // Apply theme safely without blocking
    try {
      applyTheme(theme);
    } catch (error) {
      logMessage(
        `[Sidebar] Theme application error during useEffect: ${error instanceof Error ? error.message : String(error)}`,
      );
    }

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      if (theme === 'system') {
        const changeNow = Date.now();
        if (changeNow - lastThemeChangeRef.current < 100) {
          return; // Throttle system theme changes
        }
        lastThemeChangeRef.current = changeNow;

        try {
          applyTheme('system'); // Re-apply system theme on change
        } catch (error) {
          logMessage(`[Sidebar] Theme reapplication error: ${error instanceof Error ? error.message : String(error)}`);
        }
      }
    };

    // Add listener regardless of theme, but only re-apply if theme is 'system'
    mediaQuery.addEventListener('change', handleChange);

    // Cleanup listener
    return () => {
      mediaQuery.removeEventListener('change', handleChange);
    };
  }, [theme, applyTheme]);
  // --- End Theme Application Logic ---

  // useEffect(() => {
  //   // Function to update detected tools
  //   const updateDetectedTools = () => {
  //     try {
  //       const toolDict = getMasterToolDict();
  //       const mcpTools = Object.values(toolDict) as DetectedTool[];

  //       // Update the detected tools state
  //       setDetectedTools(mcpTools);

  //       if (mcpTools.length > 0) {
  //         // logMessage(`[Sidebar] Found ${mcpTools.length} MCP tools`);
  //       }
  //     } catch (error) {
  //       // If getMasterToolDict fails, just log the error
  //       logger.error("Error updating detected tools:", error);
  //     }
  //   };

  //   // Set up interval to check for new tools
  //   const updateInterval = setInterval(updateDetectedTools, 1000);

  //   // Track URL changes to clear detected tools on navigation
  //   let lastUrl = window.location.href;

  // Apply push mode when settings change - with robust retry mechanism
  useEffect(() => {
    logMessage(
      `[Sidebar] Push mode effect triggered - visible: ${sidebarVisible}, pushMode: ${isPushMode}, minimized: ${isMinimized}, width: ${sidebarWidth}`,
    );

    // Use async function to handle the retry mechanism
    const applyPushModeSettings = async () => {
      try {
        // Wait for SidebarManager to become available with retry
        const sidebarManager = await waitForSidebarManager();

        if (sidebarManager) {
          logMessage(`[Sidebar] activeSidebarManager available: true`);

          try {
            // Apply push mode settings when visible
            if (sidebarVisible) {
              logMessage(
                `[Sidebar] Applying push mode (${isPushMode}, minimized: ${isMinimized}) and width (${sidebarWidth})`,
              );
              sidebarManager.setPushContentMode(
                isPushMode,
                isMinimized ? SIDEBAR_MINIMIZED_WIDTH : sidebarWidth,
                isMinimized,
              );
            } else {
              // Ensure push mode is disabled when sidebar is hidden
              logMessage('[Sidebar] Disabling push mode - sidebar not visible');
              sidebarManager.setPushContentMode(false);
            }
          } catch (error) {
            logMessage(
              `[Sidebar] Error applying push mode settings: ${error instanceof Error ? error.message : String(error)}`,
            );
          }
        } else {
          logMessage('[Sidebar] activeSidebarManager not available after retries - cannot apply push mode');
        }
      } catch (error) {
        logMessage(
          `[Sidebar] Error in push mode application process: ${error instanceof Error ? error.message : String(error)}`,
        );
      }
    };

    // Execute the async function
    applyPushModeSettings();
  }, [isPushMode, sidebarWidth, isMinimized, sidebarVisible, waitForSidebarManager]);

  // Cleanup: Ensure push mode is disabled when component unmounts - with retry mechanism
  useEffect(() => {
    return () => {
      // Use async cleanup with retry mechanism
      const cleanupPushMode = async () => {
        try {
          const sidebarManager = await waitForSidebarManager(5, 50); // Shorter retry for cleanup
          if (sidebarManager) {
            logMessage('[Sidebar] Component unmounting - disabling push mode');
            sidebarManager.setPushContentMode(false);
          } else {
            logMessage('[Sidebar] Component unmounting - could not access SidebarManager for cleanup');
          }
        } catch (error) {
          logMessage(
            `[Sidebar] Error during push mode cleanup: ${error instanceof Error ? error.message : String(error)}`,
          );
        }
      };

      cleanupPushMode();
    };
  }, [waitForSidebarManager]);

  // Simple transition management
  const startTransition = () => {
    // Clear any existing timer
    if (transitionTimerRef.current !== null) {
      clearTimeout(transitionTimerRef.current);
    }

    setIsTransitioning(true);

    // Add visual feedback to sidebar during transition
    if (sidebarRef.current) {
      sidebarRef.current.classList.add('sidebar-transitioning');
    }

    // Set timeout to end transition
    transitionTimerRef.current = window.setTimeout(() => {
      setIsTransitioning(false);
      if (sidebarRef.current) {
        sidebarRef.current.classList.remove('sidebar-transitioning');
      }
      transitionTimerRef.current = null;
    }, 500) as unknown as number;
  };

  const handleToggleMinimize = () => {
    startTransition();

    // Add a subtle bounce effect to the toggle
    if (sidebarRef.current) {
      sidebarRef.current.style.transform = 'scale(0.98)';
      setTimeout(() => {
        if (sidebarRef.current) {
          sidebarRef.current.style.transform = '';
        }
      }, 100);
    }

    toggleMinimize('user action');
  };

  const toggleInputMinimize = () => setIsInputMinimized(prev => !prev);

  const handleResize = useCallback(
    (width: number) => {
      // Mark as resizing to prevent unnecessary updates
      if (!isResizingRef.current) {
        isResizingRef.current = true;

        if (sidebarRef.current) {
          sidebarRef.current.classList.add('resizing');
        }
      }

      // Enforce minimum width constraint
      const constrainedWidth = Math.max(SIDEBAR_DEFAULT_WIDTH, width);

      // Update push mode styles if enabled
      if (isPushMode) {
        try {
          const sidebarManager = (window as any).activeSidebarManager;
          if (sidebarManager && typeof sidebarManager.updatePushModeStyles === 'function') {
            sidebarManager.updatePushModeStyles(constrainedWidth);
          }
        } catch (error) {
          logMessage(
            `[Sidebar] Error updating push mode styles: ${error instanceof Error ? error.message : String(error)}`,
          );
        }
      }

      // Debounce the state update for better performance
      if (window.requestAnimationFrame) {
        window.requestAnimationFrame(() => {
          resizeSidebar(constrainedWidth);

          // End resize after a short delay
          if (transitionTimerRef.current !== null) {
            clearTimeout(transitionTimerRef.current);
          }

          transitionTimerRef.current = window.setTimeout(() => {
            if (sidebarRef.current) {
              sidebarRef.current.classList.remove('resizing');
            }

            // Store current width for future reference
            previousWidthRef.current = constrainedWidth;
            isResizingRef.current = false;
            transitionTimerRef.current = null;
          }, 200) as unknown as number;
        });
      } else {
        resizeSidebar(constrainedWidth);
      }
    },
    [isPushMode],
  );

  const handlePushModeToggle = (checked: boolean) => {
    updatePreferences({ isPushMode: checked });
    logMessage(`[Sidebar] Push mode ${checked ? 'enabled' : 'disabled'}`);
  };

  const handleAutoSubmitToggle = (checked: boolean) => {
    updatePreferences({ autoSubmit: checked });
    logMessage(`[Sidebar] Auto submit ${checked ? 'enabled' : 'disabled'}`);
  };

  const handleClearTools = () => {
    logMessage('[Sidebar] Clear tools requested - functionality deprecated');
    // Note: Tool clearing is now handled by the store/MCP client
    // This is kept for UI compatibility but doesn't clear anything
  };

  const handleRefreshTools = async () => {
    logMessage('[Sidebar] Refreshing tools');
    setIsRefreshing(true);
    try {
      await refreshTools(true);
      logMessage('[Sidebar] Tools refreshed successfully');
    } catch (error) {
      logMessage(
        `[Sidebar] Error refreshing tools (non-blocking): ${error instanceof Error ? error.message : String(error)}`,
      );
      // Don't show error to user - this is a background operation
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleThemeToggle = () => {
    const currentIndex = THEME_CYCLE.indexOf(theme);
    const nextIndex = (currentIndex + 1) % THEME_CYCLE.length;
    const nextTheme = THEME_CYCLE[nextIndex];
    setTheme(nextTheme);
    logMessage(`[Sidebar] Theme toggled to: ${nextTheme}`);
  };

  // Transform availableTools to match the expected format for InstructionManager
  const formattedTools = availableTools.map(tool => ({
    name: tool.name,
    schema: tool.schema,
    description: tool.description || '', // Ensure description is always a string
  }));

  // Expose availableTools globally for popover access
  if (typeof window !== 'undefined') {
    (window as any).availableTools = availableTools;
  }

  // Helper to get the current theme icon name
  const getCurrentThemeIcon = (): 'sun' | 'moon' | 'laptop' => {
    switch (theme) {
      case 'light':
        return 'sun';
      case 'dark':
        return 'moon';
      case 'system':
        return 'laptop';
      default:
        return 'laptop'; // Default to system
    }
  };

  return (
    <div
      ref={sidebarRef}
      className={cn(
        'fixed top-0 right-0 h-screen bg-white dark:bg-slate-900 shadow-lg z-50 flex flex-col border-l border-slate-200 dark:border-slate-700 sidebar',
        isPushMode ? 'push-mode' : '',
        isResizingRef.current ? 'resizing' : '',
        isMinimized ? 'collapsed' : '',
        isTransitioning ? 'sidebar-transitioning' : '',
      )}
      style={{ width: isMinimized ? `${SIDEBAR_MINIMIZED_WIDTH}px` : `${sidebarWidth}px` }}>
      {/* Resize Handle - only visible when not minimized */}
      {!isMinimized && (
        <ResizeHandle
          onResize={handleResize}
          minWidth={SIDEBAR_DEFAULT_WIDTH}
          maxWidth={500}
          className="absolute left-0 top-0 bottom-0 w-1 cursor-ew-resize hover:bg-primary-400 dark:hover:bg-primary-600 z-[60] transition-colors duration-300"
        />
      )}

      {/* Header - Adjust content based on isMinimized */}
      <div className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-4 flex items-center justify-between flex-shrink-0 shadow-sm sidebar-header">
        {!isMinimized ? (
          <>
            <div className="flex items-center space-x-2">
              {/* Always show the header content immediately */}
              <a
                href="https://github.com/srbhptl39/TormentNexus-Extension"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="Visit TormentNexus Extension repository"
                className="block">
                {' '}
                {/* Make link block for sizing */}
                <img
                  src={chrome.runtime.getURL('icon-34.png')}
                  alt="MCP Logo"
                  className="w-8 h-8 rounded-md " // Increase size & add rounded corners
                />
              </a>
              <>
                {/* Wrap title in link */}
                <a
                  href="https://github.com/srbhptl39/TormentNexus-Extension"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-slate-800 dark:text-slate-100 hover:text-slate-600 dark:hover:text-slate-300 transition-colors duration-150 no-underline"
                  aria-label="Visit TormentNexus Extension repository">
                  <Typography variant="h4" className="font-semibold">
                    TormentNexus Extension
                  </Typography>
                </a>
                {/* Existing icon link */}
                <a
                  href="https://github.com/srbhptl39/TormentNexus-Extension"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="ml-1 text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 transition-colors duration-150"
                  aria-label="Visit TormentNexus Extension repository">
                  <Icon name="arrow-up-right" size="xs" className="inline-block align-baseline" />
                </a>
              </>
            </div>
            <div className="flex items-center space-x-2 pr-1">
              <NotificationCenter />
              {/* Theme Toggle Button */}
              <Button
                variant="ghost"
                size="icon"
                onClick={handleThemeToggle}
                aria-label={`Toggle theme (current: ${theme})`}
                className="hover:bg-slate-100 dark:hover:bg-slate-700 rounded-full transition-all duration-200 hover:scale-105">
                <Icon
                  name={getCurrentThemeIcon()}
                  size="sm"
                  className="transition-all text-primary-600 dark:text-primary-400"
                />
                <span className="sr-only">Toggle theme</span>
              </Button>
              {/* Minimize Button */}
              <Button
                variant="ghost"
                size="icon"
                onClick={handleToggleMinimize}
                aria-label="Minimize sidebar"
                className="hover:bg-slate-100 dark:hover:bg-slate-700 rounded-full transition-all duration-200 hover:scale-105">
                <Icon name="chevron-right" className="h-4 w-4 text-slate-700 dark:text-slate-300" />
              </Button>
            </div>
          </>
        ) : (
          // Expand Button when minimized
          <Button
            variant="ghost"
            size="icon"
            onClick={handleToggleMinimize}
            aria-label="Expand sidebar"
            className="mx-auto hover:bg-slate-100 dark:hover:bg-slate-700 rounded-full transition-all duration-200 hover:scale-110">
            <Icon name="chevron-left" className="h-4 w-4 text-slate-700 dark:text-slate-300" />
          </Button>
        )}
      </div>

      {/* Main Content Area - Using sliding panel approach */}
      <div className="sidebar-inner-content flex-1 relative overflow-hidden bg-white dark:bg-slate-900">
        <Onboarding />
        <ToastContainer />
        {/* Virtual slide - content always at full width */}
        <div
          ref={contentRef}
          className={cn(
            'absolute top-0 bottom-0 right-0 transition-transform duration-200 ease-in-out',
            isMinimized ? 'translate-x-full' : 'translate-x-0',
            isTransitioning ? 'will-change-transform' : '',
          )}
          style={{ width: `${sidebarWidth}px` }}>
          <div className="flex flex-col h-full">
            {/* Critical Error Display - Only show for severe failures, never block UI */}
            {initializationError && (
              <div className="bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800 p-3 flex-shrink-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-start space-x-2">
                    <Icon name="alert-triangle" size="sm" className="text-red-600 dark:text-red-400 mt-0.5" />
                    <div className="flex-1">
                      <Typography variant="subtitle" className="text-red-800 dark:text-red-200 font-medium">
                        {extensionContextInvalid ? 'Extension Reloaded' : 'Warning'}
                      </Typography>
                      <Typography variant="caption" className="text-red-700 dark:text-red-300">
                        {extensionContextInvalid
                          ? 'The extension was reloaded. Please refresh this page to restore full functionality.'
                          : `Some features may be limited: ${initializationError}`}
                      </Typography>
                      {extensionContextInvalid && (
                        <div className="mt-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => window.location.reload()}
                            className="border-red-300 dark:border-red-600 text-red-700 dark:text-red-300 hover:bg-red-100 dark:hover:bg-red-800 mr-2">
                            Refresh Page
                          </Button>
                        </div>
                      )}
                    </div>
                  </div>
                  {!extensionContextInvalid && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setInitializationError(null)}
                      className="border-red-300 dark:border-red-600 text-red-700 dark:text-red-300 hover:bg-red-100 dark:hover:bg-red-800">
                      Dismiss
                    </Button>
                  )}
                </div>
              </div>
            )}

            {/* Status and Settings section */}
            <div className="py-4 px-4 space-y-4 overflow-y-auto flex-shrink-0">
              <ServerStatus status={serverStatus} />

              {/* Settings */}
              <Card className="sidebar-card border-slate-200 dark:border-slate-700 dark:bg-slate-800 flex-shrink-0 overflow-hidden rounded-lg shadow-sm transition-shadow duration-300">
                <CardContent className="p-3 space-y-3">
                  <div className="flex items-center justify-between">
                    <Typography variant="subtitle" className="text-slate-700 dark:text-slate-300 font-medium">
                      Push Content Mode
                    </Typography>
                    <ToggleWithoutLabel
                      label="Push Content Mode"
                      checked={isPushMode}
                      onChange={handlePushModeToggle}
                    />
                  </div>
                  {/* <div className="flex items-center justify-between">
                    <Typography variant="subtitle" className="text-slate-700 dark:text-slate-300 font-medium">
                      Auto Submit Tool Results
                    </Typography>
                    <ToggleWithoutLabel
                      label="Auto Submit Tool Results"
                      checked={autoSubmit}
                      onChange={handleAutoSubmitToggle}
                    />
                  </div> */}

                  {/* DEBUG BUTTON - ONLY FOR DEVELOPMENT - REMOVE IN PRODUCTION */}
                  {process.env.NODE_ENV === 'development' && (
                    <Button
                      variant="outline"
                      size="sm"
                      className="w-full mt-2 border-slate-200 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-700"
                      onClick={() => {
                        const shadowHost = (window as any).activeSidebarManager?.getShadowHost();
                        if (shadowHost && shadowHost.shadowRoot) {
                          logMessage('Shadow DOM debug requested');
                          // Debug functionality removed - use browser dev tools instead
                        } else {
                          logMessage('Cannot debug: Shadow DOM not found');
                        }
                      }}>
                      Debug Styles
                    </Button>
                  )}
                </CardContent>
              </Card>

              {/* Tabs for Tools/Instructions */}
              <div className="border-b border-slate-200 dark:border-slate-700 mb-2">
                <div className="flex">
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'availableTools'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('availableTools')}>
                    Available Tools
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'instructions'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('instructions')}>
                    Instructions
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'activity'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('activity')}>
                    Activity
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'resources'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('resources')}>
                    Resources
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'dashboard'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('dashboard')}>
                    Dashboard
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'macros'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('macros')}>
                    Macros
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'prompts'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('prompts')}>
                    Prompts
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'debugger'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('debugger')}>
                    Debugger
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'settings'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('settings')}>
                    Settings
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'system'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('system')}>
                    System
                  </button>
                  <button
                    className={cn(
                      'py-2 px-4 font-medium text-sm transition-all duration-200',
                      activeTab === 'help'
                        ? 'border-b-2 border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                        : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-t-lg',
                    )}
                    onClick={() => setActiveTab('help')}>
                    Help
                  </button>
                </div>
              </div>
            </div>

            {/* Tab Content Area - scrollable area with flex-grow to fill available space */}
            <div className="flex-1 min-h-0 px-4 pb-4 overflow-hidden">
              {/* AvailableTools */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'availableTools' },
                )}>
                <Card className="border-slate-200 dark:border-slate-700 dark:bg-slate-800 rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow duration-300">
                  <CardContent className="p-0">
                    <AvailableTools
                      tools={availableTools}
                      onExecute={sendMessage}
                      onRefresh={handleRefreshTools}
                      isRefreshing={isRefreshing}
                    />
                  </CardContent>
                </Card>
              </div>

              {/* Instructions */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'instructions' },
                )}>
                <Card className="border-slate-200 dark:border-slate-700 dark:bg-slate-800 rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow duration-300">
                  <CardContent className="p-0">
                    <InstructionManager adapter={adapter} tools={formattedTools} />
                  </CardContent>
                </Card>
              </div>

              {/* Activity Log */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'activity' },
                )}>
                <ActivityLog />
              </div>

              {/* Resources */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'resources' },
                )}>
                <ResourceBrowser />
              </div>

              {/* Dashboard */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'dashboard' },
                )}>
                <Dashboard />
              </div>

              {/* Macros */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'macros' },
                )}>
                <MacroList onExecute={sendMessage} />
              </div>

              {/* Prompt Templates */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'prompts' },
                )}>
                <PromptTemplates />
              </div>

              {/* Debugger */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'debugger' },
                )}>
                <Debugger />
              </div>

              {/* Settings */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'settings' },
                )}>
                <Settings />
              </div>

              {/* System */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'system' },
                )}>
                <SystemInfo />
              </div>

              {/* Help */}
              <div
                className={cn(
                  'h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600 scrollbar-track-transparent',
                  { hidden: activeTab !== 'help' },
                )}>
                <Help />
              </div>
            </div>

            {/* Input Area (Smart Context) */}
            <div className="border-t border-slate-200 dark:border-slate-700 flex-shrink-0 bg-white dark:bg-slate-800 shadow-inner z-10">
              {!isInputMinimized ? (
                <InputArea
                  onSubmit={async text => {
                    // In "Push Mode" or overlay, we usually want to insert into the AI's chat box
                    // But if we have our own input area, we might want to send directly to the AI via adapter
                    try {
                      await adapter.insertTextIntoInput(text);
                      // Optional: trigger submission if configured
                      if (autoSubmit) {
                        await new Promise(resolve => setTimeout(resolve, 300));
                        await adapter.triggerSubmission();
                      }
                    } catch (e) {
                      logMessage(`[Sidebar] Error inserting text: ${e}`);
                    }
                  }}
                  onToggleMinimize={toggleInputMinimize}
                />
              ) : (
                <div className="p-2">
                  <Button variant="outline" size="sm" onClick={toggleInputMinimize} className="w-full h-8 text-xs">
                    <Icon name="chevron-up" size="xs" className="mr-2" />
                    Show Input & Context
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;
