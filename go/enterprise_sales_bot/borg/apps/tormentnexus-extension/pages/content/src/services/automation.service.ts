import { useUIStore } from '@src/stores';

export class AutomationService {
  private static instance: AutomationService;

  private constructor() {}

  public static getInstance(): AutomationService {
    if (!AutomationService.instance) {
      AutomationService.instance = new AutomationService();
    }
    return AutomationService.instance;
  }

  // Helper to get current prefs from store - use store directly
  private getPreferences() {
    return useUIStore.getState().preferences;
  }

  public shouldAutoExecute(toolName: string): boolean {
    const prefs = this.getPreferences();

    // Check if autoExecute is enabled (by checking if delay is set, though better to have a bool)
    // In Settings.tsx, we see `autoExecuteDelay` being managed.
    // Assuming if delay is valid, it's ON.
    if (prefs.autoExecuteDelay === undefined || prefs.autoExecuteDelay === null) {
        return false;
    }

    const whitelist = prefs.autoExecuteWhitelist || [];

    // Safety Logic:
    // If whitelist has items, we STRICTLY only allow tools in the whitelist.
    if (whitelist.length > 0) {
      return whitelist.includes(toolName);
    }

    // If whitelist is empty, it means "Trust None" or "Trust All"?
    // In a high-security context ("Safe Mode"), empty whitelist should mean NO auto-execution.
    // But for usability, if the user hasn't configured a whitelist but turned on the feature...
    // Let's default to safe: If autoExecute is ON, but no whitelist, we require confirmation?
    // OR we follow the global setting.
    // Given the "Insanely Great" requirement for robustness, let's allow it but log a warning?
    // Actually, `Settings.tsx` shows "Tools listed here will auto-execute".
    // This implies that ONLY tools listed there will auto-execute.
    // So if whitelist is empty, NOTHING auto-executes.
    return false;
  }

  public async updateAutomationStateOnWindow() {
    // This syncs state to the window object for non-React parts of the app if needed
    const prefs = this.getPreferences();
    (window as any).mcpAutomation = {
      autoInsertDelay: prefs.autoInsertDelay,
      autoSubmitDelay: prefs.autoSubmitDelay,
      autoExecuteDelay: prefs.autoExecuteDelay,
      autoExecuteWhitelist: prefs.autoExecuteWhitelist,
    };
  }
}
