import * as vscode from 'vscode';
import WebSocket from 'ws';

let socket: WebSocket | null = null;
let statusBarItem: vscode.StatusBarItem;
let outputChannel: vscode.OutputChannel;
let sidebarProvider: TormentNexusSidebarProvider | null = null;
let reconnectTimer: NodeJS.Timeout | null = null;
let lastActivityTime = Date.now();
let debounceTimer: NodeJS.Timeout | null = null;
let ignoreNextActivity = false;
const terminalIds = new WeakMap<vscode.Terminal, string>();
const terminalBuffers = new Map<string, string>();
const activityFeed: ActivityEntry[] = [];
let terminalSequence = 0;
const MAX_TERMINAL_BUFFER = 20_000;
const MAX_ACTIVITY_ITEMS = 12;
const MAX_CHAT_HISTORY_ITEMS = 24;
const chatHistoryFeed: ChatHistoryEntry[] = [];

const VSCODE_BRIDGE_CAPABILITIES = [
    'bridge.websocket',
    'memory.capture',
    'rag.ingest',
    'chat.inject',
    'command.execute',
    'editor.selection.read',
    'terminal.buffer.read',
];

const VSCODE_BRIDGE_HOOK_PHASES = [
    'session.start',
    'user.activity',
    'context.capture',
    'memory.capture',
    'chat.submit',
    'editor.selection',
    'terminal.output',
];

const DASHBOARD_ROUTES = {
    home: '/dashboard',
    memory: '/dashboard/memory',
    tools: '/dashboard/mcp/ai-tools',
    logs: '/dashboard/mcp/logs',
    analytics: '/dashboard/metrics',
    debate: '/dashboard/council',
    templates: '/dashboard/council',
    architecture: '/dashboard/architecture',
};

type TerminalDataEvent = {
    terminal: vscode.Terminal;
    data: string;
};

type ProposedTerminalWindow = typeof vscode.window & {
    onDidWriteTerminalData?: (listener: (event: TerminalDataEvent) => void) => vscode.Disposable;
};

type HubStatus = {
    connectionState: 'connected' | 'disconnected';
    researcher: string;
    coder: string;
    error?: string;
};

type ActivityEntry = {
    time: string;
    title: string;
    detail: string;
    kind: 'status' | 'research' | 'code' | 'memory' | 'rag' | 'tool' | 'navigation' | 'system';
};

type ChatHistoryEntry = {
    time: string;
    role: 'system' | 'user' | 'assistant';
    source: 'extension' | 'research-agent' | 'coder-agent' | 'chat-bridge' | 'editor-snapshot';
    content: string;
};

type SidebarSnapshot = {
    status: HubStatus;
    recentActivity: ActivityEntry[];
    activeEditor: string;
    activeTerminal: string;
    dashboardUrl: string;
};

function createNonce(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let value = '';
    for (let i = 0; i < 32; i++) {
        value += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return value;
}

function formatResult(value: unknown): string {
    if (typeof value === 'string') {
        return value;
    }

    try {
        return JSON.stringify(value, null, 2);
    } catch {
        return String(value);
    }
}

function log(message: string) {
    const timestamp = new Date().toISOString();
    outputChannel.appendLine(`[${timestamp}] ${message}`);
    emitInspectorLog('info', message);
}

function emitInspectorLog(level: 'info' | 'warn' | 'error', message: string, url = 'vscode://tormentnexus-vscode-extension') {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        return;
    }

    socket.send(JSON.stringify({
        type: 'BROWSER_LOG',
        level,
        content: message,
        message,
        timestamp: Date.now(),
        url,
        source: 'vscode_extension',
    }));
}

function addActivity(kind: ActivityEntry['kind'], title: string, detail: string) {
    activityFeed.unshift({
        kind,
        title,
        detail,
        time: new Date().toLocaleTimeString(),
    });

    if (activityFeed.length > MAX_ACTIVITY_ITEMS) {
        activityFeed.length = MAX_ACTIVITY_ITEMS;
    }

    log(`${title}: ${detail}`);
    void sidebarProvider?.refreshSnapshot();
}

function addChatHistory(role: ChatHistoryEntry['role'], source: ChatHistoryEntry['source'], content: string) {
    const trimmed = content.trim();
    if (!trimmed) {
        return;
    }

    chatHistoryFeed.unshift({
        role,
        source,
        content: trimmed,
        time: new Date().toLocaleTimeString(),
    });

    if (chatHistoryFeed.length > MAX_CHAT_HISTORY_ITEMS) {
        chatHistoryFeed.length = MAX_CHAT_HISTORY_ITEMS;
    }
}

function summarizeText(value: string, maxLength = 600): string {
    const normalized = value.replace(/\s+/g, ' ').trim();
    if (normalized.length <= maxLength) {
        return normalized;
    }

    return `${normalized.slice(0, maxLength)}…`;
}

function getVisibleChatEditorSnapshots(): ChatHistoryEntry[] {
    const snapshots: ChatHistoryEntry[] = [];

    for (const editor of vscode.window.visibleTextEditors) {
        const document = editor.document;
        const uriString = document.uri.toString().toLowerCase();
        const fileName = document.fileName.toLowerCase();
        const languageId = document.languageId.toLowerCase();
        const isLikelyChatDocument = uriString.includes('chat')
            || fileName.includes('chat')
            || languageId.includes('chat')
            || uriString.includes('copilot');

        if (!isLikelyChatDocument) {
            continue;
        }

        const text = summarizeText(document.getText(), 1000);
        if (!text) {
            continue;
        }

        snapshots.push({
            role: 'assistant',
            source: 'editor-snapshot',
            time: new Date().toLocaleTimeString(),
            content: `[Editor Snapshot] ${document.uri.toString()} :: ${text}`,
        });
    }

    return snapshots;
}

