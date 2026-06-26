import test from 'node:test';
import assert from 'node:assert/strict';
import { DEFAULT_SETTINGS } from './settings.js';
import { DEFAULT_SURFACE_PROFILE, resolveSurfaceProfile } from './surface_profiles.js';
import {
    classifyBrowserFamily,
    collectInspectionHints,
    detectSurfaceName,
    inspectionLooksLikeAntigravity,
    normalizeComparableLabel,
    resolveActionLabels,
    resolveChatState,
    resolveDetectedSurface
} from './decision_logic.js';
import type { ChatSurfaceInfo, UiInspection } from './ui_automation.js';

function makeInspection(partial?: Partial<UiInspection>): UiInspection {
    return {
        window: {
            title: 'Antigravity',
            processName: 'firefox',
            processId: 123
        },
        buttons: [],
        inputs: [],
        labels: [],
        ...partial
    };
}

function makeSurface(partial?: Partial<ChatSurfaceInfo>): ChatSurfaceInfo {
    return {
        title: 'Antigravity',
        processName: 'firefox',
        processPath: null,
        processId: 123,
        bounds: null,
        browserFamily: 'firefox',
        detectedSurface: 'browser-chat',
        surfaceProfile: resolveSurfaceProfile('browser-chat'),
        heuristics: [],
        ...partial
    };
}

test('normalizeComparableLabel collapses punctuation and spacing', () => {
    assert.equal(normalizeComparableLabel(' Accept   all! '), 'accept all');
});

test('collectInspectionHints includes labels, button names, and input metadata', () => {
    const inspection = makeInspection({
        labels: ['Run'],
        buttons: [{ name: 'Accept all', isEnabled: true, isOffscreen: false, hasKeyboardFocus: false }],
        inputs: [{ name: '', automationId: 'composer', className: 'chat-editor', isEnabled: true, isOffscreen: false, hasKeyboardFocus: true }]
    });

    assert.deepEqual(collectInspectionHints(inspection), ['Run', 'Accept all', 'composer', 'chat-editor']);
});

test('inspectionLooksLikeAntigravity detects antigravity approval labels', () => {
    const inspection = makeInspection({
        labels: ['Run', 'Approve']
    });

    assert.equal(inspectionLooksLikeAntigravity(inspection), true);
});

test('inspectionLooksLikeAntigravity detects terminal hint surfaces', () => {
    const inspection = makeInspection({
        inputs: [{ name: '@terminal:pwsh', automationId: null, className: 'terminal', isEnabled: true, isOffscreen: false, hasKeyboardFocus: true }]
    });

    assert.equal(inspectionLooksLikeAntigravity(inspection), true);
});

test('inspectionLooksLikeAntigravity stays false for ordinary browser noise', () => {
    const inspection = makeInspection({
        labels: ['Search', 'Menu'],
        buttons: [{ name: 'Cancel', isEnabled: true, isOffscreen: false, hasKeyboardFocus: false }]
    });

    assert.equal(inspectionLooksLikeAntigravity(inspection), false);
});

test('resolveActionLabels prefers explicit labels first', () => {
    const labels = resolveActionLabels(['Proceed'], makeSurface(), DEFAULT_SETTINGS);
    assert.deepEqual(labels, ['Proceed']);
});

test('resolveActionLabels forces default action set for antigravity', () => {
    const labels = resolveActionLabels(undefined, makeSurface({
        detectedSurface: 'antigravity',
        surfaceProfile: resolveSurfaceProfile('antigravity')
    }), DEFAULT_SETTINGS);

    assert.deepEqual(labels, [...DEFAULT_SETTINGS.actionLabels]);
});

test('resolveActionLabels falls back to surface profile outside antigravity', () => {
    const labels = resolveActionLabels(undefined, makeSurface({
        detectedSurface: 'claude-web',
        surfaceProfile: resolveSurfaceProfile('claude-web')
    }), DEFAULT_SETTINGS);

    assert.deepEqual(labels, resolveSurfaceProfile('claude-web').actionLabels);
});

test('classifyBrowserFamily detects firefox and chromium browsers', () => {
    assert.equal(classifyBrowserFamily('firefox'), 'firefox');
    assert.equal(classifyBrowserFamily('chrome'), 'chromium');
    assert.equal(classifyBrowserFamily('brave'), 'chromium');
    assert.equal(classifyBrowserFamily('msedge'), 'edge');
    assert.equal(classifyBrowserFamily('cursor'), null);
});

