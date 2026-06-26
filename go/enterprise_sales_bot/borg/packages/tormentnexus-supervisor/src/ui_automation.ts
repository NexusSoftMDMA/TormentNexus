import activeWin from 'active-win';
import { runPowerShell, runPowerShellJson, toPowerShellString } from './powershell.js';
import { DEFAULT_ACTION_LABELS, SupervisorSettings, SupervisorSettingsManager } from './settings.js';
import { DEFAULT_SURFACE_PROFILE, listSurfaceProfiles, resolveSurfaceProfile, SurfaceProfile } from './surface_profiles.js';
import { inspectionLooksLikeAntigravity, resolveActionLabels, resolveChatState, resolveDetectedSurface } from './decision_logic.js';

export interface WindowBounds {
    left: number;
    top: number;
    width: number;
    height: number;
}

export interface WindowInfo {
    title: string;
    processName?: string | null;
    processId?: number | null;
    bounds?: WindowBounds | null;
}

export interface UiElementInfo {
    name: string;
    automationId?: string | null;
    className?: string | null;
    controlType?: string | null;
    isEnabled: boolean;
    isOffscreen: boolean;
    hasKeyboardFocus: boolean;
    bounds?: WindowBounds | null;
}

export interface UiInspection {
    window: WindowInfo;
    buttons: UiElementInfo[];
    inputs: UiElementInfo[];
    labels: string[];
}

export interface ChatSurfaceInfo {
    title: string;
    processName: string | null;
    processPath: string | null;
    processId: number | null;
    bounds: WindowBounds | null;
    browserFamily: string | null;
    detectedSurface: string;
    surfaceProfile: SurfaceProfile;
    heuristics: string[];
}

export interface ChatStateInfo {
    surface: ChatSurfaceInfo;
    inspection: UiInspection;
    state: 'awaiting_action' | 'ready_for_input' | 'unknown';
    pendingActionButtons: string[];
    reasoning: string[];
}

export interface ClickActionResult {
    window: WindowInfo;
    clicked: UiElementInfo[];
    missing: string[];
}

export interface SetInputResult {
    window: WindowInfo;
    target: UiElementInfo | null;
    method: 'value-pattern' | 'focus-sendkeys';
    textLength: number;
}

export interface SubmitInputResult {
    window: WindowInfo;
    keyChord: string;
}

export interface AdvanceChatResult {
    state: ChatStateInfo;
    clicked: UiElementInfo[];
    typed: boolean;
    submitted: boolean;
    detail: string;
}

export interface SurfaceDetectionOptions {
    surfaceOverride?: string;
    windowTitle?: string;
    processName?: string;
}

