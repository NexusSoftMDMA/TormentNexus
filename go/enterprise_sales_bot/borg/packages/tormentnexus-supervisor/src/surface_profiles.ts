export interface SurfaceProfile {
    id: string;
    displayName: string;
    actionLabels: string[];
    submitKeyChord?: string;
    inputControlTypes: string[];
    notes: string[];
}

export const DEFAULT_SURFACE_PROFILE: SurfaceProfile = {
    id: 'default',
    displayName: 'Default chat surface',
    actionLabels: ['Run', 'Expand', 'Always Allow', 'Retry', 'Accept all', 'Accept', 'Allow', 'Approve', 'Proceed', 'Keep'],
    submitKeyChord: 'alt+enter',
    inputControlTypes: ['Document', 'Edit'],
    notes: [
        'Fallback profile when no fork-specific adapter matches',
        'Prefers browser-like document inputs before edit controls'
    ]
};

const SURFACE_PROFILES: Record<string, SurfaceProfile> = {
    antigravity: {
        id: 'antigravity',
        displayName: 'Antigravity browser chat',
        actionLabels: ['Run', 'Expand', 'Always Allow', 'Retry', 'Accept all', 'Accept', 'Allow', 'Approve', 'Proceed', 'Keep'],
        submitKeyChord: 'alt+enter',
        inputControlTypes: ['Document', 'Edit'],
        notes: [
            'Optimized for browser-hosted coding chats with approval buttons',
            'Keeps Alt+Enter as the default submit chord'
        ]
    },
    'gemini-web': {
        id: 'gemini-web',
        displayName: 'Gemini web chat',
        actionLabels: ['Run', 'Expand', 'Retry', 'Accept', 'Proceed', 'Keep'],
        submitKeyChord: 'alt+enter',
        inputControlTypes: ['Document', 'Edit'],
        notes: [
            'Derived from browser-hosted Gemini/Antigravity-like surfaces'
        ]
    },
    'claude-web': {
        id: 'claude-web',
        displayName: 'Claude web chat',
        actionLabels: ['Retry', 'Accept', 'Allow', 'Proceed', 'Keep'],
        submitKeyChord: 'enter',
        inputControlTypes: ['Document', 'Edit'],
        notes: [
            'Uses Enter as a safer default unless overridden by settings or tool arguments'
        ]
    },
    'chatgpt-web': {
        id: 'chatgpt-web',
        displayName: 'ChatGPT web chat',
        actionLabels: ['Retry', 'Continue', 'Proceed', 'Accept'],
        submitKeyChord: 'enter',
        inputControlTypes: ['Document', 'Edit'],
        notes: [
            'Keeps a lighter action-label set because ChatGPT surfaces usually expose fewer coding-specific approvals'
        ]
    },
    copilot: {
        id: 'copilot',
        displayName: 'Copilot chat surface',
        actionLabels: ['Run', 'Retry', 'Accept', 'Allow', 'Proceed', 'Keep'],
        submitKeyChord: 'ctrl+enter',
        inputControlTypes: ['Edit', 'Document'],
        notes: [
            'Prefers editor-like input controls and Ctrl+Enter style submission'
        ]
    },
    cursor: {
        id: 'cursor',
        displayName: 'Cursor chat surface',
        actionLabels: ['Run', 'Retry', 'Accept', 'Allow', 'Proceed', 'Keep'],
        submitKeyChord: 'ctrl+enter',
        inputControlTypes: ['Edit', 'Document'],
        notes: [
            'Prefers desktop/editor-style input controls'
        ]
    },
    'browser-chat': {
        id: 'browser-chat',
        displayName: 'Generic browser chat',
        actionLabels: ['Run', 'Expand', 'Retry', 'Accept', 'Allow', 'Proceed', 'Keep'],
        submitKeyChord: 'enter',
        inputControlTypes: ['Document', 'Edit'],
        notes: [
            'Used when the process looks like a browser but the title does not match a known fork'
        ]
    },
    vscode: {
        id: 'vscode',
        displayName: 'VS Code or editor chat',
        actionLabels: ['Run', 'Accept', 'Allow', 'Proceed', 'Keep', 'Retry'],
        submitKeyChord: 'ctrl+enter',
        inputControlTypes: ['Edit', 'Document'],
        notes: [
            'Prefers desktop editor controls and Ctrl+Enter submission'
        ]
    }
};

export function resolveSurfaceProfile(surfaceId: string): SurfaceProfile {
    return SURFACE_PROFILES[surfaceId] ?? DEFAULT_SURFACE_PROFILE;
}

export function listSurfaceProfiles(): SurfaceProfile[] {
    return [DEFAULT_SURFACE_PROFILE, ...Object.values(SURFACE_PROFILES)];
}
