import { DEFAULT_ACTION_LABELS, SupervisorSettings } from './settings.js';
import { ChatSurfaceInfo, UiInspection } from './ui_automation.js';

export const TERMINAL_TEXT_HINTS = ['@terminal:', 'pwsh', 'powershell', 'terminal', 'shell'];
export const ANTIGRAVITY_LABEL_HINTS = [
    'Run',
    'Expand',
    'Always Allow',
    'Retry',
    'Accept all',
    'Accept',
    'Allow',
    'Approve',
    'Proceed',
    'Keep',
    'Accept all changes',
    'Accept All Changes',
    'Accept All',
    'Approve All',
    'Run command',
    'Allow all'
];

export function normalizeComparableLabel(value: string | null | undefined): string {
    if (!value) {
        return '';
    }

    return value
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, ' ')
        .replace(/\s+/g, ' ')
        .trim();
}

export function collectInspectionHints(inspection: UiInspection): string[] {
    return [
        ...inspection.labels,
        ...inspection.buttons.map((button) => button.name),
        ...inspection.inputs.flatMap((input) => [input.name, input.automationId ?? '', input.className ?? ''])
    ].filter((value): value is string => Boolean(value && value.trim()));
}

export function inspectionLooksLikeAntigravity(inspection: UiInspection): boolean {
    const normalizedHints = new Set(collectInspectionHints(inspection).map((value) => normalizeComparableLabel(value)));

    for (const label of ANTIGRAVITY_LABEL_HINTS) {
        if (normalizedHints.has(normalizeComparableLabel(label))) {
            return true;
        }
    }

    return [...normalizedHints].some((value) =>
        TERMINAL_TEXT_HINTS.some((needle) => value.includes(normalizeComparableLabel(needle)))
    );
}

export function resolveActionLabels(explicitLabels: string[] | undefined, surface: ChatSurfaceInfo, settings: SupervisorSettings): string[] {
    if (explicitLabels && explicitLabels.length > 0) {
        return explicitLabels;
    }

    if (surface.detectedSurface === 'antigravity') {
        return [...DEFAULT_ACTION_LABELS];
    }

    return surface.surfaceProfile.actionLabels ?? settings.actionLabels ?? [...DEFAULT_ACTION_LABELS];
}

export function classifyBrowserFamily(processName: string | null): string | null {
    if (!processName) {
        return null;
    }

    const normalized = processName.toLowerCase();

    if (normalized.includes('firefox')) {
        return 'firefox';
    }

    if (normalized.includes('chrome') || normalized.includes('brave')) {
        return 'chromium';
    }

    if (normalized.includes('edge')) {
        return 'edge';
    }

    return null;
}

export function detectSurfaceName(title: string, processName: string | null): { detectedSurface: string; heuristics: string[] } {
    const normalizedTitle = title.toLowerCase();
    const normalizedProcess = (processName ?? '').toLowerCase();
    const heuristics: string[] = [];

    if (normalizedTitle.includes('antigravity')) {
        heuristics.push('window title contains "antigravity"');
        return { detectedSurface: 'antigravity', heuristics };
    }

    if (normalizedTitle.includes('gemini')) {
        heuristics.push('window title contains "gemini"');
        return { detectedSurface: 'gemini-web', heuristics };
    }

    if (normalizedTitle.includes('claude')) {
        heuristics.push('window title contains "claude"');
        return { detectedSurface: 'claude-web', heuristics };
    }

    if (normalizedTitle.includes('chatgpt')) {
        heuristics.push('window title contains "chatgpt"');
        return { detectedSurface: 'chatgpt-web', heuristics };
    }

    if (normalizedTitle.includes('copilot')) {
        heuristics.push('window title contains "copilot"');
        return { detectedSurface: 'copilot', heuristics };
    }

    if (normalizedTitle.includes('cursor')) {
        heuristics.push('window title contains "cursor"');
        return { detectedSurface: 'cursor', heuristics };
    }

    if (normalizedProcess.includes('firefox') || normalizedProcess.includes('chrome') || normalizedProcess.includes('msedge') || normalizedProcess.includes('brave')) {
        heuristics.push('active process looks like a browser');
        return { detectedSurface: 'browser-chat', heuristics };
    }

    if (normalizedProcess.includes('code')) {
        heuristics.push('active process looks like VS Code');
        return { detectedSurface: 'vscode', heuristics };
    }

    return { detectedSurface: 'unknown', heuristics: ['no known chat-surface heuristic matched'] };
}

export function resolveDetectedSurface(options: {
    title: string;
    processName: string | null;
    windowTargeted: boolean;
    surfaceOverride?: string;
    inspectionSuggestsAntigravity?: boolean;
}): { detectedSurface: string; browserFamily: string | null; heuristics: string[] } {
    const browserFamily = classifyBrowserFamily(options.processName);
    const detection = detectSurfaceName(options.title, options.processName);
    let detectedSurface = options.surfaceOverride ?? detection.detectedSurface;
    let heuristics = options.surfaceOverride
        ? [`surface override applied: ${options.surfaceOverride}`, ...detection.heuristics]
        : [...detection.heuristics];

    if (options.windowTargeted) {
        heuristics.unshift('surface detected from targeted window criteria');
    }

    if (!options.surfaceOverride && options.inspectionSuggestsAntigravity && (browserFamily !== null || detection.detectedSurface === 'browser-chat' || detection.detectedSurface === 'unknown')) {
        detectedSurface = 'antigravity';
        heuristics = ['inspection hints matched Antigravity approval/composer patterns', ...heuristics];
    }

    return {
        detectedSurface,
        browserFamily,
        heuristics
    };
}

export function resolveChatState(options: {
    inspection: UiInspection;
    actionLabels: string[];
    preferredInputControlTypes: string[];
}): {
    state: 'awaiting_action' | 'ready_for_input' | 'unknown';
    pendingActionButtons: string[];
    reasoning: string[];
} {
    const pendingActionButtons = options.inspection.buttons
        .map((button) => button.name)
        .filter((name) => options.actionLabels.some((label) => name?.trim().toLowerCase() === label.toLowerCase()));

    const reasoning: string[] = [];
    let state: 'awaiting_action' | 'ready_for_input' | 'unknown' = 'unknown';

    if (pendingActionButtons.length > 0) {
        state = 'awaiting_action';
        reasoning.push('Found actionable approval/continue buttons in the active window');
    } else if (options.inspection.inputs.some((input) => input.isEnabled && !input.isOffscreen)) {
        state = 'ready_for_input';
        reasoning.push(`Found an enabled visible text input and no pending action buttons; surface profile prefers ${options.preferredInputControlTypes.join(' > ')}`);
    } else {
        reasoning.push('Did not find a pending action button or a usable text input');
    }

    return {
        state,
        pendingActionButtons,
        reasoning
    };
}