const UI_AUTOMATION_BASE = String.raw`
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName UIAutomationClient | Out-Null
Add-Type -AssemblyName UIAutomationTypes | Out-Null
Add-Type @"
using System;
using System.Runtime.InteropServices;

public static class TormentNexusNativeAutomation {
    [DllImport("user32.dll")]
    public static extern IntPtr GetForegroundWindow();

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool SetCursorPos(int x, int y);

    [DllImport("user32.dll")]
    public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, UIntPtr dwExtraInfo);
}
"@ | Out-Null

function Get-ControlTypeName($element) {
    $controlType = $element.Current.ControlType
    if ($null -eq $controlType) { return $null }
    $name = $controlType.ProgrammaticName
    if ([string]::IsNullOrWhiteSpace($name)) { return $null }
    if ($name.StartsWith('ControlType.')) { return $name.Substring(12) }
    return $name
}

function Convert-Bounds($rect) {
    if ($null -eq $rect) { return $null }
    return @{
        left = [int][Math]::Round($rect.Left)
        top = [int][Math]::Round($rect.Top)
        width = [int][Math]::Round($rect.Width)
        height = [int][Math]::Round($rect.Height)
    }
}

function Convert-Element($element) {
    return @{
        name = $element.Current.Name
        automationId = $element.Current.AutomationId
        className = $element.Current.ClassName
        controlType = Get-ControlTypeName $element
        isEnabled = [bool]$element.Current.IsEnabled
        isOffscreen = [bool]$element.Current.IsOffscreen
        hasKeyboardFocus = [bool]$element.Current.HasKeyboardFocus
        bounds = Convert-Bounds $element.Current.BoundingRectangle
    }
}

function Get-ElementTextHint($element) {
    $parts = New-Object System.Collections.Generic.List[string]

    foreach ($value in @($element.Current.Name, $element.Current.AutomationId, $element.Current.ClassName)) {
        if (-not [string]::IsNullOrWhiteSpace($value)) {
            $parts.Add($value) | Out-Null
        }
    }

    $valuePattern = $null
    if ($element.TryGetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern, [ref]$valuePattern)) {
        $value = $valuePattern.Current.Value
        if (-not [string]::IsNullOrWhiteSpace($value)) {
            $parts.Add($value) | Out-Null
        }
    }

    $textPattern = $null
    if ($element.TryGetCurrentPattern([System.Windows.Automation.TextPattern]::Pattern, [ref]$textPattern)) {
        try {
            $text = $textPattern.DocumentRange.GetText(-1)
            if (-not [string]::IsNullOrWhiteSpace($text)) {
                $parts.Add($text) | Out-Null
            }
        } catch {
        }
    }

    return Normalize-Label ($parts -join ' ')
}

function Normalize-Label([string]$value) {
    if ([string]::IsNullOrWhiteSpace($value)) { return '' }
    $normalized = $value.ToLowerInvariant()
    $normalized = [System.Text.RegularExpressions.Regex]::Replace($normalized, '\s+', ' ')
    $normalized = $normalized.Trim()
    return $normalized
}

function Get-ComparableLabel([string]$value) {
    $normalized = Normalize-Label $value
    if ([string]::IsNullOrWhiteSpace($normalized)) { return '' }
    $normalized = [System.Text.RegularExpressions.Regex]::Replace($normalized, '[^a-z0-9]+', ' ')
    $normalized = [System.Text.RegularExpressions.Regex]::Replace($normalized, '\s+', ' ')
    return $normalized.Trim()
}

function Test-ExactishLabelMatch([string]$elementName, [string]$targetLabel) {
    $left = Get-ComparableLabel $elementName
    $right = Get-ComparableLabel $targetLabel
    if ([string]::IsNullOrWhiteSpace($left) -or [string]::IsNullOrWhiteSpace($right)) {
        return $false
    }

    return $left -eq $right
}

function Test-IsTerminalLikeHint([string]$textHint) {
    if ([string]::IsNullOrWhiteSpace($textHint)) {
        return $false
    }

    foreach ($needle in @('@terminal:', 'pwsh', 'powershell', 'terminal', 'shell')) {
        if ($textHint.Contains($needle)) {
            return $true
        }
    }

    return $false
}

function Test-IsLikelyChatInput($element) {
    if (-not $element.Current.IsEnabled -or $element.Current.IsOffscreen) {
        return $false
    }

    $controlType = Get-ControlTypeName $element
    if ($controlType -notin @('Document', 'Edit')) {
        return $false
    }

    $textHint = Get-ElementTextHint $element
    if (Test-IsTerminalLikeHint $textHint) {
        return $false
    }

    return $true
}

function Get-ForegroundAutomationWindow() {
    $handle = [TormentNexusNativeAutomation]::GetForegroundWindow()
    if ($handle -eq [IntPtr]::Zero) {
        throw 'No foreground window available'
    }

    return [System.Windows.Automation.AutomationElement]::FromHandle($handle)
}

function Get-WindowProcessName($window) {
    $processId = $window.Current.ProcessId
    if (-not $processId) { return $null }
    try {
        return (Get-Process -Id $processId -ErrorAction Stop).ProcessName
    } catch {
        return $null
    }
}

function Get-TargetWindow([string]$windowTitle, [string]$processName) {
    if ([string]::IsNullOrWhiteSpace($windowTitle) -and [string]::IsNullOrWhiteSpace($processName)) {
        return Get-ForegroundAutomationWindow
    }

    $windows = [System.Windows.Automation.AutomationElement]::RootElement.FindAll(
        [System.Windows.Automation.TreeScope]::Children,
        [System.Windows.Automation.Condition]::TrueCondition
    )

    foreach ($candidate in $windows) {
        $candidateTitle = $candidate.Current.Name
        $candidateProcessName = Get-WindowProcessName $candidate

        if (-not [string]::IsNullOrWhiteSpace($windowTitle)) {
            if ([string]::IsNullOrWhiteSpace($candidateTitle) -or -not $candidateTitle.ToLowerInvariant().Contains($windowTitle.ToLowerInvariant())) {
                continue
            }
        }

        if (-not [string]::IsNullOrWhiteSpace($processName)) {
            if ([string]::IsNullOrWhiteSpace($candidateProcessName) -or $candidateProcessName.ToLowerInvariant() -ne $processName.ToLowerInvariant()) {
                continue
            }
        }

        return $candidate
    }

    throw "No matching window found"
}

function Get-Inspection([System.Windows.Automation.AutomationElement]$window) {
    $descendants = $window.FindAll(
        [System.Windows.Automation.TreeScope]::Descendants,
        [System.Windows.Automation.Condition]::TrueCondition
    )

    $buttons = @()
    $inputs = @()
    $labels = New-Object System.Collections.Generic.List[string]

    foreach ($element in $descendants) {
        $dto = Convert-Element $element
        $controlType = $dto.controlType
        $name = $dto.name

        if (-not [string]::IsNullOrWhiteSpace($name) -and $labels.Count -lt 80) {
            $normalizedName = Normalize-Label $name
            if (-not [string]::IsNullOrWhiteSpace($normalizedName)) {
                $labels.Add($name)
            }
        }

        if ($controlType -in @('Button', 'Hyperlink')) {
            $buttons += $dto
            continue
        }

        if (($controlType -in @('Edit', 'Document')) -and (Test-IsLikelyChatInput $element)) {
            $inputs += $dto
        }
    }

    return @{
        window = @{
            title = $window.Current.Name
            processName = Get-WindowProcessName $window
            processId = $window.Current.ProcessId
            bounds = Convert-Bounds $window.Current.BoundingRectangle
        }
        buttons = $buttons
        inputs = $inputs
        labels = $labels
    }
}

function Get-BestMatchingElement(
    [System.Windows.Automation.AutomationElement[]]$elements,
    [string[]]$targetLabels
) {
    foreach ($element in $elements) {
        if (-not $element.Current.IsEnabled -or $element.Current.IsOffscreen) {
            continue
        }

        $elementName = Normalize-Label $element.Current.Name
        if ([string]::IsNullOrWhiteSpace($elementName)) {
            continue
        }

        foreach ($targetLabel in $targetLabels) {
            $normalizedTarget = Get-ComparableLabel $targetLabel
            if ([string]::IsNullOrWhiteSpace($normalizedTarget)) {
                continue
            }

            if (Test-ExactishLabelMatch $elementName $normalizedTarget) {
                return $element
            }
        }
    }

    return $null
}

function Get-InteractiveAutomationElements([System.Windows.Automation.AutomationElement]$window, [string[]]$controlTypes) {
    $descendants = $window.FindAll(
        [System.Windows.Automation.TreeScope]::Descendants,
        [System.Windows.Automation.Condition]::TrueCondition
    )

    $matches = New-Object System.Collections.Generic.List[System.Windows.Automation.AutomationElement]

    foreach ($element in $descendants) {
        $typeName = Get-ControlTypeName $element
        if ($controlTypes -contains $typeName) {
            if (($typeName -in @('Edit', 'Document')) -and -not (Test-IsLikelyChatInput $element)) {
                continue
            }
            $matches.Add($element)
        }
    }

    return $matches.ToArray()
}

function Invoke-Element([System.Windows.Automation.AutomationElement]$element) {
    $invokePattern = $null
    if ($element.TryGetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern, [ref]$invokePattern)) {
        $invokePattern.Invoke()
        return 'invoke-pattern'
    }

    $selectionPattern = $null
    if ($element.TryGetCurrentPattern([System.Windows.Automation.SelectionItemPattern]::Pattern, [ref]$selectionPattern)) {
        $selectionPattern.Select()
        return 'selection-pattern'
    }

    if (-not $element.Current.HasKeyboardFocus) {
        $element.SetFocus()
        Start-Sleep -Milliseconds 100
    }

    $rect = $element.Current.BoundingRectangle
    if ($rect.Width -gt 0 -and $rect.Height -gt 0) {
        $x = [int][Math]::Round($rect.Left + ($rect.Width / 2))
        $y = [int][Math]::Round($rect.Top + ($rect.Height / 2))
        [TormentNexusNativeAutomation]::SetCursorPos($x, $y) | Out-Null
        Start-Sleep -Milliseconds 50
        [TormentNexusNativeAutomation]::mouse_event(0x0002, 0, 0, 0, [UIntPtr]::Zero)
        [TormentNexusNativeAutomation]::mouse_event(0x0004, 0, 0, 0, [UIntPtr]::Zero)
        return 'mouse-click'
    }

    throw 'Unable to invoke element'
}
`;