function getChatHistoryLines(): string[] {
    const combined = [...chatHistoryFeed, ...getVisibleChatEditorSnapshots()];
    if (combined.length === 0) {
        return ['[System][extension]: No captured VS Code chat interactions yet.'];
    }

    return combined.slice(0, MAX_CHAT_HISTORY_ITEMS).map((entry) => {
        return `[${entry.role}][${entry.source}][${entry.time}] ${entry.content}`;
    });
}

function getWorkspaceState() {
    return {
        activeEditor: vscode.window.activeTextEditor
            ? vscode.workspace.asRelativePath(vscode.window.activeTextEditor.document.uri)
            : 'No active editor',
        activeTerminal: vscode.window.activeTerminal?.name ?? 'No active terminal',
    };
}

function resolveCoreHttpUrl(): string {
    const config = vscode.workspace.getConfiguration('tormentnexus');
    const wsUrl = config.get<string>('coreUrl', 'ws://localhost:3001');

    try {
        const url = new URL(wsUrl);
        url.protocol = url.protocol === 'wss:' ? 'https:' : 'http:';
        return url.toString().replace(/\/$/, '');
    } catch {
        return 'http://localhost:3001';
    }
}

function resolveDashboardBaseUrl(): string {
    const config = vscode.workspace.getConfiguration('tormentnexus');
    const configured = config.get<string>('dashboardUrl', 'http://localhost:3000');
    return configured.replace(/\/$/, '');
}

function buildDashboardUrl(route: string): string {
    return `${resolveDashboardBaseUrl()}${route}`;
}

async function openDashboardRoute(route: string, label: string) {
    const url = buildDashboardUrl(route);
    await vscode.env.openExternal(vscode.Uri.parse(url));
    addActivity('navigation', label, url);
}

