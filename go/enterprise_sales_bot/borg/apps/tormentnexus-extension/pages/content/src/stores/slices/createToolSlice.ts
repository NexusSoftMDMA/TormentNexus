import { StateCreator } from 'zustand';
import { eventBus } from '../../events';
import { getToolEnablementState, saveToolEnablementState } from '../../utils/storage';
import type { Tool, DetectedTool, ToolExecution } from '../../types/stores';
import { createLogger } from '@extension/shared/lib/logger';
import type { RootState } from '../root.store';

const logger = createLogger('ToolSlice');

export interface ToolSlice {
  tool: {
    availableTools: Tool[];
    toolsByProfile: Record<string, Tool[]>;
    detectedTools: DetectedTool[];
    toolExecutions: Record<string, ToolExecution>;
    isExecuting: boolean;
    lastExecutionId: string | null;
    enabledTools: Set<string>;
    isLoadingEnablement: boolean;
  };

  // Actions
  setAvailableTools: (profileId: string, tools: Tool[]) => void;
  addDetectedTool: (tool: DetectedTool) => void;
  clearDetectedTools: () => void;
  startToolExecution: (toolName: string, parameters: Record<string, any>) => string;
  updateToolExecution: (execution: Partial<ToolExecution> & { id: string }) => void;
  completeToolExecution: (id: string, result: any, status: 'success' | 'error', error?: string) => void;
  getToolExecution: (id: string) => ToolExecution | undefined;
  
  // Enablement actions
  enableTool: (toolName: string) => void;
  disableTool: (toolName: string) => void;
  enableAllTools: () => void;
  disableAllTools: () => void;
  isToolEnabled: (toolName: string) => boolean;
  loadToolEnablementState: () => Promise<void>;
}

const initialToolState = {
  availableTools: [],
  toolsByProfile: {},
  detectedTools: [],
  toolExecutions: {},
  isExecuting: false,
  lastExecutionId: null,
  enabledTools: new Set<string>(),
  isLoadingEnablement: false,
};