function powerShellStringArray(values: string[]): string {
    return `@(${values.map((value) => toPowerShellString(value)).join(', ')})`;
}

function buildInspectionScript(windowTitle?: string, processName?: string): string {
    return `${UI_AUTOMATION_BASE}
$window = Get-TargetWindow ${toPowerShellString(windowTitle ?? '')} ${toPowerShellString(processName ?? '')}
$result = Get-Inspection $window
$result | ConvertTo-Json -Depth 8 -Compress`;
}

function buildSurfaceProbeScript(windowTitle?: string, processName?: string): string {
    return `${UI_AUTOMATION_BASE}
$window = Get-TargetWindow ${toPowerShellString(windowTitle ?? '')} ${toPowerShellString(processName ?? '')}
$result = @{
    title = $window.Current.Name
    processName = Get-WindowProcessName $window
    processId = $window.Current.ProcessId
    bounds = Convert-Bounds $window.Current.BoundingRectangle
}
$result | ConvertTo-Json -Depth 6 -Compress`;
}

function buildClickScript(labels: string[], delays: Pick<SupervisorSettings, 'afterClickDelayMs'>, windowTitle?: string, processName?: string): string {
    return `${UI_AUTOMATION_BASE}
$window = Get-TargetWindow ${toPowerShellString(windowTitle ?? '')} ${toPowerShellString(processName ?? '')}
$targets = ${powerShellStringArray(labels)}
$buttonElements = Get-InteractiveAutomationElements $window @('Button', 'Hyperlink')
$clicked = @()
$missing = New-Object System.Collections.Generic.List[string]

foreach ($target in $targets) {
    $match = Get-BestMatchingElement $buttonElements @($target)
    if ($null -eq $match) {
        $missing.Add($target)
        continue
    }

    Invoke-Element $match | Out-Null
    $clicked += (Convert-Element $match)
    Start-Sleep -Milliseconds ${delays.afterClickDelayMs}
}

$result = @{
    window = @{
        title = $window.Current.Name
        processName = Get-WindowProcessName $window
        processId = $window.Current.ProcessId
        bounds = Convert-Bounds $window.Current.BoundingRectangle
    }
    clicked = $clicked
    missing = $missing
}

$result | ConvertTo-Json -Depth 8 -Compress`;
}

