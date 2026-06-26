'use client';

import { useState, useEffect } from 'react';
import { trpc } from '@/utils/trpc';

export default function DirectorConfig() {
    const configQuery = trpc.directorConfig.get.useQuery(undefined, {
        refetchInterval: 5000 // Refresh every 5s to see changes
    });

    const updateMutation = trpc.directorConfig.update.useMutation({
        onSuccess: () => configQuery.refetch()
    });

    const [formState, setFormState] = useState<any>({});
    const [isEditing, setIsEditing] = useState(false);
    const [diagnosticStatus, setDiagnosticStatus] = useState<string | null>(null);

    const testQuery = trpc.directorConfig.test.useQuery(undefined, {
        enabled: false,
    });

    const handleTestEndpoint = async () => {
        setDiagnosticStatus(null);
        try {
            const result = await testQuery.refetch();
            if (!result.data) {
                setDiagnosticStatus('Failed • No response payload');
                return;
            }

            const director = result.data.directorReady ? 'ready' : 'offline';
            const llm = result.data.llmServiceReady ? 'ready' : 'offline';
            setDiagnosticStatus(`OK • Director ${director} • LLM ${llm}`);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            setDiagnosticStatus(`Failed • ${message}`);
        }
    };

    // Sync form with data when loaded (only if not editing)
    useEffect(() => {
        if (configQuery.data && !isEditing) {
            setFormState(configQuery.data);
        }
    }, [configQuery.data, isEditing]);

    const handleChange = (field: string, value: any) => {
        setIsEditing(true);
        setFormState((prev: any) => ({ ...prev, [field]: value }));
    };

    const handleSave = () => {
        updateMutation.mutate(formState);
        setIsEditing(false);
    };

    if (configQuery.isLoading) return <div className="p-4 bg-gray-900/50 rounded animate-pulse">Loading config...</div>;

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 space-y-4">
            <h2 className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-400">
                Director Configuration
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Loop Timing */}
                <div className="space-y-4">
                    <h3 className="text-gray-400 font-medium border-b border-gray-800 pb-2 flex items-center gap-2">
                        Loop Timing
                        <InfoIcon tooltip="Controls how fast the Director operates loop-by-loop." />
                    </h3>

                    <ConfigSlider
                        label="Task Cooldown"
                        value={formState.taskCooldownMs}
                        onChange={(v) => handleChange('taskCooldownMs', v)}
                        min={1000} max={60000} step={1000}
                        unit="ms"
                        tooltip="Minimum wait time between finishing one task and starting the next. Prevents rapid-fire looping."
                    />

                    <ConfigSlider
                        label="Heartbeat Interval"
                        value={formState.heartbeatIntervalMs}
                        onChange={(v) => handleChange('heartbeatIntervalMs', v)}
                        min={1000} max={60000} step={1000}
                        unit="ms"
                        tooltip="How often the Director checks the system state (e.g. reading terminals, checking for user activity)."
                    />
                </div>

                {/* Features */}
                <div className="space-y-4">
                    <h3 className="text-gray-400 font-medium border-b border-gray-800 pb-2 flex items-center gap-2">
                        Behavior
                        <InfoIcon tooltip="Settings that affect how the Director interacts with the world." />
                    </h3>

                    <div className="space-y-1">
                        <div className="flex items-center gap-2">
                            <label className="text-sm text-gray-300">Default Focus Topic</label>
                            <InfoIcon tooltip="The fallback goal the Director pursues when no specific directive is active from the Council." />
                        </div>
                        <input
                            type="text"
                            value={formState.defaultTopic || ''}
                            onChange={(e) => handleChange('defaultTopic', e.target.value)}
                            onFocus={() => setIsEditing(true)}
                            placeholder="e.g. Implement Roadmap Features"
                            className="w-full bg-gray-800 rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-blue-500 border border-gray-700 placeholder-gray-600"
                        />
                        <p className="text-xs text-gray-500">Supervisor fallback goal.</p>
                    </div>

                    <ConfigSlider
                        label="Summary Interval"
                        value={formState.periodicSummaryMs}
                        onChange={(v) => handleChange('periodicSummaryMs', v)}
                        min={60000} max={600000} step={30000}
                        unit="ms"
                        tooltip="How often the Director posts a status summary ('Director Status') to the chat window."
                    />

                    <ConfigSlider
                        label="Paste Delay"
                        value={formState.pasteToSubmitDelayMs}
                        onChange={(v) => handleChange('pasteToSubmitDelayMs', v)}
                        min={0} max={5000} step={100}
                        unit="ms"
                        tooltip="Time to wait AFTER pasting text into the chat box before pressing Enter. Higher values prevent clipping/spam."
                    />

                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <span className="text-sm text-gray-300">Detection Mode</span>
                            <InfoIcon tooltip="How the Director detects if the user has accepted a suggestion. 'Polling' checks logs; 'State' listens for events." />
                        </div>
                        <select
                            value={formState.acceptDetectionMode || 'polling'}
                            onChange={(e) => handleChange('acceptDetectionMode', e.target.value)}
                            onFocus={() => setIsEditing(true)}
                            className="bg-gray-800 border border-gray-700 text-white text-sm rounded px-2 py-1"
                        >
                            <option value="polling">Polling (Logs)</option>
                            <option value="state">State (Exp)</option>
                        </select>
                    </div>
                </div>

                {/* Personality */}
                <div className="space-y-4">
                    <h3 className="text-gray-400 font-medium border-b border-gray-800 pb-2 flex items-center gap-2">
                        Personality
                        <InfoIcon tooltip="Influences the tone and style of the Director's output." />
                    </h3>
                    <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-300">Persona</span>
                        <select
                            value={formState.persona || 'default'}
                            onChange={(e) => handleChange('persona', e.target.value)}
                            onFocus={() => setIsEditing(true)}
                            className="bg-gray-800 border border-gray-700 text-white text-sm rounded px-2 py-1"
                        >
                            <option value="default">Default</option>
                            <option value="homie">Homie (Casual)</option>
                            <option value="professional">Professional</option>
                            <option value="chaos">Chaos (Creative)</option>
                        </select>
                    </div>

                    <div>
                        <div className="flex items-center gap-2 mb-1">
                            <label className="block text-sm text-gray-300">Chat Prefix</label>
                            <InfoIcon tooltip="The prefix added to every message the Director pastes into the chat." />
                        </div>
                        <input
                            type="text"
                            value={formState.chatPrefix !== undefined ? formState.chatPrefix : '[Director]:'}
                            onChange={(e) => handleChange('chatPrefix', e.target.value)}
                            onFocus={() => setIsEditing(true)}
                            placeholder="[Director]:"
                            className="w-full bg-gray-800 rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-blue-500 border border-gray-700 placeholder-gray-600 mb-4"
                        />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                        <div>
                            <div className="flex items-center gap-2 mb-1">
                                <label className="block text-sm text-gray-300">Action Prefix</label>
                                <InfoIcon tooltip="Prefix for tasks started by the Director (User or Auto)." />
                            </div>
                            <input
                                type="text"
                                value={formState.directorActionPrefix !== undefined ? formState.directorActionPrefix : '🎬 **Director Action**:'}
                                onChange={(e) => handleChange('directorActionPrefix', e.target.value)}
                                onFocus={() => setIsEditing(true)}
                                className="w-full bg-gray-800 rounded px-3 py-2 text-xs text-white focus:ring-1 focus:ring-blue-500 border border-gray-700"
                            />
                        </div>
                        <div>
                            <div className="flex items-center gap-2 mb-1">
                                <label className="block text-sm text-gray-300">Council Prefix</label>
                                <InfoIcon tooltip="Prefix for tasks initiated by the Council." />
                            </div>
                            <input
                                type="text"
                                value={formState.councilPrefix !== undefined ? formState.councilPrefix : '🏛️ [Council]:'}
                                onChange={(e) => handleChange('councilPrefix', e.target.value)}
                                onFocus={() => setIsEditing(true)}
                                className="w-full bg-gray-800 rounded px-3 py-2 text-xs text-white focus:ring-1 focus:ring-blue-500 border border-gray-700"
                            />
                        </div>
                        <div>
                            <div className="flex items-center gap-2 mb-1">
                                <label className="block text-sm text-gray-300">Status Prefix</label>
                                <InfoIcon tooltip="Prefix for periodic status updates." />
                            </div>
                            <input
                                type="text"
                                value={formState.statusPrefix !== undefined ? formState.statusPrefix : '📊 [Director Status]:'}
                                onChange={(e) => handleChange('statusPrefix', e.target.value)}
                                onFocus={() => setIsEditing(true)}
                                className="w-full bg-gray-800 rounded px-3 py-2 text-xs text-white focus:ring-1 focus:ring-blue-500 border border-gray-700"
                            />
                        </div>
                    </div>

                    <div>
                        <div className="flex items-center gap-2 mb-1">
                            <label className="block text-sm text-gray-300">Custom Instructions</label>
                            <InfoIcon tooltip="Specific instructions that override the default persona prompts." />
                        </div>
                        <textarea
                            value={formState.customInstructions || ''}
                            onChange={(e) => handleChange('customInstructions', e.target.value)}
                            onFocus={() => setIsEditing(true)}
                            placeholder="e.g. Always use TypeScript. Prefer functional programming."
                            className="w-full bg-gray-800 rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-blue-500 border border-gray-700 placeholder-gray-600 h-24 resize-none"
                        />
                    </div>
                </div>

                {/* Advanced Controls */}
                <div className="space-y-4">
                    <h3 className="text-gray-400 font-medium border-b border-gray-800 pb-2">Advanced Controls</h3>

                    <div className="flex items-center justify-between bg-gray-800/50 p-3 rounded border border-gray-700/50">
                        <div>
                            <span className="text-sm text-gray-300 block flex items-center gap-2">
                                Paste to Chat
                                <InfoIcon tooltip="If enabled, the Director will output its thoughts and commands to the Chat window." />
                            </span>
                            <span className="text-xs text-gray-500">Output to Chat Window</span>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                className="sr-only peer"
                                checked={formState.enableChatPaste !== false}
                                onChange={(e) => handleChange('enableChatPaste', e.target.checked)}
                                onFocus={() => setIsEditing(true)}
                            />
                            <div className="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-600"></div>
                        </label>
                    </div>

                    <div className="flex items-center justify-between bg-gray-800/50 p-3 rounded border border-gray-700/50">
                        <div>
                            <span className="text-sm text-gray-300 block flex items-center gap-2">
                                Autonomous Thinking
                                <InfoIcon tooltip="If enabled, the Director uses the Council to make autonomous decisions. Disable to pause thinking." />
                            </span>
                            <span className="text-xs text-gray-500">Enable Council Loops</span>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                className="sr-only peer"
                                checked={formState.enableCouncil !== false}
                                onChange={(e) => handleChange('enableCouncil', e.target.checked)}
                                onFocus={() => setIsEditing(true)}
                            />
                            <div className="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-purple-600"></div>
                        </label>
                    </div>

                    <div className="flex items-center justify-between bg-gray-800/50 p-3 rounded border border-gray-700/50">
                        <div>
                            <span className="text-sm text-gray-300 block flex items-center gap-2">
                                Auto-Submit Chat
                                <InfoIcon tooltip="If enabled, the Director will automatically press ENTER after pasting a message." />
                            </span>
                            <span className="text-xs text-gray-500">Press Enter Automatically</span>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                className="sr-only peer"
                                checked={formState.autoSubmitChat || false}
                                onChange={(e) => handleChange('autoSubmitChat', e.target.checked)}
                                onFocus={() => setIsEditing(true)}
                            />
                            <div className="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                        </label>
                    </div>

                    <ConfigSlider
                        label="LLM Timeout"
                        value={formState.lmStudioTimeoutMs || 30000}
                        onChange={(v) => handleChange('lmStudioTimeoutMs', v)}
                        min={5000} max={120000} step={5000}
                        unit="ms"
                        tooltip="Max/Timeout to wait for a response from the LLM (LMStudio) before aborting."
                    />

                    <div className="flex items-center justify-between bg-gray-800/50 p-3 rounded border border-gray-700/50">
                        <div>
                            <span className="text-sm text-gray-300 block flex items-center gap-2">
                                Emergency Stop
                                <InfoIcon tooltip="Global Kill Switch. Stops all Director activity immediately." />
                            </span>
                            <span className="text-xs text-gray-500">Halt All Director Loops</span>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                className="sr-only peer"
                                checked={formState.stopDirector || false}
                                onChange={(e) => handleChange('stopDirector', e.target.checked)}
                                onFocus={() => setIsEditing(true)}
                            />
                            <div className="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-red-600"></div>
                        </label>
                    </div>
                </div>

                {/* Diagnostics */}
                <div className="space-y-4">
                    <h3 className="text-gray-400 font-medium border-b border-gray-800 pb-2">Diagnostics</h3>
                    <div className="flex gap-4">
                        <button
                            onClick={() => window.open('http://localhost:1234', '_blank')}
                            className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 text-xs text-white rounded border border-zinc-700"
                        >
                            Open LMStudio Web UI
                        </button>
                        <button
                            onClick={handleTestEndpoint}
                            disabled={testQuery.isFetching}
                            className="px-3 py-1 bg-blue-900/60 hover:bg-blue-800 text-xs text-white rounded border border-blue-700 disabled:opacity-50"
                        >
                            {testQuery.isFetching ? 'Testing…' : 'Test Director Endpoint'}
                        </button>
                    </div>
                    {diagnosticStatus ? (
                        <div className="text-xs text-gray-300 bg-gray-800/50 border border-gray-700 rounded px-3 py-2">
                            {diagnosticStatus}
                        </div>
                    ) : null}
                </div>

            </div>

            <div className="flex justify-end pt-4 border-t border-gray-800">
                <button
                    onClick={handleSave}
                    disabled={!isEditing || updateMutation.isPending}
                    className={`px-4 py-2 rounded font-medium transition-colors ${isEditing
                        ? 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-900/20'
                        : 'bg-gray-800 text-gray-500 cursor-not-allowed'
                        }`}
                >
                    {updateMutation.isPending ? 'Saving...' : 'Apply Changes'}
                </button>
            </div>
        </div >
    );
}