async function postCoreJson<T>(path: string, body: Record<string, unknown>): Promise<T> {
    const response = await fetch(`${resolveCoreHttpUrl()}${path}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
    });

    const json = await response.json() as T & { success?: boolean; error?: string };
    if (!response.ok || ('success' in json && json.success === false)) {
        throw new Error((json as { error?: string }).error || `HTTP ${response.status}`);
    }

    return json as T;
}

async function fetchHubStatus(): Promise<HubStatus> {
    try {
        const status = await postCoreJson<{ success: true; researcher: string; coder: string }>('/expert.status', {});
        return {
            connectionState: socket?.readyState === WebSocket.OPEN ? 'connected' : 'disconnected',
            researcher: status.researcher,
            coder: status.coder,
        };
    } catch (error: unknown) {
        return {
            connectionState: socket?.readyState === WebSocket.OPEN ? 'connected' : 'disconnected',
            researcher: 'unknown',
            coder: 'unknown',
            error: error instanceof Error ? error.message : String(error),
        };
    }
}

async function createSidebarSnapshot(): Promise<SidebarSnapshot> {
    const state = getWorkspaceState();
    return {
        status: await fetchHubStatus(),
        recentActivity: [...activityFeed],
        activeEditor: state.activeEditor,
        activeTerminal: state.activeTerminal,
        dashboardUrl: buildDashboardUrl(DASHBOARD_ROUTES.home),
    };
}

class TormentNexusSidebarProvider implements vscode.WebviewViewProvider {
    public static readonly viewType = 'tormentnexus.dispatchView';
    private view?: vscode.WebviewView;

    resolveWebviewView(webviewView: vscode.WebviewView): void | Thenable<void> {
        this.view = webviewView;
        webviewView.webview.options = {
            enableScripts: true,
        };

        webviewView.webview.html = this.getHtml();
        webviewView.webview.onDidReceiveMessage(async (message) => {
            try {
                switch (message.type) {
                    case 'ready':
                    case 'refreshSnapshot':
                        await this.refreshSnapshot();
                        break;
                    case 'runResearch': {
                        const result = await dispatchResearchTask(String(message.query ?? ''), Number(message.depth ?? 2));
                        this.postMessage({ type: 'result', mode: 'research', value: formatResult(result) });
                        break;
                    }
                    case 'runCode': {
                        const result = await dispatchCodeTask(String(message.task ?? ''));
                        this.postMessage({ type: 'result', mode: 'code', value: formatResult(result) });
                        break;
                    }
                    case 'rememberSelection':
                        await rememberSelection();
                        break;
                    case 'ingestSelectionToRag':
                        await ingestSelectionToRag();
                        break;
                    case 'ingestUrl':
                        await ingestUrl();
                        break;
                    case 'openDashboard':
                        await openDashboardRoute(DASHBOARD_ROUTES.home, 'Opened dashboard');
                        break;
                    case 'showLogs':
                        await showLogs();
                        break;
                    case 'searchMemory':
                        await searchMemory();
                        break;
                    case 'listTools':
                        await listTools();
                        break;
                    case 'viewAnalytics':
                        await viewAnalytics();
                        break;
                    case 'startDebate':
                        await startDebate();
                        break;
                    case 'listDebateTemplates':
                        await listDebateTemplates();
                        break;
                    case 'architectMode':
                        await architectMode();
                        break;
                    case 'invokeTool':
                        await invokeTool();
                        break;
                    default:
                        break;
                }

                await this.refreshSnapshot();
            } catch (error) {
                const messageText = error instanceof Error ? error.message : String(error);
                this.postMessage({ type: 'error', value: messageText });
            }
        });
    }

    async refreshSnapshot() {
        if (!this.view) {
            return;
        }

        const snapshot = await createSidebarSnapshot();
        this.postMessage({ type: 'snapshot', value: snapshot });
    }

    private postMessage(message: unknown) {
        this.view?.webview.postMessage(message);
    }

    private getHtml(): string {
        const nonce = createNonce();
        const csp = `default-src 'none'; style-src 'unsafe-inline'; script-src 'nonce-${nonce}';`;

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta http-equiv="Content-Security-Policy" content="${csp}" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <style>
        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            padding: 12px;
        }
        h2, h3 { margin: 0 0 8px; }
        .card {
            border: 1px solid var(--vscode-panel-border);
            border-radius: 8px;
            padding: 12px;
            margin-bottom: 12px;
            background: color-mix(in srgb, var(--vscode-editor-background) 94%, transparent);
        }
        .grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
        }
        .row {
            display: flex;
            gap: 8px;
            margin-top: 8px;
        }
        .row > * { flex: 1; }
        button, textarea, select, input {
            width: 100%;
            box-sizing: border-box;
            font: inherit;
        }
        button {
            border: none;
            border-radius: 6px;
            padding: 8px 10px;
            cursor: pointer;
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
        }
        button.secondary {
            background: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
        }
        textarea, select, input {
            background: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            border-radius: 6px;
            padding: 8px;
        }
        textarea { min-height: 84px; resize: vertical; }
        .metric {
            border-radius: 8px;
            padding: 10px;
            background: var(--vscode-editorWidget-background);
        }
        .metric .label {
            font-size: 11px;
            opacity: 0.8;
            text-transform: uppercase;
            letter-spacing: 0.04em;
        }
        .metric .value {
            font-size: 13px;
            margin-top: 4px;
        }
        .result {
            white-space: pre-wrap;
            word-break: break-word;
            background: var(--vscode-textCodeBlock-background);
            border-radius: 6px;
            padding: 10px;
            min-height: 56px;
            font-family: var(--vscode-editor-font-family);
            font-size: 12px;
        }
        ul.feed {
            list-style: none;
            padding: 0;
            margin: 0;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }
        ul.feed li {
            padding: 8px;
            border-radius: 6px;
            background: var(--vscode-editorWidget-background);
        }
        .feed-title {
            font-size: 12px;
            font-weight: 600;
        }
        .feed-meta {
            font-size: 11px;
            opacity: 0.78;
            margin-top: 2px;
        }
        .muted {
            opacity: 0.76;
            font-size: 12px;
        }
        .pill {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 999px;
            font-size: 11px;
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
        }
    </style>
</head>
<body>
    <div class="card">
        <h2>TormentNexus Mini Dashboard</h2>
        <div class="grid">
            <div class="metric">
                <div class="label">Connection</div>
                <div class="value" id="connectionValue">Checking…</div>
            </div>
            <div class="metric">
                <div class="label">Experts</div>
                <div class="value"><span id="researcherValue">—</span> / <span id="coderValue">—</span></div>
            </div>
            <div class="metric">
                <div class="label">Active Editor</div>
                <div class="value" id="editorValue">—</div>
            </div>
            <div class="metric">
                <div class="label">Active Terminal</div>
                <div class="value" id="terminalValue">—</div>
            </div>
        </div>
        <div class="row">
            <button id="refreshBtn" class="secondary">Refresh</button>
            <button id="dashboardBtn" class="secondary">Open Dashboard</button>
            <button id="logsBtn" class="secondary">Logs</button>
        </div>
    </div>

    <div class="card">
        <h3>Quick Actions</h3>
        <div class="grid">
            <button id="rememberBtn">Remember Selection</button>
            <button id="ragBtn">Ingest to RAG</button>
            <button id="urlBtn">Ingest URL</button>
            <button id="memoryBtn">Search Memory</button>
            <button id="toolsBtn">Tools</button>
            <button id="invokeToolBtn">Invoke Tool</button>
            <button id="analyticsBtn">Analytics</button>
            <button id="debateBtn">Start Debate</button>
            <button id="templatesBtn">Debate Templates</button>
            <button id="architectBtn">Architect Mode</button>
        </div>
        <div class="muted" style="margin-top: 8px;">Dashboard-linked actions open the richer TormentNexus web UI when a direct Core endpoint does not exist yet.</div>
    </div>

    <div class="card">
        <h3>Research Agent</h3>
        <textarea id="researchQuery" placeholder="Ask TormentNexus to research a topic…"></textarea>
        <div class="row">
            <select id="researchDepth">
                <option value="1">Depth 1</option>
                <option value="2" selected>Depth 2</option>
                <option value="3">Depth 3</option>
                <option value="4">Depth 4</option>
                <option value="5">Depth 5</option>
            </select>
            <button id="researchBtn">Run Research</button>
        </div>
    </div>

    <div class="card">
        <h3>Coder Agent</h3>
        <textarea id="codeTask" placeholder="Describe a coding task for TormentNexus…"></textarea>
        <div class="row">
            <button id="codeBtn">Run Coder</button>
        </div>
    </div>

    <div class="card">
        <h3>Recent Tasks</h3>
        <ul id="feed" class="feed">
            <li><span class="muted">Waiting for TormentNexus activity…</span></li>
        </ul>
    </div>

    <div class="card">
        <h3>Latest Result</h3>
        <div id="result" class="result">Ready.</div>
        <div class="muted" id="dashboardUrl">—</div>
    </div>

    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        const resultEl = document.getElementById('result');
        const connectionValue = document.getElementById('connectionValue');
        const researcherValue = document.getElementById('researcherValue');
        const coderValue = document.getElementById('coderValue');
        const editorValue = document.getElementById('editorValue');
        const terminalValue = document.getElementById('terminalValue');
        const feedEl = document.getElementById('feed');
        const dashboardUrlEl = document.getElementById('dashboardUrl');

        function post(type, payload = {}) {
            vscode.postMessage({ type, ...payload });
        }

        function renderFeed(items) {
            if (!items || items.length === 0) {
                feedEl.innerHTML = '<li><span class="muted">No recent tasks yet.</span></li>';
                return;
            }

            feedEl.innerHTML = items.map((item) => {
                return '<li>' +
                    '<div class="feed-title">' + item.title + ' <span class="pill">' + item.kind + '</span></div>' +
                    '<div class="feed-meta">' + item.time + ' • ' + item.detail + '</div>' +
                '</li>';
            }).join('');
        }

        document.getElementById('refreshBtn').addEventListener('click', () => post('refreshSnapshot'));
        document.getElementById('dashboardBtn').addEventListener('click', () => post('openDashboard'));
        document.getElementById('logsBtn').addEventListener('click', () => post('showLogs'));
        document.getElementById('rememberBtn').addEventListener('click', () => post('rememberSelection'));
        document.getElementById('ragBtn').addEventListener('click', () => post('ingestSelectionToRag'));
        document.getElementById('urlBtn').addEventListener('click', () => post('ingestUrl'));
        document.getElementById('memoryBtn').addEventListener('click', () => post('searchMemory'));
        document.getElementById('toolsBtn').addEventListener('click', () => post('listTools'));
        document.getElementById('invokeToolBtn').addEventListener('click', () => post('invokeTool'));
        document.getElementById('analyticsBtn').addEventListener('click', () => post('viewAnalytics'));
        document.getElementById('debateBtn').addEventListener('click', () => post('startDebate'));
        document.getElementById('templatesBtn').addEventListener('click', () => post('listDebateTemplates'));
        document.getElementById('architectBtn').addEventListener('click', () => post('architectMode'));
        document.getElementById('researchBtn').addEventListener('click', () => {
            post('runResearch', {
                query: document.getElementById('researchQuery').value,
                depth: Number(document.getElementById('researchDepth').value),
            });
        });
        document.getElementById('codeBtn').addEventListener('click', () => {
            post('runCode', {
                task: document.getElementById('codeTask').value,
            });
        });

        window.addEventListener('message', (event) => {
            const message = event.data;
            if (message.type === 'snapshot') {
                const snapshot = message.value;
                connectionValue.textContent = snapshot.status.connectionState + (snapshot.status.error ? ' (' + snapshot.status.error + ')' : '');
                researcherValue.textContent = snapshot.status.researcher;
                coderValue.textContent = snapshot.status.coder;
                editorValue.textContent = snapshot.activeEditor;
                terminalValue.textContent = snapshot.activeTerminal;
                dashboardUrlEl.textContent = 'Dashboard: ' + snapshot.dashboardUrl;
                renderFeed(snapshot.recentActivity);
            } else if (message.type === 'result') {
                resultEl.textContent = '[' + message.mode + ']\n' + message.value;
            } else if (message.type === 'error') {
                resultEl.textContent = 'Error: ' + message.value;
            }
        });

        post('ready');
    </script>
</body>
</html>`;
    }
}