function buildSetTextScript(
    text: string,
    clearExisting: boolean,
    delays: Pick<SupervisorSettings, 'inputSettleDelayMs'>,
    inputControlTypes: string[],
    windowTitle?: string,
    processName?: string
): string {
    const clearMode = clearExisting ? '$true' : '$false';

    return `${UI_AUTOMATION_BASE}
$window = Get-TargetWindow ${toPowerShellString(windowTitle ?? '')} ${toPowerShellString(processName ?? '')}
$targetText = ${toPowerShellString(text)}
$clearExisting = ${clearMode}
$inputElements = Get-InteractiveAutomationElements $window ${powerShellStringArray(inputControlTypes)}
$target = $null

foreach ($candidate in $inputElements) {
    if ($candidate.Current.HasKeyboardFocus -and $candidate.Current.IsEnabled -and -not $candidate.Current.IsOffscreen) {
        $target = $candidate
        break
    }
}

if ($null -eq $target) {
    foreach ($candidate in $inputElements) {
        if ($candidate.Current.IsEnabled -and -not $candidate.Current.IsOffscreen) {
            $target = $candidate
            break
        }
    }
}

if ($null -eq $target) {
    throw 'No enabled text input found in target window'
}

$method = 'focus-sendkeys'
$valuePattern = $null
if ($target.TryGetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern, [ref]$valuePattern) -and -not $valuePattern.Current.IsReadOnly) {
    if (-not $target.Current.HasKeyboardFocus) {
        $target.SetFocus()
        Start-Sleep -Milliseconds ${delays.inputSettleDelayMs}
    }
    $nextValue = $targetText
    if (-not $clearExisting) {
        $nextValue = $valuePattern.Current.Value + $targetText
    }
    $valuePattern.SetValue($nextValue)
    $method = 'value-pattern'
} else {
    $target.SetFocus()
    Start-Sleep -Milliseconds ${delays.inputSettleDelayMs}
    $wshell = New-Object -ComObject wscript.shell
    if ($clearExisting) {
        $wshell.SendKeys('^a')
        Start-Sleep -Milliseconds 60
        $wshell.SendKeys('{BACKSPACE}')
        Start-Sleep -Milliseconds 60
    }
    $wshell.SendKeys($targetText)
}

$result = @{
    window = @{
        title = $window.Current.Name
        processName = Get-WindowProcessName $window
        processId = $window.Current.ProcessId
        bounds = Convert-Bounds $window.Current.BoundingRectangle
    }
    target = Convert-Element $target
    method = $method
    textLength = $targetText.Length
}

$result | ConvertTo-Json -Depth 8 -Compress`;
}

