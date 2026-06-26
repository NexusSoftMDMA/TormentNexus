'use client';

import { trpc } from '@/utils/trpc';
import { useState, useEffect } from 'react';
import { Card, Button, Textarea, Tabs, TabsContent, TabsList, TabsTrigger } from '@tormentnexus/ui';
import DirectorConfig from "@/components/DirectorConfig";

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
                    <h1 className="text-3xl font-bold tracking-tight text-white">System Settings</h1>
                    <p className="text-muted-foreground">Manage core configuration (.tormentnexus/config.json)</p>
                </div>
            </div>

            <Tabs defaultValue="director" className="w-full flex-1 flex flex-col min-h-0">
                <TabsList className="grid grid-cols-2 max-w-[400px] mb-4 bg-zinc-900 border border-zinc-800 p-1 rounded-lg">
                    <TabsTrigger value="director" className="text-sm font-medium py-1.5 rounded-md transition-all">Director Config</TabsTrigger>
                    <TabsTrigger value="raw" className="text-sm font-medium py-1.5 rounded-md transition-all">Raw JSON Config</TabsTrigger>
                </TabsList>

                <TabsContent value="director" className="flex-1 flex flex-col min-h-0 outline-none">
                    <Card className="flex-1 overflow-y-auto bg-zinc-900 border-zinc-800 p-6 rounded-xl">
                        <DirectorConfig />
                    </Card>
                </TabsContent>

                <TabsContent value="raw" className="flex-1 flex flex-col min-h-0 outline-none gap-4">
                    <div className="flex justify-end">
                        <Button
                            onClick={handleSave}
                            disabled={updateMutation.isPending}
                            className="bg-yellow-600 hover:bg-yellow-500 text-white font-medium px-4 py-2 rounded-lg"
                        >
                            {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                        </Button>
                    </div>

                    <Card className="flex-1 min-h-0 bg-zinc-900 border-zinc-800 flex flex-col p-4 gap-4 rounded-xl">
                        {log && (
                            <div className={`p-2 rounded text-sm font-mono ${log.startsWith('✅') ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'}`}>
                                {log}
                            </div>
                        )}

                        {settingsQuery.isPending ? (
                            <div className="text-zinc-500 font-mono">Loading configuration...</div>
                        ) : (
                            <Textarea
                                value={configJson}
                                onChange={(e) => { setConfigJson(e.target.value); setLog(''); }}
                                className="flex-1 font-mono text-sm bg-black border-zinc-800 text-green-400 leading-relaxed resize-none p-3 rounded-lg"
                                spellCheck={false}
                            />
                        )}
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