export function activate(context: vscode.ExtensionContext) {
    outputChannel = vscode.window.createOutputChannel('TormentNexus Bridge');
    sidebarProvider = new TormentNexusSidebarProvider();
    context.subscriptions.push(vscode.window.registerWebviewViewProvider(TormentNexusSidebarProvider.viewType, sidebarProvider));

    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.command = 'tormentnexus.connect';
    updateStatusBar(false);
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);
    context.subscriptions.push(outputChannel);

    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.connect', connectToCore));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.disconnect', disconnectFromCore));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.showStatus', showHubStatus));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.runAgent', runAgentDispatch));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.searchMemory', searchMemory));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.ingestSelectionToRag', ingestSelectionToRag));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.ingestUrl', ingestUrl));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.rememberSelection', rememberSelection));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.listTools', listTools));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.invokeTool', invokeTool));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.openDashboard', () => openDashboardRoute(DASHBOARD_ROUTES.home, 'Opened dashboard')));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.showLogs', showLogs));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.startDebate', startDebate));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.viewAnalytics', viewAnalytics));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.listDebateTemplates', listDebateTemplates));
    context.subscriptions.push(vscode.commands.registerCommand('tormentnexus.architectMode', architectMode));

    const config = vscode.workspace.getConfiguration('tormentnexus');
    if (config.get<boolean>('autoConnect', true)) {
        void connectToCore();
    }

    context.subscriptions.push(vscode.window.onDidChangeTextEditorSelection(() => {
        if (ignoreNextActivity) {
            return;
        }
        lastActivityTime = Date.now();
        if (debounceTimer) {
            clearTimeout(debounceTimer);
        }
        debounceTimer = setTimeout(sendActivity, 1000);
        void sidebarProvider?.refreshSnapshot();
    }));

    context.subscriptions.push(vscode.workspace.onDidChangeTextDocument(() => {
        if (ignoreNextActivity) {
            return;
        }
        lastActivityTime = Date.now();
        if (debounceTimer) {
            clearTimeout(debounceTimer);
        }
        debounceTimer = setTimeout(sendActivity, 1000);
    }));

    context.subscriptions.push(vscode.window.onDidChangeActiveTextEditor(() => {
        void sidebarProvider?.refreshSnapshot();
    }));

    context.subscriptions.push(vscode.window.onDidChangeActiveTerminal(() => {
        void sidebarProvider?.refreshSnapshot();
    }));

    context.subscriptions.push(vscode.window.onDidCloseTerminal((terminal) => {
        const terminalId = terminalIds.get(terminal);
        if (!terminalId) {
            return;
        }
        terminalBuffers.delete(terminalId);
        terminalIds.delete(terminal);
        void sidebarProvider?.refreshSnapshot();
    }));

    const terminalWindow = vscode.window as ProposedTerminalWindow;
    if (typeof terminalWindow.onDidWriteTerminalData === 'function') {
        context.subscriptions.push(terminalWindow.onDidWriteTerminalData((event) => {
            const terminalId = getTerminalId(event.terminal);
            const existing = terminalBuffers.get(terminalId) ?? '';
            terminalBuffers.set(terminalId, (existing + event.data).slice(-MAX_TERMINAL_BUFFER));
        }));
        addActivity('system', 'Terminal capture enabled', 'Using terminal write event for rolling output buffer.');
    } else {
        addActivity('system', 'Terminal capture unavailable', 'VS Code runtime does not expose terminal write event.');
    }

    addActivity('system', 'Extension activated', 'TormentNexus VS Code bridge is online.');
    addChatHistory('system', 'extension', 'VS Code extension bridge activated.');
}

