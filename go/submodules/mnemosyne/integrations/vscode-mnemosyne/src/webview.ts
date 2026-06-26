import * as vscode from 'vscode';

/**
 * Placeholder for future memory detail webview.
 */
export class MemoryWebview {
  constructor(private _extensionUri: vscode.Uri) {}

  show(memoryId: string): void {
    const panel = vscode.window.createWebviewPanel(
      'mnemosyneDetail',
      'Memory Detail',
      vscode.ViewColumn.One,
      { enableScripts: true }
    );
    panel.webview.html = `<html><body><h2>Memory ${memoryId}</h2></body></html>`;
  }
}
