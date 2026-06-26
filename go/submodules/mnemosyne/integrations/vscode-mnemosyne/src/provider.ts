import * as vscode from 'vscode';
import { MnemosyneClient, Memory, Stats } from './client';

export class MemoriesProvider implements vscode.WebviewViewProvider {
  private _view?: vscode.WebviewView;
  private _results: Memory[] | null = null;

  constructor(
    private readonly _extensionUri: vscode.Uri,
    private readonly _client: MnemosyneClient
  ) {}

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ): void {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this._extensionUri]
    };

    webviewView.webview.html = this._getHtml();
    this._refresh();

    // Handle messages from webview
    webviewView.webview.onDidReceiveMessage(async (message) => {
      switch (message.type) {
        case 'refresh':
          this._refresh();
          break;
        case 'search':
          if (message.query) {
            const results = await this._client.search(message.query);
            this._renderResults(results);
          } else {
            this._results = null;
            this._refresh();
          }
          break;
        case 'remember':
          vscode.commands.executeCommand('mnemosyne.remember');
          break;
        case 'openDetail':
          // Could open a detail view, for now just show info
          vscode.window.showInformationMessage(
            `Memory: ${(message.content || '').substring(0, 100)}...`
          );
          break;
      }
    });
  }

  refresh(): void {
    this._results = null;
    this._refresh();
  }

  showResults(results: Memory[]): void {
    this._results = results;
    this._renderResults(results);
  }

  private async _refresh(): Promise<void> {
    if (!this._view) return;

    const [stats, memories] = await Promise.all([
      this._client.getStats(),
      this._results || this._client.listRecent(20)
    ]);

    this._view.webview.html = this._getHtml(stats, memories);
  }

  private _renderResults(memories: Memory[]): void {
    if (!this._view) return;
    this._view.webview.html = this._getHtml(undefined, memories);
  }

  private _getHtml(stats?: Stats, memories?: Memory[]): string {
    const statsHtml = stats
      ? Object.entries(stats.tiers || {})
          .map(([tier, count]) => `<span class="stat">${count} ${tier.replace('_', ' ')}</span>`)
          .join('') + `<span class="stat">${stats.total} total</span>`
      : '<span class="stat">Loading...</span>';

    const memHtml = memories && memories.length > 0
      ? memories.map(m => `
        <div class="memory" onclick="vscode.postMessage({type:'openDetail',content:${JSON.stringify(m.content)}})">
          <div class="meta">
            <span class="tag ${m.tier}">${m.tier.replace('_', ' ')}</span>
            <span class="src">${m.source}</span>
            <span class="imp">${'★'.repeat(Math.max(1, Math.round(m.importance * 5)))}</span>
          </div>
          <div class="content">${this._escape(m.content)}</div>
          <div class="ts">${(m.timestamp || '').substring(0, 19)}</div>
        </div>`).join('')
      : '<div class="empty">No memories found. Use the search bar above or store your first memory.</div>';

    return `<!DOCTYPE html>
<html>
<head>
<style>
  :root {
    --bg: var(--vscode-sideBar-background);
    --text: var(--vscode-sideBar-foreground);
    --border: var(--vscode-sideBar-border);
    --accent: var(--vscode-textLink-foreground);
    --input-bg: var(--vscode-input-background);
    --input-border: var(--vscode-input-border);
    --btn-bg: var(--vscode-button-background);
    --btn-fg: var(--vscode-button-foreground);
  }
  body { margin: 0; padding: 8px; font-family: var(--vscode-font-family); font-size: 13px; color: var(--text); }
  .toolbar { display: flex; gap: 4px; margin-bottom: 8px; }
  .toolbar input { flex: 1; background: var(--input-bg); border: 1px solid var(--input-border); color: var(--text); padding: 4px 8px; border-radius: 2px; }
  .toolbar button { background: var(--btn-bg); color: var(--btn-fg); border: none; padding: 4px 8px; cursor: pointer; border-radius: 2px; font-size: 12px; }
  .stats { display: flex; gap: 12px; margin-bottom: 8px; font-size: 11px; flex-wrap: wrap; }
  .stat { opacity: 0.7; }
  .memory { padding: 8px; border: 1px solid var(--border); border-radius: 4px; margin-bottom: 6px; cursor: pointer; }
  .memory:hover { border-color: var(--accent); }
  .meta { display: flex; gap: 8px; font-size: 11px; margin-bottom: 4px; }
  .tag { padding: 1px 6px; border-radius: 3px; font-weight: 600; font-size: 10px; text-transform: uppercase; }
  .tag.working { background: #e3f2fd; color: #1565c0; }
  .tag.episodic { background: #f3e5f5; color: #7b1fa2; }
  .tag.memories { background: #e8f5e9; color: #2e7d32; }
  .src { opacity: 0.6; }
  .imp { opacity: 0.8; color: #f9a825; }
  .content { line-height: 1.4; word-break: break-word; }
  .ts { font-size: 10px; opacity: 0.4; margin-top: 4px; }
  .empty { text-align: center; padding: 20px; opacity: 0.5; }
  .actions { display: flex; gap: 4px; margin-top: 8px; }
  .actions button { background: transparent; border: 1px solid var(--border); color: var(--text); padding: 2px 8px; border-radius: 2px; cursor: pointer; font-size: 11px; }
  .actions button:hover { border-color: var(--accent); }
</style>
</head>
<body>
  <div class="toolbar">
    <input type="text" id="searchInput" placeholder="Search memories..." onkeydown="if(event.key==='Enter')search()">
    <button onclick="search()">Search</button>
    <button onclick="vscode.postMessage({type:'remember'})" title="Store a memory">+</button>
  </div>
  <div class="stats">${statsHtml}</div>
  <div id="memories">${memHtml}</div>
  <div class="actions">
    <button onclick="vscode.postMessage({type:'refresh'})">Refresh</button>
  </div>
<script>
const vscode = acquireVsCodeApi();
function search() {
  const q = document.getElementById('searchInput').value;
  vscode.postMessage({type:'search', query: q});
}
</script>
</body>
</html>`;
  }

  private _escape(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
}
