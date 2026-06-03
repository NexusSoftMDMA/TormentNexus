import React, { useState } from 'react';
import { useUserPreferences } from '@src/hooks';
import { useProfileStore } from '@src/stores';
import { useActivityStore } from '@src/stores';
import { useToastStore } from '@src/stores';
import { useConfigStore } from '@src/stores';
import { Card, CardContent, CardHeader, CardTitle } from '@src/components/ui/card';
import { Typography, Icon, Button, ToggleWithoutLabel, Toggle } from '../ui';
import { AutomationService } from '@src/services/automation.service';
import { cn } from '@src/lib/utils';
import { createLogger } from '@extension/shared/lib/logger';
import { getExtensionStorageJson, setExtensionStorageJson } from '@src/stores/extension-storage';

const logger = createLogger('Settings');

const DEFAULT_DELAYS = {
  autoInsertDelay: 2,
  autoSubmitDelay: 2,
  autoExecuteDelay: 2,
} as const;

const Settings: React.FC = () => {
  const { preferences, updatePreferences } = useUserPreferences();
  const configStore = useConfigStore();
  const [newToolInput, setNewToolInput] = useState('');
  const [activeTab, setActiveTab] = useState<'automation' | 'advanced'>('automation');

  // Handle delay input changes
  const handleDelayChange = (type: 'autoInsert' | 'autoSubmit' | 'autoExecute', value: string) => {
    const delay = Math.max(0, parseInt(value) || 0); // Ensure non-negative integer
    logger.debug(`${type} delay changed to: ${delay}`);

    // Update user preferences store with the new delay
    updatePreferences({ [`${type}Delay`]: delay });

    void (async () => {
      try {
        const storedDelays = await getExtensionStorageJson<Record<string, number>>('mcpDelaySettings', {});
        await setExtensionStorageJson('mcpDelaySettings', {
          ...storedDelays,
          [`${type}Delay`]: delay,
        });
      } catch (error) {
        logger.error('[Settings] Error storing delay settings:', error);
      }
    })();

    // Update automation state on window
    AutomationService.getInstance().updateAutomationStateOnWindow().catch(console.error);
  };

  // Load stored delays on component mount, set default to 2 seconds if not set
  React.useEffect(() => {
    void (async () => {
      try {
        const storedDelays = await getExtensionStorageJson<Record<string, number>>('mcpDelaySettings', {});
        if (Object.keys(storedDelays).length === 0) {
          updatePreferences(DEFAULT_DELAYS);
          await setExtensionStorageJson('mcpDelaySettings', DEFAULT_DELAYS);
        } else {
          updatePreferences(storedDelays);
        }
      } catch (error) {
        logger.error('[Settings] Error loading stored delay settings:', error);
        updatePreferences(DEFAULT_DELAYS);
        await setExtensionStorageJson('mcpDelaySettings', DEFAULT_DELAYS);
      }
    })();
  }, [updatePreferences]);

  const handleResetDefaults = () => {
    updatePreferences(DEFAULT_DELAYS);
    void setExtensionStorageJson('mcpDelaySettings', DEFAULT_DELAYS);
    logger.debug('[Settings] Reset to defaults');
  };

  const handleExportData = async () => {
    const data = {
      preferences: preferences,
      profiles: useProfileStore.getState().profiles,
      activeProfileId: useProfileStore.getState().activeProfileIds?.[0] || null,
      logs: useActivityStore.getState().logs,
      favorites: await getExtensionStorageJson<string[]>('mcpFavorites', []),
      version: '0.7.0',
      timestamp: new Date().toISOString(),
    };

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `tormentnexus-extension-backup-${new Date().toISOString().slice(0, 10)}.json`;
    a.download = `tormentnexus-extension-backup-${new Date().toISOString().slice(0, 10)}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    useToastStore.getState().addToast({
      title: 'Export Successful',
      message: 'Data exported to JSON file',
      type: 'success',
    });
  };

  const handleImportData = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async e => {
      try {
        const data = JSON.parse(e.target?.result as string);

        if (data.preferences) updatePreferences(data.preferences);
        if (data.profiles) useProfileStore.setState({ profiles: data.profiles, activeProfileIds: data.activeProfileId ? [data.activeProfileId] : [] });
        if (data.logs) useActivityStore.setState({ logs: data.logs });
        if (data.favorites) await setExtensionStorageJson('mcpFavorites', data.favorites);

        useToastStore.getState().addToast({
          title: 'Import Successful',
          message: 'Settings and data restored',
          type: 'success',
        });

        // Refresh page to ensure all stores rehydrate correctly if needed
        setTimeout(() => window.location.reload(), 1500);
      } catch (error) {
        logger.error('Import failed', error);
        useToastStore.getState().addToast({
          title: 'Import Failed',
          message: 'Invalid backup file',
          type: 'error',
        });
      }
    };
    reader.readAsText(file);
    // Reset input
    event.target.value = '';
  };

  return (
    <div className="p-4 space-y-6">
      <div className="flex flex-col space-y-2">
        <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100">
          Settings
        </Typography>
        <Typography variant="caption" className="text-slate-500 dark:text-slate-400">
          Configure automation behaviors and extension preferences.
        </Typography>
      </div>

      <div className="flex border-b border-slate-200 dark:border-slate-800 mb-4">
        <button
          className={cn(
            'flex-1 py-2 text-sm font-medium border-b-2 transition-colors',
            activeTab === 'automation'
              ? 'border-blue-600 text-blue-600 dark:border-blue-500 dark:text-blue-400'
              : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200',
          )}
          onClick={() => setActiveTab('automation')}>
          Automation
        </button>
        <button
          className={cn(
            'flex-1 py-2 text-sm font-medium border-b-2 transition-colors',
            activeTab === 'advanced'
              ? 'border-amber-500 text-amber-600 dark:border-amber-400 dark:text-amber-400'
              : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200',
          )}
          onClick={() => setActiveTab('advanced')}>
          Advanced
        </button>
      </div>

      <div className="space-y-4">
        {/* Automation Tab */}
        {activeTab === 'automation' && (
          <>
            {/* Automation Settings */}
            <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
              <CardHeader className="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 p-4">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-blue-100 dark:bg-blue-900/30 rounded text-blue-600 dark:text-blue-400">
                    <Icon name="lightning" size="sm" />
                  </div>
                  <CardTitle className="text-base font-medium">Automation Controls</CardTitle>
                </div>
              </CardHeader>
              <CardContent className="p-5 space-y-6">
                {/* Auto Insert Delay */}
                <div className="group">
                  <div className="flex items-center justify-between mb-2">
                    <label
                      htmlFor="auto-insert-delay"
                      className="block text-sm font-medium text-slate-700 dark:text-slate-300">
                      Auto Insert Delay
                    </label>
                    <div className="text-xs text-slate-400 dark:text-slate-500 font-mono bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded">
                      {preferences.autoInsertDelay}s
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <input
                      id="auto-insert-delay"
                      type="range"
                      min="0"
                      max="30"
                      value={preferences.autoInsertDelay || 0}
                      onChange={e => handleDelayChange('autoInsert', e.target.value)}
                      className="w-full h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer dark:bg-slate-700 accent-blue-600"
                    />
                    <input
                      type="number"
                      min="0"
                      max="60"
                      value={preferences.autoInsertDelay || 0}
                      onChange={e => handleDelayChange('autoInsert', e.target.value)}
                      className="w-16 p-1 text-sm text-center border rounded-md bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-600 text-slate-900 dark:text-slate-100"
                    />
                  </div>
                  <p className="mt-2 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                    Automatically inserts the tool result into the chat input box after execution.
                    <br />
                    <span className="text-slate-400 dark:text-slate-500 italic">
                      Wait time allows you to review the result before insertion. 0 = Instant.
                    </span>
                  </p>
                </div>

                <div className="border-t border-slate-100 dark:border-slate-800"></div>

                {/* Auto Submit Delay */}
                <div className="group">
                  <div className="flex items-center justify-between mb-2">
                    <label
                      htmlFor="auto-submit-delay"
                      className="block text-sm font-medium text-slate-700 dark:text-slate-300">
                      Auto Submit Delay
                    </label>
                    <div className="text-xs text-slate-400 dark:text-slate-500 font-mono bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded">
                      {preferences.autoSubmitDelay}s
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <input
                      id="auto-submit-delay"
                      type="range"
                      min="0"
                      max="30"
                      value={preferences.autoSubmitDelay || 0}
                      onChange={e => handleDelayChange('autoSubmit', e.target.value)}
                      className="w-full h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer dark:bg-slate-700 accent-green-600"
                    />
                    <input
                      type="number"
                      min="0"
                      max="60"
                      value={preferences.autoSubmitDelay || 0}
                      onChange={e => handleDelayChange('autoSubmit', e.target.value)}
                      className="w-16 p-1 text-sm text-center border rounded-md bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-600 text-slate-900 dark:text-slate-100"
                    />
                  </div>
                  <p className="mt-2 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                    Automatically submits the chat message after result insertion.
                    <br />
                    <span className="text-slate-400 dark:text-slate-500 italic">
                      Requires 'Auto Insert'. Wait time allows you to cancel submission.
                    </span>
                  </p>
                </div>

                <div className="border-t border-slate-100 dark:border-slate-800"></div>

                {/* Auto Execute Delay */}
                <div className="group">
                  <div className="flex items-center justify-between mb-2">
                    <label
                      htmlFor="auto-execute-delay"
                      className="block text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2">
                      Auto Execute Delay
                      <span className="px-1.5 py-0.5 text-[10px] font-bold bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300 rounded uppercase">
                        Advanced
                      </span>
                    </label>
                    <div className="text-xs text-slate-400 dark:text-slate-500 font-mono bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded">
                      {preferences.autoExecuteDelay}s
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <input
                      id="auto-execute-delay"
                      type="range"
                      min="0"
                      max="30"
                      value={preferences.autoExecuteDelay || 0}
                      onChange={e => handleDelayChange('autoExecute', e.target.value)}
                      className="w-full h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer dark:bg-slate-700 accent-amber-500"
                    />
                    <input
                      type="number"
                      min="0"
                      max="60"
                      value={preferences.autoExecuteDelay || 0}
                      onChange={e => handleDelayChange('autoExecute', e.target.value)}
                      className="w-16 p-1 text-sm text-center border rounded-md bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-600 text-slate-900 dark:text-slate-100"
                    />
                  </div>
                  <p className="mt-2 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                    Automatically runs tools when detected, skipping the "Run" click.
                    <br />
                    <span className="text-slate-400 dark:text-slate-500 italic">
                      Use with caution. Wait time allows you to cancel execution.
                    </span>
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Safety Settings (Trusted Tools) */}
            <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
              <CardHeader className="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 p-4">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-green-100 dark:bg-green-900/30 rounded text-green-600 dark:text-green-400">
                    <Icon name="check" size="sm" />
                  </div>
                  <CardTitle className="text-base font-medium">Safety & Whitelist</CardTitle>
                </div>
              </CardHeader>
              <CardContent className="p-5">
                <div className="text-sm text-slate-600 dark:text-slate-400 mb-3">
                  Tools listed here will auto-execute without confirmation (if enabled above).
                </div>
                <div className="flex gap-2 mb-3">
                  <input
                    type="text"
                    placeholder="Enter tool name (e.g., filesystem.read_file)"
                    value={newToolInput}
                    onChange={e => setNewToolInput(e.target.value)}
                    className="flex-1 px-3 py-2 text-sm border rounded-md bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-600 text-slate-900 dark:text-slate-100"
                    onKeyDown={e => {
                      if (e.key === 'Enter') {
                        const val = newToolInput.trim();
                        if (val && !(preferences.trustedTools || []).includes(val)) {
                          updatePreferences({ trustedTools: [...(preferences.trustedTools || []), val] });
                          setNewToolInput('');
                        }
                      }
                    }}
                  />
                  <Button
                    size="sm"
                    variant="outline"
                    className="border-slate-300 dark:border-slate-600"
                    onClick={() => {
                      const val = newToolInput.trim();
                      if (val && !(preferences.trustedTools || []).includes(val)) {
                        updatePreferences({ trustedTools: [...(preferences.trustedTools || []), val] });
                        setNewToolInput('');
                      }
                    }}>
                    Add
                  </Button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {(preferences.trustedTools || []).map(tool => (
                    <div
                      key={tool}
                      className="flex items-center gap-1 px-2 py-1 bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 text-xs rounded border border-green-100 dark:border-green-800">
                      <span>{tool}</span>
                      <button
                        onClick={() => {
                          const newTools = (preferences.trustedTools || []).filter(t => t !== tool);
                          updatePreferences({ trustedTools: newTools });
                        }}
                        className="hover:text-green-900 dark:hover:text-green-100">
                        <Icon name="x" size="xs" />
                      </button>
                    </div>
                  ))}
                  {(!preferences.trustedTools || preferences.trustedTools.length === 0) && (
                    <span className="text-xs text-slate-400 italic">No trusted tools configured.</span>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Data Management */}
            <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
              <CardHeader className="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 p-4">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-purple-100 dark:bg-purple-900/30 rounded text-purple-600 dark:text-purple-400">
                    <Icon name="box" size="sm" />
                  </div>
                  <CardTitle className="text-base font-medium">Data Management</CardTitle>
                </div>
              </CardHeader>
              <CardContent className="p-5">
                <div className="flex gap-3">
                  <Button
                    variant="outline"
                    className="flex-1 border-slate-300 dark:border-slate-600"
                    onClick={handleExportData}>
                    <Icon name="arrow-up-right" size="sm" className="mr-2 rotate-45" />
                    Export Data
                  </Button>
                  <div className="relative flex-1">
                    <input
                      type="file"
                      accept=".json"
                      className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                      onChange={handleImportData}
                    />
                    <Button variant="outline" className="w-full border-slate-300 dark:border-slate-600">
                      <Icon name="arrow-up-right" size="sm" className="mr-2 rotate-[-135deg]" />
                      Import Data
                    </Button>
                  </div>
                </div>
                <Typography variant="caption" className="text-slate-500 dark:text-slate-400 mt-2 block text-center">
                  Backup your settings, profiles, logs, and favorites.
                </Typography>
              </CardContent>
            </Card>
          </>
        )}

        {/* Advanced Tab */}
        {activeTab === 'advanced' && (
          <>
            {/* Feature Flags */}
            <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden border-t-amber-500 dark:border-t-amber-500 border-t-2">
              <CardHeader className="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className="p-1.5 bg-amber-100 dark:bg-amber-900/30 rounded text-amber-600 dark:text-amber-400">
                      <Icon name="tool" size="sm" />
                    </div>
                    <div>
                      <CardTitle className="text-base font-medium">Experimental Features</CardTitle>
                    </div>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="p-5 space-y-4">
                <div className="text-sm text-slate-600 dark:text-slate-400 mb-4">
                  Warning: These features are in development and may be unstable. Toggle flags to override default rollout rules.
                </div>

                {Object.keys(configStore.featureFlags).length === 0 ? (
                  <div className="text-center py-6 text-slate-400 dark:text-slate-500 italic text-sm">
                    No experimental features available.
                  </div>
                ) : (
                  <div className="space-y-4">
                    {Object.entries(configStore.featureFlags).map(([flagKey, flagObj]) => {
                      // We allow manually enabling/disabling flags if needed by mutating the store locally
                      return (
                        <div key={flagKey} className="flex items-start justify-between border border-slate-100 dark:border-slate-800 p-3 rounded-lg bg-slate-50 dark:bg-slate-800/30">
                          <div className="flex flex-col">
                            <Typography variant="body" className="font-semibold text-slate-800 dark:text-slate-200">
                              {flagKey}
                            </Typography>
                            <div className="flex items-center gap-2 mt-1">
                              <span className="text-xs bg-slate-200 dark:bg-slate-700 px-1.5 py-0.5 rounded text-slate-600 dark:text-slate-300 font-mono">
                                {flagObj.rollout}% rollout
                              </span>
                              {flagObj.config && (
                                <span className="text-xs text-slate-400">Configured</span>
                              )}
                            </div>
                          </div>
                          <Toggle
                            label="Enable Feature"
                            checked={flagObj.enabled}
                            onChange={(checked) => {
                              configStore.updateFeatureFlags({
                                [flagKey]: { ...flagObj, enabled: checked }
                              });
                            }}
                          />
                        </div>
                      );
                    })}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Notification Config Details */}
            <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
              <CardHeader className="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 p-4">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-blue-100 dark:bg-blue-900/30 rounded text-blue-600 dark:text-blue-400">
                    <Icon name="info" size="sm" />
                  </div>
                  <CardTitle className="text-base font-medium">Notification Rules</CardTitle>
                </div>
              </CardHeader>
              <CardContent className="p-5">
                <div className="space-y-3">
                  <div className="flex justify-between items-center text-sm border-b border-slate-100 dark:border-slate-800 pb-2">
                    <span className="text-slate-600 dark:text-slate-400">Remote Notifications</span>
                    <ToggleWithoutLabel
                      label="Remote Notifications Enabled"
                      size="sm"
                      checked={configStore.notificationConfig.enabled}
                      onChange={(checked) => {
                        configStore.updateNotificationConfig({ enabled: checked });
                      }}
                    />
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-600 dark:text-slate-400">Max Per Day</span>
                    <span className="font-mono text-slate-800 dark:text-slate-200">{configStore.notificationConfig.maxPerDay}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-600 dark:text-slate-400">Cooldown</span>
                    <span className="font-mono text-slate-800 dark:text-slate-200">{configStore.notificationConfig.cooldownHours}h</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </>
        )}

        {/* Info / Reset combined footer rendered globally */}
        <div className="flex items-center justify-between pt-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleResetDefaults}
            className="text-xs text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200">
            <Icon name="refresh" size="xs" className="mr-1.5" />
            Reset to Defaults
          </Button>

          <div className="text-xs text-slate-400 dark:text-slate-500 italic">Changes are saved automatically</div>
        </div>
      </div>
    </div>
  );
};

export default Settings;