export function deactivate() {
    disconnectFromCore();
}

function updateStatusBar(connected: boolean) {
    if (connected) {
        statusBarItem.text = '$(plug) TormentNexus: Connected';
        statusBarItem.tooltip = 'Connected to TormentNexus Core';
        statusBarItem.backgroundColor = undefined;
    } else {
        statusBarItem.text = '$(debug-disconnect) TormentNexus: Disconnected';
        statusBarItem.tooltip = 'Click to Connect';
        statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
    }
}

function sendActivity() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
            type: 'USER_ACTIVITY',
            lastActivityTime: Date.now(),
        }));
    }
}

async function connectToCore() {
    if (socket) {
        return;
    }

    const config = vscode.workspace.getConfiguration('tormentnexus');
    const url = config.get<string>('coreUrl', 'ws://localhost:3001');
    addActivity('system', 'Connecting to core', url);

    try {
        socket = new WebSocket(url);

        socket.on('open', () => {
            updateStatusBar(true);
            if (reconnectTimer) {
                clearTimeout(reconnectTimer);
                reconnectTimer = null;
            }
            socket?.send(JSON.stringify({
                type: 'TORMENTNEXUS_CLIENT_HELLO',
                clientType: 'vscode-extension',
                clientName: 'TormentNexus VS Code Bridge',
                platform: `${vscode.env.appName} ${vscode.version}`,
                capabilities: VSCODE_BRIDGE_CAPABILITIES,
                hookPhases: VSCODE_BRIDGE_HOOK_PHASES,
            }));
            addActivity('status', 'Connected to TormentNexus Core', url);
            void sidebarProvider?.refreshSnapshot();
        });

        socket.on('message', async (data) => {
            try {
                const message = JSON.parse(data.toString()) as Record<string, unknown>;
                await handleMessage(message);
            } catch (error) {
                log(`Failed to parse message: ${error instanceof Error ? error.message : String(error)}`);
            }
        });

        socket.on('close', () => {
            updateStatusBar(false);
            socket = null;
            addActivity('status', 'Disconnected from TormentNexus Core', 'Retrying in 5 seconds.');
            void sidebarProvider?.refreshSnapshot();
            reconnectTimer = setTimeout(() => {
                void connectToCore();
            }, 5000);
        });

        socket.on('error', (error) => {
            emitInspectorLog('error', `Socket error: ${error.message}`);
            log(`Socket error: ${error.message}`);
            socket?.close();
        });
    } catch (error) {
        addActivity('status', 'Connection setup failed', error instanceof Error ? error.message : String(error));
    }
}

function disconnectFromCore() {
    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
    }

    if (socket) {
        socket.close();
        socket = null;
    }

    updateStatusBar(false);
    addActivity('status', 'Disconnected manually', 'TormentNexus Core connection closed.');
    void sidebarProvider?.refreshSnapshot();
}

function getTerminalId(terminal: vscode.Terminal): string {
    const existing = terminalIds.get(terminal);
    if (existing) {
        return existing;
    }

    terminalSequence += 1;
    const created = `terminal-${terminalSequence}`;
    terminalIds.set(terminal, created);
    terminalBuffers.set(created, terminalBuffers.get(created) ?? '');
    return created;
}

