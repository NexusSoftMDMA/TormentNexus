'use client';

import { trpc } from '@/utils/trpc';
import { useState, useEffect } from 'react';
import { Card } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { Textarea } from '@tormentnexus/ui';

export default function SettingsDashboard() {
    const [configJson, setConfigJson] = useState('');
    const [log, setLog] = useState('');

    // Fetch settings
    const settingsQuery = trpc.settings.get.useQuery();
    const updateMutation = trpc.settings.update.useMutation();

    useEffect(() => {
        if (settingsQuery.data) {
            setConfigJson(JSON.stringify(settingsQuery.data, null, 2));
        }
    }, [settingsQuery.data]);

    const handleSave = async () => {
        try {
            const config = JSON.parse(configJson);
            await updateMutation.mutateAsync({ config });
            setLog('✅ Configuration saved successfully.');
            settingsQuery.refetch();
        } catch (e: any) {
            setLog(`❌ Error saving config: ${e.message}`);
        }
    };

    return (
        <div className="p-6 space-y-6 h-full flex flex-col">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-yellow-400">System Settings</h1>
                    <p className="text-muted-foreground">Manage core configuration (.tormentnexus/config.json)</p>
                </div>
                <Button
                    onClick={handleSave}
                    disabled={updateMutation.isPending}
                    className="bg-yellow-600 hover:bg-yellow-500"
                >
                    {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                </Button>
            </div>

            <Card className="flex-1 min-h-0 bg-gray-900 border-gray-700 flex flex-col p-4 gap-4">
                {log && (
                    <div className={`p-2 rounded text-sm font-mono ${log.startsWith('✅') ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'}`}>
                        {log}
                    </div>
                )}

                {settingsQuery.isPending ? (
                    <div className="text-gray-500 font-mono">Loading configuration...</div>
                ) : (
                    <Textarea
                        value={configJson}
                        onChange={(e) => { setConfigJson(e.target.value); setLog(''); }}
                        className="flex-1 font-mono text-sm bg-black border-gray-800 text-green-400 leading-relaxed resize-none"
                        spellCheck={false}
                    />
                )}
            </Card>
        </div>
    );
}
