"use client";
import { useState, useEffect } from "react";
import { trpc } from "@/utils/trpc";

export default function ConfigEditor() {
    const [jsonContent, setJsonContent] = useState<string>("{}");
    const [status, setStatus] = useState<string>("");
    const configQuery = trpc.settings.get.useQuery();
    const updateMutation = trpc.settings.update.useMutation();

    useEffect(() => {
        if (configQuery.data) {
            setJsonContent(JSON.stringify(configQuery.data, null, 2));
        }
    }, [configQuery.data]);

    const handleSave = async () => {
        try {
            setStatus('');
            const parsed = JSON.parse(jsonContent);
            await updateMutation.mutateAsync({ config: parsed });
            setStatus('Configuration saved successfully.');
            await configQuery.refetch();
        } catch (error: any) {
            setStatus(`Error: ${error?.message ?? 'Failed to save configuration.'}`);
        }
    };

    return (
        <div className="p-6 border rounded-lg bg-zinc-900 text-zinc-100 shadow-md w-full max-w-2xl mt-8">
            <h2 className="text-xl font-bold mb-4">⚙️ Antigravity Config (mcp.json)</h2>
            <div className="relative">
                <textarea
                    aria-label="Antigravity configuration JSON"
                    title="Antigravity configuration JSON"
                    placeholder={`{\n  "mcpServers": []\n}`}
                    className="w-full h-96 bg-black font-mono text-sm p-4 border border-zinc-700 rounded focus:border-blue-500 outline-none"
                    value={jsonContent}
                    onChange={(e) => setJsonContent(e.target.value)}
                    disabled={configQuery.isPending}
                />
            </div>
            <div className="flex justify-between items-center mt-4">
                <span className={`text-sm ${status.includes("Error") ? "text-red-400" : status ? "text-green-400" : "text-zinc-400"}`}>
                    {configQuery.isPending ? 'Loading configuration...' : status}
                </span>
                <button
                    onClick={handleSave}
                    disabled={configQuery.isPending || updateMutation.isPending}
                    className="bg-blue-600 hover:bg-blue-700 text-white py-2 px-6 rounded font-medium disabled:opacity-50"
                >
                    {updateMutation.isPending ? 'Saving...' : 'Save Config'}
                </button>
            </div>
        </div>
    );
}
