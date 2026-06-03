"use client";

import { useState, useEffect, useCallback } from "react";

interface TormentNexusServer {
    uuid: string;
    name: string;
    description: string | null;
    type: string;
    command: string | null;
    args: string[] | null;
    url: string | null;
    status: string;
    enabled: boolean;
}

interface TormentNexusStatus {
    available: boolean;
    url: string;
}

export default function TormentNexusPage() {
    const [status, setStatus] = useState<TormentNexusStatus | null>(null);
    const [servers, setServers] = useState<TormentNexusServer[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // -- Add Server form state
    const [showAddForm, setShowAddForm] = useState(false);
    const [newServer, setNewServer] = useState({
        name: "",
        description: "",
        type: "STDIO" as "STDIO" | "SSE" | "STREAMABLE_HTTP",
        command: "",
        args: "",
        url: "",
    });

    const fetchData = useCallback(async () => {
        try {
            setLoading(true);
            const [statusRes, serversRes] = await Promise.all([
                fetch("/api/trpc/mcpServers.tormentnexusStatus").then((r) => r.json()),
                fetch("/api/trpc/mcpServers.listFromTormentNexus").then((r) => r.json()),
            ]);
            setStatus(statusRes?.result?.data ?? null);
            setServers(serversRes?.result?.data ?? []);
            setError(null);
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to connect");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleAddServer = async () => {
        try {
            const payload: Record<string, unknown> = {
                name: newServer.name,
                type: newServer.type,
            };
            if (newServer.description) payload.description = newServer.description;
            if (newServer.command) payload.command = newServer.command;
            if (newServer.args)
                payload.args = newServer.args.split(",").map((s) => s.trim());
            if (newServer.url) payload.url = newServer.url;

            const res = await fetch("/api/trpc/mcpServers.createInTormentNexus", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!res.ok) throw new Error("Failed to create server");

            setShowAddForm(false);
            setNewServer({
                name: "",
                description: "",
                type: "STDIO",
                command: "",
                args: "",
                url: "",
            });
            await fetchData();
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to add server");
        }
    };

    const handleDeleteServer = async (uuid: string) => {
        try {
            await fetch("/api/trpc/mcpServers.deleteFromTormentNexus", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ uuid }),
            });
            await fetchData();
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to delete server");
        }
    };

    return (
        <div className="space-y-6 p-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">
                        TormentNexus Management
                    </h1>
                    <p className="mt-1 text-sm text-gray-400">
                        Manage MCP servers via the TormentNexus backend at port 12009
                    </p>
                </div>
                <div className="flex items-center gap-3">
                    {/* Status Indicator */}
                    <div
                        className={`flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-medium ${status?.available
                                ? "bg-green-500/20 text-green-400"
                                : "bg-red-500/20 text-red-400"
                            }`}
                    >
                        <span
                            className={`h-2 w-2 rounded-full ${status?.available ? "bg-green-400 animate-pulse" : "bg-red-400"
                                }`}
                        />
                        {status?.available ? "Connected" : "Offline"}
                    </div>
                    <button
                        onClick={() => fetchData()}
                        className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 transition-colors"
                    >
                        Refresh
                    </button>
                    <button
                        onClick={() => setShowAddForm(!showAddForm)}
                        className="rounded-lg bg-purple-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-purple-500 transition-colors"
                    >
                        + Add Server
                    </button>
                </div>
            </div>

            {/* Error Banner */}
            {error && (
                <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-300">
                    {error}
                </div>
            )}

            {/* Add Server Form */}
            {showAddForm && (
                <div className="rounded-xl border border-gray-700 bg-gray-800/50 p-4 space-y-3">
                    <h3 className="text-sm font-semibold text-white">
                        Register New MCP Server
                    </h3>
                    <div className="grid grid-cols-2 gap-3">
                        <div>
                            <label className="block text-xs text-gray-400 mb-1">Name</label>
                            <input
                                type="text"
                                value={newServer.name}
                                onChange={(e) =>
                                    setNewServer({ ...newServer, name: e.target.value })
                                }
                                className="w-full rounded-lg border border-gray-600 bg-gray-700 px-3 py-1.5 text-sm text-white placeholder-gray-500"
                                placeholder="my-mcp-server"
                            />
                        </div>
                        <div>
                            <label className="block text-xs text-gray-400 mb-1">Type</label>
                            <select
                                value={newServer.type}
                                onChange={(e) =>
                                    setNewServer({
                                        ...newServer,
                                        type: e.target.value as "STDIO" | "SSE" | "STREAMABLE_HTTP",
                                    })
                                }
                                className="w-full rounded-lg border border-gray-600 bg-gray-700 px-3 py-1.5 text-sm text-white"
                            >
                                <option value="STDIO">STDIO</option>
                                <option value="SSE">SSE</option>
                                <option value="STREAMABLE_HTTP">Streamable HTTP</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-xs text-gray-400 mb-1">
                                Command
                            </label>
                            <input
                                type="text"
                                value={newServer.command}
                                onChange={(e) =>
                                    setNewServer({ ...newServer, command: e.target.value })
                                }
                                className="w-full rounded-lg border border-gray-600 bg-gray-700 px-3 py-1.5 text-sm text-white placeholder-gray-500"
                                placeholder="npx -y @modelcontextprotocol/server-git"
                            />
                        </div>
                        <div>
                            <label className="block text-xs text-gray-400 mb-1">
                                Args (comma-separated)
                            </label>
                            <input
                                type="text"
                                value={newServer.args}
                                onChange={(e) =>
                                    setNewServer({ ...newServer, args: e.target.value })
                                }
                                className="w-full rounded-lg border border-gray-600 bg-gray-700 px-3 py-1.5 text-sm text-white placeholder-gray-500"
                                placeholder="--port, 3000"
                            />
                        </div>
                        <div className="col-span-2">
                            <label className="block text-xs text-gray-400 mb-1">
                                URL (for SSE/HTTP)
                            </label>
                            <input
                                type="text"
                                value={newServer.url}
                                onChange={(e) =>
                                    setNewServer({ ...newServer, url: e.target.value })
                                }
                                className="w-full rounded-lg border border-gray-600 bg-gray-700 px-3 py-1.5 text-sm text-white placeholder-gray-500"
                                placeholder="http://localhost:3000/sse"
                            />
                        </div>
                        <div className="col-span-2">
                            <label className="block text-xs text-gray-400 mb-1">
                                Description
                            </label>
                            <input
                                type="text"
                                value={newServer.description}
                                onChange={(e) =>
                                    setNewServer({ ...newServer, description: e.target.value })
                                }
                                className="w-full rounded-lg border border-gray-600 bg-gray-700 px-3 py-1.5 text-sm text-white placeholder-gray-500"
                                placeholder="Optional description of this server"
                            />
                        </div>
                    </div>
                    <div className="flex gap-2 pt-2">
                        <button
                            onClick={handleAddServer}
                            disabled={!newServer.name}
                            className="rounded-lg bg-green-600 px-4 py-1.5 text-xs font-medium text-white hover:bg-green-500 disabled:opacity-40 transition-colors"
                        >
                            Create Server
                        </button>
                        <button
                            onClick={() => setShowAddForm(false)}
                            className="rounded-lg bg-gray-600 px-4 py-1.5 text-xs font-medium text-white hover:bg-gray-500 transition-colors"
                        >
                            Cancel
                        </button>
                    </div>
                </div>
            )}

            {/* Server List */}
            {loading ? (
                <div className="flex items-center justify-center py-12">
                    <div className="h-6 w-6 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent" />
                    <span className="ml-3 text-sm text-gray-400">
                        Loading TormentNexus servers...
                    </span>
                </div>
            ) : servers.length === 0 ? (
                <div className="rounded-xl border border-gray-700 bg-gray-800/30 p-8 text-center">
                    <p className="text-gray-400">
                        {status?.available
                            ? "No MCP servers registered in TormentNexus yet. Click \"+ Add Server\" to register one."
                            : "TormentNexus backend is offline. Start it at http://localhost:12009 to manage servers."}
                    </p>
                </div>
            ) : (
                <div className="grid gap-3">
                    {servers.map((server) => (
                        <div
                            key={server.uuid}
                            className="rounded-xl border border-gray-700 bg-gray-800/50 p-4 hover:border-indigo-500/50 transition-colors"
                        >
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-3">
                                    {/* Status dot */}
                                    <span
                                        className={`h-2.5 w-2.5 rounded-full ${server.enabled
                                                ? "bg-green-400 animate-pulse"
                                                : "bg-gray-500"
                                            }`}
                                        title={server.enabled ? "Enabled" : "Disabled"}
                                    />
                                    <div>
                                        <h3 className="text-sm font-semibold text-white">
                                            {server.name}
                                        </h3>
                                        {server.description && (
                                            <p className="text-xs text-gray-400 mt-0.5">
                                                {server.description}
                                            </p>
                                        )}
                                    </div>
                                </div>
                                <div className="flex items-center gap-2">
                                    <span
                                        className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${server.type === "STDIO"
                                                ? "bg-blue-500/20 text-blue-300"
                                                : server.type === "SSE"
                                                    ? "bg-yellow-500/20 text-yellow-300"
                                                    : "bg-purple-500/20 text-purple-300"
                                            }`}
                                    >
                                        {server.type}
                                    </span>
                                    <button
                                        onClick={() => handleDeleteServer(server.uuid)}
                                        className="rounded-lg bg-red-600/20 px-2 py-1 text-[10px] font-medium text-red-300 hover:bg-red-600/40 transition-colors"
                                        title="Remove this server from TormentNexus"
                                    >
                                        Remove
                                    </button>
                                </div>
                            </div>
                            {/* Details row */}
                            <div className="mt-2 flex gap-4 text-[11px] text-gray-500">
                                {server.command && (
                                    <span>
                                        <strong className="text-gray-400">cmd:</strong>{" "}
                                        <code className="text-gray-300">{server.command}</code>
                                    </span>
                                )}
                                {server.url && (
                                    <span>
                                        <strong className="text-gray-400">url:</strong>{" "}
                                        <code className="text-gray-300">{server.url}</code>
                                    </span>
                                )}
                                <span>
                                    <strong className="text-gray-400">status:</strong>{" "}
                                    {server.status}
                                </span>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* TormentNexus Info Footer */}
            <div className="rounded-xl border border-gray-700/50 bg-gray-800/30 p-4 text-xs text-gray-500">
                <p>
                    <strong className="text-gray-400">TormentNexus Backend:</strong>{" "}
                    {status?.url ?? "http://localhost:12009"} •{" "}
                    <strong className="text-gray-400">Integration:</strong> HTTP Bridge
                    via TormentNexusBridgeService •{" "}
                    <strong className="text-gray-400">Protocol:</strong> tRPC over HTTP
                </p>
            </div>
        </div>
    );
}
