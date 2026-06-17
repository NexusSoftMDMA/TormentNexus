
"use client";
import { trpc } from "../utils/trpc";
import { useState } from "react";

export function BrowserToolWidget() {
    const [url, setUrl] = useState("https://google.com");
    const [tool, setTool] = useState("browser_navigate");
    const [result, setResult] = useState("");

    const executeMutation = trpc.executeTool.useMutation({
        onSuccess: (data) => setResult(typeof data === 'string' ? data : JSON.stringify(data, null, 2)),
        onError: (err) => setResult(`Error: ${err.message}`)
    });

    const handleExecute = () => {
        setResult("Executing...");
        executeMutation.mutate({
            name: tool,
            args: { url, headless: true }
        });
    };

    return (
        <div className="bg-[#1e1e1e] border border-[#333] rounded-lg p-6 shadow-xl w-full mt-6">
            <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    🌐 Browser Automation (Native)
                </h2>
                <div className="text-xs text-zinc-500 uppercase font-bold tracking-wider">
                    Playwright / Browser-use / Browserbase
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                <div className="flex flex-col gap-2">
                    <label className="text-xs font-bold text-zinc-400 uppercase">Target URL</label>
                    <input
                        type="text"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                        className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-zinc-200 focus:outline-none focus:border-blue-500"
                        placeholder="https://..."
                    />
                </div>
                <div className="flex flex-col gap-2">
                    <label className="text-xs font-bold text-zinc-400 uppercase">Tool / Alias</label>
                    <select
                        value={tool}
                        onChange={(e) => setTool(e.target.value)}
                        className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-zinc-200 focus:outline-none focus:border-blue-500"
                    >
                        <option value="browser_navigate">Browser Navigate (Standard)</option>
                        <option value="browser_screenshot">Browser Screenshot</option>
                        <option value="browser_use_navigate">Browser-use Navigate</option>
                        <option value="browsermcp_navigate">BrowserMCP Navigate</option>
                        <option value="browserbase_navigate">Browserbase Navigate</option>
                        <option value="browserbase_screenshot">Browserbase Screenshot</option>
                    </select>
                </div>
            </div>

            <button
                onClick={handleExecute}
                disabled={executeMutation.isPending}
                className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-bold rounded disabled:opacity-50 transition-colors mb-4"
            >
                {executeMutation.isPending ? "Executing..." : "Execute Tool"}
            </button>

            {(result || executeMutation.isPending) && (
                <div className="bg-black/50 border border-zinc-800 rounded-lg p-4 max-h-[300px] overflow-auto">
                    <pre className="text-xs font-mono text-zinc-300 whitespace-pre-wrap break-all">
                        {result}
                    </pre>
                </div>
            )}
        </div>
    );
}