test('detectSurfaceName prefers explicit title matches', () => {
    assert.deepEqual(detectSurfaceName('Antigravity - Task', 'firefox'), {
        detectedSurface: 'antigravity',
        heuristics: ['window title contains "antigravity"']
    });
    assert.deepEqual(detectSurfaceName('Claude', 'firefox'), {
        detectedSurface: 'claude-web',
        heuristics: ['window title contains "claude"']
    });
});

test('detectSurfaceName falls back to process-based browser/editor detection', () => {
    assert.deepEqual(detectSurfaceName('Untitled', 'firefox'), {
        detectedSurface: 'browser-chat',
        heuristics: ['active process looks like a browser']
    });
    assert.deepEqual(detectSurfaceName('Workspace', 'Code'), {
        detectedSurface: 'vscode',
        heuristics: ['active process looks like VS Code']
    });
    assert.deepEqual(detectSurfaceName('Untitled', 'unknown'), {
        detectedSurface: 'unknown',
        heuristics: ['no known chat-surface heuristic matched']
    });
});

test('resolveDetectedSurface honors explicit surface override', () => {
    assert.deepEqual(resolveDetectedSurface({
        title: 'Untitled',
        processName: 'firefox',
        windowTargeted: false,
        surfaceOverride: 'claude-web',
        inspectionSuggestsAntigravity: true
    }), {
        detectedSurface: 'claude-web',
        browserFamily: 'firefox',
        heuristics: ['surface override applied: claude-web', 'active process looks like a browser']
    });
});

test('resolveDetectedSurface adds targeted-window heuristic note', () => {
    assert.deepEqual(resolveDetectedSurface({
        title: 'Untitled',
        processName: 'firefox',
        windowTargeted: true
    }), {
        detectedSurface: 'browser-chat',
        browserFamily: 'firefox',
        heuristics: ['surface detected from targeted window criteria', 'active process looks like a browser']
    });
});

test('resolveDetectedSurface promotes browser-like surfaces to antigravity when inspection hints match', () => {
    assert.deepEqual(resolveDetectedSurface({
        title: 'Untitled',
        processName: 'firefox',
        windowTargeted: false,
        inspectionSuggestsAntigravity: true
    }), {
        detectedSurface: 'antigravity',
        browserFamily: 'firefox',
        heuristics: ['inspection hints matched Antigravity approval/composer patterns', 'active process looks like a browser']
    });
});

test('resolveDetectedSurface does not promote non-browser unknown surfaces without browser fallback', () => {
    assert.deepEqual(resolveDetectedSurface({
        title: 'Local App',
        processName: 'customtool',
        windowTargeted: false,
        inspectionSuggestsAntigravity: true
    }), {
        detectedSurface: 'antigravity',
        browserFamily: null,
        heuristics: ['inspection hints matched Antigravity approval/composer patterns', 'no known chat-surface heuristic matched']
    });
});

test('resolveChatState returns awaiting_action when an approval button is present', () => {
    const result = resolveChatState({
        inspection: makeInspection({
            buttons: [{ name: 'Run', isEnabled: true, isOffscreen: false, hasKeyboardFocus: false }]
        }),
        actionLabels: ['Run', 'Accept all'],
        preferredInputControlTypes: ['Document', 'Edit']
    });

    assert.deepEqual(result, {
        state: 'awaiting_action',
        pendingActionButtons: ['Run'],
        reasoning: ['Found actionable approval/continue buttons in the active window']
    });
});

test('resolveChatState returns ready_for_input when no approval button exists but a usable input does', () => {
    const result = resolveChatState({
        inspection: makeInspection({
            inputs: [{ name: 'composer', automationId: 'chat-input', className: 'editor', isEnabled: true, isOffscreen: false, hasKeyboardFocus: true }]
        }),
        actionLabels: ['Run', 'Accept all'],
        preferredInputControlTypes: ['Document', 'Edit']
    });

    assert.deepEqual(result, {
        state: 'ready_for_input',
        pendingActionButtons: [],
        reasoning: ['Found an enabled visible text input and no pending action buttons; surface profile prefers Document > Edit']
    });
});

test('resolveChatState returns unknown when neither approval buttons nor usable input exist', () => {
    const result = resolveChatState({
        inspection: makeInspection(),
        actionLabels: ['Run', 'Accept all'],
        preferredInputControlTypes: ['Document', 'Edit']
    });

    assert.deepEqual(result, {
        state: 'unknown',
        pendingActionButtons: [],
        reasoning: ['Did not find a pending action button or a usable text input']
    });
});
