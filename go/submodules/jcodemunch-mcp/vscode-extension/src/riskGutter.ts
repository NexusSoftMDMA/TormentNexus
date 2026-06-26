// Risk-density gutter — paints colored bars in the editor gutter at each
// risky function/method's start line. Composite risk comes from
// `jcodemunch-mcp file-risk <abs>`, which runs the server-side
// `get_file_risk` tool.
//
// Design properties:
//   - Green = invisible (no gutter mark) — keeps signal-to-noise high.
//   - Yellow / orange / red carry the actual signal.
//   - Refresh on file open + on save. Typing does not refresh; cyclomatic
//     doesn't move with whitespace edits, so the cost isn't worth it.
//   - Hover at a decorated line shows the per-axis breakdown.
//
// Single-file scope: each editor tracks its own decorations + risk map.

import * as vscode from "vscode";
import { spawn } from "child_process";

interface AxisScores {
    complexity: number;
    exposure: number;
    churn: number;
    test_gap: number;
}

interface RiskInfo {
    composite: number;
    level: "green" | "yellow" | "orange" | "red";
    axes: AxisScores;
}

interface SymbolRisk {
    symbol_id: string;
    name: string;
    kind: string;
    line: number;
    end_line: number;
    cyclomatic: number;
    risk: RiskInfo;
}

interface FileMetrics {
    incoming_files: number;
    churn_30d: number;
    has_tests: boolean;
}

interface FileRiskResponse {
    file: string;
    language: string;
    file_metrics: FileMetrics;
    symbols: SymbolRisk[];
    error?: string;
}

// Per-level decoration types (created once, reused).
const decorationTypes: Map<string, vscode.TextEditorDecorationType> = new Map();

// Per-document risk maps so the hover provider can answer fast.
// Key: document URI string. Value: line (1-based) -> SymbolRisk.
const riskByDoc: Map<string, Map<number, SymbolRisk>> = new Map();

// Per-document refresh debouncers so rapid save sequences collapse.
const refreshTimers: Map<string, NodeJS.Timeout> = new Map();

let outputChannel: vscode.OutputChannel | undefined;

function getChannel(): vscode.OutputChannel {
    if (!outputChannel) {
        outputChannel = vscode.window.createOutputChannel("jCodeMunch Risk Gutter");
    }
    return outputChannel;
}

function svgDot(fill: string): string {
    // 16x16 SVG, vertically centered solid circle. The icon shows up in
    // the gutter at the symbol's line.
    const svg =
        `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16">` +
        `<circle cx="8" cy="8" r="4" fill="${fill}"/>` +
        `</svg>`;
    return "data:image/svg+xml;base64," + Buffer.from(svg, "utf8").toString("base64");
}

function getDecorationType(level: string): vscode.TextEditorDecorationType | undefined {
    if (decorationTypes.has(level)) {
        return decorationTypes.get(level);
    }
    let fill = "";
    let overview = "";
    switch (level) {
        case "yellow": fill = "#e5c07b"; overview = "rgba(229,192,123,0.6)"; break;
        case "orange": fill = "#d19a66"; overview = "rgba(209,154,102,0.6)"; break;
        case "red":    fill = "#e06c75"; overview = "rgba(224,108,117,0.6)"; break;
        default: return undefined;  // green — no decoration
    }
    const dt = vscode.window.createTextEditorDecorationType({
        gutterIconPath: vscode.Uri.parse(svgDot(fill)),
        gutterIconSize: "contain",
        overviewRulerColor: overview,
        overviewRulerLane: vscode.OverviewRulerLane.Right,
    });
    decorationTypes.set(level, dt);
    return dt;
}