function buildSubmitScript(
    keyChord: string,
    delays: Pick<SupervisorSettings, 'focusDelayMs'>,
    inputControlTypes: string[],
    windowTitle?: string,
    processName?: string
): string {
    const normalizedKeyChord = keyChord.toLowerCase();

    const sendKeysExpression = normalizedKeyChord === 'alt+enter'
        ? "'%{ENTER}'"
        : normalizedKeyChord === 'ctrl+enter' || normalizedKeyChord === 'control+enter'
            ? "'^{ENTER}'"
            : normalizedKeyChord === 'shift+enter'
                ? "'+{ENTER}'"
                : normalizedKeyChord === 'enter'
                    ? "'{ENTER}'"
                    : toPowerShellString(keyChord);

    return `${UI_AUTOMATION_BASE}
$window = Get-TargetWindow ${toPowerShellString(windowTitle ?? '')} ${toPowerShellString(processName ?? '')}
$inputElements = Get-InteractiveAutomationElements $window ${powerShellStringArray(inputControlTypes)}
$target = $null

foreach ($candidate in $inputElements) {
    if ($candidate.Current.HasKeyboardFocus -and $candidate.Current.IsEnabled -and -not $candidate.Current.IsOffscreen) {
        $target = $candidate
        break
    }
}

if ($null -eq $target) {
    throw 'No enabled chat input found in target window for submission'
}

if (-not $target.Current.HasKeyboardFocus) {
    $target.SetFocus()
    Start-Sleep -Milliseconds ${delays.focusDelayMs}
}

$wshell = New-Object -ComObject wscript.shell
$wshell.SendKeys(${sendKeysExpression})

$result = @{
    window = @{
        title = $window.Current.Name
        processName = Get-WindowProcessName $window
        processId = $window.Current.ProcessId
        bounds = Convert-Bounds $window.Current.BoundingRectangle
    }
    keyChord = ${toPowerShellString(keyChord)}
}

$result | ConvertTo-Json -Depth 8 -Compress`;
}

export class UiAutomationManager {
    private readonly settingsManager: SupervisorSettingsManager;
    private bumpIndex = 0;

    constructor(settingsManager?: SupervisorSettingsManager) {
        this.settingsManager = settingsManager ?? new SupervisorSettingsManager();
    }

    async getSettings(): Promise<SupervisorSettings> {
        return this.settingsManager.getSettings();
    }

    async updateSettings(update: Partial<SupervisorSettings>): Promise<SupervisorSettings> {
        return this.settingsManager.updateSettings(update);
    }

    listSurfaceProfiles(): SurfaceProfile[] {
        return listSurfaceProfiles();
    }

    async inspectWindow(windowTitle?: string, processName?: string): Promise<UiInspection> {
        return runPowerShellJson<UiInspection>(buildInspectionScript(windowTitle, processName));
    }