async function showHubStatus() {
    const status = await fetchHubStatus();
    addActivity('status', 'Hub status checked', `${status.connectionState}; researcher=${status.researcher}; coder=${status.coder}`);
    addChatHistory('system', 'extension', `Hub status checked: ${status.connectionState}; researcher=${status.researcher}; coder=${status.coder}`);
    void vscode.window.showInformationMessage(`TormentNexus Hub ${status.connectionState} — Researcher: ${status.researcher}, Coder: ${status.coder}`);
}

async function dispatchResearchTask(query: string, depth: number): Promise<unknown> {
    const trimmedQuery = query.trim();
    if (!trimmedQuery) {
        throw new Error('Research query is required');
    }

    addChatHistory('user', 'research-agent', `Research query: ${trimmedQuery}`);
    addActivity('research', 'Research task dispatched', trimmedQuery);
    const result = await postCoreJson<{ success: true; kind: 'research'; result: unknown }>('/expert.dispatch', {
        kind: 'research',
        query: trimmedQuery,
        depth,
        breadth: 3,
    });

    log(`Research result: ${formatResult(result.result)}`);
    addChatHistory('assistant', 'research-agent', `Research result: ${summarizeText(formatResult(result.result), 900)}`);
    addActivity('research', 'Research task completed', trimmedQuery);
    return result.result;
}

async function dispatchCodeTask(task: string): Promise<unknown> {
    const trimmedTask = task.trim();
    if (!trimmedTask) {
        throw new Error('Coding task is required');
    }

    addChatHistory('user', 'coder-agent', `Coding task: ${trimmedTask}`);
    addActivity('code', 'Coding task dispatched', trimmedTask);
    const result = await postCoreJson<{ success: true; kind: 'code'; result: unknown }>('/expert.dispatch', {
        kind: 'code',
        task: trimmedTask,
    });

    log(`Coder result: ${formatResult(result.result)}`);
    addChatHistory('assistant', 'coder-agent', `Coder result: ${summarizeText(formatResult(result.result), 900)}`);
    addActivity('code', 'Coding task completed', trimmedTask);
    return result.result;
}

async function runAgentDispatch() {
    const target = await vscode.window.showQuickPick([
        { label: 'Research Agent', value: 'research' },
        { label: 'Coder Agent', value: 'code' },
    ], {
        placeHolder: 'Choose which TormentNexus expert to run',
    });

    if (!target) {
        return;
    }

    if (target.value === 'research') {
        const query = await vscode.window.showInputBox({
            prompt: 'Research query',
            placeHolder: 'Compare current MCP router orchestration patterns',
            ignoreFocusOut: true,
        });
        if (!query?.trim()) {
            return;
        }

        const depthPick = await vscode.window.showQuickPick(['1', '2', '3', '4', '5'], {
            placeHolder: 'Research depth',
        });
        if (!depthPick) {
            return;
        }

        const result = await dispatchResearchTask(query, Number(depthPick));
        outputChannel.show(true);
        void vscode.window.showInformationMessage(`Research complete. ${formatResult(result).slice(0, 120)}`);
        return;
    }

    const seededTask = vscode.window.activeTextEditor?.document.getText(vscode.window.activeTextEditor.selection).trim() ?? '';
    const task = await vscode.window.showInputBox({
        prompt: 'Coding task',
        value: seededTask,
        placeHolder: 'Refactor the active module or add a helper function',
        ignoreFocusOut: true,
    });
    if (!task?.trim()) {
        return;
    }

    const result = await dispatchCodeTask(task);
    outputChannel.show(true);
    void vscode.window.showInformationMessage(`Coder complete. ${formatResult(result).slice(0, 120)}`);
}

async function rememberSelection() {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        void vscode.window.showWarningMessage('TormentNexus Core is not connected. Connect first to save context.');
        return;
    }

    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        void vscode.window.showWarningMessage('No active editor to capture.');
        return;
    }

    const selectionText = editor.document.getText(editor.selection).trim();
    const fullText = editor.document.getText().trim();
    const content = selectionText || fullText;

    if (!content) {
        void vscode.window.showWarningMessage('There is no editor content to save.');
        return;
    }

    const title = selectionText
        ? `VS Code Selection: ${vscode.workspace.asRelativePath(editor.document.uri)}`
        : `VS Code File: ${vscode.workspace.asRelativePath(editor.document.uri)}`;

    socket.send(JSON.stringify({
        type: 'KNOWLEDGE_CAPTURE',
        requestId: `knowledge-capture-${Date.now()}`,
        title,
        url: editor.document.uri.toString(),
        source: 'vscode_extension',
        timestamp: Date.now(),
        content,
    }));

    addActivity('memory', 'Saved editor context to memory', vscode.workspace.asRelativePath(editor.document.uri));
    void vscode.window.showInformationMessage(`Saved context from ${vscode.workspace.asRelativePath(editor.document.uri)} to TormentNexus memory.`);
}

