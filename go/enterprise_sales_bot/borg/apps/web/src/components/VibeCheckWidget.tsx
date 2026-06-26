
"use client";
import { trpc } from "../utils/trpc";
import { useState } from "react";

export function VibeCheckWidget() {
    const [code, setCode] = useState("// TODO: Implement this\nfunction foo() {\n  return 'bar';\n}");
    const [tool, setTool] = useState("vibe_check");
    const [result, setResult] = useState("");

    const executeMutation = trpc.executeTool.useMutation({
        onSuccess: (data) => setResult(typeof data === 'string' ? data : JSON.stringify(data, null, 2)),
        onError: (err) => setResult(`Error: ${err.message}`)
    });

    const handleExecute = () => {
        setResult("Analyzing vibe...");
        executeMutation.mutate({
            name: tool,
            args: { code }
        });
    };

    return (
        <div className="bg-[#1e1e1e] border border-[#333] rounded-lg p-6 shadow-xl w-full mt-6">
            <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    ✨ Vibe Check (Native)
                </h2>
                <div className="text-xs text-zinc-500 uppercase font-bold tracking-wider">
                    vibe-coder-mcp
                </div>
            </div>

            <div className="flex flex-col gap-2 mb-4">
                <label className="text-xs font-bold text-zinc-400 uppercase">Code Snippet</label>
                <textarea
                    rows={5}
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-zinc-200 font-mono text-sm focus:outline-none focus:border-blue-500"
                    placeholder="Enter code to check for vibe issues..."
                />
            </div>

            <div className="flex flex-col gap-2 mb-4">
                <label className="text-xs font-bold text-zinc-400 uppercase">Tool / Alias</label>
                <select
                    value={tool}
                    onChange={(e) => setTool(e.target.value)}
                    className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-zinc-200 focus:outline-none focus:border-blue-500"
                >
                    <option value="vibe_check">Vibe Check Analyze (Standard)</option>
                    <option value="vibe_quick">Vibe Quick Check</option>
                </select>
            </div>

            <button
                onClick={handleExecute}
                disabled={executeMutation.isPending}
                className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-bold rounded disabled:opacity-50 transition-colors mb-4"
            >
                {executeMutation.isPending ? "Analyzing..." : "Execute Vibe Check"}
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