    async detectChatSurface(options?: SurfaceDetectionOptions): Promise<ChatSurfaceInfo> {
        let title = '';
        let processName: string | null = null;
        let processPath: string | null = null;
        let processId: number | null = null;
        let bounds: WindowBounds | null = null;

        if (options?.windowTitle || options?.processName) {
            const probedWindow = await runPowerShellJson<{
                title?: string;
                processName?: string | null;
                processId?: number | null;
                bounds?: WindowBounds | null;
            }>(buildSurfaceProbeScript(options.windowTitle, options.processName));

            title = probedWindow.title ?? '';
            processName = probedWindow.processName ?? null;
            processId = probedWindow.processId ?? null;
            bounds = probedWindow.bounds ?? null;
        } else {
            const activeWindow = await activeWin();
            title = activeWindow?.title ?? '';
            processName = activeWindow?.owner?.name ?? null;
            processPath = activeWindow?.owner?.path ?? null;
            processId = activeWindow?.owner?.processId ?? null;
            bounds = activeWindow?.bounds
                ? {
                    left: activeWindow.bounds.x,
                    top: activeWindow.bounds.y,
                    width: activeWindow.bounds.width,
                    height: activeWindow.bounds.height
                }
                : null;
        }

        const initialSurface = resolveDetectedSurface({
            title,
            processName,
            windowTargeted: Boolean(options?.windowTitle || options?.processName),
            surfaceOverride: options?.surfaceOverride
        });

        let inspectionSuggestsAntigravity = false;

        if (!options?.surfaceOverride && (initialSurface.browserFamily !== null || initialSurface.detectedSurface === 'browser-chat' || initialSurface.detectedSurface === 'unknown')) {
            try {
                const inspection = await this.inspectWindow(options?.windowTitle, options?.processName);
                inspectionSuggestsAntigravity = inspectionLooksLikeAntigravity(inspection);
            } catch {
            }
        }

        const {
            detectedSurface: resolvedSurfaceId,
            browserFamily,
            heuristics
        } = resolveDetectedSurface({
            title,
            processName,
            windowTargeted: Boolean(options?.windowTitle || options?.processName),
            surfaceOverride: options?.surfaceOverride,
            inspectionSuggestsAntigravity
        });

        const surfaceProfile = resolveSurfaceProfile(resolvedSurfaceId);

        return {
            title,
            processName,
            processPath,
            processId,
            bounds,
            browserFamily,
            detectedSurface: resolvedSurfaceId,
            surfaceProfile,
            heuristics
        };
    }

    async detectChatState(windowTitle?: string, processName?: string, actionLabels?: string[], options?: SurfaceDetectionOptions): Promise<ChatStateInfo> {
        const [surface, inspection, settings] = await Promise.all([
            this.detectChatSurface({ ...options, windowTitle, processName }),
            this.inspectWindow(windowTitle, processName),
            this.getSettings()
        ]);
        const resolvedActionLabels = resolveActionLabels(actionLabels, surface, settings);
        const stateResult = resolveChatState({
            inspection,
            actionLabels: resolvedActionLabels,
            preferredInputControlTypes: surface.surfaceProfile.inputControlTypes
        });

        return {
            surface,
            inspection,
            state: stateResult.state,
            pendingActionButtons: stateResult.pendingActionButtons,
            reasoning: stateResult.reasoning
        };
    }

    async clickActionButtons(
        labels?: string[],
        windowTitle?: string,
        processName?: string,
        options?: SurfaceDetectionOptions
    ): Promise<ClickActionResult> {
        const [settings, surface] = await Promise.all([
            this.getSettings(),
            this.detectChatSurface({ ...options, windowTitle, processName })
        ]);
        const resolvedLabels = resolveActionLabels(labels, surface, settings);
        return runPowerShellJson<ClickActionResult>(buildClickScript(resolvedLabels, settings, windowTitle, processName));
    }

    async setChatInput(text: string, options?: {
        clearExisting?: boolean;
        windowTitle?: string;
        processName?: string;
        surfaceOverride?: string;
    }): Promise<SetInputResult> {
        const [settings, surface] = await Promise.all([
            this.getSettings(),
            this.detectChatSurface({ surfaceOverride: options?.surfaceOverride, windowTitle: options?.windowTitle, processName: options?.processName })
        ]);
        const inputControlTypes = surface.surfaceProfile.inputControlTypes.length > 0
            ? surface.surfaceProfile.inputControlTypes
            : DEFAULT_SURFACE_PROFILE.inputControlTypes;
        return runPowerShellJson<SetInputResult>(
            buildSetTextScript(
                text,
                options?.clearExisting ?? true,
                settings,
                inputControlTypes,
                options?.windowTitle,
                options?.processName
            )
        );
    }