async function ingestSelectionToRag() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        void vscode.window.showWarningMessage('No active editor to ingest into RAG.');
        return;
    }

    const selectionText = editor.document.getText(editor.selection).trim();
    const fullText = editor.document.getText().trim();
    const text = selectionText || fullText;
    if (!text) {
        void vscode.window.showWarningMessage('There is no editor content to ingest into RAG.');
        return;
    }

    const relativePath = vscode.workspace.asRelativePath(editor.document.uri);
    const sourceName = selectionText
        ? `VS Code Selection: ${relativePath}`
        : `VS Code File: ${relativePath}`;

    const response = await postCoreJson<{ success: boolean; chunksIngested: number }>('/rag.ingest-text', {
        text,
        sourceName,
        userId: 'default',
        chunkSize: 1000,
        chunkOverlap: 200,
        strategy: 'recursive',
    });

    addActivity('rag', 'Ingested editor content into RAG', `${relativePath} (${response.chunksIngested} chunks)`);
    void vscode.window.showInformationMessage(`Ingested ${relativePath} into TormentNexus RAG (${response.chunksIngested} chunks).`);
}

async function ingestUrl() {
    const seededUrl = (() => {
        const candidate = vscode.window.activeTextEditor?.document.getText(vscode.window.activeTextEditor.selection).trim();
        if (candidate && /^https?:\/\//i.test(candidate)) {
            return candidate;
        }
        return '';
    })();

    const url = await vscode.window.showInputBox({
        prompt: 'URL to ingest into TormentNexus Knowledge',
        placeHolder: 'https://example.com/docs',
        value: seededUrl,
        ignoreFocusOut: true,
        validateInput: (value) => {
            const trimmed = value.trim();
            if (!trimmed) {
                return 'URL is required';
            }

            return /^https?:\/\//i.test(trimmed) ? undefined : 'URL must start with http:// or https://';
        },
    });

    if (!url?.trim()) {
        return;
    }

    const trimmedUrl = url.trim();
    const response = await postCoreJson<{ success: boolean; result: string }>('/knowledge.ingest-url', {
        url: trimmedUrl,
        source: 'vscode_extension',
    });

    addChatHistory('user', 'extension', `URL ingest requested: ${trimmedUrl}`);
    addChatHistory('assistant', 'extension', `URL ingest result: ${summarizeText(response.result, 900)}`);
    addActivity('memory', 'Ingested URL into Knowledge', trimmedUrl);
    void vscode.window.showInformationMessage(`Ingested URL into TormentNexus Knowledge: ${trimmedUrl}`);
}

async function searchMemory() {
    const query = await vscode.window.showInputBox({
        prompt: 'Memory search query',
        placeHolder: 'e.g. orchestration, council, mcp, memory',
        ignoreFocusOut: true,
    });

    const route = query?.trim()
        ? `${DASHBOARD_ROUTES.memory}?q=${encodeURIComponent(query.trim())}`
        : DASHBOARD_ROUTES.memory;
    await openDashboardRoute(route, 'Opened memory dashboard');
}

async function listTools() {
    await openDashboardRoute(DASHBOARD_ROUTES.tools, 'Opened tools dashboard');
}

async function invokeTool() {
    const name = await vscode.window.showInputBox({
        prompt: 'Tool name',
        placeHolder: 'Enter the TormentNexus tool name to invoke',
        ignoreFocusOut: true,
    });
    if (!name?.trim()) {
        return;
    }

    const argsText = await vscode.window.showInputBox({
        prompt: 'Tool arguments as JSON (optional)',
        value: '{}',
        ignoreFocusOut: true,
    });
    if (argsText === undefined) {
        return;
    }

    let args: Record<string, unknown> = {};
    try {
        args = argsText.trim() ? JSON.parse(argsText) as Record<string, unknown> : {};
    } catch {
        void vscode.window.showErrorMessage('Tool arguments must be valid JSON.');
        return;
    }

    const response = await postCoreJson<{ result: { data: unknown } }>('/tool/execute', {
        name: name.trim(),
        args,
    });

    outputChannel.show(true);
    outputChannel.appendLine(`\n=== Tool Invocation: ${name.trim()} ===\n${formatResult(response.result.data)}\n`);
    addActivity('tool', 'Invoked tool', name.trim());
    void vscode.window.showInformationMessage(`Tool ${name.trim()} executed. See TormentNexus Bridge output for details.`);
}

async function showLogs() {
    outputChannel.show(true);
    await openDashboardRoute(DASHBOARD_ROUTES.logs, 'Opened logs dashboard');
}

async function startDebate() {
    const topic = await vscode.window.showInputBox({
        prompt: 'Debate topic (optional)',
        placeHolder: 'Mission plan, architecture choice, tool policy, etc.',
        ignoreFocusOut: true,
    });

    const route = topic?.trim()
        ? `${DASHBOARD_ROUTES.debate}?topic=${encodeURIComponent(topic.trim())}`
        : DASHBOARD_ROUTES.debate;
    await openDashboardRoute(route, 'Opened council debate dashboard');
}

async function viewAnalytics() {
    await openDashboardRoute(DASHBOARD_ROUTES.analytics, 'Opened analytics dashboard');
}

async function listDebateTemplates() {
    await openDashboardRoute(DASHBOARD_ROUTES.templates, 'Opened debate templates dashboard');
}

