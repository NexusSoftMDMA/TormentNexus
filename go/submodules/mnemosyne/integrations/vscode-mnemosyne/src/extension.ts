import * as vscode from 'vscode';
import { MnemosyneClient } from './client';
import { MemoriesProvider } from './provider';
import { MemoryWebview } from './webview';

let client: MnemosyneClient;
let webview: MemoryWebview;

export function activate(context: vscode.ExtensionContext) {
  client = new MnemosyneClient();
  webview = new MemoryWebview(context.extensionUri);

  // Register sidebar webview
  const provider = new MemoriesProvider(context.extensionUri, client);
  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider('mnemosyneMemories', provider)
  );

  // Commands
  context.subscriptions.push(
    vscode.commands.registerCommand('mnemosyne.refresh', async () => {
      provider.refresh();
      vscode.window.showInformationMessage('Mnemosyne memories refreshed');
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('mnemosyne.search', async () => {
      const query = await vscode.window.showInputBox({
        placeHolder: 'Search memories...',
        prompt: 'Search your Mnemosyne memories by semantic similarity or keyword'
      });
      if (query) {
        const results = await client.search(query);
        provider.showResults(results);
      }
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('mnemosyne.remember', async () => {
      const content = await vscode.window.showInputBox({
        placeHolder: 'What do you want to remember?',
        prompt: 'Store a new memory in Mnemosyne',
        validateInput: (text: string) => text.trim() ? null : 'Content cannot be empty'
      });
      if (content) {
        await client.remember(content);
        vscode.window.showInformationMessage('Memory stored ✓');
        provider.refresh();
      }
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('mnemosyne.openMcpConfig', async () => {
      const configUri = vscode.Uri.joinPath(
        vscode.workspace.workspaceFolders?.[0]?.uri || vscode.Uri.file('.'),
        '.cursor/mcp.json'
      );
      const doc = await vscode.workspace.openTextDocument(configUri);
      vscode.window.showTextDocument(doc);
    })
  );

  // Status bar
  const statusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Left,
    100
  );
  statusBar.text = '$(database) Mnemosyne';
  statusBar.tooltip = 'Click to view memory stats';
  statusBar.command = 'mnemosyne.refresh';
  statusBar.show();
  context.subscriptions.push(statusBar);

  // Auto-refresh on save
  if (vscode.workspace.getConfiguration('mnemosyne').get('autoRefresh')) {
    context.subscriptions.push(
      vscode.workspace.onDidSaveTextDocument(() => provider.refresh())
    );
  }

  console.log('Mnemosyne extension activated');
}

export function deactivate() {
  console.log('Mnemosyne extension deactivated');
}