export const createToolSlice: StateCreator<RootState, [], [], ToolSlice> = (set, get) => ({
  tool: initialToolState,

  setAvailableTools: (profileId: string, tools: Tool[]) => {
    set((state: RootState) => {
      const newToolsByProfile = { ...state.tool.toolsByProfile, [profileId]: tools };
      const newAvailableTools = Object.values(newToolsByProfile).flat() as Tool[];
      
      logger.debug(`[ToolSlice] Available tools updated for profile ${profileId}`);
      eventBus.emit('tool:list-updated', { tools: newAvailableTools });
      
      return { tool: { ...state.tool, toolsByProfile: newToolsByProfile, availableTools: newAvailableTools } };
    });

    get().loadToolEnablementState();
  },

  addDetectedTool: (newTool: DetectedTool) => {
    set((state: RootState) => ({ tool: { ...state.tool, detectedTools: [...state.tool.detectedTools, newTool] } }));
    logger.debug('[ToolSlice] Tool detected:', newTool);
    eventBus.emit('tool:detected', { tools: [newTool], source: newTool.source || 'unknown' });
  },

  clearDetectedTools: () => {
    set((state: RootState) => ({ tool: { ...state.tool, detectedTools: [] } }));
    logger.debug('[ToolSlice] Detected tools cleared.');
  },

  startToolExecution: (toolName: string, parameters: Record<string, any>): string => {
    const executionId = `exec_${toolName}_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`;
    const newExecution: ToolExecution = {
      id: executionId,
      toolName,
      parameters,
      status: 'pending',
      timestamp: Date.now(),
      result: null,
    };
    set((state: RootState) => ({
      tool: {
        ...state.tool,
        toolExecutions: { ...state.tool.toolExecutions, [executionId]: newExecution },
        isExecuting: true,
        lastExecutionId: executionId,
      }
    }));
    logger.debug(`Starting execution for ${toolName} (ID: ${executionId})`, parameters);
    eventBus.emit('tool:execution-started', { toolName, callId: executionId });
    return executionId;
  },

  updateToolExecution: (executionUpdate: Partial<ToolExecution> & { id: string }) => {
    const { id, ...updateData } = executionUpdate;
    const existingExecution = get().tool.toolExecutions[id];
    if (existingExecution) {
      const updatedExecution = { ...existingExecution, ...updateData, timestamp: Date.now() };
      set((state: RootState) => ({
        tool: {
          ...state.tool,
          toolExecutions: { ...state.tool.toolExecutions, [id]: updatedExecution },
          isExecuting: updatedExecution.status === 'pending',
        }
      }));
      logger.debug(`Execution updated (ID: ${id}):`, updatedExecution);
      if (updatedExecution.status === 'success' || updatedExecution.status === 'error') {
        eventBus.emit('tool:execution-completed', { execution: updatedExecution });
      }
    } else {
      logger.warn(`Attempted to update non-existent execution (ID: ${id})`);
    }
  },

  completeToolExecution: (id: string, result: any, status: 'success' | 'error', error?: string) => {
    const execution = get().tool.toolExecutions[id];
    if (execution) {
      const completedExecution: ToolExecution = {
        ...execution,
        result,
        status,
        error,
        timestamp: Date.now(),
      };
      set((state: RootState) => {
        const newExecutions = { ...state.tool.toolExecutions, [id]: completedExecution };
        return {
          tool: {
            ...state.tool,
            toolExecutions: newExecutions,
            isExecuting: Object.values(newExecutions).some((ex: any) => ex.id !== id && ex.status === 'pending'),
          }
        };
      });
      logger.debug(`Execution ${status} (ID: ${id}):`, completedExecution);
      eventBus.emit('tool:execution-completed', { execution: completedExecution });
      if (status === 'error') {
        eventBus.emit('tool:execution-failed', {
          toolName: execution.toolName,
          error: error || 'Unknown execution error',
          callId: id,
        });
      }
    } else {
      logger.warn(`Attempted to complete non-existent execution (ID: ${id})`);
    }
  },

  getToolExecution: (id: string): ToolExecution | undefined => {
    return get().tool.toolExecutions[id];
  },

  enableTool: (toolName: string) => {
    set((state: RootState) => {
      const newEnabledTools = new Set<string>([...state.tool.enabledTools, toolName]);
      saveToolEnablementState(newEnabledTools).catch(err =>
        logger.error('[ToolSlice] Failed to save tool enablement state:', err),
      );
      return { tool: { ...state.tool, enabledTools: newEnabledTools } };
    });
    logger.debug(`Tool enabled: ${toolName}`);
  },

  disableTool: (toolName: string) => {
    set((state: RootState) => {
      const newEnabledTools = new Set<string>(state.tool.enabledTools);
      newEnabledTools.delete(toolName);
      saveToolEnablementState(newEnabledTools).catch(err =>
        logger.error('[ToolSlice] Failed to save tool enablement state:', err),
      );
      return { tool: { ...state.tool, enabledTools: newEnabledTools } };
    });
    logger.debug(`Tool disabled: ${toolName}`);
  },

  enableAllTools: () => {
    set((state: RootState) => {
      const newEnabledTools = new Set<string>(state.tool.availableTools.map((t: Tool) => t.name));
      saveToolEnablementState(newEnabledTools).catch(err =>
        logger.error('[ToolSlice] Failed to save tool enablement state:', err),
      );
      return { tool: { ...state.tool, enabledTools: newEnabledTools } };
    });
    logger.debug('[ToolSlice] All tools enabled');
  },

  disableAllTools: () => {
    const newEnabledTools = new Set<string>();
    set((state: RootState) => ({ tool: { ...state.tool, enabledTools: newEnabledTools } }));
    saveToolEnablementState(newEnabledTools).catch(err =>
      logger.error('[ToolSlice] Failed to save tool enablement state:', err),
    );
    logger.debug('[ToolSlice] All tools disabled');
  },

  isToolEnabled: (toolName: string): boolean => {
    return get().tool.enabledTools.has(toolName);
  },

  loadToolEnablementState: async () => {
    set((state: RootState) => ({ tool: { ...state.tool, isLoadingEnablement: true } }));
    try {
      const storedEnabledTools = await getToolEnablementState();
      const state = get();

      if (storedEnabledTools.size === 0 && state.tool.availableTools.length > 0) {
        const allToolsEnabled = new Set<string>(state.tool.availableTools.map((t: Tool) => t.name));
        set((s: RootState) => ({ tool: { ...s.tool, enabledTools: allToolsEnabled, isLoadingEnablement: false } }));
        await saveToolEnablementState(allToolsEnabled);
        logger.debug('[ToolSlice] No stored state found, enabled all tools by default');
      } else {
        set((s: RootState) => ({ tool: { ...s.tool, enabledTools: storedEnabledTools, isLoadingEnablement: false } }));
        logger.debug(`Tool enablement state loaded: ${storedEnabledTools.size} tools enabled`);
      }
    } catch (error) {
      logger.error('[ToolSlice] Failed to load tool enablement state:', error);
      set((state: RootState) => ({ tool: { ...state.tool, isLoadingEnablement: false } }));
    }
  },
});