async function architectMode() {
    const task = await vscode.window.showInputBox({
        prompt: 'Architect mode task',
        placeHolder: 'Describe the feature or architecture problem to analyze',
        ignoreFocusOut: true,
    });
    if (!task?.trim()) {
        await openDashboardRoute(DASHBOARD_ROUTES.architecture, 'Opened architecture dashboard');
        return;
    }

    const researchPrompt = `Architecture research brief: ${task.trim()}`;
    const codePrompt = `Architect mode implementation plan. First propose architecture options, then recommend one, then outline implementation steps for: ${task.trim()}`;

    const research = await dispatchResearchTask(researchPrompt, 2);
    const code = await dispatchCodeTask(codePrompt);

    outputChannel.show(true);
    outputChannel.appendLine(`\n=== Architect Mode Research ===\n${formatResult(research)}\n`);
    outputChannel.appendLine(`\n=== Architect Mode Plan ===\n${formatResult(code)}\n`);
    addActivity('code', 'Architect mode completed', task.trim());
    void vscode.window.showInformationMessage('Architect mode completed. See TormentNexus Bridge output for research and plan details.');
}

async function handleMessage(msg: Record<string, unknown>) {
    if (msg.type === 'TORMENTNEXUS_CORE_MANIFEST') {
        const manifest = msg.manifest as { connectedClients?: Array<{ clientName?: string }>; supportedHookPhases?: string[] } | undefined;
        addActivity(
            'system',
            'Core bridge manifest received',
            `${manifest?.connectedClients?.length ?? 0} registered clients · ${manifest?.supportedHookPhases?.length ?? 0} hook phases advertised.`,
        );
        return;
    }

    if (msg.type === 'GET_USER_ACTIVITY') {
        sendResponse(String(msg.requestId ?? ''), {
            lastActivityTime,
            isIdle: (Date.now() - lastActivityTime) > 5000,
        });
        return;
    }

    if (msg.type === 'VSCODE_COMMAND') {
        const command = String(msg.command ?? '');
        const args = Array.isArray(msg.args) ? msg.args : [];
        try {
            const result = await vscode.commands.executeCommand(command, ...args);
            sendResponse(String(msg.requestId ?? ''), { success: true, result });
        } catch (error) {
            sendResponse(String(msg.requestId ?? ''), {
                success: false,
                error: error instanceof Error ? error.message : String(error),
            });
        }
        return;
    }

    if (msg.type === 'GET_STATUS') {
        sendResponse(String(msg.requestId ?? ''), {
            status: {
                activeEditor: vscode.window.activeTextEditor?.document.fileName ?? null,
                activeTerminal: vscode.window.activeTerminal?.name ?? null,
                workspace: vscode.workspace.workspaceFolders?.map((folder) => folder.uri.fsPath) ?? [],
            },
        });
        return;
    }

    if (msg.type === 'GET_SELECTION') {
        const editor = vscode.window.activeTextEditor;
        sendResponse(String(msg.requestId ?? ''), {
            content: editor?.document.getText(editor.selection) ?? '',
        });
        return;
    }

    if (msg.type === 'GET_CHAT_HISTORY') {
        const history = getChatHistoryLines();
        addActivity('system', 'Chat history requested', `${history.length} entries returned to Core.`);
        sendResponse(String(msg.requestId ?? ''), {
            history,
        });
        return;
    }

    if (msg.type === 'GET_TERMINAL') {
        const terminal = vscode.window.activeTerminal;
        if (!terminal) {
            sendResponse(String(msg.requestId ?? ''), { content: 'No active terminal.' });
            return;
        }

        const terminalId = getTerminalId(terminal);
        const captured = terminalBuffers.get(terminalId) ?? '';
        sendResponse(String(msg.requestId ?? ''), {
            content: captured.trim().length > 0 ? captured : `[No terminal output captured yet for ${terminal.name}]`,
            terminalName: terminal.name,
            terminalId,
        });
        return;
    }

    if (msg.type === 'PASTE_INTO_CHAT' || msg.type === 'SUBMIT_CHAT') {
        try {
            if ((Date.now() - lastActivityTime) < 2000 && !ignoreNextActivity) {
                log('[ABORT] User active within 2s, aborting auto-paste.');
                return;
            }

            ignoreNextActivity = true;
            setTimeout(() => {
                ignoreNextActivity = false;
            }, 1500);

            const text = typeof msg.text === 'string' ? msg.text : '';
            if (text) {
                addChatHistory('user', 'chat-bridge', `Injected into VS Code chat: ${summarizeText(text, 700)}`);
                await vscode.env.clipboard.writeText(text);
                await vscode.commands.executeCommand('workbench.action.chat.open');
                await new Promise((resolve) => setTimeout(resolve, 300));
                await vscode.commands.executeCommand('workbench.action.chat.focusInput');
                await new Promise((resolve) => setTimeout(resolve, 200));
                await vscode.commands.executeCommand('editor.action.clipboardPasteAction');
            }

            if (msg.submit === true || msg.type === 'SUBMIT_CHAT') {
                addChatHistory('system', 'chat-bridge', 'Triggered VS Code chat submission.');
                await new Promise((resolve) => setTimeout(resolve, 800));
                for (const command of [
                    'workbench.action.chat.submit',
                    'workbench.action.chat.send',
                    'interactive.acceptChanges',
                    'workbench.action.terminal.chat.accept',
                    'inlineChat.accept',
                ]) {
                    try {
                        await vscode.commands.executeCommand(command);
                    } catch {
                        // Ignore unsupported commands.
                    }
                }
            }
        } catch (error) {
            const message = `Failed chat action: ${error instanceof Error ? error.message : String(error)}`;
            emitInspectorLog('error', message);
            log(message);
            ignoreNextActivity = false;
        }
    }
}

function sendResponse(requestId: string, payload: Record<string, unknown>) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
            type: 'RESPONSE',
            requestId,
            ...payload,
        }));
    }
}
