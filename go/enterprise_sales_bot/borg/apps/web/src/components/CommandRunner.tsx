
"use client";

import { useState } from 'react';
import { trpc } from '../utils/trpc';

export function CommandRunner() {
    const [command, setCommand] = useState("");
    const [output, setOutput] = useState("");

    // We invoke this manually, not automatically
    const executeMutation = trpc.commands.execute.useMutation({
        onSuccess: (data) => {
            setOutput(data.output || 'Command completed.');
        },
        onError: (err: any) => {
            setOutput(`Error: ${err.message}`);
        }
    });

    const handleRun = async () => {
        if (!command) return;
        setOutput("Running...");
        executeMutation.mutate({ input: command });
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleRun();
        }
    };

    return (
        <div className="p-6 bg-[#1e1e1e] rounded-xl border border-[#333] shadow-lg flex flex-col mt-6">
            <h2 className="text-xl font-bold text-white mb-4">Command Runner</h2>

            <div className="flex gap-2 mb-4">
                <input
                    type="text"
                    value={command}
                    onChange={(e) => setCommand(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Enter command (e.g. dir, git status)..."
                    className="flex-1 bg-[#111] text-gray-300 border border-[#444] rounded px-3 py-2 font-mono text-sm focus:outline-none focus:border-blue-500"
                />
                <button
                    onClick={handleRun}
                    disabled={executeMutation.isPending || !command}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    {executeMutation.isPending ? 'Running...' : 'Run'}
                </button>
            </div>

            <div className="bg-[#111] rounded p-4 overflow-x-auto min-h-[100px] max-h-[300px] overflow-y-auto font-mono text-sm text-gray-300 whitespace-pre-wrap">
                {output || <span className="text-gray-600 italic">Output will appear here...</span>}
            </div>
        </div>
    );
}
