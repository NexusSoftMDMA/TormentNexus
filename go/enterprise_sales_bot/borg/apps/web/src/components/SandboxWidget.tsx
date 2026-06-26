"use client";

import React, { useState } from 'react';
import { trpc } from '../utils/trpc';

export function SandboxWidget() {
    const [language, setLanguage] = useState<'python' | 'node'>('python');
    const [code, setCode] = useState('print("Hello from Secure Sandbox")');
    const [output, setOutput] = useState('');
    const [isError, setIsError] = useState(false);

    // @ts-ignore
    const executeMutation = (trpc as any).sandbox.execute.useMutation({
        onSuccess: (data: any) => {
            if (data.error) {
                setOutput(data.error);
                setIsError(true);
            } else {
                setOutput(data.output);
                setIsError(false);
            }
        },
        onError: (err: any) => {
            setOutput(err.message);
            setIsError(true);
        }
    });

    const handleRun = () => {
        setOutput('Running...');
        setIsError(false);
        executeMutation.mutate({ language, code });
    };

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 flex flex-col h-full">
            <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-mono font-semibold text-green-400">📦 Secure Sandbox</h3>
                <select
                    value={language}
                    onChange={(e) => {
                        setLanguage(e.target.value as 'python' | 'node');
                        setCode(e.target.value === 'python' ? 'print("Hello from Python")' : 'console.log("Hello from Node")');
                    }}
                    className="bg-gray-800 border-gray-700 text-xs rounded p-1"
                >
                    <option value="python">Python 3.10</option>
                    <option value="node">Node.js 18</option>
                </select>
                <div className="flex gap-2">
                    <button
                        onClick={handleRun}
                        disabled={executeMutation.isLoading}
                        className="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm disabled:opacity-50"
                    >
                        {executeMutation.isLoading ? 'Running...' : 'Run Code'}
                    </button>
                    <button
                        onClick={() => setCode('')}
                        className="text-gray-400 hover:text-white px-3 py-1 text-sm"
                    >
                        Clear
                    </button>
                </div>
            </div>

            {isError && (
                <div className="bg-red-900/50 text-red-200 p-2 text-xs font-mono mb-2">
                    Execution Error
                </div>
            )}
            <textarea
                value={code}
                onChange={(e) => setCode(e.target.value)}
                className="flex-1 bg-black/50 border border-gray-700 p-2 font-mono text-sm text-gray-300 resize-none rounded mb-2 focus:outline-none focus:border-green-500"
                spellCheck={false}
            />

            <div className="bg-black border border-gray-800 rounded p-2 h-32 overflow-auto font-mono text-xs mb-2">
                {output ? (
                    <pre className={isError ? "text-red-400" : "text-gray-300"}>{output}</pre>
                ) : (
                    <span className="text-gray-600 italic">Ready to execute.</span>
                )}
            </div>

            <button
                onClick={handleRun}
                disabled={executeMutation.isLoading}
                className="bg-green-700 hover:bg-green-600 text-white py-1 rounded text-sm font-semibold disabled:opacity-50"
            >
                {executeMutation.isLoading ? 'Executing...' : 'Run Code'}
            </button>
        </div>
    );
}