async function fetchFileRisk(filePath: string): Promise<FileRiskResponse | null> {
    const cfg = vscode.workspace.getConfiguration("jcodemunch.riskGutter");
    if (!cfg.get<boolean>("enabled", true)) return null;

    const cmdCfg = vscode.workspace.getConfiguration("jcodemunch.indexOnSave");
    const cmd = cmdCfg.get<string>("command", "jcodemunch-mcp");
    const ch = getChannel();

    return new Promise<FileRiskResponse | null>((resolve) => {
        const child = spawn(cmd, ["file-risk", filePath], {
            stdio: ["ignore", "pipe", "pipe"],
            windowsHide: true,
        });
        let stdout = "";
        let stderr = "";
        child.stdout?.on("data", (d) => { stdout += d.toString(); });
        child.stderr?.on("data", (d) => { stderr += d.toString(); });
        child.on("error", (err) => {
            ch.appendLine(`[error] ${cmd} file-risk failed: ${err.message}`);
            resolve(null);
        });
        child.on("exit", (code) => {
            if (code !== 0) {
                ch.appendLine(`[exit ${code}] file-risk ${filePath}: ${stderr.trim()}`);
                resolve(null);
                return;
            }
            try {
                resolve(JSON.parse(stdout));
            } catch (e) {
                ch.appendLine(`[parse] file-risk JSON: ${(e as Error).message}`);
                resolve(null);
            }
        });
    });
}

function clearDecorations(editor: vscode.TextEditor) {
    for (const dt of decorationTypes.values()) {
        editor.setDecorations(dt, []);
    }
}

async function refreshEditor(editor: vscode.TextEditor) {
    if (editor.document.uri.scheme !== "file") return;
    const docKey = editor.document.uri.toString();
    const filePath = editor.document.uri.fsPath;

    const result = await fetchFileRisk(filePath);
    if (!result || result.error) {
        // No data — clear any prior decorations + risk map.
        clearDecorations(editor);
        riskByDoc.delete(docKey);
        return;
    }

    // Build per-line lookup for the hover provider.
    const lineMap: Map<number, SymbolRisk> = new Map();
    const buckets: Record<string, vscode.Range[]> = { yellow: [], orange: [], red: [] };
    for (const sym of result.symbols) {
        if (sym.risk.level === "green") continue;
        // VS Code line numbers are 0-based; the index records 1-based.
        const startLine = Math.max(0, sym.line - 1);
        const range = new vscode.Range(startLine, 0, startLine, 0);
        if (sym.risk.level in buckets) {
            buckets[sym.risk.level].push(range);
        }
        lineMap.set(sym.line, sym);
    }
    riskByDoc.set(docKey, lineMap);

    // Apply decorations bucket by bucket; each level has its own type.
    for (const level of ["yellow", "orange", "red"]) {
        const dt = getDecorationType(level);
        if (!dt) continue;
        editor.setDecorations(dt, buckets[level]);
    }
}

function scheduleRefresh(editor: vscode.TextEditor, debounceMs: number) {
    const docKey = editor.document.uri.toString();
    const existing = refreshTimers.get(docKey);
    if (existing) clearTimeout(existing);
    const timer = setTimeout(() => {
        refreshTimers.delete(docKey);
        refreshEditor(editor).catch((e) => {
            getChannel().appendLine(`[error] refresh: ${(e as Error).message}`);
        });
    }, debounceMs);
    refreshTimers.set(docKey, timer);
}

export function buildHoverMarkdown(sym: SymbolRisk, fileMetrics: FileMetrics): vscode.MarkdownString {
    const md = new vscode.MarkdownString();
    md.supportThemeIcons = true;
    md.isTrusted = false;
    const emoji = sym.risk.level === "red" ? "🔴" :
                  sym.risk.level === "orange" ? "🟠" :
                  sym.risk.level === "yellow" ? "🟡" : "🟢";
    md.appendMarkdown(`### ${emoji} jCodemunch risk — \`${sym.name}\`\n\n`);
    md.appendMarkdown(`**Composite:** ${sym.risk.composite.toFixed(1)}/100 (${sym.risk.level})\n\n`);
    md.appendMarkdown("| Axis | Score | Signal |\n|---|---:|---|\n");
    md.appendMarkdown(`| complexity | ${sym.risk.axes.complexity.toFixed(0)} | cyclomatic ${sym.cyclomatic} |\n`);
    md.appendMarkdown(`| exposure | ${sym.risk.axes.exposure.toFixed(0)} | ${fileMetrics.incoming_files} importing files |\n`);
    md.appendMarkdown(`| churn | ${sym.risk.axes.churn.toFixed(0)} | ${fileMetrics.churn_30d} commits in 30 days |\n`);
    md.appendMarkdown(`| test_gap | ${sym.risk.axes.test_gap.toFixed(0)} | tests reference this module: ${fileMetrics.has_tests ? "yes" : "no"} |\n\n`);
    md.appendMarkdown(`_Drill in: \`get_call_hierarchy\`, \`get_pr_risk_profile\` for \`${sym.symbol_id}\`._`);
    return md;
}