function InfoIcon({ tooltip }: { tooltip: string }) {
    return (
        <div className="group relative inline-flex items-center justify-center cursor-help">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-500 hover:text-blue-400 transition-colors">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="16" x2="12" y2="12"></line>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
            <div className="absolute bottom-full mb-2 hidden group-hover:block w-48 p-2 bg-black text-xs text-gray-200 rounded border border-gray-700 shadow-xl z-50">
                {tooltip}
                <div className="absolute top-full left-1/2 -ml-1 border-4 border-transparent border-t-black"></div>
            </div>
        </div>
    );
}

function ConfigSlider({ label, value, onChange, min, max, step, unit, tooltip }: {
    label: string,
    value: number,
    onChange: (v: number) => void,
    min: number,
    max: number,
    step: number,
    unit: string,
    tooltip?: string
}) {
    return (
        <div className="space-y-1">
            <div className="flex justify-between text-sm">
                <span className="text-gray-300 flex items-center gap-2">
                    {label}
                    {tooltip && <InfoIcon tooltip={tooltip} />}
                </span>
                <span className="text-blue-400 font-mono">{value} {unit}</span>
            </div>
            <input
                type="range"
                min={min} max={max} step={step}
                value={value || 0}
                onChange={(e) => onChange(Number(e.target.value))}
                className="w-full h-2 bg-gray-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
            />
        </div>
    );
}

function SectionHeader({ title }: { title: string }) {
    return <h3 className="text-gray-400 font-medium border-b border-gray-800 pb-2 pt-2">{title}</h3>;
}
