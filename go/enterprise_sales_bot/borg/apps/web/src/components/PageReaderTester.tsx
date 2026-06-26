
"use client";
import { trpc } from "../utils/trpc";
import { useState } from "react";

export function PageReaderTester() {
    const [url, setUrl] = useState("https://example.com");
    const [result, setResult] = useState("");
    const executeMutation = trpc.executeTool.useMutation({
        onSuccess: (data) => setResult(data),
        onError: (err) => setResult(`Error: ${err.message}`)
    });

    const handleRead = () => {
        setResult("Reading...");
        executeMutation.mutate({
            name: "read_page",
            args: { url }
        });
    };

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 shadow-xl w-full">
            <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    🌐 Page Reader
                </h2>
                <div className="text-xs text-zinc-500 uppercase font-bold tracking-wider">
                    Scraper Tool
                </div>
            </div>

            <div className="flex gap-2 mb-4">
                <input
                    type="text"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    className="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-blue-500 transition-colors"
                    placeholder="https://..."
                />
                <button
                    onClick={handleRead}
                    disabled={executeMutation.isPending}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-bold rounded disabled:opacity-50 transition-colors"
                >
                    {executeMutation.isPending ? "Reading..." : "Read URL"}
                </button>
            </div>

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