    async submitChatInput(
        keyChord = 'alt+enter',
        windowTitle?: string,
        processName?: string,
        options?: SurfaceDetectionOptions
    ): Promise<SubmitInputResult> {
        const [settings, surface] = await Promise.all([
            this.getSettings(),
            this.detectChatSurface({ ...options, windowTitle, processName })
        ]);
        const inputControlTypes = surface.surfaceProfile.inputControlTypes.length > 0
            ? surface.surfaceProfile.inputControlTypes
            : DEFAULT_SURFACE_PROFILE.inputControlTypes;
        return runPowerShellJson<SubmitInputResult>(buildSubmitScript(keyChord, settings, inputControlTypes, windowTitle, processName));
    }

    async advanceChat(options?: {
        bumpText?: string;
        actionLabels?: string[];
        windowTitle?: string;
        processName?: string;
        surfaceOverride?: string;
    }): Promise<AdvanceChatResult> {
        const [settings, surface] = await Promise.all([
            this.getSettings(),
            this.detectChatSurface({ surfaceOverride: options?.surfaceOverride, windowTitle: options?.windowTitle, processName: options?.processName })
        ]);
        const profile = surface.surfaceProfile ?? DEFAULT_SURFACE_PROFILE;
        const actionLabels = options?.actionLabels ?? profile.actionLabels ?? settings.actionLabels;

        let bumpText = options?.bumpText ?? settings.bumpText;
        if (!options?.bumpText && settings.bumpSentences && settings.bumpSentences.length > 0) {
            bumpText = settings.bumpSentences[this.bumpIndex % settings.bumpSentences.length];
            this.bumpIndex++;
        }
        const submitKeyChord = profile.id === 'antigravity'
            ? 'alt+enter'
            : profile.submitKeyChord ?? 'alt+enter';
        const windowTitle = options?.windowTitle;
        const processName = options?.processName;
        const state = await this.detectChatState(windowTitle, processName, actionLabels, { surfaceOverride: options?.surfaceOverride });

        if (state.state === 'awaiting_action') {
            const clickResult = await this.clickActionButtons(actionLabels, windowTitle, processName, { surfaceOverride: options?.surfaceOverride });
            return {
                state,
                clicked: clickResult.clicked,
                typed: false,
                submitted: false,
                detail: clickResult.clicked.length > 0
                    ? `Clicked ${clickResult.clicked.map((item) => item.name).join(', ')}`
                    : 'No matching action buttons were clicked'
            };
        }

        if (state.state === 'ready_for_input' && bumpText) {
            await this.setChatInput(bumpText, {
                clearExisting: true,
                windowTitle,
                processName,
                surfaceOverride: options?.surfaceOverride
            });

            await this.submitChatInput(submitKeyChord, windowTitle, processName, { surfaceOverride: options?.surfaceOverride });
            return {
                state,
                clicked: [],
                typed: true,
                submitted: true,
                detail: `Typed bump text and submitted with ${submitKeyChord} for ${profile.id}`
            };
        }

        return {
            state,
            clicked: [],
            typed: false,
            submitted: false,
            detail: `No action taken because the current UI state was unknown or no bump text was provided for ${profile.id}`
        };
    }

    async sendKeys(keys: string, windowTitle?: string): Promise<string> {
        const keyMap: Record<string, string> = {
            'ctrl+r': '^{r}',
            'f5': '{F5}',
            'enter': '{ENTER}',
            'esc': '{ESC}',
            'control+enter': '^{ENTER}',
            'ctrl+enter': '^{ENTER}',
            'shift+enter': '+{ENTER}',
            'alt+enter': '%{ENTER}'
        };

        const command = keyMap[keys.toLowerCase()] || keys;
        const activationScript = windowTitle
            ? `
$window = Get-TargetWindow ${toPowerShellString(windowTitle)} ''
$window.SetFocus()
Start-Sleep -Milliseconds 100
`
            : '';

        await runPowerShell(`${UI_AUTOMATION_BASE}
${activationScript}
$wshell = New-Object -ComObject wscript.shell
$wshell.SendKeys(${toPowerShellString(command)})`);

        return `Successfully sent keys: ${keys}${windowTitle ? ` to '${windowTitle}'` : ''}`;
    }
}
