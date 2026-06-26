import React from 'react';

interface SystemStatusProps {
    status: any;
}

export function SystemStatus({ status }: SystemStatusProps) {
    if (!status) return null;

    return (
        <div className="bg-gray-800 p-4 rounded-lg mb-6 border border-gray-700">
            <h2 className="text-lg font-semibold mb-3 flex items-center">
                <span className={`w-3 h-3 rounded-full mr-2 ${status.status === 'operational' ? 'bg-green-500' : 'bg-red-500'}`}></span>
                System Status
            </h2>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <div className="bg-gray-700 p-3 rounded">
                    <div className="text-gray-400 text-sm">Uptime</div>
                    <div className="text-xl font-mono">{Math.floor(status.uptime)}s</div>
                </div>
                <div className="bg-gray-700 p-3 rounded">
                    <div className="text-gray-400 text-sm">Active Clients</div>
                    <div className="text-xl">{status.connections?.total || 0}</div>
                    <div className="text-xs text-gray-500 mt-1">
                        VSCode: {status.connections?.vscode} | Browser: {status.connections?.browser}
                    </div>
                </div>
                <div className="bg-gray-700 p-3 rounded">
                    <div className="text-gray-400 text-sm">MCP Servers</div>
                    <div className="text-xl">{status.mcp?.running} / {status.mcp?.total}</div>
                </div>
                <div className="bg-gray-700 p-3 rounded">
                    <div className="text-gray-400 text-sm">Tokens (Est.)</div>
                    <div className="text-xl">
                        {status.usage ? (status.usage.tokensIn + status.usage.tokensOut).toLocaleString() : 0}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                        In: {status.usage?.tokensIn || 0} | Out: {status.usage?.tokensOut || 0}
                    </div>
                </div>
                 <div className="bg-gray-700 p-3 rounded">
                    <div className="text-gray-400 text-sm">Cost (Est.)</div>
                    <div className="text-xl text-green-400">
                        ${status.usage?.cost.toFixed(4) || '0.0000'}
                    </div>
                </div>
            </div>
        </div>
    );
}
