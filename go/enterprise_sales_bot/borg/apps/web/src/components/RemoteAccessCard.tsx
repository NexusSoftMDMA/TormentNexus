"use client";
import { useEffect, useState } from "react";
import { trpc } from "@/utils/trpc";

export default function RemoteAccessCard() {
    const executeTool = trpc.executeTool.useMutation();
    const [isActive, setIsActive] = useState(false);
    const [url, setUrl] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const refreshStatus = async () => {
        try {
            const raw = await executeTool.mutateAsync({
                name: 'get_remote_access_status',
                args: {},
            });
            const parsed = JSON.parse(raw) as { active?: boolean; url?: string | null };
            setIsActive(Boolean(parsed.active));
            setUrl(parsed.url ?? null);
            setError(null);
        } catch (e: any) {
            setError(e?.message ?? 'Unable to fetch remote access status.');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        refreshStatus();
    }, []);

    const handleToggle = async () => {
        setError(null);
        try {
            await executeTool.mutateAsync({
                name: isActive ? 'stop_remote_access' : 'start_remote_access',
                args: isActive ? {} : { port: 3000, label: 'tormentnexus-dashboard' },
            });
            await refreshStatus();
        } catch (e: any) {
            setError(e?.message ?? 'Failed to toggle remote access.');
        }
    };

    return (
        <div className="p-6 border rounded-lg bg-zinc-900 text-zinc-100 shadow-md w-full max-w-md">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold">📡 Remote Access</h2>
                <div className={`px-2 py-1 rounded text-xs font-bold ${isActive ? 'bg-green-700/50 text-green-200' : 'bg-zinc-700'}`}>
                    {isActive ? 'ONLINE' : 'OFFLINE'}
                </div>
            </div>

            <p className="text-sm text-zinc-400 mb-4">
                Expose your TormentNexus Dashboard securely via Cloudflare Tunnel to access it from your mobile device.
            </p>

            {error && (
                <div className="mb-4 p-2 bg-red-900/50 text-red-200 text-sm rounded">
                    {error}
                </div>
            )}

            {url && (
                <div className="mb-4 p-2 bg-zinc-800 text-zinc-200 text-sm rounded break-all">
                    <div className="text-xs text-zinc-400 mb-1">Tunnel URL</div>
                    <a className="text-blue-400 hover:underline" href={url} target="_blank" rel="noreferrer">{url}</a>
                </div>
            )}

            <button
                onClick={handleToggle}
                disabled={isLoading || executeTool.isPending}
                title={isActive ? 'Disable remote access tunnel' : 'Enable remote access tunnel'}
                className="w-full py-2 px-4 rounded font-medium bg-blue-600 text-white disabled:opacity-50"
            >
                {isLoading || executeTool.isPending
                    ? 'Updating Remote Access...'
                    : isActive
                        ? 'Disable Remote Access'
                        : 'Enable Remote Access'}
            </button>
        </div>
    );
}