class RiskHoverProvider implements vscode.HoverProvider {
    // Keep a per-document file_metrics cache populated by refreshEditor.
    private fileMetricsByDoc: Map<string, FileMetrics> = new Map();

    setFileMetrics(docKey: string, m: FileMetrics) {
        this.fileMetricsByDoc.set(docKey, m);
    }

    provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
    ): vscode.ProviderResult<vscode.Hover> {
        const lineMap = riskByDoc.get(document.uri.toString());
        if (!lineMap) return undefined;
        // VS Code position.line is 0-based; index lines are 1-based.
        const line1 = position.line + 1;
        const sym = lineMap.get(line1);
        if (!sym) return undefined;
        const metrics = this.fileMetricsByDoc.get(document.uri.toString());
        if (!metrics) return undefined;
        return new vscode.Hover(buildHoverMarkdown(sym, metrics), new vscode.Range(position.line, 0, position.line, 0));
    }
}

let hoverProvider: RiskHoverProvider | undefined;

export function activateRiskGutter(context: vscode.ExtensionContext) {
    const ch = getChannel();
    ch.appendLine("jCodeMunch risk gutter active.");

    hoverProvider = new RiskHoverProvider();
    context.subscriptions.push(
        vscode.languages.registerHoverProvider({ scheme: "file" }, hoverProvider),
    );

    // Wrapped refresh that also updates the hover provider's file_metrics.
    const refreshAndCache = async (editor: vscode.TextEditor) => {
        if (editor.document.uri.scheme !== "file") return;
        const result = await fetchFileRisk(editor.document.uri.fsPath);
        if (!result || result.error) return;
        if (hoverProvider) hoverProvider.setFileMetrics(editor.document.uri.toString(), result.file_metrics);
        await refreshEditor(editor);
    };

    // On open
    context.subscriptions.push(
        vscode.window.onDidChangeActiveTextEditor((editor) => {
            if (editor) refreshAndCache(editor).catch(() => undefined);
        }),
    );

    // On save (debounced — same file might be saved several times in quick succession).
    context.subscriptions.push(
        vscode.workspace.onDidSaveTextDocument((doc) => {
            if (doc.uri.scheme !== "file") return;
            const editor = vscode.window.visibleTextEditors.find((e) => e.document === doc);
            if (!editor) return;
            const cfg = vscode.workspace.getConfiguration("jcodemunch.riskGutter");
            const debounceMs = cfg.get<number>("debounceMs", 600);
            scheduleRefresh(editor, debounceMs);
            // Also pre-warm the hover cache.
            fetchFileRisk(doc.uri.fsPath).then((r) => {
                if (r && hoverProvider) {
                    hoverProvider.setFileMetrics(doc.uri.toString(), r.file_metrics);
                }
            }).catch(() => undefined);
        }),
    );

    // Initial pass for any editors already open at activation.
    for (const editor of vscode.window.visibleTextEditors) {
        refreshAndCache(editor).catch(() => undefined);
    }
}

export function deactivateRiskGutter() {
    for (const t of refreshTimers.values()) clearTimeout(t);
    refreshTimers.clear();
    for (const dt of decorationTypes.values()) dt.dispose();
    decorationTypes.clear();
    riskByDoc.clear();
}
